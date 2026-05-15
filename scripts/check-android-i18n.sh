#!/usr/bin/env bash
set -euo pipefail

base="android/app/src/main/res/values/strings.xml"
en="android/app/src/main/res/values-en/strings.xml"
zh="android/app/src/main/res/values-zh-rCN/strings.xml"

extract_keys() {
  sed -n 's/.*<string name="\([^"]*\)".*/\1/p' "$1" | sort -u
}

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

extract_keys "$base" > "$tmp_dir/base"
extract_keys "$en" > "$tmp_dir/en"
extract_keys "$zh" > "$tmp_dir/zh"

missing_en="$(comm -23 "$tmp_dir/base" "$tmp_dir/en" || true)"
missing_zh="$(comm -23 "$tmp_dir/base" "$tmp_dir/zh" || true)"

if [[ -n "$missing_en" || -n "$missing_zh" ]]; then
  if [[ -n "$missing_en" ]]; then
    echo "Missing English string keys:"
    echo "$missing_en"
  fi
  if [[ -n "$missing_zh" ]]; then
    echo "Missing Simplified Chinese string keys:"
    echo "$missing_zh"
  fi
  exit 1
fi

echo "Android i18n key check passed."
