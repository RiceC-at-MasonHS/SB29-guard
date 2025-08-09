package sheets

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

// FetchCSVPolicy downloads a published Google Sheets CSV (public link) and converts it into a *policy.Policy.
// Required headers (case-insensitive): domain, classification, rationale, last_review, status
// Optional headers: source_ref, notes, expires, tags
// Extra columns are ignored.
func FetchCSVPolicy(url string, client *http.Client) (*policy.Policy, error) {
	if url == "" {
		return nil, errors.New("csv url empty")
	}
	if client == nil {
		client = http.DefaultClient
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(backoff(attempt))
			continue
		}
		var pOut *policy.Policy
		func() {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
				return
			}
			b, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20)) // 5MB safety limit
			if err != nil {
				lastErr = fmt.Errorf("read csv: %w", err)
				return
			}
			p, perr := parseCSV(string(b))
			if perr != nil {
				lastErr = perr
				return
			}
			p.Metadata.Source = "csv"
			pOut = p
		}()
		if pOut != nil {
			return pOut, nil
		}
		time.Sleep(backoff(attempt))
	}
	return nil, fmt.Errorf("download csv failed: %w", lastErr)
}

func backoff(attempt int) time.Duration {
	switch attempt {
	case 0:
		return 500 * time.Millisecond
	case 1:
		return 1 * time.Second
	default:
		return 2 * time.Second
	}
}

// FetchCSVPolicyCached fetches CSV with conditional requests and a simple on-disk cache.
// It stores two files under cacheDir/sheets: <id>.csv and <id>.meta.json, where <id> is a hash of the URL.
// Returns the policy and whether it was served from cache (true) or network (false).
func FetchCSVPolicyCached(url, cacheDir string, client *http.Client) (*policy.Policy, bool, error) {
	if url == "" {
		return nil, false, errors.New("csv url empty")
	}
	if client == nil {
		client = http.DefaultClient
	}
	if cacheDir == "" {
		cacheDir = "cache"
	}
	sheetDir := cacheDir + string(os.PathSeparator) + "sheets"
	_ = os.MkdirAll(sheetDir, 0o755)
	id := urlHash(url)
	csvPath := sheetDir + string(os.PathSeparator) + id + ".csv"
	metaPath := sheetDir + string(os.PathSeparator) + id + ".meta.json"
	var etag, lastMod string
	// Load meta if exists
	if b, err := os.ReadFile(metaPath); err == nil {
		var m struct{ ETag, LastModified string }
		_ = json.Unmarshal(b, &m)
		etag, lastMod = m.ETag, m.LastModified
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastMod != "" {
		req.Header.Set("If-Modified-Since", lastMod)
	}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusNotModified {
		// Use cache
		if data, rerr := os.ReadFile(csvPath); rerr == nil {
			_ = resp.Body.Close()
			p, perr := parseCSV(string(data))
			if perr != nil {
				return nil, false, perr
			}
			p.Metadata.Source = "csv-cache"
			return p, true, nil
		}
		// fallthrough to fetch fresh if cache missing
	}
	if err != nil {
		return nil, false, fmt.Errorf("download csv: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, false, fmt.Errorf("read csv: %w", err)
	}
	// write cache best-effort
	_ = os.WriteFile(csvPath, b, 0o644)
	// store meta
	m := struct{ ETag, LastModified string }{ETag: resp.Header.Get("ETag"), LastModified: resp.Header.Get("Last-Modified")}
	if mb, jerr := json.Marshal(m); jerr == nil {
		_ = os.WriteFile(metaPath, mb, 0o644)
	}
	p, perr := parseCSV(string(b))
	if perr != nil {
		return nil, false, perr
	}
	p.Metadata.Source = "csv"
	return p, false, nil
}

func urlHash(s string) string {
	// simple FNV-1a 64-bit for file naming
	var h uint64 = 1469598103934665603
	const prime uint64 = 1099511628211
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime
	}
	return fmt.Sprintf("%x", h)
}

func parseCSV(data string) (*policy.Policy, error) {
	r := csv.NewReader(strings.NewReader(data))
	r.FieldsPerRecord = -1
	head, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	index := map[string]int{}
	for i, h := range head {
		index[strings.ToLower(strings.TrimSpace(h))] = i
	}
	req := []string{"domain", "classification", "rationale", "last_review", "status"}
	for _, k := range req {
		if _, ok := index[k]; !ok {
			return nil, fmt.Errorf("missing required column %s", k)
		}
	}
	var records []policy.Record
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		// Skip empty domain rows
		if strings.TrimSpace(row[index["domain"]]) == "" {
			continue
		}
		rec := policy.Record{
			Domain:         strings.TrimSpace(row[index["domain"]]),
			Classification: strings.TrimSpace(row[index["classification"]]),
			Rationale:      strings.TrimSpace(row[index["rationale"]]),
			LastReview:     strings.TrimSpace(row[index["last_review"]]),
			Status:         strings.TrimSpace(row[index["status"]]),
		}
		if i, ok := index["source_ref"]; ok {
			rec.SourceRef = strings.TrimSpace(row[i])
		}
		if i, ok := index["notes"]; ok {
			rec.Notes = strings.TrimSpace(row[i])
		}
		if i, ok := index["expires"]; ok {
			rec.Expires = strings.TrimSpace(row[i])
		}
		if i, ok := index["tags"]; ok {
			raw := strings.TrimSpace(row[i])
			if raw != "" {
				parts := strings.Split(raw, ",")
				for _, p := range parts {
					rec.Tags = append(rec.Tags, strings.TrimSpace(p))
				}
				sort.Strings(rec.Tags)
			}
		}
		records = append(records, rec)
	}
	p := &policy.Policy{
		Version: "0.1.0", // placeholder; could derive from sheet metadata in future
		Updated: time.Now().UTC().Format("2006-01-02"),
		Records: records,
		Metadata: &struct {
			GeneratedHash string `yaml:"generated_hash,omitempty" json:"generated_hash,omitempty"`
			Source        string `yaml:"source,omitempty" json:"source,omitempty"`
			Notes         string `yaml:"notes,omitempty" json:"notes,omitempty"`
		}{Source: "csv"},
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}
