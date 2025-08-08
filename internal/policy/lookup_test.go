package policy

import "testing"

func TestLookup(t *testing.T) {
	p, err := Load([]byte(samplePolicyYAML))
	if err != nil || p == nil {
		t.Fatalf("load error: %v", err)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("validate error: %v", err)
	}
	// Exact
	if rec, ok := p.Lookup("exampletool.com"); !ok || rec.Classification != "NO_DPA" {
		t.Fatalf("expected NO_DPA exact match got %#v ok=%v", rec, ok)
	}
	// With www prefix (no wildcard, should miss)
	if rec, ok := p.Lookup("www.exampletool.com"); ok {
		if rec.Domain == "exampletool.com" {
			// This would only match if we implemented implicit www collapsing.
			t.Fatalf("unexpected match for www.exampletool.com without wildcard")
		}
	}
	// Wildcard base domain
	if rec, ok := p.Lookup("trackingwidgets.io"); !ok || rec.Classification != "EXPIRED_DPA" {
		t.Fatalf("expected EXPIRED_DPA wildcard base match got %#v ok=%v", rec, ok)
	}
	// Wildcard subdomain
	if rec, ok := p.Lookup("api.trackingwidgets.io"); !ok || rec.Classification != "EXPIRED_DPA" {
		t.Fatalf("expected EXPIRED_DPA wildcard subdomain match got %#v ok=%v", rec, ok)
	}
	// Negative
	if _, ok := p.Lookup("nonexistent.domain.test"); ok {
		t.Fatalf("expected miss for nonexistent domain")
	}
}
