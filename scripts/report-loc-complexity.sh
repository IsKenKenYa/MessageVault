#!/usr/bin/env bash
set -euo pipefail

echo "Largest source files by line count:"
find android backend msglayer -type f \( -name '*.kt' -o -name '*.go' -o -name '*.json' -o -name '*.xml' \) \
  -not -path '*/build/*' \
  -not -path '*/.gradle/*' \
  -print0 |
  xargs -0 wc -l |
  sort -nr |
  head -30

echo
echo "Large file review threshold: 500 lines for app/source files, 1000 lines for generated schemas/examples."
