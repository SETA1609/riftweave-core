#!/usr/bin/env bash
# Build the frontend, its API, and the ruleset data into one image and run it
# (container/app.Dockerfile, docs/backend/frontend.md#running-it). Ctrl-C stops it.
# Needs docker or podman; RIFTWEAVE_ENGINE=docker|podman picks one.
# Usage: ./scripts/app-container.sh [--port 8080] [--build-only]
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
port=8080
build_only=false
image=localhost/riftweave-app:local  # a full name, since podman does not resolve short ones

while [[ $# -gt 0 ]]; do
  case $1 in
    --port) port=$2; shift 2 ;;
    --build-only) build_only=true; shift ;;
    -h | --help) sed -n '2,6p' "$0"; exit 0 ;;
    *) echo "app-container.sh: unknown argument $1" >&2; exit 2 ;;
  esac
done

engine=${RIFTWEAVE_ENGINE:-}
if [[ -z $engine ]]; then
  for candidate in docker podman; do
    command -v "$candidate" > /dev/null && { engine=$candidate; break; }
  done
fi
[[ -n $engine ]] || { echo "app-container.sh: install docker or podman" >&2; exit 1; }
if [[ $engine == podman && $(id -u) != 0 && -z $(awk -F: -v user="$(id -un)" '$1 == user' /etc/subuid 2> /dev/null) ]]; then
  echo "app-container.sh: rootless podman needs subordinate ids for $(id -un) in /etc/subuid and /etc/subgid" >&2
  echo "  (sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $(id -un); podman system migrate)" >&2
  exit 1
fi

"$engine" build -f "$root/container/app.Dockerfile" -t "$image" "$root"
$build_only && exit 0

terminal=()
[[ -t 0 && -t 1 ]] && terminal=(-it)
echo "app-container.sh: the app is at http://127.0.0.1:$port" >&2
exec "$engine" run --rm --init "${terminal[@]}" -p "127.0.0.1:$port:8080" "$image"
