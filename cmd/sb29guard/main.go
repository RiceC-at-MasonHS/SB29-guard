// Command sb29guard provides CLI subcommands to validate policies, compute hashes,
// generate DNS outputs, and serve the explanation web server.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/dnsgen"
	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
	"github.com/RiceC-at-MasonHS/SB29-guard/internal/server"
	"github.com/RiceC-at-MasonHS/SB29-guard/internal/sheets"
)

// version info is injected via -ldflags at release time.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	sub := os.Args[1]
	switch sub {
	case "validate":
		cmdValidate(os.Args[2:])
	case "hash":
		cmdHash(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "generate-dns":
		cmdGenerateDNS(os.Args[2:])
	case "generate-proxy":
		cmdGenerateProxy(os.Args[2:])
	case "generate-explain-static":
		cmdGenerateExplainStatic(os.Args[2:])
	case "version":
		cmdVersion()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("sb29guard <command> [flags]")
	fmt.Println("commands: validate, hash, serve, generate-dns, generate-proxy, generate-explain-static, version")
	fmt.Println("generate-dns formats: hosts|bind|unbound|rpz|dnsmasq|domain-list|winps")
}

func cmdVersion() {
	// Print a compact one-line and JSON for machine-readability
	v := version
	if v == "" {
		v = "dev"
	}
	fmt.Printf("sb29guard %s", v)
	if commit != "" {
		fmt.Printf(" (%s)", commit[:minInt(7, len(commit))])
	}
	if date != "" {
		fmt.Printf(" %s", date)
	}
	fmt.Println()
	// JSON line
	fmt.Printf("{\"version\":%q,\"commit\":%q,\"date\":%q}\n", v, commit, date)
}

// minInt returns the smaller of two ints (named to avoid builtin-name linters)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	policyPath := fs.String("policy", "policy/domains.yaml", "Path to policy file")
	sheetCSV := fs.String("sheet-csv", "", "Published Google Sheet CSV URL (overrides --policy)")
	strict := fs.Bool("strict", true, "Enforce JSON Schema validation")
	_ = fs.Parse(args)
	var p *policy.Policy
	var err error
	if *sheetCSV != "" {
		p, _, err = sheets.FetchCSVPolicyCached(*sheetCSV, "", &http.Client{Timeout: 15 * 1e9})
		if err != nil {
			fmt.Printf("{\"status\":\"error\",\"message\":%q}\n", err.Error())
			os.Exit(1)
		}
	} else {
		data, rerr := os.ReadFile(*policyPath)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", rerr)
			os.Exit(2)
		}
		policy.StrictValidation = *strict
		p, err = policy.Load(data)
		if err != nil {
			fmt.Printf("{\"status\":\"error\",\"message\":%q}\n", err.Error())
			os.Exit(1)
		}
	}
	if err := p.Validate(); err != nil {
		fmt.Printf("{\"status\":\"error\",\"message\":%q}\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("{\"status\":\"ok\",\"records\":%d,\"version\":%q}\n", len(p.Records), p.Version)
}

func cmdHash(args []string) {
	fs := flag.NewFlagSet("hash", flag.ExitOnError)
	policyPath := fs.String("policy", "policy/domains.yaml", "Path to policy file")
	sheetCSV := fs.String("sheet-csv", "", "Published Google Sheet CSV URL (overrides --policy)")
	strict := fs.Bool("strict", true, "Enforce JSON Schema validation")
	_ = fs.Parse(args)
	var p *policy.Policy
	var err error
	if *sheetCSV != "" {
		p, _, err = sheets.FetchCSVPolicyCached(*sheetCSV, "", &http.Client{Timeout: 15 * 1e9})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		b, rerr := os.ReadFile(*policyPath)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", rerr)
			os.Exit(2)
		}
		policy.StrictValidation = *strict
		p, err = policy.Load(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	if err := p.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	h := p.CanonicalHash()
	fmt.Printf("{\"hash\":%q,\"records\":%d}\n", h, len(p.Records))
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	policyPath := fs.String("policy", "policy/domains.yaml", "Path to policy file")
	sheetCSV := fs.String("sheet-csv", "", "Published Google Sheet CSV URL (overrides --policy)")
	listen := fs.String("listen", ":8080", "Listen address host:port")
	refreshAt := fs.String("refresh-at", "23:59", "Daily refresh time local (HH:MM), only with --sheet-csv")
	refreshEvery := fs.Duration("refresh-every", 0, "If >0, refresh policy at this interval instead of daily time (only with --sheet-csv)")
	templatesDir := fs.String("templates", "", "Optional templates directory to override embedded templates")
	_ = fs.Parse(args)
	var p *policy.Policy
	var err error
	if *sheetCSV != "" {
		p, fromCache, err := sheets.FetchCSVPolicyCached(*sheetCSV, "", &http.Client{Timeout: 15 * 1e9})
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid sheet csv: %v\n", err)
			os.Exit(1)
		}
		// Build server with optional template overrides
		var srv *server.Server
		if *templatesDir != "" {
			t := templateFromDir(*templatesDir)
			srv = server.NewWithTemplates(*listen, p, t, cssFromDir(*templatesDir))
		} else {
			srv = server.New(*listen, p)
		}
		src := "csv"
		if fromCache {
			src = "csv-cache"
		}
		fmt.Printf("{\"event\":\"server.start\",\"listen\":%q,\"records\":%d,\"source\":%q}\n", *listen, len(p.Records), src)
		// Start background refresh
		go scheduleCSVRefresh(srv, *sheetCSV, *refreshAt, *refreshEvery)
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	data, rerr := os.ReadFile(*policyPath)
	if rerr != nil {
		fmt.Fprintf(os.Stderr, "error reading policy: %v\n", rerr)
		os.Exit(2)
	}
	p, err = policy.Load(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid policy: %v\n", err)
		os.Exit(1)
	}
	var srv *server.Server
	if *templatesDir != "" {
		t := templateFromDir(*templatesDir)
		srv = server.NewWithTemplates(*listen, p, t, cssFromDir(*templatesDir))
	} else {
		srv = server.New(*listen, p)
	}
	fmt.Printf("{\"event\":\"server.start\",\"listen\":%q,\"records\":%d,\"source\":%q}\n", *listen, len(p.Records), "file")
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// scheduleCSVRefresh refreshes the policy either at a daily HH:MM time or every interval if provided.
func scheduleCSVRefresh(srv *server.Server, csvURL, at string, every time.Duration) {
	client := &http.Client{Timeout: 15 * time.Second}
	// helper to perform one refresh
	doRefresh := func() {
		fmt.Printf("{\"event\":\"policy.refresh.start\",\"time\":%q}\n", time.Now().Format(time.RFC3339))
		p, fromCache, err := sheets.FetchCSVPolicyCached(csvURL, "", client)
		if err != nil {
			fmt.Printf("{\"event\":\"policy.refresh.error\",\"message\":%q}\n", err.Error())
			srv.RecordRefreshError(err.Error())
			return
		}
		srv.UpdatePolicy(p)
		src := "csv"
		if fromCache {
			src = "csv-cache"
		}
		srv.RecordRefreshSuccess(src)
		fmt.Printf("{\"event\":\"policy.refresh.success\",\"records\":%d,\"source\":%q,\"version\":%q}\n", len(p.Records), src, p.Version)
	}

	if every > 0 {
		fmt.Printf("{\"event\":\"policy.refresh.mode\",\"interval\":%q}\n", every.String())
		ticker := time.NewTicker(every)
		for range ticker.C {
			doRefresh()
		}
	} else {
		// Parse HH:MM
		hour, minute := 23, 59
		if len(at) >= 4 {
			if t, err := time.Parse("15:04", at); err == nil {
				hour, minute = t.Hour(), t.Minute()
			}
		}
		for {
			next := nextDailyTime(hour, minute)
			fmt.Printf("{\"event\":\"policy.refresh.scheduled\",\"next\":%q}\n", next.Format(time.RFC3339))
			time.Sleep(time.Until(next))
			doRefresh()
		}
	}
}

func nextDailyTime(hour, minute int) time.Time {
	now := time.Now()
	loc := now.Location()
	candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if now.Before(candidate) {
		return candidate
	}
	return candidate.Add(24 * time.Hour)
}

func cmdGenerateDNS(args []string) {
	fs := flag.NewFlagSet("generate-dns", flag.ExitOnError)
	policyPath := fs.String("policy", "policy/domains.yaml", "Path to policy file")
	sheetCSV := fs.String("sheet-csv", "", "Published Google Sheet CSV URL (overrides --policy)")
	out := fs.String("out", "", "Output file path (required unless --dry-run)")
	format := fs.String("format", "hosts", "Output format: hosts|bind|unbound|rpz|dnsmasq|domain-list|winps")
	mode := fs.String("mode", "a-record", "Mode a-record|cname")
	redirectIPv4 := fs.String("redirect-ipv4", "", "Redirect IPv4 address (required for a-record/hosts)")
	redirectHost := fs.String("redirect-host", "blocked.guard.local", "Redirect host (for cname mode)")
	ttl := fs.Int("ttl", 300, "Record TTL seconds")
	serialStrategy := fs.String("serial-strategy", "date", "Serial strategy for bind/rpz: date|epoch|hash")
	dryRun := fs.Bool("dry-run", false, "Print to stdout instead of writing file")
	_ = fs.Parse(args)
	var p *policy.Policy
	var err error
	if *sheetCSV != "" {
		p, _, err = sheets.FetchCSVPolicyCached(*sheetCSV, "", &http.Client{Timeout: 15 * 1e9})
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid sheet csv: %v\n", err)
			os.Exit(1)
		}
	} else {
		if *policyPath == "" {
			fmt.Fprintln(os.Stderr, "--policy required (or use --sheet-csv)")
			os.Exit(2)
		}
		data, rerr := os.ReadFile(*policyPath)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "error reading policy: %v\n", rerr)
			os.Exit(2)
		}
		p, err = policy.Load(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid policy: %v\n", err)
			os.Exit(1)
		}
	}
	opts := dnsgen.Options{Format: *format, Mode: *mode, RedirectIPv4: *redirectIPv4, RedirectHost: *redirectHost, TTL: *ttl, SerialStrategy: *serialStrategy}
	content, err := dnsgen.Generate(p, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generation error: %v\n", err)
		os.Exit(1)
	}
	if *dryRun {
		fmt.Print(string(content))
		return
	}
	if *out == "" {
		fmt.Fprintln(os.Stderr, "--out required unless --dry-run")
		os.Exit(2)
	}
	if err := os.MkdirAll(dirOf(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir error: %v\n", err)
		os.Exit(2)
	}
	if err := os.WriteFile(*out, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("{\"status\":\"ok\",\"format\":%q,\"mode\":%q,\"bytes\":%d}\n", *format, *mode, len(content))
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

// templateFromDir loads layout.html, explain.html, and root.html from dir, in that order.
func templateFromDir(dir string) *template.Template {
	files := []string{
		filepath.Join(dir, "layout.html"),
		filepath.Join(dir, "explain.html"),
		filepath.Join(dir, "root.html"),
	}
	t := template.New("layout.html")
	tmpl, err := t.ParseFiles(files...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "template parse error: %v\n", err)
		os.Exit(1)
	}
	return tmpl
}

// cssFromDir returns the contents of style.css if present; otherwise returns an empty string.
func cssFromDir(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, "style.css"))
	if err != nil {
		return ""
	}
	return string(b)
}

// cmdGenerateProxy emits reverse proxy config snippets for School Mode.
// Supports two modes: header-injection (reverse proxy to backend) and redirect (302 to static explain page).
func cmdGenerateProxy(args []string) {
	fs := flag.NewFlagSet("generate-proxy", flag.ExitOnError)
	format := fs.String("format", "caddy", "caddy|nginx|haproxy|apache")
	mode := fs.String("mode", "header-injection", "header-injection|redirect")
	siteHost := fs.String("site-host", "blocked.example", "Virtual host name handling blocked flows")
	backendURL := fs.String("backend-url", "http://127.0.0.1:8080", "Backend SB29 Guard URL (header-injection mode)")
	explainURL := fs.String("explain-url", "https://explain.example/explain", "Public explain page URL (redirect mode)")
	out := fs.String("out", "", "Output file (optional; prints to stdout when empty)")
	dryRun := fs.Bool("dry-run", false, "Print to stdout even if --out is set")
	bundleDir := fs.String("bundle-dir", "", "If set and format=nginx, write a ready-to-use bundle into this directory")
	tlsCert := fs.String("tls-cert", "", "TLS certificate path for HTTPS vhost (nginx bundle)")
	tlsKey := fs.String("tls-key", "", "TLS key path for HTTPS vhost (nginx bundle)")
	policyPath := fs.String("policy", "", "Policy file to derive selective routing map (optional)")
	sheetCSV := fs.String("sheet-csv", "", "Published Google Sheet CSV URL to derive map (optional)")
	redirectUnknown := fs.Bool("redirect-unknown", false, "In nginx bundle, intercept 404 from guard and redirect to static explain at --explain-url?d=$host")
	_ = fs.Parse(args)

	// Bundle mode for nginx
	if *bundleDir != "" {
		switch strings.ToLower(*format) {
		case "nginx":
			if err := writeNginxBundle(*bundleDir, *mode, *siteHost, *backendURL, *explainURL, *tlsCert, *tlsKey, *policyPath, *sheetCSV, *redirectUnknown); err != nil {
				fmt.Fprintf(os.Stderr, "bundle error: %v\n", err)
				os.Exit(1)
			}
		case "caddy":
			if err := writeCaddyBundle(*bundleDir, *mode, *siteHost, *backendURL, *explainURL); err != nil {
				fmt.Fprintf(os.Stderr, "bundle error: %v\n", err)
				os.Exit(1)
			}
		case "haproxy":
			if err := writeHAProxyBundle(*bundleDir, *mode, *siteHost, *backendURL, *explainURL, *policyPath, *sheetCSV); err != nil {
				fmt.Fprintf(os.Stderr, "bundle error: %v\n", err)
				os.Exit(1)
			}
		case "apache":
			if err := writeApacheBundle(*bundleDir, *mode, *siteHost, *backendURL, *explainURL); err != nil {
				fmt.Fprintf(os.Stderr, "bundle error: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unsupported format for bundle: %s\n", *format)
			os.Exit(2)
		}
		fmt.Printf("{\"status\":\"ok\",\"bundle\":%q,\"format\":%q}\n", *bundleDir, *format)
		return
	}

	cfg, err := renderProxySnippet(*format, *mode, *siteHost, *backendURL, *explainURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	if *dryRun || *out == "" {
		fmt.Print(cfg)
		return
	}
	if err := os.MkdirAll(dirOf(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir error: %v\n", err)
		os.Exit(2)
	}
	if err := os.WriteFile(*out, []byte(cfg), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("{\"status\":\"ok\",\"format\":%q,\"mode\":%q,\"bytes\":%d}\n", *format, *mode, len(cfg))
}

func renderProxySnippet(format, mode, siteHost, backendURL, explainURL string) (string, error) {
	f := strings.ToLower(strings.TrimSpace(format))
	m := strings.ToLower(strings.TrimSpace(mode))
	switch f {
	case "caddy":
		if m == "header-injection" {
			return fmt.Sprintf(`# Caddyfile: explanatory vhost for blocked traffic
%s {
	encode zstd gzip
	@all {
		path *
	}
	reverse_proxy @all %s {
		header_up X-Original-Host {host}
		header_up X-Forwarded-Host {host}
	}
}
`, siteHost, backendURL), nil
		}
		if m == "redirect" {
			return fmt.Sprintf(`# Caddyfile: 302 to static explain page with display-only params
%s {
	@any path *
	redir @any %s?d={host} 302
}
`, siteHost, explainURL), nil
		}
	case "nginx":
		if m == "header-injection" {
			return fmt.Sprintf(`# nginx: explanatory server for blocked traffic
server {
	listen 80;
	server_name %s;
	location / {
		proxy_set_header X-Original-Host $host;
		proxy_set_header X-Forwarded-Host $host;
		proxy_pass %s;
	}
}
`, siteHost, backendURL), nil
		}
		if m == "redirect" {
			return fmt.Sprintf(`# nginx: 302 to static explain page
server {
	listen 80;
	server_name %s;
	location / {
		return 302 %s?d=$host;
	}
}
`, siteHost, explainURL), nil
		}
	case "haproxy":
		if m == "header-injection" {
			return fmt.Sprintf(`# HAProxy: explanatory frontend/backend
frontend fe_explain
	bind *:80
	mode http
	acl vhost hdr(host) -i %s
	use_backend be_guard if vhost

backend be_guard
	mode http
	http-request set-header X-Original-Host %%[req.hdr(host)]
	http-request set-header X-Forwarded-Host %%[req.hdr(host)]
	server s1 %s
`, siteHost, strings.TrimPrefix(strings.TrimPrefix(backendURL, "http://"), "https://")), nil
		}
		if m == "redirect" {
			return fmt.Sprintf(`# HAProxy: 302 redirect to static explain
frontend fe_explain
	bind *:80
	mode http
	acl vhost hdr(host) -i %s
	http-request redirect code 302 location %s?d=%%[req.hdr(host)] if vhost
`, siteHost, explainURL), nil
		}
	case "apache":
		if m == "header-injection" {
			return fmt.Sprintf(`# Apache httpd: explanatory vhost
<VirtualHost *:80>
	ServerName %s
	ProxyPreserveHost On
	RequestHeader set X-Original-Host "%%{Host}i"
	RequestHeader set X-Forwarded-Host "%%{Host}i"
	ProxyPass / %s
	ProxyPassReverse / %s
</VirtualHost>
`, siteHost, ensureTrailingSlash(backendURL), ensureTrailingSlash(backendURL)), nil
		}
		if m == "redirect" {
			return fmt.Sprintf(`# Apache httpd: 302 redirect to static explain
<VirtualHost *:80>
	ServerName %s
	Redirect 302 / %s?d=%%{HTTP_HOST}
</VirtualHost>
`, siteHost, explainURL), nil
		}
	}
	return "", fmt.Errorf("unsupported format %q or mode %q", format, mode)
}

func ensureTrailingSlash(u string) string {
	if strings.HasSuffix(u, "/") {
		return u
	}
	return u + "/"
}

// writeNginxBundle assembles a ready-to-use directory with site.conf, optional blocked_map.conf, smoke.ps1, and README.md
func writeNginxBundle(dir, mode, siteHost, backendURL, explainURL, tlsCert, tlsKey, policyPath, sheetCSV string, redirectUnknown bool) error {
	// Create dir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	// Build server block with optional TLS and intercept
	sb := &strings.Builder{}
	// HTTP server: either redirect to HTTPS if TLS provided or serve normally
	if strings.TrimSpace(tlsCert) != "" && strings.TrimSpace(tlsKey) != "" {
		fmt.Fprintf(sb, "server {\n    listen 80;\n    server_name %s;\n    return 301 https://$host$request_uri;\n}\n\n", siteHost)
		// HTTPS server
		fmt.Fprintf(sb, "server {\n    listen 443 ssl;\n    server_name %s;\n    ssl_certificate %s;\n    ssl_certificate_key %s;\n", siteHost, tlsCert, tlsKey)
	} else {
		fmt.Fprintf(sb, "server {\n    listen 80;\n    server_name %s;\n", siteHost)
	}
	// Common server content depending on mode
	if strings.ToLower(mode) == "header-injection" {
		if redirectUnknown {
			fmt.Fprintln(sb, "    proxy_intercept_errors on;")
			fmt.Fprintln(sb, "    error_page 404 = @static_explain;")
		}
		fmt.Fprintln(sb, "    location / {")
		fmt.Fprintln(sb, "        proxy_set_header X-Original-Host $host;")
		fmt.Fprintln(sb, "        proxy_set_header X-Forwarded-Host $host;")
		fmt.Fprintf(sb, "        proxy_pass %s;\n", backendURL)
		fmt.Fprintln(sb, "    }")
		if redirectUnknown {
			fmt.Fprintf(sb, "    location @static_explain { return 302 %s?d=$host; }\n", explainURL)
		}
	} else { // redirect mode
		fmt.Fprintln(sb, "    location / {\n        return 302 "+explainURL+"?d=$host;\n    }")
	}
	fmt.Fprintln(sb, "}")
	if err := os.WriteFile(filepath.Join(dir, "site.conf"), []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("write site.conf: %w", err)
	}

	// Optional: write blocked_map.conf if policy provided
	var mapContent string
	if strings.TrimSpace(policyPath) != "" || strings.TrimSpace(sheetCSV) != "" {
		p, err := loadPolicyFromInputs(policyPath, sheetCSV)
		if err != nil {
			return fmt.Errorf("load policy: %w", err)
		}
		mc := &strings.Builder{}
		fmt.Fprintln(mc, "# Include this file inside the 'http {}' context in nginx.conf")
		fmt.Fprintln(mc, "# Example usage: if ($sb29_blocked) { return 302 https://your-guard-host/; }")
		fmt.Fprintln(mc, "map $host $sb29_blocked {")
		fmt.Fprintln(mc, "    default 0;")
		for _, r := range p.Records {
			if strings.TrimSpace(r.Status) == "suspended" {
				continue
			}
			d := strings.ToLower(strings.TrimSpace(r.Domain))
			if strings.HasPrefix(d, "*.") {
				base := strings.TrimPrefix(d, "*.")
				// Regex: match base domain or any subdomain of it
				re := "~^(?:.*\\.)?" + nginxRegexEscape(base) + "$"
				fmt.Fprintf(mc, "    %s 1;\n", re)
			} else {
				fmt.Fprintf(mc, "    %s 1;\n", d)
			}
		}
		fmt.Fprintln(mc, "}")
		mapContent = mc.String()
		if err := os.WriteFile(filepath.Join(dir, "blocked_map.conf"), []byte(mapContent), 0o644); err != nil {
			return fmt.Errorf("write blocked_map.conf: %w", err)
		}
	}

	// Smoke test script (PowerShell)
	smoke := fmt.Sprintf(`param(
  [string]$Guard = "%s",
  [string]$HostName = "%s"
)
Write-Host "Checking /health on $Guard"
try { $h = Invoke-WebRequest -UseBasicParsing -Uri "$Guard/health"; Write-Host "Health: $($h.StatusCode)" } catch { Write-Host "Health failed: $_" }
Write-Host "Checking /explain with X-Original-Host=$HostName"
try {
  $r = Invoke-WebRequest -UseBasicParsing -Uri "$Guard/explain" -Headers @{ 'X-Original-Host'=$HostName }
  Write-Host "Explain: $($r.StatusCode)"
} catch { Write-Host "Explain failed: $_" }
`, strings.TrimSuffix(backendURL, "/"), siteHost)
	if err := os.WriteFile(filepath.Join(dir, "smoke.ps1"), []byte(smoke), 0o644); err != nil {
		return fmt.Errorf("write smoke.ps1: %w", err)
	}

	// README
	rd := &strings.Builder{}
	fmt.Fprintf(rd, "# SB29 Guard NGINX Bundle\n\n")
	fmt.Fprintf(rd, "Files:\n- site.conf: server block for %s (%s)\n", siteHost, mode)
	if mapContent != "" {
		fmt.Fprintln(rd, "- blocked_map.conf: map of denylisted hosts -> $sb29_blocked=1 (include under http {})")
	}
	fmt.Fprintln(rd, "- smoke.ps1: quick health check")
	fmt.Fprintln(rd, "\nQuick Start:\n1) Place site.conf in your NGINX sites-available and enable it.\n2) If using TLS, ensure cert/key paths are correct.\n3) Reload NGINX.\n4) Run smoke.ps1 -Guard 'http://127.0.0.1:8081' -HostName 'exampletool.com'.")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(rd.String()), 0o644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}
	return nil
}

func nginxRegexEscape(s string) string {
	// Escape dots and other regex meta; keep hyphens
	replacer := strings.NewReplacer(
		`.`, `\\.`,
		`+`, `\\+`,
		`?`, `\\?`,
		`*`, `\\*`,
		`^`, `\\^`,
		`$`, `\\$`,
		`(`, `\\(`,
		`)`, `\\)`,
		`[`, `\\[`,
		`]`, `\\]`,
		`{`, `\\{`,
		`}`, `\\}`,
		`|`, `\\|`,
		`\`, `\\\\`,
	)
	return replacer.Replace(s)
}

// loadPolicyFromInputs mirrors load logic from other commands
func loadPolicyFromInputs(policyPath, sheetCSV string) (*policy.Policy, error) {
	if strings.TrimSpace(sheetCSV) != "" {
		p, _, err := sheets.FetchCSVPolicyCached(sheetCSV, "", &http.Client{Timeout: 15 * 1e9})
		return p, err
	}
	if strings.TrimSpace(policyPath) == "" {
		return nil, fmt.Errorf("--policy or --sheet-csv required for map generation")
	}
	b, err := os.ReadFile(policyPath)
	if err != nil {
		return nil, err
	}
	return policy.Load(b)
}

// writeCaddyBundle emits a minimal Caddyfile and README
func writeCaddyBundle(dir, mode, siteHost, backendURL, explainURL string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var caddy string
	if strings.ToLower(mode) == "header-injection" {
		caddy = fmt.Sprintf(`%s {
	encode zstd gzip
	@all {
		path *
	}
	reverse_proxy @all %s {
		header_up X-Original-Host {host}
		header_up X-Forwarded-Host {host}
	}
}
`, siteHost, backendURL)
	} else {
		caddy = fmt.Sprintf(`%s {
	@any path *
	redir @any %s?d={host} 302
}
`, siteHost, explainURL)
	}
	if err := os.WriteFile(filepath.Join(dir, "Caddyfile"), []byte(caddy), 0o644); err != nil {
		return err
	}
	readme := fmt.Sprintf("# SB29 Guard Caddy Bundle\n\n- Caddyfile for %s (%s)\n- Run: caddy run --config Caddyfile\n", siteHost, mode)
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}

// writeHAProxyBundle emits haproxy.cfg and optional map of blocked hosts
func writeHAProxyBundle(dir, mode, siteHost, backendURL, explainURL, policyPath, sheetCSV string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Basic haproxy config focusing on the guard vhost
	cfg := &strings.Builder{}
	fmt.Fprint(cfg, "global\n    maxconn 1000\n")
	fmt.Fprint(cfg, "defaults\n    mode http\n    timeout connect 5s\n    timeout client  30s\n    timeout server  30s\n")
	fmt.Fprintf(cfg, "frontend fe_guard\n    bind *:80\n    acl vhost hdr(host) -i %s\n", siteHost)
	if strings.ToLower(mode) == "header-injection" {
		fmt.Fprintln(cfg, "    use_backend be_guard if vhost")
		fmt.Fprintf(cfg, "backend be_guard\n    http-request set-header X-Original-Host %%[req.hdr(host)]\n    http-request set-header X-Forwarded-Host %%[req.hdr(host)]\n    server s1 %s\n", strings.TrimPrefix(strings.TrimPrefix(backendURL, "http://"), "https://"))
	} else {
		// redirect mode
		fmt.Fprintf(cfg, "    http-request redirect code 302 location %s?d=%%[req.hdr(host)] if vhost\n", explainURL)
	}
	if err := os.WriteFile(filepath.Join(dir, "haproxy.cfg"), []byte(cfg.String()), 0o644); err != nil {
		return err
	}
	// Optional map file for blocked hosts (for selective routing in user’s wider config)
	if strings.TrimSpace(policyPath) != "" || strings.TrimSpace(sheetCSV) != "" {
		p, err := loadPolicyFromInputs(policyPath, sheetCSV)
		if err != nil {
			return err
		}
		mb := &strings.Builder{}
		for _, r := range p.Records {
			if strings.TrimSpace(r.Status) == "suspended" {
				continue
			}
			d := strings.ToLower(strings.TrimSpace(r.Domain))
			if strings.HasPrefix(d, "*.") {
				base := strings.TrimPrefix(d, "*.")
				// Map semantics: wildcard noted with a leading dot to match subdomains per HAProxy's -i -m dom
				fmt.Fprintf(mb, ".%s 1\n", base)
				fmt.Fprintf(mb, "%s 1\n", base)
			} else {
				fmt.Fprintf(mb, "%s 1\n", d)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "blocked.map"), []byte(mb.String()), 0o644); err != nil {
			return err
		}
	}
	readme := "# SB29 Guard HAProxy Bundle\n\n- haproxy.cfg: guard frontend/backend for \n- blocked.map: optional host map for selective routing in your config\n"
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}

// writeApacheBundle emits a vhost conf and README
func writeApacheBundle(dir, mode, siteHost, backendURL, explainURL string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var conf string
	if strings.ToLower(mode) == "header-injection" {
		conf = fmt.Sprintf(`<VirtualHost *:80>
	ServerName %s
	ProxyPreserveHost On
	RequestHeader set X-Original-Host "%%{Host}i"
	RequestHeader set X-Forwarded-Host "%%{Host}i"
	ProxyPass / %s
	ProxyPassReverse / %s
</VirtualHost>
`, siteHost, ensureTrailingSlash(backendURL), ensureTrailingSlash(backendURL))
	} else {
		conf = fmt.Sprintf(`<VirtualHost *:80>
	ServerName %s
	Redirect 302 / %s?d=%%{HTTP_HOST}
</VirtualHost>
`, siteHost, explainURL)
	}
	if err := os.WriteFile(filepath.Join(dir, "guard.conf"), []byte(conf), 0o644); err != nil {
		return err
	}
	readme := fmt.Sprintf("# SB29 Guard Apache Bundle\n\n- guard.conf for %s (%s)\n- Enable required modules: proxy, proxy_http, headers\n", siteHost, mode)
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644)
}

// cmdGenerateExplainStatic writes a static HTML bundle that reads d,c,v,h query params client-side.
func cmdGenerateExplainStatic(args []string) {
	fs := flag.NewFlagSet("generate-explain-static", flag.ExitOnError)
	outDir := fs.String("out-dir", "dist/explain", "Output directory for static bundle")
	title := fs.String("title", "SB29 Guard", "Page title")
	lawURL := fs.String("law-url", "https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/", "Law reference URL")
	inlineCSS := fs.Bool("inline-css", true, "Inline CSS into index.html (else writes style.css)")
	_ = fs.Parse(args)

	if strings.TrimSpace(*outDir) == "" {
		fmt.Fprintln(os.Stderr, "--out-dir is required")
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir error: %v\n", err)
		os.Exit(2)
	}
	// Build index.html content
	css := defaultStaticCSS
	if !*inlineCSS {
		if err := os.WriteFile(filepath.Join(*outDir, "style.css"), []byte(css), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write style.css error: %v\n", err)
			os.Exit(2)
		}
		css = ""
	}
	html := renderStaticExplainHTML(*title, *lawURL, css)
	if err := os.WriteFile(filepath.Join(*outDir, "index.html"), []byte(html), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write index.html error: %v\n", err)
		os.Exit(2)
	}
	// README.md with quick usage
	readme := `# Static Explain Page

This bundle renders an explanation page using URL parameters:
- d: original domain (hostname)
- c: classification (optional, display-only)
- v: policy version (optional)
- h: policy hash short (optional)

Example:
  https://` + "${YOUR_HOST}" + `/index.html?d=example.com&c=NO_DPA&v=0.1.0

Security notes:
- Values are sanitized client-side. Do not include untrusted HTML.
- For strict CSP, serve with appropriate headers on your web server.
`
	if err := os.WriteFile(filepath.Join(*outDir, "README.md"), []byte(readme), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write README.md error: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("{\"status\":\"ok\",\"out_dir\":%q}\n", *outDir)
}

const defaultStaticCSS = `:root{color-scheme:dark light;font-family:system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,sans-serif;--bg:#0f1115;--fg:#e7e9ee;--accent:#62a8ff;--badge:#ff6b60;--panel:#151923;--muted:#9aa4b2;--ring:#2a3342}
@media (prefers-color-scheme:light){:root{--bg:#ffffff;--fg:#222;--accent:#004c99;--badge:#d9534f;--panel:#f5f7fa;--muted:#5b6673;--ring:#dfe6ef}}
body{margin:0;background:var(--bg);color:var(--fg);line-height:1.5}
header,main,footer{max-width:900px;margin:0 auto;padding:1rem}
header{border-bottom:1px solid #ccc2}
footer{border-top:1px solid #ccc2;margin-top:2rem;font-size:.85rem;opacity:.85}
h1{font-size:1.5rem;margin:.2rem 0 .6rem}
h2{font-size:1.1rem;margin:1rem 0 .5rem}
code{background:var(--panel);padding:2px 6px;border-radius:6px;font-size:.875em}
.muted{color:var(--muted)}
.explain-page{display:grid;grid-template-columns:1fr;gap:1rem;align-items:start}
.card{background:var(--panel);border:1px solid var(--ring);border-radius:14px;padding:1rem 1.1rem;box-shadow:0 6px 24px rgba(0,0,0,.25)}
.card-title{margin:.2rem 0 .6rem;font-size:1.2rem;line-height:1.25}
.badge{display:inline-block;padding:.35rem .7rem;background:var(--badge);color:#fff;border-radius:999px;font-size:.75rem;font-weight:700;letter-spacing:.4px;text-transform:uppercase}
`

func renderStaticExplainHTML(title, lawURL, inlineCSS string) string {
	// very small JS to parse query params and safely inject as text
	// CSP: if inline CSS used, a meta CSP cannot allow style-src unsafe-inline; advise setting headers server-side
	cssTag := ""
	if inlineCSS != "" {
		cssTag = "<style>" + inlineCSS + "</style>"
	} else {
		cssTag = "<link rel=\"stylesheet\" href=\"style.css\">"
	}
	return "<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">" +
		"<title>" + template.HTMLEscapeString(title) + "</title>" + cssTag +
		"</head><body><header><h1>SB29 Guard</h1><p class=\"muted\">Access Redirected</p></header><main>" +
		`<section class="card">
  <h2 class="card-title"><span class="muted">Access to</span> <span id="domain" class="domain"></span> <span class="muted">is restricted</span></h2>
  <p class="chips"><span id="classification" class="badge"></span></p>
  <dl class="meta-grid">
	<div><dt>Policy</dt><dd id="policyVersion"></dd></div>
	<div><dt>UTC</dt><dd id="now"></dd></div>
  </dl>
  <p class="contact">See your instructional technology team for help.</p>
</section>` +
		"</main><footer><p>Policy reference: <a href=\"" + template.HTMLEscapeString(lawURL) + "\">SB29</a></p></footer>" +
		`<script>(function(){
  function qp(k){const u=new URL(window.location.href);return (u.searchParams.get(k)||"").trim();}
  function setText(id,val){var el=document.getElementById(id); if(!el) return; el.textContent = val || ''}
  function sanitizeHost(h){try{h=h.replace(/^\s+|\s+$/g,''); if(h.indexOf('://')>=0){var u=new URL(h); return u.host.toLowerCase();} return h.toLowerCase();}catch(e){return h;}}
  var d = qp('d') || qp('domain') || qp('original') || qp('url');
  if(d){ d = sanitizeHost(d); if(d.startsWith('www.')) d = d.substring(4); }
  setText('domain', d||'');
  var c = qp('c'); if(c){ setText('classification', c); }
  var v = qp('v'); if(v){ setText('policyVersion', 'v'+v); }
  setText('now', new Date().toISOString());
})();</script>` +
		"</body></html>"
}
