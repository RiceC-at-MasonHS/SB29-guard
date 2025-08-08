package dnsgen

import (
	"strings"
	"testing"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

func testPolicy() *policy.Policy {
	return &policy.Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []policy.Record{
		{Domain: "exampletool.com", Classification: "NO_DPA", Rationale: "x", LastReview: "2025-08-01", Status: "active"},
		{Domain: "*.trackingwidgets.io", Classification: "EXPIRED_DPA", Rationale: "x", LastReview: "2025-07-15", Status: "active"},
	}}
}

func TestGenerateHosts(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "hosts", Mode: "a-record", RedirectIPv4: "10.10.10.50"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "exampletool.com") {
		t.Fatalf("missing domain")
	}
	if strings.Contains(out, "*.trackingwidgets.io") {
		t.Fatalf("wildcard not stripped")
	}
}

func TestGenerateBindA(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "bind", Mode: "a-record", RedirectIPv4: "10.10.10.50", RedirectHost: "blocked.guard.local"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "IN A 10.10.10.50") {
		t.Fatalf("missing A record")
	}
}

func TestGenerateBindCNAME(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "bind", Mode: "cname", RedirectHost: "blocked.guard.local"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "CNAME blocked.guard.local.") {
		t.Fatalf("missing CNAME")
	}
}

func TestGenerateRPZ(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "rpz", Mode: "cname", RedirectHost: "blocked.guard.local", RedirectIPv4: "10.10.10.50"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "exampletool.com.") {
		t.Fatalf("missing domain in rpz")
	}
	if !strings.Contains(string(b), "A 10.10.10.50") {
		t.Fatalf("missing redirect A")
	}
}

func TestGenerateUnbound(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "unbound", Mode: "a-record", RedirectIPv4: "10.10.10.50"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "local-zone: \"exampletool.com\"") {
		t.Fatalf("missing local-zone")
	}
}

func TestSerialStrategies(t *testing.T) {
	p := testPolicy()
	strategies := []string{"date", "epoch", "hash"}
	for _, s := range strategies {
		opt := Options{Format: "bind", Mode: "a-record", RedirectIPv4: "10.10.10.50", RedirectHost: "blocked.guard.local", SerialStrategy: s}
		b, err := Generate(p, opt)
		if err != nil {
			t.Fatalf("strategy %s error: %v", s, err)
		}
		out := string(b)
		if !strings.Contains(out, "IN SOA") {
			t.Fatalf("strategy %s missing SOA", s)
		}
		// Basic shape check: serial present as number
		if !strings.Contains(out, " hostmaster.") {
			t.Fatalf("strategy %s unexpected SOA format", s)
		}
	}
}
