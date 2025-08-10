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
	fmt.Println("commands: validate, hash, serve, generate-dns, version")
	fmt.Println("generate-dns formats: hosts|bind|unbound|rpz")
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
	format := fs.String("format", "hosts", "Output format: hosts|bind|unbound|rpz")
	mode := fs.String("mode", "a-record", "Mode a-record|cname (cname only for bind/unbound)")
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
