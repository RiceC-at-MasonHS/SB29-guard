package server

import (
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

func testPolicy() *policy.Policy {
	return &policy.Policy{Version: "0.1.0", Updated: "2025-08-08", Records: []policy.Record{
		{Domain: "exampletool.com", Classification: "NO_DPA", Rationale: "r", LastReview: "2025-08-01", Status: "active"},
		{Domain: "*.trackingwidgets.io", Classification: "EXPIRED_DPA", Rationale: "r2", LastReview: "2025-08-01", Status: "active"},
	}}
}

func newTestServer(t *testing.T) *Server {
	p := testPolicy()
	srv := New(":0", p) // not starting network listener for handler tests
	return srv
}

func TestHandleHealth(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.handleHealth(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type %s", ct)
	}
	if body := rr.Body.String(); !strings.Contains(body, "\"status\":\"ok\"") {
		t.Fatalf("missing ok status body=%s", body)
	}
}

func TestHandleRoot(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	srv.handleRoot(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct == "" {
		t.Fatalf("missing content-type")
	}
	if body := rr.Body.String(); !strings.Contains(body, "Policy v0.1.0") {
		t.Fatalf("expected footer with policy version; got: %s", body)
	}
}

func TestHandleExplainSuccess(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain?domain=exampletool.com", nil)
	srv.handleExplain(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	if body := rr.Body.String(); !strings.Contains(body, "exampletool.com") {
		t.Fatalf("missing domain in body")
	}
}

func TestHandleExplainMissingParam(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain", nil)
	srv.handleExplain(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rr.Code)
	}
}

func TestHostFallbackIsOffByDefault(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain", nil)
	// Simulate only Host header present; flag should be off by default
	req.Host = "exampletool.com"
	srv.handleExplain(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when only Host is present and fallback disabled, got %d", rr.Code)
	}
}

func TestHandleExplainNotFound(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain?domain=missing.example", nil)
	srv.handleExplain(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rr.Code)
	}
}

func TestExtractOriginalDomainFromHeaders_Precedence(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/explain", nil)
	// Set multiple headers, ensure precedence: X-Original-Host > X-Forwarded-Host > Referer > Host
	r.Header.Set("X-Forwarded-Host", "xfw.example, other")
	r.Header.Set("Referer", "https://ref.example/path?q=1")
	r.Host = "host.example"
	if got := extractOriginalDomainFromHeaders(r, false); got != "xfw.example" {
		t.Fatalf("expected xfw.example, got %s", got)
	}
	r.Header.Set("X-Original-Host", "orig.example")
	if got := extractOriginalDomainFromHeaders(r, false); got != "orig.example" {
		t.Fatalf("expected orig.example, got %s", got)
	}
	r.Header.Del("X-Original-Host")
	r.Header.Del("X-Forwarded-Host")
	if got := extractOriginalDomainFromHeaders(r, false); got != "ref.example" {
		t.Fatalf("expected ref.example from Referer, got %s", got)
	}
	r.Header.Del("Referer")
	if got := extractOriginalDomainFromHeaders(r, false); got != "" {
		t.Fatalf("expected empty when no informative headers, got %s", got)
	}
	// With feature flag, Host is allowed as last resort
	if got := extractOriginalDomainFromHeaders(r, true); got != "host.example" {
		t.Fatalf("expected host.example from Host with flag, got %s", got)
	}
}

func TestHandleExplain_UsesHeadersWhenNoQueryParam(t *testing.T) {
	srv := newTestServer(t)
	// Ensure policy contains a record matching header-derived domain
	// Modify the test policy to include xfw.example
	srv.UpdatePolicy(&policy.Policy{Version: "H1", Updated: "2025-08-09", Records: []policy.Record{
		{Domain: "xfw.example", Classification: "NO_DPA", Rationale: "r", LastReview: "2025-08-01", Status: "active"},
	}})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain", nil)
	req.Header.Set("X-Forwarded-Host", "xfw.example")
	srv.handleExplain(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "xfw.example") {
		t.Fatalf("response did not include derived domain: %s", rr.Body.String())
	}
}

func TestHandleExplain_HeaderNormalization(t *testing.T) {
	srv := newTestServer(t)
	srv.UpdatePolicy(&policy.Policy{Version: "H2", Updated: "2025-08-09", Records: []policy.Record{
		{Domain: "exampletool.com", Classification: "NO_DPA", Rationale: "r", LastReview: "2025-08-01", Status: "active"},
	}})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain", nil)
	// Include scheme, port, and www.
	req.Header.Set("Referer", "https://www.exampletool.com:8443/path")
	srv.handleExplain(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "exampletool.com") {
		t.Fatalf("normalized domain not found in response: %s", rr.Body.String())
	}
}

func TestHeaderWhitespaceCasingAndMultipleHosts(t *testing.T) {
	// Ensure trimming and case-folding of header-derived values happens downstream
	r := httptest.NewRequest(http.MethodGet, "/explain", nil)
	r.Header.Set("X-Forwarded-Host", "  XFW.EXAMPLE:443  , ignored.example ")
	got := extractOriginalDomainFromHeaders(r, false)
	if strings.TrimSpace(got) != "XFW.EXAMPLE:443" { // raw extract preserves case/port; normalization happens later in handler
		t.Fatalf("unexpected raw header extraction: %q", got)
	}
	// Referer with trailing spaces
	r = httptest.NewRequest(http.MethodGet, "/explain", nil)
	r.Header.Set("Referer", " https://Ref.Example:443/path ")
	got = extractOriginalDomainFromHeaders(r, false)
	if got != "Ref.Example:443" {
		t.Fatalf("unexpected referer host extract: %q", got)
	}
}

func TestNormalizeHostPortAndWWW(t *testing.T) {
	srv := newTestServer(t)
	srv.UpdatePolicy(&policy.Policy{Version: "H3", Updated: "2025-08-09", Records: []policy.Record{{Domain: "ref.example", Classification: "NO_DPA", Rationale: "r", LastReview: "2025-08-01", Status: "active"}}})
	// Case: X-Forwarded-Host includes port; handler should strip port and lowercase, then trim www.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain", nil)
	req.Header.Set("X-Forwarded-Host", "WWW.Ref.Example:443")
	srv.handleExplain(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ref.example") {
		t.Fatalf("expected normalized domain in response: %s", rr.Body.String())
	}
}

func TestIPv6BracketPortHandling_NoLookup(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain", nil)
	// Not expected in our flow, but ensure no panic: bracketed IPv6 with port
	req.Header.Set("X-Original-Host", "[2001:db8::1]:8080")
	srv.handleExplain(rr, req)
	// No policy for IPv6, expect 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when extracted host not in policy, got %d", rr.Code)
	}
}

func TestQueryParamPrecedenceOverHeaders(t *testing.T) {
	srv := newTestServer(t)
	srv.UpdatePolicy(&policy.Policy{Version: "H4", Updated: "2025-08-09", Records: []policy.Record{
		{Domain: "param.example", Classification: "NO_DPA", Rationale: "r", LastReview: "2025-08-01", Status: "active"},
		{Domain: "header.example", Classification: "NO_DPA", Rationale: "r", LastReview: "2025-08-01", Status: "active"},
	}})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain?domain=param.example", nil)
	req.Header.Set("X-Forwarded-Host", "header.example")
	srv.handleExplain(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "param.example") {
		t.Fatalf("expected body to reflect query param domain, got: %s", rr.Body.String())
	}
}

func TestHtmlEscape(t *testing.T) {
	in := `<script>alert("x")</script>&'"`
	out := htmlEscape(in)
	// Ensure key characters escaped
	wants := []string{"&lt;script&gt;", "&quot;", "&#39;", "&amp;"}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("expected escape %s in %s", w, out)
		}
	}
}

func TestUpdatePolicyHotSwap(t *testing.T) {
	// Policy A
	pA := &policy.Policy{Version: "A", Updated: "2025-08-08", Records: []policy.Record{
		{Domain: "a.com", Classification: "NO_DPA", Rationale: "rA", LastReview: "2025-08-01", Status: "active"},
	}}
	// Policy B
	pB := &policy.Policy{Version: "B", Updated: "2025-08-09", Records: []policy.Record{
		{Domain: "b.com", Classification: "EXPIRED_DPA", Rationale: "rB", LastReview: "2025-08-02", Status: "active"},
	}}
	srv := New(":0", pA)
	// Check /health and / reflect Policy A
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.handleHealth(rr, req)
	if !strings.Contains(rr.Body.String(), "A") {
		t.Fatalf("expected version A in health, got %s", rr.Body.String())
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	srv.handleRoot(rr, req)
	if !strings.Contains(rr.Body.String(), "Policy vA") {
		t.Fatalf("expected Policy vA in root, got %s", rr.Body.String())
	}
	// Hot-swap to Policy B
	srv.UpdatePolicy(pB)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.handleHealth(rr, req)
	if !strings.Contains(rr.Body.String(), "B") {
		t.Fatalf("expected version B in health after swap, got %s", rr.Body.String())
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	srv.handleRoot(rr, req)
	if !strings.Contains(rr.Body.String(), "Policy vB") {
		t.Fatalf("expected Policy vB in root after swap, got %s", rr.Body.String())
	}
}

func TestUpdatePolicyConcurrency(t *testing.T) {
	pA := &policy.Policy{Version: "A", Updated: "2025-08-08", Records: []policy.Record{{Domain: "a.com", Classification: "NO_DPA", Rationale: "rA", LastReview: "2025-08-01", Status: "active"}}}
	pB := &policy.Policy{Version: "B", Updated: "2025-08-09", Records: []policy.Record{{Domain: "b.com", Classification: "EXPIRED_DPA", Rationale: "rB", LastReview: "2025-08-02", Status: "active"}}}
	srv := New(":0", pA)
	done := make(chan struct{})
	// Start goroutine to swap policy repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				srv.UpdatePolicy(pA)
			} else {
				srv.UpdatePolicy(pB)
			}
		}
		close(done)
	}()
	// While swapping, hit /health and /
	for i := 0; i < 100; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		srv.handleHealth(rr, req)
		body := rr.Body.String()
		if !strings.Contains(body, "A") && !strings.Contains(body, "B") {
			t.Fatalf("unexpected version in health: %s", body)
		}
		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		srv.handleRoot(rr, req)
		body = rr.Body.String()
		if !strings.Contains(body, "Policy vA") && !strings.Contains(body, "Policy vB") {
			t.Fatalf("unexpected version in root: %s", body)
		}
	}
	<-done
}

func TestServerStartInvalidAddr(t *testing.T) {
	// Reset default mux to avoid handler re-registration panics across tests
	http.DefaultServeMux = http.NewServeMux()
	srv := New("bad:addr", testPolicy())
	if err := srv.Start(); err == nil {
		t.Fatalf("expected error for invalid addr")
	}
}

func TestServerStartPortInUse(t *testing.T) {
	// Reset default mux
	http.DefaultServeMux = http.NewServeMux()
	// Grab a port and keep it open
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	defer ln.Close()
	srv := New(addr, testPolicy())
	if err := srv.Start(); err == nil {
		t.Fatalf("expected error when port already in use")
	}
}

func TestServerStartAndServeHealth(t *testing.T) {
	// Reset default mux
	http.DefaultServeMux = http.NewServeMux()
	// Preselect a free port (close it to free before starting)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	srv := New(addr, testPolicy())
	done := make(chan struct{})
	go func() {
		// Start blocks; ignore eventual shutdown (test process exit)
		_ = srv.Start()
		close(done)
	}()
	// Poll for readiness
	url := "http://" + addr + "/health"
	var lastErr error
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url)
		if err == nil {
			if resp.StatusCode != 200 {
				t.Fatalf("unexpected status %d", resp.StatusCode)
			}
			return
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server did not become ready: %v", lastErr)
}

func TestHandleRootTemplateError(t *testing.T) {
	// Construct a server with a template missing "layout.html" to force ExecuteTemplate error
	badTmpl := template.New("bad")
	srv := &Server{addr: ":0", policy: testPolicy(), tmpl: badTmpl, inlineCSS: ""}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	srv.handleRoot(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "template error") {
		t.Fatalf("expected template error message, got: %s", rr.Body.String())
	}
}

func TestHandleExplainTemplateError(t *testing.T) {
	// Use a bad template (no "layout.html") so ExecuteTemplate fails even when lookup succeeds
	badTmpl := template.New("bad")
	srv := &Server{addr: ":0", policy: testPolicy(), tmpl: badTmpl, inlineCSS: ""}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain?domain=exampletool.com", nil)
	srv.handleExplain(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "template error") {
		t.Fatalf("expected template error message, got: %s", rr.Body.String())
	}
}

func TestMetricsEndpoint(t *testing.T) {
	srv := newTestServer(t)
	// Record an error then a success to populate metrics
	srv.RecordRefreshError("network error")
	srv.RecordRefreshSuccess("csv-cache")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	srv.handleMetrics(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "\"status\":\"ok\"") {
		t.Fatalf("unexpected metrics body: %s", body)
	}
	if !strings.Contains(body, "csv-cache") {
		t.Fatalf("expected last_refresh_source csv-cache in metrics: %s", body)
	}
}

func TestHandleLawRedirect(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/law", nil)
	srv.handleLaw(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if loc == "" {
		t.Fatalf("missing Location header")
	}
	// Default should be the LIS PDF unless overridden via env
	if !strings.Contains(loc, "search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/") {
		t.Fatalf("unexpected law redirect target: %s", loc)
	}
}

func TestClassifyEndpointFoundAndNotFound(t *testing.T) {
	srv := newTestServer(t)
	// Found
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/classify?d=exampletool.com", nil)
	srv.handleClassify(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "\"found\":true") || !strings.Contains(body, "\"classification\":") {
		t.Fatalf("unexpected classify body: %s", body)
	}
	// Not found
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/classify?d=missing.example", nil)
	srv.handleClassify(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "\"found\":false") {
		t.Fatalf("expected found=false: %s", rr.Body.String())
	}
}

func TestDomainListEndpoint(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/domain-list", nil)
	srv.handleDomainList(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	body := rr.Body.String()
	// Should contain exampletool.com
	if !strings.Contains(body, "exampletool.com") {
		t.Fatalf("missing exampletool.com in domain list: %s", body)
	}
	// And wildcard represented as base and .base
	if !strings.Contains(body, "trackingwidgets.io") || !strings.Contains(body, ".trackingwidgets.io") {
		t.Fatalf("wildcard representation missing: %s", body)
	}
}
