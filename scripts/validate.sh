#!/usr/bin/env bash
# Run ruleset/scripts/validate.py inside the toolchain container.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
# Usage: ./scripts/validate.sh [--keep-cache]
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops-tools container --settings-file "$(dirname "$0")/../tools/devops/pyproject.toml" validate "$@"
