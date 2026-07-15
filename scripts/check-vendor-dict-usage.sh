#!/usr/bin/env bash
# check-vendor-dict-usage.sh - CI gate: every vendor dictionary package under
# internal/radiusd/vendors/ must be imported somewhere outside itself (parser,
# enhancer, EAP handler, test, ...). A zero-reference dictionary is dead weight
# in the repo (see docs/vendor-vsa-gap-baseline.md); either wire it into a
# parser/enhancer or delete it and regenerate on demand.
#
# Usage: scripts/check-vendor-dict-usage.sh
# Exits non-zero and lists offending packages if any dictionary is unreferenced.
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
VENDORS_DIR="$ROOT_DIR/internal/radiusd/vendors"
MODULE_PREFIX="github.com/talkincode/toughradius/v9/internal/radiusd/vendors"

fail=0
for dir in "$VENDORS_DIR"/*/; do
    pkg=$(basename "$dir")
    import_path="$MODULE_PREFIX/$pkg"
    # Count Go files outside the package itself that import it. The trailing
    # `|| true` keeps `set -o pipefail` from aborting when grep finds nothing.
    refs=$(grep -rl --include='*.go' "\"$import_path\"" "$ROOT_DIR" 2>/dev/null \
        | grep -v "internal/radiusd/vendors/$pkg/" | wc -l | tr -d ' ' || true)
    if [ "$refs" -eq 0 ]; then
        echo "::error::vendor dictionary '$pkg' has zero external references ($import_path)"
        fail=1
    else
        echo "ok: $pkg ($refs external reference(s))"
    fi
done

if [ "$fail" -ne 0 ]; then
    echo
    echo "Zero-reference vendor dictionaries detected. Wire each one into a" >&2
    echo "parser/enhancer (see internal/radiusd/plugins/vendorparsers/parsers/)" >&2
    echo "or remove the package; it can be regenerated when actually needed." >&2
    exit 1
fi
