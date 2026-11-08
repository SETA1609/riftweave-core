#!/usr/bin/env bash
# Build the devops-tools agent image and tag it riftweave-agent:devcontainer, the image
# .devcontainer/devcontainer.json names (its initializeCommand). The agent tag carries a
# content hash, so the devcontainer cannot name it directly. Needs uv and docker or podman.
set -euo pipefail
cd "$(dirname "$0")/.."
./scripts/agent.sh true
if docker info >/dev/null 2>&1; then engine=docker; else engine=podman; fi
version="$(cat .toolchain-python-version)"
image="$("$engine" images riftweave-agent --format '{{.Repository}}:{{.Tag}}' | grep -F ":python-$version-" | head -n 1)"
[[ -n "$image" ]] || { echo "no riftweave-agent image after ./scripts/agent.sh" >&2; exit 1; }
"$engine" tag "$image" riftweave-agent:devcontainer
echo "riftweave-agent:devcontainer -> $image"
