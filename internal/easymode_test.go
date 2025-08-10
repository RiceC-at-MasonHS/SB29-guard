//go:build easymode && integration
// +build easymode,integration

package internal

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// This test spins up the easy-mode docker-compose stack and verifies /explain works.
// It is opt-in: run with build tags and an env var to avoid CI usage by default.
// Example:
//
//	go test -tags "easymode integration" -run TestEasyMode_Stack
func TestEasyMode_Stack(t *testing.T) {
	if os.Getenv("SB29_EASYMODE_TEST") != "1" {
		t.Skip("set SB29_EASYMODE_TEST=1 and build tags 'easymode integration' to run")
	}
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("pwd: %v", err)
	}
	compose := filepath.Join(repoRoot, "easy-mode", "docker-compose.yml")
	if _, err := os.Stat(compose); err != nil {
		t.Skip("easy-mode compose not found")
	}
	// Ensure policy exists
	policyDir := filepath.Join(repoRoot, "easy-mode", "policy")
	_ = os.MkdirAll(policyDir, 0o755)
	from := filepath.Join(repoRoot, "policy", "domains.example.yaml")
	to := filepath.Join(policyDir, "domains.yaml")
	b, err := os.ReadFile(from)
	if err != nil {
		t.Fatalf("read example policy: %v", err)
	}
	if err := os.WriteFile(to, b, 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	// Write .env
	envPath := filepath.Join(repoRoot, "easy-mode", ".env")
	if err := os.WriteFile(envPath, []byte("SB29_DOMAIN=localhost\nACME_EMAIL=admin@example.org\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	// Start stack
	cmd := exec.Command("docker", "compose", "-f", compose, "up", "-d")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compose up: %v\n%s", err, string(out))
	}
	defer func() {
		down := exec.Command("docker", "compose", "-f", compose, "down")
		down.Dir = repoRoot
		_, _ = down.CombinedOutput()
	}()
	// Poll for readiness (Caddy + app)
	client := &http.Client{Timeout: 2 * time.Second}
	ok := false
	for i := 0; i < 60; i++ {
		resp, err := client.Get("http://localhost/explain?domain=exampletool.com")
		if err == nil {
			if resp.StatusCode == 200 {
				ok = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	if !ok {
		t.Fatalf("/explain did not return 200 via easy-mode stack")
	}
}
