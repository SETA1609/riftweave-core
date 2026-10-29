#!/usr/bin/env bash
# Open an interactive bash in the toolchain container, with the repo at /src.
# Only session start, end, and exit code are logged to cache/logs/. Needs uv on the host.
# Usage: ./scripts/shell.sh
#        ./scripts/shell.sh --keep-cache
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops shell "$@"
