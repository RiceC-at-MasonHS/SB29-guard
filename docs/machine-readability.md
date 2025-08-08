# Machine Readability Guidelines

Status: Draft

## Goals
Ensure all critical project documentation and outputs can be parsed programmatically to facilitate automation, validation, and reproducibility.

## Structured Artifacts
| Artifact | Format | Path | Purpose |
|----------|--------|------|---------|
| Policy dataset | YAML | `policy/domains.yaml` | Canonical domain list |
| Policy schema | JSON Schema | `internal/policy/policy.schema.json` | Validate policy (embedded at build) |
| Requirements | Markdown + embedded key blocks | `docs/requirements.md` | Human + machine extraction |
| CLI design | Markdown (structured sections) | `docs/cli-design.md` | Generate scaffolding |
| API contract | Markdown (until OpenAPI) | `docs/api-contract.md` | Basis for OpenAPI generation |
| OpenAPI spec (planned) | JSON | `dist/api/openapi.json` | API tooling |
| Aggregated logs | JSON | `logs/aggregates/*.json` | Reporting |
| DNS artifacts | zone/hosts/text | `dist/dns/` | Deployment |

## Conventions
1. All JSON objects use snake_case keys unless external spec mandates otherwise.
2. Timestamps in ISO8601 UTC (`YYYY-MM-DDTHH:MM:SSZ`).
3. Policy versioning uses semantic version + hash prefix where needed.
4. All generated files begin with a metadata comment:
```
# sb29guard policy_version=<ver> hash=<sha256> generated=<iso8601> tool_version=<tool>
```
5. Markdown documents include machine-scrapable sections headed by `##` with predictable titles.
6. Enumerated requirements prefixed `FR-`, `NFR-`, `TST-` for automated extraction.

## Extraction Strategy (Example)
A script can:
- Parse `requirements.md` lines matching `^FR-[0-9]+:` -> JSON array.
- Derive OpenAPI by scanning `api-contract.md` endpoint blocks.
- Validate policy via embedded schema.

## Example Requirements JSON (Derived)
```
[
  {"id":"FR-1","text":"Maintain a machine-readable policy file (default: policy/domains.yaml)."},
  {"id":"NFR-1","text":"Privacy first: No storage of client IP..."}
]
```

## Schema Versioning
- Increment `$id` if breaking changes.
- Maintain backward compatibility or provide migration script.

## Hashing Policy
Canonical form: sort active records by domain; for each record concatenate selected normalized fields with `\n`; SHA-256 of result hex-lowercase.

## Validation Pipeline (CI)
1. Lint YAML (policy).
2. Validate JSON Schema.
3. Extract requirements & ensure IDs unique & contiguous.
4. Regenerate OpenAPI (future); compare with committed version (diff fail if drift).
5. Run unit tests (policy & dnsgen packages must meet coverage gates).

## Tooling Suggestions
- Python: `pyyaml`, `jsonschema`.
- Node.js: `ajv` for schema validation.

## Future
- Provide a `machine-index.json` summarizing all structured assets & their versions.
