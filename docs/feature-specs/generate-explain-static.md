# Feature Spec: sb29guard generate-explain-static

Goal
Emit a static explain page bundle for simple hosting.

CLI
- name: generate-explain-static
- flags:
  - --out-dir: directory (required)
  - --title: default "SB29 Guard"
  - --law-url: optional override
  - --inline-css: default true

Bundle
- index.html: reads d,c,v,h from URL (display-only), no JS required; server-side-friendly markup.
- style.css: same visual language as dynamic page.
- README.md: deploy instructions and param contract.

Validation
- Sanitize title; omit external scripts; CSP-friendly content.

Acceptance
- PX-2: Writes all files; renders without JS; parameters affect display only.
