package wageproof

import (
	"errors"
	"math/big"
	"testing"

	"github.com/google/uuid"
)

func memoryKeys() (func() (Keys, bool, error), func(Keys) error) {
	var saved *Keys
	load := func() (Keys, bool, error) {
		if saved == nil {
			return Keys{}, false, nil
		}
		return *saved, true, nil
	}
	store := func(keys Keys) error {
		saved = &keys
		return nil
	}
	return load, store
}

func sampleLeaves(t *testing.T) []Leaf {
	t.Helper()
	leaves := make([]Leaf, 0, 3)
	for _, worker := range []struct {
		pay  int64
		days int
	}{{3_000_000, 25}, {2_400_000, 20}, {12_000_000, 25}} {
		salt, err := NewSalt()
		if err != nil {
			t.Fatal(err)
		}
		leaves = append(leaves, Leaf{Worker: WorkerKey(uuid.New()), Pay: worker.pay, Days: worker.days, Salt: salt})
	}
	return leaves
}

func TestProofVerifiesOnlyForTheCommittedRootAndMinimum(t *testing.T) {
	load, store := memoryKeys()
	prover := NewProver(load, store)
	leaves := sampleLeaves(t)
	const minimum = 2_878_646

	root, proof, err := prover.Prove(leaves, minimum)
	if err != nil {
		t.Fatalf("prove: %v", err)
	}
	if ok, err := prover.Verify(proof, root, minimum, len(leaves)); err != nil || !ok {
		t.Fatalf("valid proof rejected: %v", err)
	}
	if ok, _ := prover.Verify(proof, root, minimum+1, len(leaves)); ok {
		t.Fatal("proof must not verify for a higher minimum wage")
	}
	if ok, _ := prover.Verify(proof, new(big.Int).Add(root, big.NewInt(1)), minimum, len(leaves)); ok {
		t.Fatal("proof must not verify for a different payroll root")
	}
	if ok, _ := prover.Verify(proof, root, minimum, len(leaves)-1); ok {
		t.Fatal("proof must not verify for a different worker count")
	}

	reloaded := NewProver(load, store)
	if ok, err := reloaded.Verify(proof, root, minimum, len(leaves)); err != nil || !ok {
		t.Fatalf("stored keys must verify old proofs: %v", err)
	}
}

func TestNoProofForAWageBelowTheMinimum(t *testing.T) {
	load, store := memoryKeys()
	leaves := sampleLeaves(t)
	leaves[1].Pay = 2_000_000
	if _, _, err := NewProver(load, store).Prove(leaves, 2_878_646); !errors.Is(err, ErrBelowMinimum) {
		t.Fatalf("got %v, want ErrBelowMinimum", err)
	}
}

func TestWorkerReceiptPathLeadsToTheRoot(t *testing.T) {
	leaves := sampleLeaves(t)
	tree, err := BuildTree(leaves)
	if err != nil {
		t.Fatal(err)
	}
	for index, leaf := range leaves {
		if !VerifyPath(leaf, index, tree.Path(index), tree.Root()) {
			t.Fatalf("worker %d receipt does not lead to the root", index)
		}
	}
	forged := leaves[1]
	forged.Pay = 9_999_999
	if VerifyPath(forged, 1, tree.Path(1), tree.Root()) {
		t.Fatal("a changed pay must not match the receipt")
	}
}
