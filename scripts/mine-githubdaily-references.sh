#!/usr/bin/env bash
set -euo pipefail

limit="${1:-80}"
source_dir="references/GitHubDaily"

if [[ ! -d "$source_dir" ]]; then
  echo "Missing $source_dir. Initialize the GitHubDaily submodule first." >&2
  exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

rg -n -i 'github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+|Go|Golang|SQLite|PostgreSQL|MySQL|self-host|自托管|高并发|并发|agent|workflow|queue|observability|plugin|database|auth' \
  "$source_dir" \
  -g '*.md' \
  > "$tmp" || true

echo -e "repo\tsignals\tsource"

awk -v limit="$limit" -F: '
  {
    line=$0
    while (match(line, /github\.com\/[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+/)) {
      repo=substr(line, RSTART, RLENGTH)
      gsub(/[),，。；;]+$/, "", repo)
      signals=""
      lower=tolower($0)
      if (lower ~ /go|golang/) signals=signals "go,"
      if (lower ~ /sqlite|postgresql|mysql|database|数据库/) signals=signals "database,"
      if (lower ~ /auth|权限|认证/) signals=signals "auth,"
      if (lower ~ /agent|workflow|plugin|observability|queue/) signals=signals "agent-workflow,"
      if (lower ~ /self-host|自托管/) signals=signals "self-hosted,"
      if (lower ~ /高并发|并发|concurrency/) signals=signals "concurrency,"
      if (signals == "") signals="needs-review,"
      sub(/,$/, "", signals)
      print repo "\t" signals "\t" $1 ":" $2
      line=substr(line, RSTART + RLENGTH)
    }
  }
' "$tmp" | sort -u | awk -v limit="$limit" 'NR <= limit { print }'

echo
echo "Review license and language before adding any candidate to references/README.md or .gitmodules." >&2
