# Migration: Legacy easy-mode to Proxy-first

Summary
- easy-mode is removed. Use proxy/gateway integration instead.

Steps
1) Choose model: header-injection (preferred) or redirect to static explain site.
2) Use sb29guard generate-proxy to get a starting config.
3) If redirect model, host the static bundle from generate-explain-static.
4) Keep Host fallback disabled unless you explicitly rely on DNS-only overrides.

Verification
- Header-injection: curl with X-Original-Host returns 200 and expected content.
- Redirect: 302 to /explain with d= param; static page renders.

Notes
- HTTPS cert warnings disappear when proxy terminates TLS for your domain.
