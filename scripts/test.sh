#!/usr/bin/env bash
# Run the ruleset unit tests inside the toolchain container.
# With no arguments this discovers ruleset/scripts/test_*.py.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
# Usage: ./scripts/test.sh [--keep-cache] [unittest args...]
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops test "$@"
