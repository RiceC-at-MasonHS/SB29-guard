package policy

import "testing"

func TestLoadValidateHash(t *testing.T) {
	p, err := Load([]byte(samplePolicyYAML))
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("validate error: %v", err)
	}
	h := p.CanonicalHash()
	if h == "" {
		t.Fatalf("empty hash")
	}
	if len(p.Records) != 2 {
		t.Fatalf("expected 2 records got %d", len(p.Records))
	}
}
