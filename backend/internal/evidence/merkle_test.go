package evidence

import (
	"strings"
	"testing"
)

func TestDigestIgnoresMapOrderButCatchesAnyChange(t *testing.T) {
	first := Digest("daily_report", map[string]any{"workers": 68, "workDone": "Pengecoran kolom"})
	second := Digest("daily_report", map[string]any{"workDone": "Pengecoran kolom", "workers": 68})
	if first != second {
		t.Fatal("same content must give the same digest")
	}
	if Digest("daily_report", map[string]any{"workers": 69, "workDone": "Pengecoran kolom"}) == first {
		t.Fatal("a changed field must change the digest")
	}
	if Digest("site_photo", map[string]any{"workers": 68, "workDone": "Pengecoran kolom"}) == first {
		t.Fatal("the kind is part of the digest")
	}
}

func TestMerkleRootIsOrderIndependentAndSensitiveToEachLeaf(t *testing.T) {
	leaves := []leaf{{"payment", "b", "01"}, {"daily_report", "a", "02"}, {"site_photo", "c", "03"}}
	reversed := []leaf{leaves[2], leaves[1], leaves[0]}
	root := MerkleRoot(leaves)
	if MerkleRoot(reversed) != root {
		t.Fatal("root must not depend on input order")
	}
	if len(root) != 66 || !strings.HasPrefix(root, "0x") {
		t.Fatalf("root %q is not a 32-byte hex value", root)
	}
	changed := []leaf{leaves[0], leaves[1], {"site_photo", "c", "04"}}
	if MerkleRoot(changed) == root {
		t.Fatal("changing one leaf must change the root")
	}
	if MerkleRoot(leaves[:2]) == root {
		t.Fatal("removing a leaf must change the root")
	}
}

func TestCalldataCarriesPrefixAndRoot(t *testing.T) {
	root := MerkleRoot([]leaf{{"payment", "a", "01"}})
	if got := Calldata(root); got != "0x424e474e"+root[2:] {
		t.Fatalf("calldata = %s", got)
	}
}
