#!/usr/bin/env bash
# Remove dangling images and unused build cache. Tagged images and volumes stay.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops-tools container --settings-file "$(dirname "$0")/../tools/devops/pyproject.toml" prune "$@"
