#!/bin/sh
# Runs as the container entrypoint. The repo is mounted at /src.
set -eu

export HOME="${HOME:-/tmp/python-home}"
mkdir -p "$HOME"

if [ -e /src ]; then
    git config --global --add safe.directory /src >/dev/null 2>&1 || true
fi

if [ "$#" -eq 0 ]; then
    set -- python ruleset/scripts/validate.py
fi

exec "$@"
