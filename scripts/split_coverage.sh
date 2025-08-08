#!/usr/bin/env bash
set -euo pipefail

# Packages to measure (exclude docs & vendor)
PACKAGES=(
  ./internal/policy
  ./internal/dnsgen
)

COVER_MODE=atomic
OUT_DIR=coverage
THRESHOLD_POLICY=70
THRESHOLD_DNSGEN=70

rm -rf "$OUT_DIR" && mkdir -p "$OUT_DIR"

summary_file="$OUT_DIR/summary.txt"
json_file="$OUT_DIR/summary.json"
echo "Per-Package Coverage:" > "$summary_file"
echo '{"packages":[' > "$json_file"
first=1

fail=0
for pkg in "${PACKAGES[@]}"; do
  base=$(basename "$pkg")
  profile="$OUT_DIR/coverage.${base}.out"
  echo "-- Running coverage for $pkg" >&2
  go test -count=1 -covermode=$COVER_MODE -coverprofile="$profile" "$pkg" >/dev/null
  line=$(go tool cover -func="$profile" | tail -n 1)
  pct=$(echo "$line" | awk '{print $3}' | sed 's/%//')
  printf "%s\n" "$line" >> "$summary_file"
  # JSON append
  if [ $first -eq 0 ]; then echo ',' >> "$json_file"; fi
  first=0
  echo "  {\"package\":\"$pkg\",\"coverage_percent\":$pct}" >> "$json_file"
  # Threshold check
  case "$base" in
    policy)
      if (( ${pct%.*} < THRESHOLD_POLICY )); then
        echo "Coverage below threshold for policy: $pct% < $THRESHOLD_POLICY%" >&2
        fail=1
      fi
      ;;
    dnsgen)
      if (( ${pct%.*} < THRESHOLD_DNSGEN )); then
        echo "Coverage below threshold for dnsgen: $pct% < $THRESHOLD_DNSGEN%" >&2
        fail=1
      fi
      ;;
  esac
done

echo ']}' >> "$json_file"

if [ $fail -ne 0 ]; then
  echo "One or more coverage thresholds not met" >&2
  cat "$summary_file" >&2
  exit 1
fi

echo "\nCoverage Summary:" >> "$summary_file"
cat "$summary_file"
