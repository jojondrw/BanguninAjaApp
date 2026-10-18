package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const calldataPrefix = "424e474e"

func Digest(kind string, fields map[string]any) string {
	payload, _ := json.Marshal(map[string]any{"kind": kind, "fields": fields})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

type leaf struct {
	kind    string
	subject string
	digest  string
}

func (l leaf) hash() []byte {
	return ethereum.Keccak256([]byte(l.kind + ":" + l.subject + ":" + l.digest))
}

func sortLeaves(leaves []leaf) {
	sort.Slice(leaves, func(a, b int) bool {
		if leaves[a].kind != leaves[b].kind {
			return leaves[a].kind < leaves[b].kind
		}
		if leaves[a].subject != leaves[b].subject {
			return leaves[a].subject < leaves[b].subject
		}
		return leaves[a].digest < leaves[b].digest
	})
}

func MerkleRoot(leaves []leaf) string {
	sorted := append([]leaf(nil), leaves...)
	sortLeaves(sorted)
	if len(sorted) == 0 {
		return "0x" + hex.EncodeToString(make([]byte, 32))
	}
	level := make([][]byte, len(sorted))
	for index, item := range sorted {
		level[index] = item.hash()
	}
	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for index := 0; index < len(level); index += 2 {
			if index+1 == len(level) {
				next = append(next, level[index])
				continue
			}
			next = append(next, ethereum.Keccak256(level[index], level[index+1]))
		}
		level = next
	}
	return "0x" + hex.EncodeToString(level[0])
}

func Calldata(root string) string {
	return "0x" + calldataPrefix + root[2:]
}
