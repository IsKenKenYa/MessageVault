#!/usr/bin/env bash
set -euo pipefail

tracked_generated_candidates="$(
  git ls-files \
    '.DS_Store' \
    '**/.DS_Store' \
    'android/**/build/**' \
    'android/.gradle/**' \
    'android/app/debug/**' \
    'backend/data/**' \
    'web/node_modules/**' \
    '*.apk' \
    '*.aab' \
    '*.log' \
    '**/AI_EDIT_LOG.md'
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

invalid_skill_docs="$(
  find .agents/skills -type f \( \
    -name 'README.md' -o \
    -name 'CHANGELOG.md' -o \
    -name 'AGENTS.md' -o \
    -name 'INSTALLATION_GUIDE.md' -o \
    -name 'QUICK_REFERENCE.md' \
  \) -print 2>/dev/null || true
)"

if [[ -n "$invalid_skill_docs" ]]; then
  echo "Skills contain non-standard documentation files:"
  echo "$invalid_skill_docs"
  exit 1
fi

echo "Repo hygiene check passed."
