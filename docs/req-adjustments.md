This tool is hopefully going to be implemented at a building or district level. For that reason, I would imagine that the DNS record tools will need to be managed in the already-existing router. Therefore, we will need some easy-to-recommend changes in a router. I saw that you recommended PiHole (which is a great tool) and a few other deployment methods, but let's make sure this entire toolset is easy for teachers to pitch to their administrators and I.T. teams - so keep expanding the list of deployment methods to cover as many common cases as possible. I want this to be a helpful open source tool that my school can implement and others can hopefully use, with as little friction as possible. 

Another item. I think that we're going to need to use a `.env` file (and an `.env.example`)to connect this tool to a list of 'blacklisted' sites that is managed by a Google Sheet. I would estimate that the person managing the Digital Privacy Agreements will be a school administrator, hwo is (very likely) not used to managing version-controlled configuration files. So, let's help describe that method of administration clearly in the machine-focused requirements and the human-focused README.

---
## Adjustments / Additions Requested (Deployment Coverage)
Goal: Minimize friction for IT approval by providing copy‑paste or near turnkey instructions for the most common on‑prem and cloud DNS / edge platforms used in K‑12.

### Target Platform Matrix (Planned Documentation)
| Platform / Product | Method | Artifact Type Needed | Priority |
|--------------------|--------|----------------------|----------|
| BIND (recursive) | Local zone / RPZ | zone / rpz.zone | Done (initial) |
| Unbound | local-zone / RPZ | .conf / rpz.zone | Done (initial) |
| Pi-hole / dnsmasq | hosts / domain list | hosts/domain list | Done (initial) |
| Windows DNS (AD) | Primary zones + A/CNAME | PowerShell script | Draft (initial) |
| pfSense (Unbound + DNS Resolver) | Host overrides / custom conf include | unbound fragment | Planned |
| OPNsense | DNS Overrides | host override list | Planned |
| Cisco/Meraki (Cloud managed) | Custom DNS forward to internal resolver OR Layer 7 block page (fallback) | Guidance doc | Planned |
| Fortinet FortiGate | Local DNS filter / DNS database / Policy redirect | zone snippet + policy steps | Planned |
| Palo Alto (PAN-OS) | DNS sinkhole (custom DNS) + response page | IP list + runbook | Planned |
| Infoblox | RPZ import | rpz.zone + import runbook | Planned |
| Azure DNS Private Resolver | Private zone override + internal load balancer | Azure CLI script | Planned |
| AWS Route53 Resolver (Private) | Private hosted zone with A/CNAME records | Terraform/CLI template | Planned |
| Google Cloud DNS (Private) | Private managed zone overrides | gcloud script | Planned |
| Cloudflare (Gateway / DNS) | Block list with custom block page URL | Domain list + URL params mapping | Planned |
| Umbrella (Cisco) | Custom block list + redirect page | Domain CSV + param strategy | Planned |
| Lightspeed / GoGuardian (if DNS layer) | Importable domain block list | Plain domain list | Planned |
| Consumer Routers (limited firmware) | Use upstream internal resolver only | Advisory note | Planned |

### Documentation Deliverables
1. Each platform gets: Overview, Prereqs, Security/Privacy notes, Step-by-step, Verification, Rollback, Automation hook.
2. Provide a common “Verification Checklist” snippet reused across pages.
3. Provide risk/impact statement (non-invasive; reversible; no PII).
4. Provide estimated implementation time (e.g., “~10 minutes after prerequisites”).

### Generator Enhancements (New Requirements to Incorporate)
- FR-21 (proposed): `generate-dns` supports `--format pfSense-unbound` outputting include file with header metadata.
- FR-22 (proposed): `generate-dns` supports `--format opnsense-unbound` identical to pfSense with naming differences.
- FR-23 (proposed): `generate-dns` supports `--format infoblox-rpz` (ensures SOA & NS fields Infoblox-friendly + optional CSV import variant).
- FR-24 (proposed): `generate-dns` supports `--format route53-json` producing a change batch JSON for AWS CLI.
- FR-25 (proposed): `generate-dns` supports `--format azure-cli` producing shell commands (idempotent) to create record sets.
- FR-26 (proposed): `generate-dns` supports `--format gcloud-dns` producing `gcloud dns record-sets transaction` script template.
- FR-27 (proposed): `generate-dns` supports domain-only plain list variant for cloud security products (Umbrella, Cloudflare Gateway) with optional classification suffix comment.
- FR-28 (proposed): Add `--classification-filter` to export subset for phased rollout.
- FR-29 (proposed): Add `--inactive-exclude` (default true) to skip suspended records.
- FR-30 (proposed): Add manifest file generation: `dist/dns/manifest.json` enumerating outputs with hashes.

### Router / Edge Device Adoption Patterns
1. Prefer NOT to edit production resolver directly; stage in test environment, verify limited sample, then expand.
2. TTL strategy: low (300s) initial, raise (1800s) after stable.
3. Provide rollback by retaining prior artifact + date label (tool includes hash).
4. Provide change log summarizing domains added/removed since last generation.
5. Validate no overlap with internally hosted domains (collision detection step in validator).

### Risk Mitigation Talking Points (Teacher -> Admin)
- Reversible: Single file removal restores original behavior.
- Transparent: Redirect page clearly states reason; fosters vendor compliance.
- Privacy-preserving: No student identifiers stored.
- Auditable: Hash & version stamped on every artifact.
- Scoped: Only affects explicitly listed domains (no wildcard overreach beyond leftmost label).

### Additional Automation Hooks
- GitHub Actions / local CI job publishes signed artifacts to internal share.
- Optional Slack/email notification: summary of new/removed domains awaiting approval (future).
- Daily cron: pull updated policy repo -> validate -> generate -> atomic deploy.

---
## Outstanding Questions for Stakeholders
1. Which DNS platforms are currently in production (list vendor + version)?
2. Is RPZ already used (impact on DNSSEC)?
3. Preferred change approval workflow (ticket? email?)
4. Requirement for signed artifacts / chain of custody?
5. Need for TLS/HTTPS for blocked domains (certificate strategy)?

---
## Next Actions
- Integrate proposed FR-21..FR-30 into main `requirements.md` (pending approval).
- Create initial pfSense & OPNsense deployment docs.
- Add manifest & classification filter options to CLI design doc.
- Provide teacher-facing one-page “Pitch” PDF (future non-code asset).