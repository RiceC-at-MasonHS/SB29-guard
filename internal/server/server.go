// Package server provides the HTTP explanation, landing, and health endpoints.
package server

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

//go:embed templates/*.html templates/*.txt
var templateFS embed.FS

// defaultCSS is the built-in stylesheet served inline; dark by default, mobile-first.
const defaultCSS = `/* Dark by default, mobile-first */
:root{color-scheme:dark light;font-family:system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,sans-serif;--bg:#0f1115;--fg:#e7e9ee;--accent:#62a8ff;--badge:#ff6b60;--panel:#151923;--muted:#9aa4b2;--ring:#2a3342}
@media (prefers-color-scheme:light){:root{--bg:#ffffff;--fg:#222;--accent:#004c99;--badge:#d9534f;--panel:#f5f7fa;--muted:#5b6673;--ring:#dfe6ef}}
body{margin:0;background:var(--bg);color:var(--fg);line-height:1.5}
header,main,footer{max-width:900px;margin:0 auto;padding:1rem}
header{border-bottom:1px solid #ccc2}
footer{border-top:1px solid #ccc2;margin-top:2rem;font-size:.85rem;opacity:.85}
h1{font-size:1.5rem;margin:.2rem 0 .6rem}
h2{font-size:1.1rem;margin:1rem 0 .5rem}
code{background:var(--panel);padding:2px 6px;border-radius:6px;font-size:.875em}
.muted{color:var(--muted)}
.tagline{margin:.1rem 0 0;color:var(--muted);font-size:.9rem}

/* Explainer layout */
.explain-page{display:grid;grid-template-columns:1fr;gap:1rem;align-items:start}
.ohio-ascii{display:none}
.card{background:var(--panel);border:1px solid var(--ring);border-radius:14px;padding:1rem 1.1rem;box-shadow:0 6px 24px rgba(0,0,0,.25)}
.card-title{margin:.2rem 0 .6rem;font-size:1.2rem;line-height:1.25}
.card-title .domain{display:inline-block;font-weight:700;color:var(--fg)}
.rationale{margin:.5rem 0 0}
.source-ref{margin:.25rem 0 0;font-size:.95rem;color:var(--muted)}
.meta-grid{display:grid;grid-template-columns:1fr 1fr;gap:.5rem;margin:1rem 0 0}
.meta-grid div{background:transparent;border:1px dashed var(--ring);border-radius:8px;padding:.5rem .6rem}
.meta-grid dt{font-weight:600;font-size:.8rem;color:var(--muted);margin:0}
.meta-grid dd{margin:0;font-family:ui-monospace,SFMono-Regular,Consolas,Monaco,monospace}
.chips{margin:.4rem 0 .2rem}
.badge{display:inline-block;padding:.35rem .7rem;background:var(--badge);color:#fff;border-radius:999px;font-size:.75rem;font-weight:700;letter-spacing:.4px;text-transform:uppercase}
.badge-lg{font-size:.85rem}

/* Larger viewports: show decorative ASCII, place beside card */
@media (min-width: 720px){
	.explain-page{grid-template-columns:minmax(180px,1fr) minmax(380px,540px);gap:2rem}
	.ohio-ascii{display:block;white-space:pre;line-height:1;opacity:.18;color:var(--muted);user-select:none}
	.card{padding:1.25rem 1.4rem}
	.card-title{font-size:1.35rem}
}`

// Server hosts HTTP endpoints rendering policy-based explanations.
type Server struct {
	addr              string
	policy            *policy.Policy
	tmpl              *template.Template
	inlineCSS         template.CSS
	ohioASCII         string
	lawURL            string
	allowHostFallback bool
	mu                sync.RWMutex

	// refresh/metrics fields
	refreshMu         sync.RWMutex
	lastRefreshTime   time.Time
	lastRefreshSource string
	refreshCount      int
	refreshErrorCount int
	lastRefreshError  string
}

// New creates a new Server bound to addr using the supplied policy.
func New(addr string, p *policy.Policy) *Server {
	t := template.New("layout.html").Funcs(template.FuncMap{})
	// Parse layout, then explain, then root (root last => its blocks override for landing page)
	tmpl, err := t.ParseFS(templateFS, "templates/layout.html", "templates/explain.html", "templates/root.html")
	if err != nil {
		panic(fmt.Sprintf("template parse error: %v", err))
	}
	ascii := ""
	if b, err := templateFS.ReadFile("templates/ohio.ascii-art.txt"); err == nil {
		ascii = string(b)
	}
	// Determine law URL target (configurable via env var; stable default to ORC section)
	law := os.Getenv("SB29_LAW_URL")
	if strings.TrimSpace(law) == "" {
		law = "https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/"
	}
	// Feature flag: allow Host header as last-resort fallback (default: false)
	allowHost := strings.EqualFold(strings.TrimSpace(os.Getenv("SB29_ALLOW_HOST_FALLBACK")), "true")
	return &Server{addr: addr, policy: p, tmpl: tmpl, inlineCSS: template.CSS(defaultCSS), ohioASCII: ascii, lawURL: law, allowHostFallback: allowHost}
}

// NewWithTemplates creates a new Server using caller-supplied templates and CSS.
// tmpl must include templates named layout.html, explain.html, and root.html.
func NewWithTemplates(addr string, p *policy.Policy, tmpl *template.Template, css string) *Server {
	return &Server{addr: addr, policy: p, tmpl: tmpl, inlineCSS: template.CSS(css)}
}

// Start begins serving HTTP until the listener stops.
func (s *Server) Start() error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/metrics", s.handleMetrics)
	http.HandleFunc("/law", s.handleLaw)
	http.HandleFunc("/explain", s.handleExplain)
	http.HandleFunc("/", s.handleRoot)
	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	p := s.getPolicy()
	_, _ = fmt.Fprintf(w, `{"status":"ok","policy_version":%q}`, p.Version)
}

func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	p := s.getPolicy()
	data := map[string]interface{}{
		"CSS":             s.inlineCSS,
		"Title":           "SB29 Guard",
		"Header":          "SB29 Guard",
		"ContentTemplate": "root_content",
		"RecordCount":     len(p.Records),
		"Year":            time.Now().Year(),
		"PolicyVersion":   p.Version,
		"Page":            "root",
		"OhioASCII":       s.ohioASCII,
		// Footer law link uses internal redirect for stability
		"LawURL": "/law",
	}
	if err := s.tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "template error: %v", err)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	p := s.getPolicy()
	s.refreshMu.RLock()
	lrTime := s.lastRefreshTime
	lrSrc := s.lastRefreshSource
	rCount := s.refreshCount
	eCount := s.refreshErrorCount
	lastErr := s.lastRefreshError
	s.refreshMu.RUnlock()
	ts := ""
	if !lrTime.IsZero() {
		ts = lrTime.UTC().Format(time.RFC3339)
	}
	_, _ = fmt.Fprintf(w, `{"status":"ok","policy_version":%q,"records":%d,"last_refresh_time":%q,"last_refresh_source":%q,"refresh_count":%d,"refresh_error_count":%d,"last_refresh_error":%q}`,
		p.Version, len(p.Records), ts, lrSrc, rCount, eCount, lastErr)
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	q := r.URL.Query()
	orig := firstNonEmpty(q.Get("original_domain"), q.Get("original"), q.Get("domain"), q.Get("url"))
	if orig == "" {
		// No query param provided—attempt to infer from headers (DNS redirect scenario)
		orig = extractOriginalDomainFromHeaders(r, s.allowHostFallback)
		if orig == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(w, "<p>Missing original domain (provide ?domain= or ensure X-Original-Host or X-Forwarded-Host headers are set)</p>")
			return
		}
	}
	if strings.Contains(orig, "://") { // full URL passed
		if u, err := url.Parse(orig); err == nil && u.Host != "" {
			orig = u.Host
		}
	}
	// Normalize: trim port (including bracketed IPv6), www., lowercase
	orig = strings.ToLower(strings.TrimSpace(orig))
	// Attempt robust host:port splitting
	if strings.Contains(orig, ":") {
		if strings.HasPrefix(orig, "[") {
			if h, _, err := net.SplitHostPort(orig); err == nil {
				orig = strings.Trim(h, "[]")
			}
		} else {
			if h, _, err := net.SplitHostPort(orig); err == nil {
				orig = h
			}
		}
	}
	lookupDomain := strings.TrimPrefix(orig, "www.")
	p := s.getPolicy()
	rec, ok := p.Lookup(lookupDomain)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "<html><body><h1>Not Classified</h1><p>The domain %s is not present in the active policy set.</p><p>Policy Version: %s</p></body></html>", orig, p.Version)
		return
	}

	// Security / privacy oriented headers
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'unsafe-inline'; frame-ancestors 'none'; base-uri 'none'; form-action 'self';")

	data := map[string]interface{}{
		"CSS":             s.inlineCSS,
		"Title":           fmt.Sprintf("Blocked: %s", orig),
		"Header":          "Access Redirected",
		"ContentTemplate": "explain_content",
		"Original":        orig,
		"Classification":  rec.Classification,
		"Rationale":       htmlEscape(rec.Rationale),
		"SourceRef":       htmlEscape(rec.SourceRef),
		"PolicyVersion":   p.Version,
		"Now":             time.Now().UTC().Format(time.RFC3339),
		"Year":            time.Now().Year(),
		"Page":            "explain",
		"OhioASCII":       s.ohioASCII,
		// Footer law link uses internal redirect for stability
		"LawURL": "/law",
	}
	if err := s.tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "template error: %v", err)
	}
}

// handleLaw performs a simple redirect to the configured law URL target.
func (s *Server) handleLaw(w http.ResponseWriter, r *http.Request) {
	// Security headers consistent with explain
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	target := s.lawURL
	http.Redirect(w, r, target, http.StatusFound)
}

// UpdatePolicy swaps the in-memory policy used by the server.
func (s *Server) UpdatePolicy(p *policy.Policy) {
	s.mu.Lock()
	s.policy = p
	s.mu.Unlock()
}

// RecordRefreshSuccess records a successful policy refresh with the given source (e.g., "csv" or "csv-cache").
func (s *Server) RecordRefreshSuccess(source string) {
	s.refreshMu.Lock()
	s.lastRefreshTime = time.Now()
	s.lastRefreshSource = source
	s.refreshCount++
	s.lastRefreshError = ""
	s.refreshMu.Unlock()
}

// RecordRefreshError records a refresh error message for metrics.
func (s *Server) RecordRefreshError(msg string) {
	s.refreshMu.Lock()
	s.lastRefreshTime = time.Now()
	s.refreshErrorCount++
	s.lastRefreshError = msg
	s.refreshMu.Unlock()
}

func (s *Server) getPolicy() *policy.Policy {
	s.mu.RLock()
	p := s.policy
	s.mu.RUnlock()
	return p
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Minimal HTML escaping for rationale/source fields (kept lightweight)
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

// baseCSS is embedded from templates/style.css

// extractOriginalDomainFromHeaders attempts to infer the original requested domain
// from common headers that survive DNS redirects or reverse proxies.
// Precedence (first match wins):
// - X-Original-Host
// - X-Forwarded-Host (first value)
// - Referer (host portion)
// - Host (last resort; may be the redirect host)
func extractOriginalDomainFromHeaders(r *http.Request, allowHostFallback bool) string {
	// X-Original-Host: non-standard but commonly set by proxies
	if v := strings.TrimSpace(r.Header.Get("X-Original-Host")); v != "" {
		return v
	}
	// X-Forwarded-Host: may be a comma-separated list; use first
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
		return v
	}
	// Referer: parse and use host
	if ref := strings.TrimSpace(r.Header.Get("Referer")); ref != "" {
		if u, err := url.Parse(ref); err == nil {
			if u.Host != "" {
				return u.Host
			}
		}
	}
	if allowHostFallback && r.Host != "" {
		return r.Host
	}
	return ""
}
