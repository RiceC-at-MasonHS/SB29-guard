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

func TestHandleExplainNotFound(t *testing.T) {
	srv := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/explain?domain=missing.example", nil)
	srv.handleExplain(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rr.Code)
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
