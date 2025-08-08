// Package server provides the HTTP explanation, landing, and health endpoints.
package server

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed templates/style.css
var baseCSS string

// Server hosts HTTP endpoints rendering policy-based explanations.
type Server struct {
	addr      string
	policy    *policy.Policy
	tmpl      *template.Template
	inlineCSS string
}

// New creates a new Server bound to addr using the supplied policy.
func New(addr string, p *policy.Policy) *Server {
	t := template.New("layout.html").Funcs(template.FuncMap{})
	// Parse layout, then explain, then root (root last => its blocks override for landing page)
	tmpl, err := t.ParseFS(templateFS, "templates/layout.html", "templates/explain.html", "templates/root.html")
	if err != nil {
		panic(fmt.Sprintf("template parse error: %v", err))
	}
	return &Server{addr: addr, policy: p, tmpl: tmpl, inlineCSS: baseCSS}
}

// Start begins serving HTTP until the listener stops.
func (s *Server) Start() error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/explain", s.handleExplain)
	http.HandleFunc("/", s.handleRoot)
	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"status":"ok","policy_version":%q}`, s.policy.Version)
}

func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"CSS":           s.inlineCSS,
		"RecordCount":   len(s.policy.Records),
		"Year":          time.Now().Year(),
		"PolicyVersion": s.policy.Version,
	}
	if err := s.tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "template error: %v", err)
	}
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	q := r.URL.Query()
	orig := firstNonEmpty(q.Get("original_domain"), q.Get("original"), q.Get("domain"), q.Get("url"))
	if orig == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "<p>Missing required parameter ?original_domain= (or original/domain/url)</p>")
		return
	}
	if strings.Contains(orig, "://") { // full URL passed
		if u, err := url.Parse(orig); err == nil && u.Host != "" {
			orig = u.Host
		}
	}
	orig = strings.ToLower(strings.TrimSpace(orig))
	lookupDomain := strings.TrimPrefix(orig, "www.")
	rec, ok := s.policy.Lookup(lookupDomain)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "<html><body><h1>Not Classified</h1><p>The domain %s is not present in the active policy set.</p><p>Policy Version: %s</p></body></html>", orig, s.policy.Version)
		return
	}

	// Security / privacy oriented headers
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self';")

	data := map[string]interface{}{
		"CSS":            s.inlineCSS,
		"Original":       orig,
		"Classification": rec.Classification,
		"Rationale":      htmlEscape(rec.Rationale),
		"SourceRef":      htmlEscape(rec.SourceRef),
		"PolicyVersion":  s.policy.Version,
		"Now":            time.Now().UTC().Format(time.RFC3339),
		"Year":           time.Now().Year(),
	}
	if err := s.tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "template error: %v", err)
	}
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
