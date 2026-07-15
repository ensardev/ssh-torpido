#!/usr/bin/env bash
# Build torpido for Linux, upload it, and restart the service.
# Usage: ./deploy/deploy.sh [ssh-target]   (default: root@torpido.dev)
set -euo pipefail

HOST="${1:-root@torpido.dev}"
cd "$(dirname "$0")/.."

echo "▶ building linux/amd64 binary…"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/torpido-linux-amd64 .

echo "▶ uploading to ${HOST}…"
scp dist/torpido-linux-amd64 "${HOST}:/opt/torpido/torpido.new"

echo "▶ swapping in the new binary and restarting…"
ssh "${HOST}" 'set -e
  mv /opt/torpido/torpido.new /opt/torpido/torpido
  chown torpido:torpido /opt/torpido/torpido
  chmod +x /opt/torpido/torpido
  systemctl restart torpido
  systemctl --no-pager status torpido | head -5'

echo "✓ deployed"
