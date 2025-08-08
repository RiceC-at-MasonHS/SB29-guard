package policy

import (
	"strings"
	"testing"
)

// Additional validation & hash edge cases to increase coverage.
func TestValidateErrorCases(t *testing.T) {
	t.Run("missing version rejected at load (schema)", func(t *testing.T) {
		bad := `updated: 2025-08-08
records:
  - domain: "example.com"
    classification: NO_DPA
    rationale: r
    last_review: 2025-08-01
    status: active
`
		if _, err := Load([]byte(bad)); err == nil {
			// Expect schema validation error before Validate() stage
			t.Fatalf("expected error for missing version")
		}
	})

	t.Run("missing updated caught by Validate", func(t *testing.T) {
		p := &Policy{Version: "0.1.0", Records: []Record{{Domain: "example.com", Classification: "NO_DPA", Status: "active", LastReview: "2025-08-01", Rationale: "r"}}}
		if err := p.Validate(); err == nil {
			t.Fatalf("expected missing updated error")
		}
	})

	t.Run("invalid classification", func(t *testing.T) {
		p := &Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []Record{{Domain: "example.com", Classification: "BAD", Status: "active", LastReview: "2025-08-01", Rationale: "r"}}}
		if err := p.Validate(); err == nil {
			t.Fatalf("expected invalid classification error")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		p := &Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []Record{{Domain: "example.com", Classification: "NO_DPA", Status: "weird", LastReview: "2025-08-01", Rationale: "r"}}}
		if err := p.Validate(); err == nil {
			t.Fatalf("expected invalid status error")
		}
	})

	t.Run("duplicate domain+classification", func(t *testing.T) {
		p := &Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []Record{
			{Domain: "example.com", Classification: "NO_DPA", Status: "active", LastReview: "2025-08-01", Rationale: "r"},
			{Domain: "example.com", Classification: "NO_DPA", Status: "active", LastReview: "2025-08-01", Rationale: "r2"},
		}}
		if err := p.Validate(); err == nil {
			t.Fatalf("expected duplicate detection error")
		}
	})
}

func TestCanonicalHashSkipsSuspended(t *testing.T) {
	p := &Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []Record{
		{Domain: "a.example.com", Classification: "NO_DPA", Status: "active", LastReview: "2025-08-01", Rationale: "r"},
		{Domain: "b.example.com", Classification: "NO_DPA", Status: "suspended", LastReview: "2025-08-01", Rationale: "r"},
		{Domain: "c.example.com", Classification: "NO_DPA", Status: "", LastReview: "2025-08-01", Rationale: "r"}, // empty treated as active
	}}
	h1 := p.CanonicalHash()
	if h1 == "" {
		t.Fatalf("empty hash")
	}
	// Remove suspended record and hash again; should be identical.
	p2 := &Policy{Version: p.Version, Updated: p.Updated, Records: []Record{p.Records[0], p.Records[2]}}
	h2 := p2.CanonicalHash()
	if h1 != h2 {
		// Suspended record should not influence hash.
		// Use Fatal to enforce invariance.
		// nolint:revive
		//
		// If this fails due to future design change, update test accordingly.
		//

		t.Fatalf("expected identical hash without suspended record; h1=%s h2=%s", h1, h2)
	}
}

func TestLookupSkipsSuspended(t *testing.T) {
	p := &Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []Record{
		{Domain: "allowed.example.com", Classification: "NO_DPA", Status: "active", LastReview: "2025-08-01", Rationale: "r"},
		{Domain: "hidden.example.com", Classification: "NO_DPA", Status: "suspended", LastReview: "2025-08-01", Rationale: "r"},
	}}
	for i := range p.Records {
		p.Records[i].Domain = strings.ToLower(p.Records[i].Domain)
	}
	if rec, ok := p.Lookup("hidden.example.com"); ok || rec != nil {
		// Should not return suspended
		// nolint:revive
		//

		t.Fatalf("expected no match for suspended domain")
	}
}
