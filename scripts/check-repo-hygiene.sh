#!/usr/bin/env bash
set -euo pipefail

is_handwritten_source() {
  local path="$1"
  case "$path" in
    backend/*|android/*|web/*)
      ;;
    *)
      return 1
      ;;
  esac

  case "$path" in
    *.go|*.kt|*.java|*.ts|*.tsx|*.vue)
      ;;
    *)
      return 1
      ;;
  esac

  case "$path" in
    */build/*|*/.gradle/*|*/node_modules/*|*/dist/*|*/coverage/*|*/sqlc/gen/*|*/schema/*|*/examples/*|*/fixture/*|*/fixtures/*)
      return 1
      ;;
  esac

  return 0
}

resolve_compare_base() {
  if [[ -n "${GITHUB_BASE_REF:-}" ]] && git show-ref --verify --quiet "refs/remotes/origin/${GITHUB_BASE_REF}"; then
    echo "refs/remotes/origin/${GITHUB_BASE_REF}"
    return 0
  fi

  local origin_head
  origin_head="$(git symbolic-ref -q --short refs/remotes/origin/HEAD 2>/dev/null || true)"
  if [[ -n "$origin_head" ]]; then
    echo "$origin_head"
    return 0
  fi

  if git show-ref --verify --quiet refs/heads/main; then
    echo "main"
    return 0
  fi

  return 1
}

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

tmp_candidates="$(mktemp)"
trap 'rm -f "$tmp_candidates"' EXIT

compare_base="$(resolve_compare_base || true)"
if [[ -n "$compare_base" ]]; then
  merge_base="$(git merge-base HEAD "$compare_base" 2>/dev/null || true)"
  if [[ -n "$merge_base" ]]; then
    git diff --name-only --diff-filter=ACMR "$merge_base"...HEAD >> "$tmp_candidates"
  fi
fi

git diff --name-only --diff-filter=ACMR >> "$tmp_candidates"
git diff --cached --name-only --diff-filter=ACMR >> "$tmp_candidates"
git ls-files --others --exclude-standard >> "$tmp_candidates"

oversized_sources=""
while IFS= read -r path; do
  [[ -n "$path" ]] || continue
  is_handwritten_source "$path" || continue
  [[ -f "$path" ]] || continue
  lines="$(wc -l < "$path")"
  if (( lines > 500 )); then
    oversized_sources+="${lines} ${path}"$'\n'
  fi
done < <(sort -u "$tmp_candidates")

if [[ -n "$oversized_sources" ]]; then
  echo "Changed hand-written source files exceeded the 500-line hard limit:"
  echo "$oversized_sources" | sort -nr
  exit 1
fi

echo "Repo hygiene check passed."
