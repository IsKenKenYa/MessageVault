#!/usr/bin/env bash
set -euo pipefail

tracked_generated_candidates="$(
  git ls-files \
    'android/**/build/**' \
    'android/.gradle/**' \
    'backend/data/**' \
    'web/node_modules/**' \
    '*.apk' \
    '*.aab' \
    '*.log'
)"

tracked_generated=""
while IFS= read -r path; do
  if [[ -n "$path" && -e "$path" ]]; then
    tracked_generated+="$path"$'\n'
  fi
done <<< "$tracked_generated_candidates"

if [[ -n "$tracked_generated" ]]; then
  echo "Tracked generated or local-only files were found:"
  echo "$tracked_generated"
  exit 1
fi

echo "Repo hygiene check passed."
