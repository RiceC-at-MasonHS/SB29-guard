//go:build easymode && integration && !windows
// +build easymode,integration,!windows

package easymodeint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Test the Bash wrapper (easy-mode/gen-dns.sh) generates DNS files without docker exec.
func TestEasyMode_DNSGen_Bash_Works(t *testing.T) {
	if os.Getenv("SB29_EASYMODE_TEST") != "1" {
		t.Skip("set SB29_EASYMODE_TEST=1 and build tags 'easymode integration' to run")
	}

	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not found")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found")
	}
	if out, err := exec.Command("docker", "compose", "version").CombinedOutput(); err != nil {
		t.Skipf("docker compose not available: %v (%s)", err, string(out))
	}
	if out, err := exec.Command("docker", "info").CombinedOutput(); err != nil {
		t.Skipf("docker engine not reachable: %v (%s)", err, string(out))
	}

	// Find repo root (look for go.mod upwards from current package dir)
	wd, _ := os.Getwd()
	repoRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot { // reached filesystem root
			break
		}
		repoRoot = parent
	}
	easy := filepath.Join(repoRoot, "easy-mode")
	sh := filepath.Join(easy, "gen-dns.sh")
	policyDir := filepath.Join(easy, "policy")
	outDir := filepath.Join(easy, "out")
	_ = os.MkdirAll(policyDir, 0o755)
	_ = os.MkdirAll(outDir, 0o755)

	// Ensure policy exists
	example := filepath.Join(repoRoot, "policy", "domains.example.yaml")
	target := filepath.Join(policyDir, "domains.yaml")
	b, err := os.ReadFile(example)
	if err != nil {
		t.Fatalf("read example policy: %v", err)
	}
	if err := os.WriteFile(target, b, 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	// Minimal .env for easy-mode
	if err := os.WriteFile(filepath.Join(easy, ".env"), []byte("SB29_DOMAIN=localhost\nACME_EMAIL=admin@example.org\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	// 1) hosts
	hostsFile := filepath.Join(outDir, "hosts.txt")
	_ = os.Remove(hostsFile)
	run := exec.Command("bash", sh, "hosts", "a-record", "10.10.10.50")
	run.Dir = repoRoot
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("gen-dns hosts failed: %v\n%s", err, string(out))
	}
	data, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatalf("hosts output missing: %v", err)
	}
	if !strings.Contains(string(data), "exampletool.com") {
		t.Fatalf("hosts output did not include exampletool.com\n---\n%s\n---", string(data))
	}

	// 2) domain-list
	listFile := filepath.Join(outDir, "domains.txt")
	_ = os.Remove(listFile)
	run2 := exec.Command("bash", sh, "domain-list")
	run2.Dir = repoRoot
	if out, err := run2.CombinedOutput(); err != nil {
		t.Fatalf("gen-dns domain-list failed: %v\n%s", err, string(out))
	}
	data2, err := os.ReadFile(listFile)
	if err != nil {
		t.Fatalf("domain-list output missing: %v", err)
	}
	if !strings.Contains(string(data2), "exampletool.com") {
		t.Fatalf("domain-list output did not include exampletool.com\n---\n%s\n---", string(data2))
	}
}
