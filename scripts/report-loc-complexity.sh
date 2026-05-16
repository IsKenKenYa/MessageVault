#!/usr/bin/env bash
set -euo pipefail

tmp_handwritten="$(mktemp)"
tmp_generated="$(mktemp)"
trap 'rm -f "$tmp_handwritten" "$tmp_generated"' EXIT

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

echo
echo "Hand-written source warning threshold: 350 lines. Hard limit: 500 lines."
awk '$1 > 350 { print }' "$tmp_handwritten" || true

echo
echo "Generated/schema/resource review threshold: 1000 lines."
awk '$1 > 1000 { print }' "$tmp_generated" || true
