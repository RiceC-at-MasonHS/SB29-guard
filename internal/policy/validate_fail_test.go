package policy

import "testing"

func TestValidateFailures(t *testing.T) {
	bad := `version: 0.1.0\nupdated: 2025-08-08\nrecords:\n - domain: "Bad Domain"\n   classification: NO_DPA\n   rationale: r\n   last_review: 2025-08-01\n   status: active\n`
	p, err := Load([]byte(bad))
	if err != nil {
		// load may still succeed
	}
	if p != nil {
		if err2 := p.Validate(); err2 == nil {
			// Expect failure due to invalid domain
			t.Fatalf("expected validation error for bad domain")
		}
	}
}
