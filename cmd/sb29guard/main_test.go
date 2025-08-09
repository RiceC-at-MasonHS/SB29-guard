package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// helper to write a temporary policy file
func writeTempPolicy(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	p := filepath.Join(d, "policy.yaml")
	content := "version: 0.1.0\n" +
		"updated: 2025-08-08\n" +
		"records:\n" +
		"  - domain: \"example.com\"\n" +
		"    classification: NO_DPA\n" +
		"    rationale: valid rationale\n" +
		"    last_review: 2025-08-01\n" +
		"    status: active\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp policy: %v", err)
	}
	return p
}

func TestCLIValidate(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "validate", "--policy", policyPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "\"status\":\"ok\"") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCLIHash(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "hash", "--policy", policyPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hash failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "\"hash\":") {
		t.Fatalf("unexpected hash output: %s", out)
	}
}

func TestCLIGenerateDNSDryRun(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "generate-dns", "--policy", policyPath, "--format", "hosts", "--redirect-ipv4", "10.0.0.1", "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate-dns dry-run failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "example.com") {
		t.Fatalf("expected domain in output: %s", out)
	}
}

func TestCLIUnknown(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "unknown-cmd")
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error for unknown command")
	}
}

func TestCLIValidateSheetCSV(t *testing.T) {
	// Serve a small valid CSV
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("domain,classification,rationale,last_review,status\nexample.com,NO_DPA,Reason,2025-08-01,active\n"))
	}))
	defer ts.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "validate", "--sheet-csv", ts.URL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate sheet-csv failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "\"status\":\"ok\"") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCLIHashSheetCSV(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("domain,classification,rationale,last_review,status\nexample.com,NO_DPA,Reason,2025-08-01,active\n"))
	}))
	defer ts.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "hash", "--sheet-csv", ts.URL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hash sheet-csv failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "\"hash\":") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCLIValidateMissingPolicy(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "validate", "--policy", "nonexistent.yaml")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for missing policy. out=%s", out)
	}
}

func TestCLIHashMissingPolicy(t *testing.T) {
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "hash", "--policy", "nonexistent.yaml")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for missing policy. out=%s", out)
	}
}

func TestCLIGenerateDNSMissingOut(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	// No --out and not --dry-run should fail
	cmd := exec.Command(bin, "generate-dns", "--policy", policyPath, "--format", "hosts", "--redirect-ipv4", "10.0.0.1")
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected failure without --out and no --dry-run")
	}
}

func TestCLIGenerateDNSWritesFile(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	td := t.TempDir()
	outPath := filepath.Join(td, "dist", "dns", "hosts.txt")
	cmd := exec.Command(bin, "generate-dns", "--policy", policyPath, "--format", "hosts", "--redirect-ipv4", "10.0.0.1", "--out", outPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate-dns write failed: %v output=%s", err, out)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file written: %v", err)
	}
}

func TestCLIServePortInUse(t *testing.T) {
	// Occupy a port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	defer ln.Close()
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "serve", "--policy", policyPath, "--listen", addr)
	// Expect failure because port is in use
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected serve to fail when port is in use")
	}
}

func TestDirOf(t *testing.T) {
	if dir := dirOf("/a/b/c.txt"); dir != "/a/b" {
		t.Fatalf("unexpected dir: %s", dir)
	}
	if dir := dirOf("a.txt"); dir != "." {
		t.Fatalf("unexpected dir: %s", dir)
	}
	// Windows-style
	if dir := dirOf("C:\\foo\\bar\\baz.txt"); !strings.HasSuffix(strings.ToLower(dir), "c:\\foo\\bar") {
		t.Fatalf("unexpected dir: %s", dir)
	}
}

func TestCLIGenerateDNSCNAME_DryRun(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "generate-dns", "--policy", policyPath, "--format", "bind", "--mode", "cname", "--redirect-host", "blocked.guard.local", "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate-dns cname dry-run failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "blocked.guard.local") {
		t.Fatalf("expected redirect host in output: %s", out)
	}
}

func TestCLIGenerateDNSRPZ_DryRun(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "generate-dns", "--policy", policyPath, "--format", "rpz", "--mode", "cname", "--redirect-host", "blocked.guard.local", "--serial-strategy", "hash", "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate-dns rpz dry-run failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "blocked.guard.local") {
		t.Fatalf("expected redirect host in rpz: %s", out)
	}
}

func TestCLIValidateStrictFalse(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "validate", "--policy", policyPath, "--strict=false")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("validate strict=false failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "\"status\":\"ok\"") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCLIHashStrictFalse(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "hash", "--policy", policyPath, "--strict=false")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hash strict=false failed: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "\"hash\":") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCLIGenDNSMissingRedirectIP(t *testing.T) {
	policyPath := writeTempPolicy(t)
	bin := buildTestBinary(t)
	// hosts (a-record) without redirect-ipv4 and dry-run should error from generator
	cmd := exec.Command(bin, "generate-dns", "--policy", policyPath, "--format", "hosts", "--dry-run")
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected error when redirect-ipv4 missing for hosts format")
	}
}

func TestCLIValidateSheetCSVError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer ts.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "validate", "--sheet-csv", ts.URL)
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected non-zero exit for sheet-csv error")
	}
}

func TestCLIHashSheetCSVError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer ts.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "hash", "--sheet-csv", ts.URL)
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected non-zero exit for sheet-csv error")
	}
}

func TestCLIServeSheetCSVHappy(t *testing.T) {
	// CSV server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("domain,classification,rationale,last_review,status\nexample.com,NO_DPA,Reason,2025-08-01,active\n"))
	}))
	defer ts.Close()
	// choose a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "serve", "--sheet-csv", ts.URL, "--listen", addr)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start serve: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	// poll /health
	url := "http://" + addr + "/health"
	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			if resp.StatusCode == 200 {
				return
			}
			lastErr = nil
		} else {
			lastErr = err
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("serve did not become healthy: %v", lastErr)
}

// helper to write a minimal templates directory for override
func writeTempTemplates(t *testing.T, marker string) string {
	t.Helper()
	d := t.TempDir()
	// layout renders a recognizable marker and policy version
	layout := "<html><body>OVERRIDE:" + marker + " Pv={{.PolicyVersion}}</body></html>"
	if err := os.WriteFile(filepath.Join(d, "layout.html"), []byte(layout), 0o644); err != nil {
		t.Fatalf("write layout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(d, "explain.html"), []byte(""), 0o644); err != nil {
		t.Fatalf("write explain: %v", err)
	}
	if err := os.WriteFile(filepath.Join(d, "root.html"), []byte(""), 0o644); err != nil {
		t.Fatalf("write root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(d, "style.css"), []byte(":root{--accent:#123456;}"), 0o644); err != nil {
		t.Fatalf("write css: %v", err)
	}
	return d
}

func TestCLIServeWithTemplatesOverride(t *testing.T) {
	policyPath := writeTempPolicy(t)
	tmplDir := writeTempTemplates(t, "T1")
	// choose a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "serve", "--policy", policyPath, "--templates", tmplDir, "--listen", addr)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start serve: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	// poll root for marker
	url := "http://" + addr + "/"
	deadline := time.Now().Add(3 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			b, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if strings.Contains(string(b), "OVERRIDE:T1") && strings.Contains(string(b), "Pv=0.1.0") {
				return
			}
		} else {
			lastErr = err
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("templates override not observed at root: %v", lastErr)
}

func TestCLIServeSheetCSVWithRefreshEveryMetrics(t *testing.T) {
	// CSV server returns a small stable CSV
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("domain,classification,rationale,last_review,status\nexample.com,NO_DPA,Reason,2025-08-01,active\n"))
	}))
	defer ts.Close()
	// choose a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	bin := buildTestBinary(t)
	cmd := exec.Command(bin, "serve", "--sheet-csv", ts.URL, "--listen", addr, "--refresh-every", "200ms")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start serve: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	// poll /metrics for refresh_count >= 1
	metricsURL := "http://" + addr + "/metrics"
	deadline := time.Now().Add(4 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(metricsURL)
		if err == nil {
			if resp.StatusCode == 200 {
				b, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				body := string(b)
				if strings.Contains(body, "\"refresh_count\":") {
					// crude check: ensure non-zero by searching for :0
					if !strings.Contains(body, "\"refresh_count\":0") {
						return
					}
				}
			}
		} else {
			lastErr = err
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("metrics did not show refresh_count > 0: %v", lastErr)
}

// Direct invocation tests (in-process) for coverage of command functions
func TestCmdValidateFunction(t *testing.T) {
	policyPath := writeTempPolicy(t)
	out := captureOutput(t, func() { cmdValidate([]string{"--policy", policyPath}) })
	if !strings.Contains(out, "\"status\":\"ok\"") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestCmdHashFunction(t *testing.T) {
	policyPath := writeTempPolicy(t)
	out := captureOutput(t, func() { cmdHash([]string{"--policy", policyPath}) })
	if !strings.Contains(out, "\"hash\":") {
		t.Fatalf("missing hash output: %s", out)
	}
}

func TestCmdGenerateDNSFunction(t *testing.T) {
	policyPath := writeTempPolicy(t)
	out := captureOutput(t, func() {
		cmdGenerateDNS([]string{"--policy", policyPath, "--format", "hosts", "--redirect-ipv4", "10.1.2.3", "--dry-run"})
	})
	if !strings.Contains(out, "example.com") {
		t.Fatalf("expected domain in output: %s", out)
	}
}

// captureOutput captures stdout produced by f.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan struct{})
	var sb strings.Builder
	go func() {
		_, _ = io.Copy(&sb, r)
		close(done)
	}()
	f()
	_ = w.Close()
	os.Stdout = old
	<-done
	return sb.String()
}

// buildTestBinary builds the current package binary into a test temp dir.
func buildTestBinary(t *testing.T) string {
	t.Helper()
	td := t.TempDir()
	outPath := filepath.Join(td, "sb29guard-test-bin")
	if runtime.GOOS == "windows" {
		outPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", outPath, ".")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v output=%s", err, b)
	}
	return outPath
}
