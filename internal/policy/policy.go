package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Record represents a single policy entry in the policy dataset
type Record struct {
	Domain         string   `yaml:"domain" json:"domain"`
	Classification string   `yaml:"classification" json:"classification"`
	Rationale      string   `yaml:"rationale" json:"rationale"`
	LastReview     string   `yaml:"last_review" json:"last_review"`
	Status         string   `yaml:"status" json:"status"`
	Notes          string   `yaml:"notes,omitempty" json:"notes,omitempty"`
	SourceRef      string   `yaml:"source_ref,omitempty" json:"source_ref,omitempty"`
	Expires        string   `yaml:"expires,omitempty" json:"expires,omitempty"`
	Tags           []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Policy is the root structure for a policy file
type Policy struct {
	Version  string   `yaml:"version" json:"version"`
	Updated  string   `yaml:"updated" json:"updated"`
	Records  []Record `yaml:"records" json:"records"`
	Metadata *struct {
		GeneratedHash string `yaml:"generated_hash,omitempty" json:"generated_hash,omitempty"`
		Source        string `yaml:"source,omitempty" json:"source,omitempty"`
		Notes         string `yaml:"notes,omitempty" json:"notes,omitempty"`
	} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

var domainPattern = regexp.MustCompile(`^(\*\.)?([a-z0-9-]{1,63}\.)+[a-z]{2,63}$`)

// CanonicalHash computes a deterministic SHA-256 over active records (simple normalization of select fields)
func (p *Policy) CanonicalHash() string {
	var lines []string
	for _, r := range p.Records {
		if r.Status != "active" && r.Status != "" { // include empty as active until full validation exists
			continue
		}
		lines = append(lines, fmt.Sprintf("%s|%s|%s|%s|%s", r.Domain, r.Classification, r.Rationale, r.LastReview, r.Status))
	}
	sort.Strings(lines)
	h := sha256.Sum256([]byte(fmt.Sprintln(lines)))
	return hex.EncodeToString(h[:])
}

// Validate applies schema-like checks (lightweight) pending full JSON Schema integration
func (p *Policy) Validate() error {
	if p.Version == "" {
		return errors.New("version required")
	}
	if p.Updated == "" {
		return errors.New("updated required")
	}
	seen := map[string]struct{}{}
	for i, r := range p.Records {
		if r.Domain == "" {
			return fmt.Errorf("record %d: domain empty", i)
		}
		d := strings.ToLower(r.Domain)
		if !domainPattern.MatchString(d) {
			return fmt.Errorf("record %d: domain invalid: %s", i, r.Domain)
		}
		if _, ok := seen[d+"|"+r.Classification]; ok {
			return fmt.Errorf("record %d: duplicate domain+classification combo: %s", i, r.Domain)
		}
		seen[d+"|"+r.Classification] = struct{}{}
		// Basic classification/status checks
		switch r.Classification {
		case "NO_DPA", "PENDING_REVIEW", "EXPIRED_DPA", "LEGAL_HOLD", "OTHER":
		default:
			return fmt.Errorf("record %d: invalid classification %s", i, r.Classification)
		}
		switch r.Status {
		case "active", "suspended", "": // empty treated as active
		default:
			return fmt.Errorf("record %d: invalid status %s", i, r.Status)
		}
		p.Records[i].Domain = d
	}
	return nil
}

// Lookup attempts to find a record for the provided domain (case-insensitive),
// supporting wildcard entries of the form "*.example.com". Matching rules:
//  1. Exact domain match (after lowercasing)
//  2. Wildcard record where record.Domain is "*.example.com" matches either
//     the base domain "example.com" or any subdomain that ends with ".example.com".
//
// Suspended records are ignored. Returns the first matching active record
// (prioritizing exact match over wildcard matches). If multiple wildcard
// records could match (should not happen under validation rules), the first
// encountered is returned.
func (p *Policy) Lookup(domain string) (*Record, bool) {
	d := strings.ToLower(strings.TrimSpace(domain))
	if d == "" {
		return nil, false
	}
	var wildcardMatch *Record
	for i := range p.Records {
		r := &p.Records[i]
		if r.Status == "suspended" { // ignore suspended
			continue
		}
		if r.Domain == d { // exact
			return r, true
		}
		if strings.HasPrefix(r.Domain, "*.") {
			base := strings.TrimPrefix(r.Domain, "*.")
			if d == base || strings.HasSuffix(d, "."+base) {
				if wildcardMatch == nil { // keep first wildcard match
					wildcardMatch = r
				}
			}
		}
	}
	if wildcardMatch != nil {
		return wildcardMatch, true
	}
	return nil, false
}
