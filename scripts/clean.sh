#!/usr/bin/env bash
# End-of-work cleanup: remove Python caches and the pip cache volume, then prune.
# Run it only when the user asks or says the work is finished. Nothing else calls it.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops clean "$@"
