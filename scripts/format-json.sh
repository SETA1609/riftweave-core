#!/usr/bin/env bash
# Lay out JSON files in the one format (CONTRIBUTING.md#json-format) inside the toolchain container.
# With no paths it formats those listed in ruleset/scripts/format_paths.txt; --check only lists
# the files it would change and fails if there are any. Needs uv on the host.
# Usage: ./scripts/format-json.sh [--check] [PATH ...]
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops-tools container --settings-file "$(dirname "$0")/../tools/devops/pyproject.toml" format-json "$@"
