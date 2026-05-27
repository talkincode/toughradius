#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
WEB_DIR="$ROOT_DIR/web"

cd "$WEB_DIR"

if [ ! -d node_modules ]; then
  echo "[frontend] Installing dependencies..."
  npm install
fi

echo "[frontend] Building React Admin bundle..."
npm run build

echo "[frontend] Assets written to $WEB_DIR/dist"
