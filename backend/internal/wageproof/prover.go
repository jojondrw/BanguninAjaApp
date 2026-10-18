package wageproof

import (
	"bytes"
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

var ErrBelowMinimum = errors.New("a wage is below the minimum, so no proof exists")

type Keys struct {
	Proving   []byte
	Verifying []byte
}

type Prover struct {
	once    sync.Once
	err     error
	circuit constraint.ConstraintSystem
	pk      groth16.ProvingKey
	vk      groth16.VerifyingKey
	load    func() (Keys, bool, error)
	store   func(Keys) error
}

func NewProver(load func() (Keys, bool, error), store func(Keys) error) *Prover {
	return &Prover{load: load, store: store}
}

func (p *Prover) ready() error {
	p.once.Do(func() {
		p.circuit, p.err = frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &WageCircuit{})
		if p.err != nil {
			return
		}
		keys, found, err := p.load()
		if err != nil {
			p.err = err
			return
		}
		if found {
			p.pk = groth16.NewProvingKey(ecc.BN254)
			p.vk = groth16.NewVerifyingKey(ecc.BN254)
			if _, p.err = p.pk.ReadFrom(bytes.NewReader(keys.Proving)); p.err != nil {
				return
			}
			_, p.err = p.vk.ReadFrom(bytes.NewReader(keys.Verifying))
			return
		}
		if p.pk, p.vk, p.err = groth16.Setup(p.circuit); p.err != nil {
			return
		}
		var proving, verifying bytes.Buffer
		if _, p.err = p.pk.WriteTo(&proving); p.err != nil {
			return
		}
		if _, p.err = p.vk.WriteTo(&verifying); p.err != nil {
			return
		}
		p.err = p.store(Keys{Proving: proving.Bytes(), Verifying: verifying.Bytes()})
	})
	return p.err
}

func assignment(leaves []Leaf, root *big.Int, minimumWage int64) *WageCircuit {
	witness := &WageCircuit{Root: root, MinimumWage: minimumWage, Count: len(leaves)}
	for index := 0; index < MaxWorkers; index++ {
		witness.Worker[index], witness.Pay[index], witness.Days[index], witness.Salt[index] = 0, 0, 1, 0
		if index < len(leaves) {
			leaf := leaves[index]
			witness.Worker[index], witness.Pay[index], witness.Days[index], witness.Salt[index] = leaf.Worker, leaf.Pay, leaf.Days, leaf.Salt
		}
	}
	return witness
}

func Meets(leaf Leaf, minimumWage int64) bool {
	return leaf.Days >= 1 && leaf.Days <= maxDays && leaf.Pay*workDaysBase >= minimumWage*int64(leaf.Days)
}

func (p *Prover) Prove(leaves []Leaf, minimumWage int64) (*big.Int, []byte, error) {
	for _, leaf := range leaves {
		if !Meets(leaf, minimumWage) {
			return nil, nil, ErrBelowMinimum
		}
	}
	if err := p.ready(); err != nil {
		return nil, nil, err
	}
	tree, err := BuildTree(leaves)
	if err != nil {
		return nil, nil, err
	}
	full, err := frontend.NewWitness(assignment(leaves, tree.Root(), minimumWage), ecc.BN254.ScalarField())
	if err != nil {
		return nil, nil, err
	}
	proof, err := groth16.Prove(p.circuit, p.pk, full)
	if err != nil {
		return nil, nil, err
	}
	var encoded bytes.Buffer
	if _, err := proof.WriteTo(&encoded); err != nil {
		return nil, nil, err
	}
	return tree.Root(), encoded.Bytes(), nil
}

func (p *Prover) Verify(encoded []byte, root *big.Int, minimumWage int64, count int) (bool, error) {
	if err := p.ready(); err != nil {
		return false, err
	}
	proof := groth16.NewProof(ecc.BN254)
	if _, err := proof.ReadFrom(bytes.NewReader(encoded)); err != nil {
		return false, nil
	}
	public, err := frontend.NewWitness(&WageCircuit{Root: root, MinimumWage: minimumWage, Count: count}, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return false, err
	}
	return groth16.Verify(proof, p.vk, public) == nil, nil
}

func (p *Prover) VerifyingKey() ([]byte, error) {
	if err := p.ready(); err != nil {
		return nil, err
	}
	var encoded bytes.Buffer
	_, err := p.vk.WriteTo(&encoded)
	return encoded.Bytes(), err
}
