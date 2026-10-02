package location

import "testing"

func TestDecodeRiskFlags(t *testing.T) {
	if flags := decodeRiskFlags(nil); flags != nil {
		t.Fatalf("nil column should stay unknown, got %v", flags)
	}

	empty := "[]"
	if flags := decodeRiskFlags(&empty); flags == nil || len(flags) != 0 {
		t.Fatalf("empty list should decode to empty, got %v", flags)
	}

	stored := `[{"code":"rawan_longsor","severity":"medium","message":"Indeks bahaya longsor sedang."}]`
	flags := decodeRiskFlags(&stored)
	if len(flags) != 1 || flags[0].Code != "rawan_longsor" || flags[0].Severity != "medium" {
		t.Fatalf("unexpected flags: %+v", flags)
	}

	broken := "{"
	if flags := decodeRiskFlags(&broken); flags != nil {
		t.Fatalf("broken json should be treated as unknown, got %v", flags)
	}
}
