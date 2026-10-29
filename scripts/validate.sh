#!/usr/bin/env bash
# Run ruleset/scripts/validate.py inside the toolchain container. The repo is mounted at /src.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
# Usage: ./scripts/validate.sh [--keep-cache]
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops validate "$@"
