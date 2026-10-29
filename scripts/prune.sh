#!/usr/bin/env bash
# Remove dangling images and unused build cache. The tagged image and the pip cache volume stay.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops prune "$@"
