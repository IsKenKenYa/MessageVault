#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source_dir="$repo_root/.agents/skills"
target_dir="$repo_root/.claude/skills"
mode="${1:-sync}"

if [[ ! -d "$source_dir" ]]; then
  echo "Missing source skills directory: $source_dir" >&2
  exit 1
fi

invalid_files="$(
  find "$source_dir" -type f \( \
    -name 'README.md' -o \
    -name 'CHANGELOG.md' -o \
    -name 'AGENTS.md' -o \
    -name 'INSTALLATION_GUIDE.md' -o \
    -name 'QUICK_REFERENCE.md' \
  \) -print
)"

if [[ -n "$invalid_files" ]]; then
  echo "Skills contain non-standard documentation files:"
  echo "$invalid_files"
  exit 1
fi

missing_skill_md="$(
  find "$source_dir" -mindepth 1 -maxdepth 1 -type d -print | while IFS= read -r skill_dir; do
    [[ -f "$skill_dir/SKILL.md" ]] || echo "$skill_dir"
  done
)"

if [[ -n "$missing_skill_md" ]]; then
  echo "Skills missing SKILL.md:"
  echo "$missing_skill_md"
  exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

mkdir -p "$target_dir"
rsync -a --delete \
  --exclude '.DS_Store' \
  --exclude 'node_modules/' \
  --exclude '.git/' \
  --exclude 'build/' \
  --exclude 'dist/' \
  "$source_dir/" "$tmp_dir/skills/"

if [[ "$mode" == "--check" ]]; then
  if ! diff -qr "$tmp_dir/skills" "$target_dir" >/tmp/commory-skills-diff.txt; then
    echo ".claude/skills is out of sync with .agents/skills."
    cat /tmp/commory-skills-diff.txt
    exit 1
  fi
  echo "Agent skills mirror is in sync."
  exit 0
fi

rsync -a --delete "$tmp_dir/skills/" "$target_dir/"
echo "Synced .agents/skills to .claude/skills."
