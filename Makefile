# Convenience targets for local examples (optional)

.PHONY: examples nginx caddy haproxy apache explain clean

examples: nginx caddy haproxy apache explain

nginx:
	go run ./cmd/sb29guard generate-proxy --format nginx --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/nginx

caddy:
	go run ./cmd/sb29guard generate-proxy --format caddy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/caddy

haproxy:
	go run ./cmd/sb29guard generate-proxy --format haproxy --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/haproxy

apache:
	go run ./cmd/sb29guard generate-proxy --format apache --mode header-injection --site-host blocked.school.local --backend-url http://127.0.0.1:8080 --bundle-dir dist/apache

explain:
	go run ./cmd/sb29guard generate-explain-static --out-dir dist/explain

# Remove generated bundles (keep dist/README.md)
clean:
	@echo Cleaning dist/
	@find dist -mindepth 1 -maxdepth 1 -type d -exec rm -rf {} +
	@find dist -maxdepth 1 -type f ! -name README.md -exec rm -f {} +
