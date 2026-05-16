#!/usr/bin/env bash
set -euo pipefail

web_baseline_manifest="web/reference-baseline-paths.txt"

is_web_baseline_path() {
  local path="$1"
  [[ -f "$web_baseline_manifest" ]] || return 1

  while IFS= read -r prefix; do
    [[ -n "$prefix" ]] || continue
    [[ "$prefix" =~ ^# ]] && continue
    if [[ "$path" == "$prefix"* ]]; then
      return 0
    fi
  done < "$web_baseline_manifest"

  return 1
}

tmp_handwritten="$(mktemp)"
tmp_generated="$(mktemp)"
tmp_commory_owned="$(mktemp)"
tmp_web_baseline="$(mktemp)"
trap 'rm -f "$tmp_handwritten" "$tmp_generated" "$tmp_commory_owned" "$tmp_web_baseline"' EXIT

find backend android web -type f \( \
  -name '*.go' -o \
  -name '*.kt' -o \
  -name '*.java' -o \
  -name '*.ts' -o \
  -name '*.tsx' -o \
  -name '*.vue' \
\) \
  -not -path '*/build/*' \
  -not -path '*/.gradle/*' \
  -not -path '*/node_modules/*' \
  -not -path '*/dist/*' \
  -not -path '*/coverage/*' \
  -not -path '*/sqlc/gen/*' \
  -not -path '*/schema/*' \
  -not -path '*/examples/*' \
  -not -path '*/fixture/*' \
  -not -path '*/fixtures/*' \
  -print0 |
  xargs -0 wc -l |
  awk '$2 != "total"' |
  sort -nr > "$tmp_handwritten"

find android backend web msglayer -type f \( \
  -name '*.json' -o \
  -name '*.xml' -o \
  -name '*.sql' \
\) \
  -not -path '*/build/*' \
  -not -path '*/.gradle/*' \
  -not -path '*/node_modules/*' \
  -not -path '*/dist/*' \
  -not -path '*/coverage/*' \
  -print0 |
  xargs -0 wc -l |
  awk '$2 != "total"' |
  sort -nr > "$tmp_generated"

echo "Largest hand-written source files by line count:"
head -30 "$tmp_handwritten"

while read -r lines path; do
  [[ -n "$lines" && -n "$path" ]] || continue
  if is_web_baseline_path "$path"; then
    echo "$lines $path" >> "$tmp_web_baseline"
  else
    echo "$lines $path" >> "$tmp_commory_owned"
  fi
done < "$tmp_handwritten"

echo
echo "Commory-owned hand-written source warning threshold: 350 lines. Hard limit: 500 lines."
awk '$1 > 350 { print }' "$tmp_commory_owned" || true

echo
echo "Reference-derived web baseline over 350 lines (tracked via $web_baseline_manifest):"
awk '$1 > 350 { print }' "$tmp_web_baseline" || true

echo
echo "Generated/schema/resource review threshold: 1000 lines."
awk '$1 > 1000 { print }' "$tmp_generated" || true
