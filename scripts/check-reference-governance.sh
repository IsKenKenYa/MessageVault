#!/usr/bin/env bash
set -euo pipefail

catalog="references/README.md"
if [[ ! -f "$catalog" ]]; then
  echo "Missing $catalog"
  exit 1
fi

expected_paths=(
  "references/art-design-pro"
  "references/new-api"
  "references/hermes-agent"
  "references/openclaw"
  "references/GitHubDaily"
  "references/memos"
  "references/gitea"
  "references/pocketbase"
  "references/sqlc"
)

for path in "${expected_paths[@]}"; do
  if ! git config --file .gitmodules --get-regexp '^submodule\..*\.path$' | awk '{print $2}' | grep -Fxq "$path"; then
    echo "Missing reference submodule path in .gitmodules: $path"
    exit 1
  fi
  if ! grep -Fq "\`$path\`" "$catalog"; then
    echo "Missing reference catalog entry: $path"
    exit 1
  fi
done

restricted_patterns=(
  "references/ClaudeCodeSource.*clean-room"
  "references/GitHubDaily.*discovery index"
  "references/new-api.*AGPL read-only"
)

for pattern in "${restricted_patterns[@]}"; do
  if ! grep -Eq "$pattern" "$catalog"; then
    echo "Missing restricted reference policy matching: $pattern"
    exit 1
  fi
done

if grep -RIn --exclude-dir='.git' 'references/ClaudeCodeSource' backend android web .github 2>/dev/null; then
  echo "ClaudeCodeSource must only be referenced in governance documentation, not product code."
  exit 1
fi

for path in "${expected_paths[@]}"; do
  if [[ -d "$path" && -e "$path/.git" || -f "$path/.git" ]]; then
    license_file="$(find "$path" -maxdepth 1 \( -iname 'LICENSE*' -o -iname 'COPYING*' \) -type f | head -n 1 || true)"
    case "$path" in
      references/new-api)
        if [[ -n "$license_file" ]] && ! grep -qi 'affero general public license' "$license_file"; then
          echo "Expected AGPL license for $path"
          exit 1
        fi
        ;;
      references/GitHubDaily)
        ;;
      *)
        if [[ -z "$license_file" ]]; then
          echo "Missing license file for initialized reference: $path"
          exit 1
        fi
        if ! grep -Eqi 'MIT License|Apache License|BSD [0-9]-Clause|Permission is hereby granted' "$license_file"; then
          echo "Reference is not recognized as permissive: $path ($license_file)"
          exit 1
        fi
        ;;
    esac
  fi
done

echo "Reference governance check passed."
