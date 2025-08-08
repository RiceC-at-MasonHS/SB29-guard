// Package server provides the HTTP explanation and health endpoints.
package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

// Server hosts HTTP endpoints rendering policy-based explanations.
type Server struct {
	addr   string
	policy *policy.Policy
}

// New creates a new Server bound to addr using the supplied policy.
func New(addr string, p *policy.Policy) *Server {
	return &Server{addr: addr, policy: p}
}

// Start begins serving HTTP until the listener stops.
func (s *Server) Start() error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/explain", s.handleExplain)
	// root -> simple landing
	http.HandleFunc("/", s.handleRoot)
	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"status":"ok","policy_version":%q}`, s.policy.Version)
}

func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, "<html><head><title>SB29 Guard</title></head><body><h1>SB29 Guard</h1><p>Records loaded: %d</p></body></html>", len(s.policy.Records))
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	q := r.URL.Query()
	// Accept multiple aliases for the original input
	orig := firstNonEmpty(
		q.Get("original_domain"),
		q.Get("original"),
		q.Get("domain"),
		q.Get("url"),
	)
	if orig == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "<p>Missing required parameter ?original_domain= (or original/domain/url)</p>")
		return
	}
	// If a full URL was passed, parse and extract host
	if strings.Contains(orig, "://") {
		if u, err := url.Parse(orig); err == nil && u.Host != "" {
			orig = u.Host
		}
	}
	orig = strings.ToLower(strings.TrimSpace(orig))
	// Trim any leading 'www.' (common pattern) for lookup attempt
	lookupDomain := strings.TrimPrefix(orig, "www.")
	rec, ok := s.policy.Lookup(lookupDomain)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "<html><body><h1>Not Classified</h1><p>The domain %s is not present in the active policy set.</p><p>Policy Version: %s</p></body></html>", orig, s.policy.Version)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self';")
	_, _ = fmt.Fprintf(w, "<html><head><title>Blocked: %s</title></head><body><h1>Access Redirected</h1><p>Domain: %s</p><p>Classification: %s</p>", orig, orig, rec.Classification)
	if rec.Rationale != "" {
		_, _ = fmt.Fprintf(w, "<p>Rationale: %s</p>", htmlEscape(rec.Rationale))
	}
	if rec.SourceRef != "" {
		_, _ = fmt.Fprintf(w, "<p>Source: %s</p>", htmlEscape(rec.SourceRef))
	}
	_, _ = fmt.Fprintf(w, "<p>Policy Version: %s</p><p>Time: %s</p></body></html>", s.policy.Version, time.Now().UTC().Format(time.RFC3339))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Minimal HTML escaping for rationale/source fields (avoid importing html/template yet)
func htmlEscape(s string) string {
	repl := []struct{ old, new string }{
		{"&", "&amp;"},
		{"<", "&lt;"},
		{">", "&gt;"},
		{"\"", "&quot;"},
		{"'", "&#39;"},
	}
	out := s
	for _, r := range repl {
		out = strings.ReplaceAll(out, r.old, r.new)
	}
	return out
}
