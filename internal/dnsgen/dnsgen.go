// Package dnsgen generates DNS artifacts (hosts, BIND, Unbound, RPZ, dnsmasq, domain-list, Windows DNS PowerShell) from the policy.
package dnsgen

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/RiceC-at-MasonHS/SB29-guard/internal/policy"
)

// Options controls DNS artifact generation.
type Options struct {
	Format         string
	Mode           string // a-record|cname
	RedirectIPv4   string
	RedirectHost   string
	TTL            int
	SerialStrategy string // date|epoch|hash
}

// Generate produces DNS content for the given policy according to Options.
func Generate(p *policy.Policy, o Options) ([]byte, error) {
	if o.Format == "" {
		return nil, errors.New("format required")
	}
	if o.Mode == "" {
		o.Mode = "a-record"
	}
	if o.TTL <= 0 {
		o.TTL = 300
	}
	records := activeDomains(p)
	switch o.Format {
	case "hosts":
		return genHosts(records, o)
	case "bind":
		return genBindZone(records, p, o)
	case "unbound":
		return genUnbound(records, p, o)
	case "rpz":
		return genRPZ(records, p, o)
	case "dnsmasq":
		return genDnsmasq(records, o)
	case "domain-list":
		return genDomainList(records)
	case "winps":
		return genWinPS(records, p, o)
	default:
		return nil, fmt.Errorf("unsupported format: %s", o.Format)
	}
}

func activeDomains(p *policy.Policy) []policy.Record {
	var out []policy.Record
	for _, r := range p.Records {
		if r.Status == "suspended" {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Domain < out[j].Domain })
	return out
}

func genHosts(recs []policy.Record, o Options) ([]byte, error) {
	if o.RedirectIPv4 == "" {
		return nil, errors.New("redirect-ipv4 required for hosts format")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# sb29guard format=hosts mode=%s\n", o.Mode)
	for _, r := range recs {
		// hosts file ignores wildcard marker; strip '*.'
		domain := strings.TrimPrefix(r.Domain, "*.")
		fmt.Fprintf(&b, "%s %s\n", o.RedirectIPv4, domain)
	}
	return []byte(b.String()), nil
}

func genBindZone(recs []policy.Record, p *policy.Policy, o Options) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "; sb29guard format=bind mode=%s policy_version=%s\n", o.Mode, p.Version)
	fmt.Fprintf(&b, "$TTL %d\n", o.TTL)
	serial := computeSerial(p, o)
	fmt.Fprintf(&b, "@ IN SOA %s. hostmaster.%s. (%s 3600 900 604800 %d)\n", o.RedirectHost, o.RedirectHost, serial, o.TTL)
	fmt.Fprintf(&b, "@ IN NS %s.\n", o.RedirectHost)
	for _, r := range recs {
		name := strings.TrimPrefix(r.Domain, "*.")
		if o.Mode == "cname" {
			fmt.Fprintf(&b, "%s %d IN CNAME %s.\n", name, o.TTL, o.RedirectHost)
		} else {
			if o.RedirectIPv4 == "" {
				return nil, errors.New("redirect-ipv4 required for a-record mode")
			}
			fmt.Fprintf(&b, "%s %d IN A %s\n", name, o.TTL, o.RedirectIPv4)
		}
	}
	return []byte(b.String()), nil
}

func genUnbound(recs []policy.Record, p *policy.Policy, o Options) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# sb29guard format=unbound mode=%s policy_version=%s\n", o.Mode, p.Version)
	for _, r := range recs {
		name := strings.TrimPrefix(r.Domain, "*.")
		if o.Mode == "cname" {
			fmt.Fprintf(&b, "local-data: \"%s CNAME %s\"\n", name, o.RedirectHost)
		} else {
			if o.RedirectIPv4 == "" {
				return nil, errors.New("redirect-ipv4 required for a-record mode")
			}
			fmt.Fprintf(&b, "local-zone: \"%s\" redirect\n", name)
			fmt.Fprintf(&b, "local-data: \"%s A %s\"\n", name, o.RedirectIPv4)
		}
	}
	return []byte(b.String()), nil
}

func genRPZ(recs []policy.Record, p *policy.Policy, o Options) ([]byte, error) {
	if o.RedirectHost == "" {
		return nil, errors.New("redirect-host required for rpz format")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "; sb29guard format=rpz policy_version=%s mode=%s\n", p.Version, o.Mode)
	fmt.Fprintf(&b, "$TTL %d\n", o.TTL)
	serial := computeSerial(p, o)
	fmt.Fprintf(&b, "@ IN SOA %s. hostmaster.%s. (%s 3600 900 604800 %d)\n", o.RedirectHost, o.RedirectHost, serial, o.TTL)
	fmt.Fprintf(&b, "@ IN NS %s.\n", o.RedirectHost)
	for _, r := range recs {
		name := r.Domain
		// keep wildcard as-is for RPZ (policy trigger)
		fmt.Fprintf(&b, "%s. CNAME %s.\n", name, o.RedirectHost)
	}
	if o.RedirectIPv4 != "" {
		fmt.Fprintf(&b, "%s. A %s\n", o.RedirectHost, o.RedirectIPv4)
	}
	return []byte(b.String()), nil
}

// genDnsmasq outputs dnsmasq config lines.
// a-record mode: address=/example.com/10.10.10.50
// cname mode: cname=example.com,blocked.guard.local
func genDnsmasq(recs []policy.Record, o Options) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# sb29guard format=dnsmasq mode=%s\n", o.Mode)
	switch o.Mode {
	case "a-record":
		if o.RedirectIPv4 == "" {
			return nil, errors.New("redirect-ipv4 required for dnsmasq a-record mode")
		}
		for _, r := range recs {
			name := strings.TrimPrefix(r.Domain, "*.")
			fmt.Fprintf(&b, "address=/%s/%s\n", name, o.RedirectIPv4)
		}
	case "cname":
		if o.RedirectHost == "" {
			return nil, errors.New("redirect-host required for dnsmasq cname mode")
		}
		for _, r := range recs {
			name := strings.TrimPrefix(r.Domain, "*.")
			fmt.Fprintf(&b, "cname=%s,%s\n", name, o.RedirectHost)
		}
	default:
		return nil, fmt.Errorf("unsupported mode for dnsmasq: %s", o.Mode)
	}
	return []byte(b.String()), nil
}

// genDomainList outputs one domain per line (wildcards stripped), for adlist-style consumers.
func genDomainList(recs []policy.Record) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# sb29guard format=domain-list\n")
	for _, r := range recs {
		name := strings.TrimPrefix(r.Domain, "*.")
		fmt.Fprintf(&b, "%s\n", name)
	}
	return []byte(b.String()), nil
}

// genWinPS emits a PowerShell script to create per-domain zones and A/CNAME records on Windows DNS.
func genWinPS(recs []policy.Record, p *policy.Policy, o Options) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# sb29guard format=winps mode=%s policy_version=%s\n", o.Mode, p.Version)
	fmt.Fprintf(&b, "$ErrorActionPreference = 'Stop'\n")
	// Parameters baked-in from options for simplicity
	if o.TTL <= 0 {
		o.TTL = 300
	}
	fmt.Fprintf(&b, "$ttl = New-TimeSpan -Seconds %d\n", o.TTL)
	switch o.Mode {
	case "a-record":
		if o.RedirectIPv4 == "" {
			return nil, errors.New("redirect-ipv4 required for winps a-record mode")
		}
		fmt.Fprintf(&b, "$ip = '%s'\n", o.RedirectIPv4)
		for _, r := range recs {
			name := strings.TrimPrefix(r.Domain, "*.")
			fmt.Fprintf(&b, "if (-not (Get-DnsServerZone -Name '%s' -ErrorAction SilentlyContinue)) { Add-DnsServerPrimaryZone -Name '%s' -ZoneFile '%s.dns' -DynamicUpdate None }\n", name, name, name)
			fmt.Fprintf(&b, "try { Add-DnsServerResourceRecordA -ZoneName '%s' -Name '@' -IPv4Address $ip -TimeToLive $ttl -AllowUpdateAny:$false -CreatePtr:$false } catch {}\n", name)
		}
	case "cname":
		if o.RedirectHost == "" {
			return nil, errors.New("redirect-host required for winps cname mode")
		}
		fmt.Fprintf(&b, "$target = '%s'\n", o.RedirectHost)
		for _, r := range recs {
			name := strings.TrimPrefix(r.Domain, "*.")
			fmt.Fprintf(&b, "if (-not (Get-DnsServerZone -Name '%s' -ErrorAction SilentlyContinue)) { Add-DnsServerPrimaryZone -Name '%s' -ZoneFile '%s.dns' -DynamicUpdate None }\n", name, name, name)
			fmt.Fprintf(&b, "try { Add-DnsServerResourceRecordCName -ZoneName '%s' -Name '@' -HostNameAlias $target } catch {}\n", name)
		}
	default:
		return nil, fmt.Errorf("unsupported mode for winps: %s", o.Mode)
	}
	return []byte(b.String()), nil
}

// computeSerial returns a BIND/RPZ serial based on strategy.
// date: YYYYMMDDNN where NN is hash-derived (00-99)
// epoch: Unix timestamp
// hash: first 4 bytes of policy hash interpreted as big-endian uint32
func computeSerial(p *policy.Policy, o Options) string {
	strategy := o.SerialStrategy
	if strategy == "" {
		strategy = "date"
	}
	switch strategy {
	case "epoch":
		return strconv.FormatInt(time.Now().UTC().Unix(), 10)
	case "hash":
		h := p.CanonicalHash()
		if len(h) >= 8 {
			if b, err := hex.DecodeString(h[:8]); err == nil && len(b) == 4 {
				n := binary.BigEndian.Uint32(b)
				return strconv.FormatUint(uint64(n), 10)
			}
		}
		// fallback to date
	}
	today := time.Now().UTC().Format("20060102")
	h := p.CanonicalHash()
	suffix := "00"
	if len(h) >= 2 {
		if v, err := strconv.ParseInt(h[:2], 16, 64); err == nil {
			suffix = fmt.Sprintf("%02d", v%100)
		}
	}
	return today + suffix
}
