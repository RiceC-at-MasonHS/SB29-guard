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

func TestGenerateDnsmasqA(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "dnsmasq", Mode: "a-record", RedirectIPv4: "10.10.10.50"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "address=/exampletool.com/10.10.10.50") {
		t.Fatalf("missing dnsmasq address line")
	}
}

func TestGenerateDnsmasqCNAME(t *testing.T) {
	p := testPolicy()
	opt := Options{Format: "dnsmasq", Mode: "cname", RedirectHost: "blocked.guard.local"}
	b, err := Generate(p, opt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "cname=exampletool.com,blocked.guard.local") {
		t.Fatalf("missing dnsmasq cname line")
	}
}

func TestGenerateDomainList(t *testing.T) {
	p := testPolicy()
	b, err := Generate(p, Options{Format: "domain-list"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "exampletool.com") || strings.Contains(s, "*.trackingwidgets.io") {
		t.Fatalf("domain-list content unexpected: %s", s)
	}
}

func TestGenerateWinPS_A(t *testing.T) {
	p := testPolicy()
	b, err := Generate(p, Options{Format: "winps", Mode: "a-record", RedirectIPv4: "10.10.10.50"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "Add-DnsServerResourceRecordA") {
		t.Fatalf("missing A record PS command")
	}
}

func TestGenerateWinPS_CNAME(t *testing.T) {
	p := testPolicy()
	b, err := Generate(p, Options{Format: "winps", Mode: "cname", RedirectHost: "blocked.guard.local"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(string(b), "Add-DnsServerResourceRecordCName") {
		t.Fatalf("missing CNAME PS command")
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

func TestGenerateErrors(t *testing.T) {
	p := testPolicy()
	// Unsupported format
	if _, err := Generate(p, Options{Format: "bogus"}); err == nil {
		t.Fatalf("expected error for unsupported format")
	}
	// Hosts missing redirect IPv4
	if _, err := Generate(p, Options{Format: "hosts"}); err == nil {
		t.Fatalf("expected error for hosts missing redirect-ipv4")
	}
	// Bind a-record missing redirect IPv4
	if _, err := Generate(p, Options{Format: "bind", Mode: "a-record", RedirectHost: "blocked.example"}); err == nil {
		t.Fatalf("expected error for bind a-record missing redirect-ipv4")
	}
	// RPZ missing redirect host
	if _, err := Generate(p, Options{Format: "rpz"}); err == nil {
		t.Fatalf("expected error for rpz missing redirect host")
	}
}
