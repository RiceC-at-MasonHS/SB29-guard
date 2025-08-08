package hash

import "testing"

func TestSHA256Hex(t *testing.T) {
	out1 := SHA256Hex([]byte("abc"))
	out2 := SHA256Hex([]byte("abc"))
	if out1 != out2 {
		t.Fatalf("hash not deterministic")
	}
	if len(out1) != 64 {
		t.Fatalf("expected 64 hex chars got %d", len(out1))
	}
	if out1 == SHA256Hex([]byte("abcd")) {
		t.Fatalf("different input should differ")
	}
}
