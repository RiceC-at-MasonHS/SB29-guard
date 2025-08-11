# Test Plan (Proxy-first)

Scope
- Unit: policy, dnsgen, server handlers.
- CLI: generate-proxy, generate-explain-static.
- Integration (manual): proxy snippets in lab setups.

Cases
- PX-1: For each format and mode, snippet contains required directives and compiles in linter/smoke tools.
- PX-2: Static page shows d/c/v/h when present, ignores invalid inputs, and still looks correct without params.
- Header precedence: params vs headers (display vs lookup behavior).
- Security headers present and unchanged.
- Edge: long domains, invalid chars, mixed case, www+port, IPv6 brackets.

Non-goals
- End-to-end proxy performance testing (separate effort).
