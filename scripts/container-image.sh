#!/usr/bin/env bash
# Build the Python toolchain image, then prune dangling layers and build cache.
# The tagged image riftweave:python-<version> is kept. The version is .toolchain-python-version.
# Output is timestamped and logged to cache/logs/. Needs uv on the host.
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops image "$@"
