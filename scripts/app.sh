#!/usr/bin/env bash
# Run the frontend and its API on the host (docs/backend/frontend.md#running-it): build and start
# riftweave-server, then `ng serve` with a proxy to it. Ctrl-C stops both.
# Needs Go and Node. With Node below 24.15 (Angular 22's minimum) it runs Node 24.22.1 through npx.
# Usage: ./scripts/app.sh [--port 4200] [--api-port 8080]
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
port=4200
api_port=8080
node_version=24.22.1
npm_version=11.21.1

while [[ $# -gt 0 ]]; do
  case $1 in
    --port) port=$2; shift 2 ;;
    --api-port) api_port=$2; shift 2 ;;
    -h | --help) sed -n '2,6p' "$0"; exit 0 ;;
    *) echo "app.sh: unknown argument $1" >&2; exit 2 ;;
  esac
done

for tool in go node npx curl; do
  command -v "$tool" > /dev/null || { echo "app.sh: $tool is not installed" >&2; exit 1; }
done

# Angular 22 needs Node 24.15 or later; an older Node gets a pinned one for this run only.
if ! node -e 'const [a, b] = process.versions.node.split(".").map(Number); process.exit(a > 24 || (a === 24 && b >= 15) ? 0 : 1)'; then
  echo "app.sh: Node $(node --version) is below 24.15; using Node $node_version through npx" >&2
  node_dir=$(npx --yes -p "node@$node_version" -c 'dirname "$(command -v node)"' | tail -n 1)
  PATH="$node_dir:$PATH"
fi

mkdir -p "$root/cache/app"
echo "app.sh: building riftweave-server" >&2
(cd "$root/backend" && go build -o "$root/cache/app/riftweave-server" ./cmd/riftweave-server)

"$root/cache/app/riftweave-server" --data "$root/ruleset/data" --addr "127.0.0.1:$api_port" &
server=$!
trap 'kill "$server" 2> /dev/null || true' EXIT INT TERM

for _ in $(seq 1 30); do
  curl -fs -o /dev/null "http://127.0.0.1:$api_port/v1/commands" && break
  kill -0 "$server" 2> /dev/null || { echo "app.sh: the server stopped; is port $api_port free?" >&2; exit 1; }
  sleep 1
done
echo "app.sh: API at http://127.0.0.1:$api_port" >&2

cd "$root/frontend"
if [[ ! -f node_modules/.package-lock.json || package-lock.json -nt node_modules/.package-lock.json ]]; then
  echo "app.sh: installing the frontend's packages" >&2
  npx --yes "npm@$npm_version" ci --no-audit --no-fund
fi

proxy="$root/cache/app/proxy.conf.json"
printf '{ "/v1": { "target": "http://127.0.0.1:%s" }, "/openapi.json": { "target": "http://127.0.0.1:%s" } }\n' "$api_port" "$api_port" > "$proxy"
npx ng serve --port "$port" --proxy-config "$proxy"
