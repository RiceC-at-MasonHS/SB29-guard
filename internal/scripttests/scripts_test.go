package scripttests

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
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

// TestLinuxScript_FetchAndGate exercises the linux-fetch-and-reload.sh script against a mocked server.
// It requires a Unix-like environment with /bin/bash available. The test is skipped on Windows.
func TestLinuxScript_FetchAndGate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows")
	}
	// Mock server for /metrics and /domain-list
	policyVer := "v1.2.3"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metrics":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, fmt.Sprintf(`{"policy_version":%q}`, policyVer))
		case "/domain-list":
			w.Header().Set("Content-Type", "text/plain")
			io.WriteString(w, "exampletool.com\n")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Prepare temp workspace
	tmp := t.TempDir()
	outFile := filepath.Join(tmp, "blocked.txt")
	verFile := outFile + ".ver"

	// Create a fake jq shim that extracts policy_version
	shimDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(shimDir, 0o755); err != nil {
		t.Fatalf("mkdir shim: %v", err)
	}
	jqShimPath := filepath.Join(shimDir, "jq")
	jqShim := "#!/usr/bin/env bash\nawk -F'\"' '/policy_version/{print $4}'\n"
	if err := os.WriteFile(jqShimPath, []byte(jqShim), 0o755); err != nil {
		t.Fatalf("write jq shim: %v", err)
	}

	// Run the script forcing ONLY_WHEN_CHANGED=false to always fetch
	scriptPath := filepath.FromSlash("docs/implementers/scripts/linux-fetch-and-reload.sh")
	cmd := exec.Command("bash", scriptPath)
	cmd.Env = append(os.Environ(),
		"GUARD_BASE="+srv.URL,
		"OUT_FILE="+outFile,
		"ONLY_WHEN_CHANGED=false",
		"PATH="+shimDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("script run failed: %v\n%s", err, string(out))
	}
	// Validate outputs
	b, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("missing out file: %v", err)
	}
	if !strings.Contains(string(b), "exampletool.com") {
		t.Fatalf("out file missing domain list content: %q", string(b))
	}
	vb, err := os.ReadFile(verFile)
	if err != nil {
		t.Fatalf("missing version file: %v", err)
	}
	if got := strings.TrimSpace(string(vb)); got != policyVer {
		t.Fatalf("version mismatch: got %q want %q", got, policyVer)
	}

	// Re-run with ONLY_WHEN_CHANGED=true; verify file content is unchanged
	h1 := sha256.Sum256(b)
	// small sleep to avoid mtime equality confusion
	time.Sleep(50 * time.Millisecond)
	cmd2 := exec.Command("bash", scriptPath)
	cmd2.Env = append(os.Environ(),
		"GUARD_BASE="+srv.URL,
		"OUT_FILE="+outFile,
		"ONLY_WHEN_CHANGED=true",
		"PATH="+shimDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	out2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("script rerun failed: %v\n%s", err, string(out2))
	}
	b2, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read out file again: %v", err)
	}
	h2 := sha256.Sum256(b2)
	if h1 != h2 {
		t.Fatalf("out file changed despite unchanged policy_version")
	}
}

// TestWindowsScript_Static checks core contents of the PowerShell script to catch regressions.
func TestWindowsScript_Static(t *testing.T) {
	path := filepath.FromSlash("docs/implementers/scripts/windows-fetch-and-import.ps1")
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("windows script not found: %v", err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	content, _ := io.ReadAll(rd)
	s := string(content)
	required := []string{
		"/metrics",
		"/domain-list",
		"Get-PolicyVersion",
		"Invoke-WebRequest",
	}
	for _, want := range required {
		if !strings.Contains(s, want) {
			t.Fatalf("windows script missing %q", want)
		}
	}
}
