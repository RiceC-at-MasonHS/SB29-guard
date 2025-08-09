# Customizing the Redirect Page  
[README](./README.md) • [Technical Reference](./TECHNICAL.md) • [Contributing](./CONTRIBUTING.md)

Institutions can adjust the HTML and CSS of the redirect (explanation) pages to match branding, accessibility standards, or tone. This guide shows lightweight, safe approaches.

## What Ships by Default
Templates (embedded at build time) live in `internal/server/templates/`:
  - `layout.html` – base page structure (includes `<head>`, header, footer, shared CSS variable slot)
  - `root.html` – landing page (`/`)
  - `explain.html` – explanation / redirect page (`/explain`)

Styling lives in `internal/server/templates/style.css` and is embedded (no extra HTTP request). The content is exposed to templates as the `CSS` variable.

Snapshot copies for reference (not used at runtime) are stored under `docs/templates/` so you can review or diff template changes without digging into internal code paths.

## Approaches to Customize
### 1. Simple (Edit and Rebuild)
If you are comfortable rebuilding Go binaries:
1. Edit the template files in `internal/server/templates/`.
2. (Optional) Adjust `baseCSS` inside `internal/server/server.go`.
3. Rebuild: `go build ./cmd/sb29guard` (or run `go build ./...`).
4. Deploy the new binary.

Pros: Single self‑contained binary.  
Cons: Requires Go toolchain and rebuild per change.

### 2. Patch CSS Only (Minimal Code Change)
Edit `internal/server/templates/style.css` directly, then rebuild. All pages will pick up the change because the file is embedded.

### 3. External Template Override (Future Option)
A future enhancement may allow a runtime flag like:
```
sb29guard serve --templates /etc/sb29guard/templates
```
If present, disk templates would override embedded ones. Track this in the project issues if you need it.

## Accessibility & Usability Tips
- Maintain sufficient color contrast (WCAG AA: contrast ratio >= 4.5:1 for normal text).
- Keep rationale text short and scannable (aim < 120 words).
- Prefer system fonts for fast rendering unless you embed a font safely.
- Avoid auto‑refresh or script heavy components (page is static explanation).
- Provide clear heading hierarchy (`h1` once; `h2` for sections like Summary / Why / Reference / Meta).

## Adding a District Logo
Option A (Future Static Handler): add a static directory and serve `/static/logo.svg` (not yet implemented).

Option B (Data URI Now): embed a base64 SVG/PNG as a background-image in CSS (simplest today).

## Changing the Tone / Wording
Edit `explain.html` content blocks. Keep variable placeholders like `{{.Original}}`, `{{.Classification}}`, `{{.PolicyVersion}}` intact so dynamic data still renders.

## Variables Available in Templates
| Key | Meaning |
| --- | ------- |
| CSS | Inline stylesheet content |
| RecordCount | Total loaded policy records (root page) |
| Year | Current year (footer) |
| PolicyVersion | Policy version string or date from file header |
| Original | Original domain requested (explain page) |
| Classification | Policy classification assigned |
| Rationale | Optional rationale text (HTML‑escaped) |
| SourceRef | Optional reference / ticket ID (HTML‑escaped) |
| Now | Current UTC timestamp (RFC3339) |

## Safe HTML Practices
- Leave interpolation (`{{...}}`) intact; Go `html/template` auto-escapes variables.
- Only introduce raw HTML if you are certain of safety (avoid untrusted input). Rationale and SourceRef are already escaped.

## Example: Color Theme Adjustment
Change accent & badge colors:
```css
:root {
  --accent:#0b4d91;
  --badge:#ac2e24;
  --panel:#eef4fa;
}
```
Add just after the existing `:root{...}` rule or override later in CSS.

## Testing Locally
1. Run `sb29guard serve --policy policy/domains.yaml`.
2. Visit: `http://localhost:8080/explain?domain=exampletool.com`.
3. Refresh after each template/CSS change (rebuild if embedded).

## Rollback Strategy
Keep a copy of original templates or rely on Git to revert changes quickly.

## Next Steps / Ideas
- Runtime `--templates` directory override
- Separate CSS file served with ETag for better caching
- Dark-mode specific palette variables
- Optional static assets directory (logo, favicon) with CSP adjustments

If you need one of these sooner, open an issue describing your use case.

---
Need help? Open an issue or PR with a draft of your desired change and we can guide you.
