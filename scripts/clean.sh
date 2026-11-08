#!/usr/bin/env bash
# End-of-work cleanup: remove the clean-directories and the uv cache volume, then prune.
# Run it only when the user asks or says the work is finished. Nothing else calls it.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops-tools container --settings-file "$(dirname "$0")/../tools/devops/pyproject.toml" clean "$@"
