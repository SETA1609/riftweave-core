"""
Constants shared by the devops CLI modules.
"""

from __future__ import annotations

import re
from pathlib import Path

# Container
TOOLCHAIN_IMAGE_REPOSITORY = "riftweave"
AGENT_IMAGE_REPOSITORY = "riftweave-agent"
# Not `.python-version`: pyenv and uv would treat a root pin as the host interpreter.
PYTHON_VERSION_FILE_NAME = ".toolchain-python-version"
CONTAINER_HOME_DIRECTORY = "/tmp/python-home"
CONTAINER_PIP_CACHE_DIRECTORY = f"{CONTAINER_HOME_DIRECTORY}/pip-cache"
CONTAINER_REPOSITORY_MOUNT = "/src"
CONTAINER_KEPT_PIP_CACHE_MOUNT = "/pip-cache"
PIP_CACHE_VOLUME_NAME = "riftweave-pip-cache"
KEEP_PIP_CACHE_FLAG = "--keep-cache"
SUPPORTED_CONTAINER_BINARIES = ("docker", "podman")
DOCKERFILE_PYTHON_VERSION_PATTERN = re.compile(r"^ARG PYTHON_VERSION=(\S+)\s*$", re.MULTILINE)
DEVCONTAINER_BASE_IMAGE_PATTERN = re.compile(r'"BASE_IMAGE"\s*:\s*"riftweave:python-([^"]+)"')

# Agent container
# Generic on purpose: every repo's agent container mounts the same home, so one login serves all.
DEFAULT_AGENT_HOME_VOLUME_NAME = "personal-agent-home"
CONTAINER_AGENT_HOME_DIRECTORY = "/home/agent"
CONTAINER_AGENT_THROWAWAY_PIP_CACHE_DIRECTORY = "/tmp/pip-cache"
REBUILD_AGENT_IMAGE_FLAG = "--rebuild-image"
DEFAULT_AGENT_COMMAND = ("bash",)

# Python cache cleanup
PYTHON_CACHE_DIRECTORY_NAMES = ("__pycache__", ".pytest_cache", ".ruff_cache", ".venv", "venv")
PROTECTED_RELATIVE_DIRECTORIES = (Path("cache/logs"),)

# Run log
EXIT_CODE_PROGRAM_NOT_FOUND = 127
EXIT_CODE_SIGNAL_BASE = 128

# CLI exit codes
EXIT_CODE_SETUP_ERROR = 1
EXIT_CODE_USAGE_ERROR = 2


def toolchain_image_tag(python_version: str) -> str:
    return f"{TOOLCHAIN_IMAGE_REPOSITORY}:python-{python_version}"


def agent_image_tag(python_version: str) -> str:
    return f"{AGENT_IMAGE_REPOSITORY}:python-{python_version}"
