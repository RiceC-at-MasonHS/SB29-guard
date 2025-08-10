#!/usr/bin/env bash
# Generate DNS artifacts using the sb29guard container (easy-mode)
# Usage examples:
#   ./gen-dns.sh hosts a-record 10.10.10.50
#   ./gen-dns.sh bind cname blocked.guard.local
#   ./gen-dns.sh domain-list
#   ./gen-dns.sh dnsmasq a-record 10.10.10.50
#   ./gen-dns.sh winps cname blocked.guard.local
set -euo pipefail

FORMAT=${1:-}
MODE=${2:-}
REDIRECT=${3:-}
OUTFILE=${4:-}

if [[ -z "$FORMAT" ]]; then
  echo "Usage: $0 <format> [a-record|cname] [redirect] [outfile]" >&2
  exit 1
fi

COMPOSE_FILE="$(cd "$(dirname "$0")" && pwd)/docker-compose.yml"
POLICY_DIR="$(cd "$(dirname "$0")" && pwd)/policy"
OUT_DIR="$(cd "$(dirname "$0")" && pwd)/out"
if [[ ! -f "$COMPOSE_FILE" ]]; then echo "Compose file not found: $COMPOSE_FILE" >&2; exit 1; fi
if [[ ! -d "$POLICY_DIR" ]]; then echo "Policy dir not found: $POLICY_DIR" >&2; exit 1; fi
mkdir -p "$OUT_DIR"

# Default output file
if [[ -z "${OUTFILE}" ]]; then
  case "$FORMAT" in
    hosts) OUTFILE="$OUT_DIR/hosts.txt" ;;
    bind) OUTFILE="$OUT_DIR/zone.db" ;;
    unbound) OUTFILE="$OUT_DIR/unbound.conf" ;;
    rpz) OUTFILE="$OUT_DIR/policy.rpz" ;;
    dnsmasq) OUTFILE="$OUT_DIR/dnsmasq.conf" ;;
    domain-list) OUTFILE="$OUT_DIR/domains.txt" ;;
    winps) OUTFILE="$OUT_DIR/windows-dns.ps1" ;;
    *) OUTFILE="$OUT_DIR/output.txt" ;;
  esac
fi

CMD=(compose -f "$COMPOSE_FILE" run --rm sb29guard generate-dns --policy /app/policy/domains.yaml --format "$FORMAT")
if [[ -n "$MODE" ]]; then CMD+=(--mode "$MODE"); fi
if [[ -n "$REDIRECT" ]]; then
  if [[ "$MODE" == "a-record" ]]; then CMD+=(--redirect-ipv4 "$REDIRECT"); fi
  if [[ "$MODE" == "cname" ]]; then CMD+=(--redirect-host "$REDIRECT"); fi
fi
BASENAME=$(basename "$OUTFILE")
CMD+=(--out "/out/$BASENAME")

echo "Running: docker ${CMD[*]}"
docker "${CMD[@]}"

echo "Wrote: $OUTFILE"
