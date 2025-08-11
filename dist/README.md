# Example proxy bundles

This folder may contain generated example bundles produced by:
- sb29guard generate-proxy --format <nginx|caddy|haproxy|apache> --mode header-injection --bundle-dir dist/<name>
- sb29guard generate-explain-static --out-dir dist/explain

Notes
- These are examples for operators to copy/paste and adapt.
- Re-running the commands will overwrite files in these directories.
- The dist/ folder is ignored by git except for this README (and optional .gitkeep), to avoid noise from local experiments.
- For convenience, you can generate all examples with: `make examples`
