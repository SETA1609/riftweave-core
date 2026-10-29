#!/usr/bin/env bash
# Run a command (default: bash) in the agent container: the toolchain image plus Claude Code
# and uv, with its own login in the home volume personal-agent-home.
# The host's ~/.claude and ~/.gitconfig are not mounted. Needs uv on the host.
# Usage: ./scripts/agent.sh
#        ./scripts/agent.sh claude
#        ./scripts/agent.sh --keep-cache --rebuild-image claude
set -euo pipefail
exec uv run --project "$(dirname "$0")/../tools/devops" --locked --no-dev --quiet devops agent "$@"
