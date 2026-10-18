package wageproof

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnmimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/google/uuid"
)

const fieldBytes = 31

var errTooManyWorkers = errors.New("too many workers for one proof")

type Leaf struct {
	Worker *big.Int
	Pay    int64
	Days   int
	Salt   *big.Int
}

func WorkerKey(employeeID uuid.UUID) *big.Int {
	sum := sha256.Sum256([]byte("banguninaja-worker:" + employeeID.String()))
	return new(big.Int).SetBytes(sum[:fieldBytes])
}

func NewSalt() (*big.Int, error) {
	raw := make([]byte, fieldBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(raw), nil
}

func mimcHash(values ...*big.Int) *big.Int {
	hasher := bnmimc.NewMiMC()
	for _, value := range values {
		var element fr.Element
		element.SetBigInt(value)
		bytes := element.Bytes()
		_, _ = hasher.Write(bytes[:])
	}
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

func (l Leaf) Hash() *big.Int {
	return mimcHash(l.Worker, big.NewInt(l.Pay), big.NewInt(int64(l.Days)), l.Salt)
}

type Tree struct {
	levels [][]*big.Int
}

func BuildTree(leaves []Leaf) (Tree, error) {
	if len(leaves) > MaxWorkers {
		return Tree{}, errTooManyWorkers
	}
	level := make([]*big.Int, MaxWorkers)
	for index := range level {
		level[index] = new(big.Int)
		if index < len(leaves) {
			level[index] = leaves[index].Hash()
		}
	}
	levels := [][]*big.Int{level}
	for depth := 0; depth < treeDepth; depth++ {
		next := make([]*big.Int, len(level)/2)
		for index := range next {
			next[index] = mimcHash(level[2*index], level[2*index+1])
		}
		levels = append(levels, next)
		level = next
	}
	return Tree{levels: levels}, nil
}

func (t Tree) Root() *big.Int {
	return t.levels[len(t.levels)-1][0]
}

func (t Tree) Path(index int) []*big.Int {
	path := make([]*big.Int, 0, treeDepth)
	for depth := 0; depth < treeDepth; depth++ {
		path = append(path, t.levels[depth][index^1])
		index /= 2
	}
	return path
}

func VerifyPath(leaf Leaf, index int, path []*big.Int, root *big.Int) bool {
	if len(path) != treeDepth {
		return false
	}
	current := leaf.Hash()
	for _, sibling := range path {
		if index%2 == 0 {
			current = mimcHash(current, sibling)
		} else {
			current = mimcHash(sibling, current)
		}
		index /= 2
	}
	return current.Cmp(root) == 0
}

func Hex(value *big.Int) string {
	var element fr.Element
	element.SetBigInt(value)
	bytes := element.Bytes()
	return "0x" + hex.EncodeToString(bytes[:])
}

func ParseHex(value string) (*big.Int, bool) {
	if len(value) < 3 || value[:2] != "0x" {
		return nil, false
	}
	parsed, ok := new(big.Int).SetString(value[2:], 16)
	return parsed, ok
}
