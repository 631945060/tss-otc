#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

case "${1:-help}" in
  test)
    go test ./... -count=1
    ;;
  backend)
    SIGNER_MODE=local go run ./apps/tss-wallet-service/cmd httpServer
    ;;
  frontend)
    echo "前端是内置静态资源，无需 Node。请启动 backend 后打开 http://127.0.0.1:8080"
    ;;
  tss1)
    go run ./apps/tss-wallet-service/cmd tssNode --node-id node-1 --addr 127.0.0.1:9441
    ;;
  tss2)
    go run ./apps/tss-wallet-service/cmd tssNode --node-id node-2 --addr 127.0.0.1:9442
    ;;
  tss3)
    go run ./apps/tss-wallet-service/cmd tssNode --node-id node-3 --addr 127.0.0.1:9443
    ;;
  consumers)
    go run ./apps/tss-wallet-service/cmd startConsumers
    ;;
  cron)
    go run ./apps/tss-wallet-service/cmd cron
    ;;
  docker)
    docker compose -f deployments/docker/docker-compose.yml up --build
    ;;
  help|*)
    echo "用法: $0 {test|backend|frontend|tss1|tss2|tss3|consumers|cron|docker}"
    echo "前三个 TSS 节点请分别在三个终端启动。"
    ;;
esac
