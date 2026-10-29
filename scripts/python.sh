#!/usr/bin/env bash
# Run python inside the toolchain container, with the repo at /src.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
# Usage: ./scripts/python.sh --version
#        ./scripts/python.sh ruleset/scripts/validate.py
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops python "$@"
