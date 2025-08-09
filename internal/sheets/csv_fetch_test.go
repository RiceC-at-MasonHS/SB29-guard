package sheets

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func sampleCSV() string {
	return "domain,classification,rationale,last_review,status\n" +
		"example.com,NO_DPA,Reason,2025-08-01,active\n"
}

func TestFetchCSVPolicy_SuccessWithRetry(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&calls, 1)
		if c == 1 {
			http.Error(w, "try again", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleCSV()))
	}))
	defer ts.Close()

	start := time.Now()
	p, err := FetchCSVPolicy(ts.URL, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Records) != 1 || p.Records[0].Domain != "example.com" {
		t.Fatalf("unexpected parsed policy: %+v", p)
	}
	if p.Metadata == nil || p.Metadata.Source != "csv" {
		t.Fatalf("expected metadata source csv, got %+v", p.Metadata)
	}
	// Expect at least one backoff sleep (~500ms) due to initial 500
	if time.Since(start) < 400*time.Millisecond {
		t.Fatalf("expected retry backoff to take noticeable time")
	}
}

func TestFetchCSVPolicy_ErrorAfterRetries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()
	_, err := FetchCSVPolicy(ts.URL, &http.Client{Timeout: 5 * time.Second})
	if err == nil || !strings.Contains(err.Error(), "download csv failed") {
		t.Fatalf("expected retry failure, got %v", err)
	}
}

func TestFetchCSVPolicyCached_ColdThen304FromCache(t *testing.T) {
	// temp cache dir
	dir := t.TempDir()
	etag := "W/\"v1\""
	lastMod := time.Now().UTC().Format(http.TimeFormat)
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&calls, 1)
		if c == 1 {
			w.Header().Set("ETag", etag)
			w.Header().Set("Last-Modified", lastMod)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(sampleCSV()))
			return
		}
		// assert conditional headers are sent
		if r.Header.Get("If-None-Match") == etag || r.Header.Get("If-Modified-Since") != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		http.Error(w, "unexpected", http.StatusBadRequest)
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	// First fetch writes cache
	p1, fromCache1, err := FetchCSVPolicyCached(ts.URL, dir, client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fromCache1 {
		t.Fatalf("first fetch should not be from cache")
	}
	if p1.Metadata == nil || p1.Metadata.Source != "csv" {
		t.Fatalf("expected source csv on cold fetch")
	}
	// Ensure cache files exist
	id := urlHash(ts.URL)
	if _, err := os.Stat(filepath.Join(dir, "sheets", id+".csv")); err != nil {
		t.Fatalf("expected csv cache file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sheets", id+".meta.json")); err != nil {
		t.Fatalf("expected meta cache file: %v", err)
	}
	// Second fetch should 304 and hit cache
	p2, fromCache2, err := FetchCSVPolicyCached(ts.URL, dir, client)
	if err != nil {
		t.Fatalf("unexpected error on 304: %v", err)
	}
	if !fromCache2 {
		t.Fatalf("expected from cache on 304")
	}
	if p2.Metadata == nil || p2.Metadata.Source != "csv-cache" {
		t.Fatalf("expected source csv-cache on 304")
	}
}

func TestBackoffValues(t *testing.T) {
	if d := backoff(0); d < 400*time.Millisecond || d > 700*time.Millisecond {
		t.Fatalf("unexpected backoff(0): %v", d)
	}
	if d := backoff(1); d < 900*time.Millisecond || d > 1100*time.Millisecond {
		t.Fatalf("unexpected backoff(1): %v", d)
	}
	if d := backoff(2); d < 1800*time.Millisecond || d > 2200*time.Millisecond {
		t.Fatalf("unexpected backoff(2+): %v", d)
	}
}

func TestURLHashStability(t *testing.T) {
	a := urlHash("https://example.com/x?y=1")
	b := urlHash("https://example.com/x?y=1")
	c := urlHash("https://example.com/x?y=2")
	if a != b {
		t.Fatalf("expected stable hash, got %s vs %s", a, b)
	}
	if a == c {
		t.Fatalf("expected different hash for different input")
	}
}
