#!/usr/bin/env bash
# Build the core to WebAssembly for the frontend (docs/backend/frontend.md#the-core-in-the-browser).
# It copies ruleset/data, ruleset/modules, and ruleset/schemas beside the entry point, so they are
# embedded in the binary and the core, its data, and the schemas author checks against come from
# one commit, then writes frontend/public/riftweave.wasm
# and the wasm_exec.js of the same Go version. Needs Go on the host.
# Usage: ./scripts/build-wasm.sh
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
entry="$root/backend/cmd/riftweave-wasm"
public="$root/frontend/public"

rm -rf "$entry/ruleset"
mkdir -p "$entry/ruleset" "$public"
cp -r "$root/ruleset/data" "$entry/ruleset/data"
cp -r "$root/ruleset/modules" "$entry/ruleset/modules"
cp -r "$root/ruleset/schemas" "$entry/ruleset/schemas"

(cd "$root/backend" && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$public/riftweave.wasm" ./cmd/riftweave-wasm)
# The same Go version must build the module and provide its JavaScript support file.
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$public/wasm_exec.js"

bytes=$(wc -c < "$public/riftweave.wasm")
gzipped=$(gzip -9 -c "$public/riftweave.wasm" | wc -c)
echo "frontend/public/riftweave.wasm: $bytes bytes, $gzipped gzipped"
