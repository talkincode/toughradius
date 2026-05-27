#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
CMD_DIR="$ROOT_DIR/cmd"
BIN_DIR="$ROOT_DIR/bin"
GO_BIN=${GO_BIN:-go}

if ! command -v "$GO_BIN" >/dev/null 2>&1; then
    echo "[build-cmds] '$GO_BIN' command not found. Please install Go or adjust GO_BIN." >&2
    exit 1
fi

mkdir -p "$BIN_DIR"

cd "$ROOT_DIR"

declare -A PACKAGE_SET=()

while IFS= read -r pkg; do
    [[ -z "$pkg" ]] && continue
    PACKAGE_SET["$pkg"]=1
done < <(GOTOOLCHAIN=${GOTOOLCHAIN:-auto} "$GO_BIN" list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./cmd/... | sed '/^$/d')

while IFS= read -r dir; do
    [[ -z "$dir" ]] && continue
    rel_path=${dir#$ROOT_DIR/}
    if [[ "$rel_path" == "$dir" ]]; then
        rel_path="cmd/$(basename "$dir")"
    fi
    pkg=$(GOTOOLCHAIN=${GOTOOLCHAIN:-auto} "$GO_BIN" list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' "./$rel_path" 2>/dev/null || true)
    [[ -z "$pkg" ]] && continue
    PACKAGE_SET["$pkg"]=1
done < <(find "$CMD_DIR" -mindepth 1 -maxdepth 1 -type d -name 'testdata' -print 2>/dev/null)

if [ ${#PACKAGE_SET[@]} -eq 0 ]; then
    echo "[build-cmds] No main packages detected under cmd/. Nothing to build."
    exit 0
fi

mapfile -t MAIN_PACKAGES < <(printf '%s\n' "${!PACKAGE_SET[@]}" | sort)

for pkg in "${MAIN_PACKAGES[@]}"; do
    binary_name=$(basename "$pkg")
    output_path="$BIN_DIR/$binary_name"
    echo "[build-cmds] Building $pkg -> $output_path"
    CGO_ENABLED=${CGO_ENABLED:-0} GOTOOLCHAIN=${GOTOOLCHAIN:-auto} "$GO_BIN" build -o "$output_path" "$pkg"
    echo "[build-cmds] Done: $output_path"
    echo
done

echo "[build-cmds] Completed building ${#MAIN_PACKAGES[@]} command(s) into $BIN_DIR"
