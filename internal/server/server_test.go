package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
