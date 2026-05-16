#!/usr/bin/env bash
set -euo pipefail

# Commory 开发模式一键启动
# 同时运行 Go backend (:3000) 和 Web dev server (:3006)
# 用法: bash scripts/dev.sh

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

CYAN='\033[36m'
YELLOW='\033[33m'
RED='\033[31m'
RESET='\033[0m'

cleanup() {
  echo -e "\n${RED}[dev]${RESET} 正在停止所有服务..."
  kill 0 2>/dev/null || true
  wait 2>/dev/null || true
  echo -e "${RED}[dev]${RESET} 已停止。"
}

trap cleanup INT TERM

echo -e "${CYAN}[dev]${RESET} 启动 Go backend (端口 :3000)..."
(cd "$REPO_ROOT/backend" && go run ./cmd/commory serve) 2>&1 \
  | sed -u "s/.*/${CYAN}[backend]${RESET} &/" &

echo -e "${YELLOW}[dev]${RESET} 启动 Web dev server (端口 :3006)..."
(cd "$REPO_ROOT/web" && pnpm dev) 2>&1 \
  | sed -u "s/.*/${YELLOW}[web]${RESET}     &/" &

echo -e "${RED}[dev]${RESET} 按 Ctrl+C 停止所有服务。"
wait
