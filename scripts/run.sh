#!/usr/bin/env bash
# Build and serve the stegno web app. Container has no system Go/node -> use the tools bundle.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export PATH="/home/krxsna/.hermes/tools/go-1.27.2-linux-x64/bin:/home/krxsna/.hermes/tools/node-26.7.0-linux-x64/bin:$PATH"
export GOPATH="${GOPATH:-/home/krxsna/.hermes/cache/scratch/gopath}"
export GOCACHE="${GOCACHE:-/home/krxsna/.hermes/cache/scratch/gocache}"
export GOTOOLCHAIN=local
export GOFLAGS=-mod=mod

cd "$ROOT"

WEB_DIST="$ROOT/web/dist"
if [ ! -f "$WEB_DIST/index.html" ]; then
  if [ -d "$ROOT/web" ]; then
    echo "[run] building frontend -> web/dist"
    cd "$ROOT/web"
    npm install --no-audit --no-fund
    npm run build
    cd "$ROOT"
  else
    echo "[run] no web/ directory, serving API only"
    WEB_DIST=""
  fi
fi

echo "[run] go build"
go build -o "$ROOT/bin/stegno-server" ./cmd/server
go build -o "$ROOT/bin/stegno" ./cmd/stegno

echo "[run] listening on :8080 (web dir: ${WEB_DIST:-none})"
exec "$ROOT/bin/stegno-server" -addr :8080 -web "${WEB_DIST:-}"
