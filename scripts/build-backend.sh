#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
OUTPUT_DIR=${OUTPUT_DIR:-"$ROOT_DIR/release"}
BINARY_NAME=${BINARY_NAME:-toughradius}

if [ ! -d "$ROOT_DIR/web/dist" ]; then
    echo "[backend] React Admin bundle missing (web/dist). Run scripts/build-frontend.sh first." >&2
    exit 1
fi

mkdir -p "$OUTPUT_DIR"

echo "[backend] Running Go tests..."
GOTOOLCHAIN=${GOTOOLCHAIN:-auto} go test ./...

echo "[backend] Building Go binary..."
CGO_ENABLED=0 GOTOOLCHAIN=${GOTOOLCHAIN:-auto} go build -o "$OUTPUT_DIR/$BINARY_NAME" ./

echo "[backend] Binary written to $OUTPUT_DIR/$BINARY_NAME"
