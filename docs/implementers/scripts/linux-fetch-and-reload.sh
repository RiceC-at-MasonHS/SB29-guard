#!/usr/bin/env bash
# SB29-guard – Linux nightly fetch and proxy reload
# Purpose: After sb29-guard refreshes its Google Sheet (default ~23:59 local),
#          fetch the latest domain list and import into your proxy, then reload.
# How to schedule (cron)
# 1) Save as /usr/local/bin/sb29-fetch-and-reload.sh and chmod +x /usr/local/bin/sb29-fetch-and-reload.sh
# 2) Edit crontab: crontab -e
# 3) Add line (10 minutes after nightly refresh):
#    10 0 * * * /usr/local/bin/sb29-fetch-and-reload.sh

# Env overrides (optional):
#   GUARD_BASE=https://guard.school.internal
#   OUT_FILE=/etc/proxy/blocked.txt
#   ONLY_WHEN_CHANGED=true

set -euo pipefail
# Allow env overrides; fall back to sensible defaults
GUARD_BASE="${GUARD_BASE:-https://guard.school.internal}"
OUT_FILE="${OUT_FILE:-/etc/proxy/blocked.txt}"
ONLY_WHEN_CHANGED="${ONLY_WHEN_CHANGED:-true}"

get_policy_version(){
  curl -fsS "$GUARD_BASE/metrics" | jq -r '.policy_version // ""' 2>/dev/null || true
}

prev_ver_file="${OUT_FILE}.ver"
prev_ver=""; [[ -f "$prev_ver_file" ]] && prev_ver="$(cat "$prev_ver_file" || true)"
cur_ver="$(get_policy_version)"
if [[ "$ONLY_WHEN_CHANGED" == "true" && -n "$cur_ver" && "$cur_ver" == "$prev_ver" ]]; then
  echo "No policy change (version $cur_ver). Skipping."
  exit 0
fi

echo "Fetching domain list from $GUARD_BASE/domain-list"
curl -fsS "$GUARD_BASE/domain-list" -o "$OUT_FILE"

# PROXY: Replace with your product's import/reload.
# NGINX example using rsync + reload on the proxy host:
#   PROXY_HOST=proxy
#   REMOTE_PATH=/etc/nginx/sb29/blocked.txt
#   rsync -az "$OUT_FILE" "$PROXY_HOST:$REMOTE_PATH"
#   ssh "$PROXY_HOST" 'nginx -s reload'
#
# HAProxy Runtime API map update (no reload):
#   # Pre-req in haproxy.cfg: stats socket ipv4@127.0.0.1:9999 level admin
#   HAPROXY_HOST=127.0.0.1
#   HAPROXY_PORT=9999
#   MAP_PATH=/etc/haproxy/blocked.map
#   echo "clear map $MAP_PATH" | nc $HAPROXY_HOST $HAPROXY_PORT
#   while IFS= read -r d; do
#     [ -n "$d" ] && printf 'add map %s %s 1\n' "$MAP_PATH" "$d" | nc $HAPROXY_HOST $HAPROXY_PORT
#   done < "$OUT_FILE"
#
# Squid / Others: import list per product docs, then reload service.

[[ -n "$cur_ver" ]] && echo -n "$cur_ver" > "$prev_ver_file"
echo "Done."
