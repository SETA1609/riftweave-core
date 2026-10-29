"""
The Python toolchain container: finding the repo, the runtime, and building docker/podman calls.
"""

from __future__ import annotations

import os
import shutil
import subprocess
from collections.abc import Callable, Sequence
from pathlib import Path

from devops.constants import (
    CONTAINER_AGENT_HOME_DIRECTORY,
    CONTAINER_AGENT_THROWAWAY_PIP_CACHE_DIRECTORY,
    CONTAINER_HOME_DIRECTORY,
    CONTAINER_KEPT_PIP_CACHE_MOUNT,
    CONTAINER_PIP_CACHE_DIRECTORY,
    CONTAINER_REPOSITORY_MOUNT,
    DEFAULT_AGENT_HOME_VOLUME_NAME,
    DEVCONTAINER_BASE_IMAGE_PATTERN,
    DOCKERFILE_PYTHON_VERSION_PATTERN,
    KEEP_PIP_CACHE_FLAG,
    PIP_CACHE_VOLUME_NAME,
    PYTHON_VERSION_FILE_NAME,
    REBUILD_AGENT_IMAGE_FLAG,
    SUPPORTED_CONTAINER_BINARIES,
    agent_image_tag,
    toolchain_image_tag,
)
from devops.exceptions import ContainerSetupError
from devops.models import AgentContainerRunSettings, ContainerRunSettings


def find_repository_root(start_directory: Path) -> Path:
    """
    Walk up from `start_directory` to the directory holding the toolchain pin and the Dockerfile.
    """
    for candidate_directory in [start_directory, *start_directory.parents]:
        has_python_pin = (candidate_directory / PYTHON_VERSION_FILE_NAME).is_file()
        has_dockerfile = (candidate_directory / "container" / "Dockerfile").is_file()
        if has_python_pin and has_dockerfile:
            return candidate_directory
    raise ContainerSetupError(
        f"no repository root above {start_directory}: "
        f"expected {PYTHON_VERSION_FILE_NAME} and container/Dockerfile"
    )


def get_dockerfile_path(repository_root: Path) -> Path:
    return repository_root / "container" / "Dockerfile"


def get_devcontainer_path(repository_root: Path) -> Path:
    return repository_root / ".devcontainer" / "devcontainer.json"


def check_python_version_matches_dockerfile(repository_root: Path) -> str:
    """
    Return the pinned Python version.
    Fail if the Dockerfile's first `ARG PYTHON_VERSION=` or the devcontainer `BASE_IMAGE` differs.
    """
    pinned_version = (
        (repository_root / PYTHON_VERSION_FILE_NAME).read_text(encoding="utf-8").strip()
    )
    dockerfile_path = get_dockerfile_path(repository_root)
    dockerfile_text = dockerfile_path.read_text(encoding="utf-8")
    version_match = DOCKERFILE_PYTHON_VERSION_PATTERN.search(dockerfile_text)
    if version_match is None:
        raise ContainerSetupError(f"{dockerfile_path} has no ARG PYTHON_VERSION= line")
    dockerfile_version = version_match.group(1)
    if dockerfile_version != pinned_version:
        raise ContainerSetupError(
            f"{PYTHON_VERSION_FILE_NAME} is {pinned_version} but "
            f"{dockerfile_path} pins {dockerfile_version}"
        )
    devcontainer_path = get_devcontainer_path(repository_root)
    if not devcontainer_path.is_file():
        raise ContainerSetupError(f"{devcontainer_path} is missing")
    base_image_match = DEVCONTAINER_BASE_IMAGE_PATTERN.search(
        devcontainer_path.read_text(encoding="utf-8")
    )
    if base_image_match is None:
        raise ContainerSetupError(
            f"{devcontainer_path} has no BASE_IMAGE riftweave:python-<version>"
        )
    devcontainer_version = base_image_match.group(1)
    if devcontainer_version != pinned_version:
        raise ContainerSetupError(
            f"{PYTHON_VERSION_FILE_NAME} is {pinned_version} but {devcontainer_path} "
            f"pins python-{devcontainer_version}"
        )
    return pinned_version


def is_container_daemon_reachable(container_binary: str) -> bool:
    if shutil.which(container_binary) is None:
        return False
    info_result = subprocess.run(
        [container_binary, "info"],
        stdin=subprocess.DEVNULL,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        check=False,
    )
    return info_result.returncode == 0


def detect_container_binary(
    check_daemon_reachable: Callable[[str], bool] = is_container_daemon_reachable,
) -> str:
    """
    Docker if its daemon answers, else podman.
    """
    for container_binary in SUPPORTED_CONTAINER_BINARIES:
        if check_daemon_reachable(container_binary):
            return container_binary
    raise ContainerSetupError("docker or podman is required and the daemon must be running")


def get_container_image(python_version: str) -> str:
    return os.environ.get("IMAGE", toolchain_image_tag(python_version))


def get_host_user_and_group_ids() -> tuple[int, int] | None:
    """
    Run as the host user so files under the repo stay theirs. `CONTAINER_AS_ROOT=1` opts out.
    """
    if os.environ.get("CONTAINER_AS_ROOT"):
        return None
    return (os.getuid(), os.getgid())


def is_container_image_present(container_binary: str, image: str) -> bool:
    inspect_result = subprocess.run(
        [container_binary, "image", "inspect", image],
        stdin=subprocess.DEVNULL,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        check=False,
    )
    return inspect_result.returncode == 0


def build_image_build_arguments(
    container_binary: str, repository_root: Path, image: str, python_version: str
) -> list[str]:
    return [
        container_binary, "build",
        "--file", str(get_dockerfile_path(repository_root)),
        "--build-arg", f"PYTHON_VERSION={python_version}",
        "--tag", image,
        str(repository_root),
    ]  # fmt: skip


def build_prune_dangling_images_arguments(container_binary: str) -> list[str]:
    return [container_binary, "image", "prune", "-f"]


def build_prune_build_cache_arguments(container_binary: str) -> list[str]:
    return [container_binary, "builder", "prune", "-f"]


def build_image_summary_arguments(container_binary: str, image: str) -> list[str]:
    return [container_binary, "image", "inspect", image, "--format", "id={{.Id}} size={{.Size}}"]


def build_container_run_arguments(
    run_settings: ContainerRunSettings,
    command_in_container: Sequence[str],
    keep_stdin_open: bool = False,
    allocate_terminal: bool = False,
) -> list[str]:
    """
    `<binary> run --rm` with the repo mounted at /src and a throwaway home.
    With `keep_pip_cache`, the pip cache lives in the named volume instead of the
    throwaway home, so it survives the run. Without it, nothing outside the repo is kept.
    """
    user_arguments: list[str] = []
    if run_settings.user_and_group_ids is not None:
        user_id, group_id = run_settings.user_and_group_ids
        user_arguments = ["--user", f"{user_id}:{group_id}"]
    terminal_arguments = (["-i"] if keep_stdin_open else []) + (["-t"] if allocate_terminal else [])
    pip_cache_arguments = ["-e", f"PIP_CACHE_DIR={CONTAINER_PIP_CACHE_DIRECTORY}"]
    if run_settings.keep_pip_cache:
        pip_cache_arguments = [
            "-e", f"PIP_CACHE_DIR={CONTAINER_KEPT_PIP_CACHE_MOUNT}",
            "-v", f"{PIP_CACHE_VOLUME_NAME}:{CONTAINER_KEPT_PIP_CACHE_MOUNT}",
        ]  # fmt: skip
    return [
        run_settings.container_binary, "run", "--rm",
        *terminal_arguments,
        *user_arguments,
        "-e", f"HOME={CONTAINER_HOME_DIRECTORY}",
        *pip_cache_arguments,
        "-v", f"{run_settings.repository_root}:{CONTAINER_REPOSITORY_MOUNT}",
        "-w", CONTAINER_REPOSITORY_MOUNT,
        run_settings.image,
        *command_in_container,
    ]  # fmt: skip


def split_keep_pip_cache_flag(extra_arguments: Sequence[str]) -> tuple[bool, list[str]]:
    """
    Take a leading `--keep-cache` off the arguments; everything after it goes to the command.
    Only a leading flag counts, so `python script.py -- --keep-cache` still reaches the program.
    """
    remaining_arguments = list(extra_arguments)
    keep_pip_cache = False
    while remaining_arguments and remaining_arguments[0] == KEEP_PIP_CACHE_FLAG:
        keep_pip_cache = True
        remaining_arguments.pop(0)
    return keep_pip_cache, remaining_arguments


def is_pip_cache_volume_present(container_binary: str) -> bool:
    inspect_result = subprocess.run(
        [container_binary, "volume", "inspect", PIP_CACHE_VOLUME_NAME],
        stdin=subprocess.DEVNULL,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        check=False,
    )
    return inspect_result.returncode == 0


def build_pip_cache_volume_create_arguments(container_binary: str) -> list[str]:
    return [container_binary, "volume", "create", PIP_CACHE_VOLUME_NAME]


def build_pip_cache_volume_remove_arguments(container_binary: str) -> list[str]:
    return [container_binary, "volume", "rm", PIP_CACHE_VOLUME_NAME]


def build_pip_cache_volume_chown_arguments(
    container_binary: str, image: str, user_and_group_ids: tuple[int, int]
) -> list[str]:
    """
    A new volume is root-owned; hand it to the host user once, as root, so pip can write.
    """
    user_id, group_id = user_and_group_ids
    return [
        container_binary, "run", "--rm",
        "--user", "0:0",
        "-v", f"{PIP_CACHE_VOLUME_NAME}:{CONTAINER_KEPT_PIP_CACHE_MOUNT}",
        "--entrypoint", "chown",
        image,
        f"{user_id}:{group_id}", CONTAINER_KEPT_PIP_CACHE_MOUNT,
    ]  # fmt: skip


def build_pip_cache_volume_size_arguments(container_binary: str, image: str) -> list[str]:
    """
    `du -sb` over the volume; prints `<bytes>\t<path>`.
    """
    return [
        container_binary, "run", "--rm",
        "-v", f"{PIP_CACHE_VOLUME_NAME}:{CONTAINER_KEPT_PIP_CACHE_MOUNT}",
        "--entrypoint", "du",
        image,
        "-sb", CONTAINER_KEPT_PIP_CACHE_MOUNT,
    ]  # fmt: skip


def measure_pip_cache_volume_size_bytes(container_binary: str, image: str) -> int | None:
    """
    Bytes in the volume, or None when it cannot be measured (no image, no volume).
    """
    size_result = subprocess.run(
        build_pip_cache_volume_size_arguments(container_binary, image),
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
        check=False,
    )
    if size_result.returncode != 0:
        return None
    try:
        return int(size_result.stdout.split()[0])
    except (IndexError, ValueError):
        return None


def get_agent_dockerfile_path(repository_root: Path) -> Path:
    return repository_root / "container" / "agent.Dockerfile"


def get_agent_container_image(python_version: str) -> str:
    return os.environ.get("AGENT_IMAGE", agent_image_tag(python_version))


def get_agent_home_volume_name() -> str:
    return os.environ.get("AGENT_HOME_VOLUME", DEFAULT_AGENT_HOME_VOLUME_NAME)


def build_agent_image_build_arguments(
    container_binary: str,
    repository_root: Path,
    agent_image: str,
    base_image: str,
    user_and_group_ids: tuple[int, int],
) -> list[str]:
    """
    Build the agent image on the toolchain image, with the agent user at the host's ids.
    """
    user_id, group_id = user_and_group_ids
    return [
        container_binary, "build",
        "--file", str(get_agent_dockerfile_path(repository_root)),
        "--build-arg", f"BASE_IMAGE={base_image}",
        "--build-arg", f"AGENT_USER_ID={user_id}",
        "--build-arg", f"AGENT_GROUP_ID={group_id}",
        "--tag", agent_image,
        str(repository_root),
    ]  # fmt: skip


def build_agent_container_run_arguments(
    agent_run_settings: AgentContainerRunSettings,
    command_in_container: Sequence[str],
    allocate_terminal: bool = False,
) -> list[str]:
    """
    `<binary> run --rm -i` of the agent image. The home volume keeps the Claude login and
    git identity; nothing from the host's home is mounted. Podman maps the host user with keep-id.
    """
    terminal_arguments = ["-i"] + (["-t"] if allocate_terminal else [])
    user_namespace_arguments = (
        ["--userns=keep-id"] if agent_run_settings.container_binary == "podman" else []
    )
    pip_cache_arguments = [
        "-e", f"PIP_CACHE_DIR={CONTAINER_AGENT_THROWAWAY_PIP_CACHE_DIRECTORY}",
    ]  # fmt: skip
    if agent_run_settings.keep_pip_cache:
        pip_cache_arguments = [
            "-e", f"PIP_CACHE_DIR={CONTAINER_KEPT_PIP_CACHE_MOUNT}",
            "-v", f"{PIP_CACHE_VOLUME_NAME}:{CONTAINER_KEPT_PIP_CACHE_MOUNT}",
        ]  # fmt: skip
    return [
        agent_run_settings.container_binary, "run", "--rm",
        *terminal_arguments,
        *user_namespace_arguments,
        "-v", f"{agent_run_settings.home_volume_name}:{CONTAINER_AGENT_HOME_DIRECTORY}",
        *pip_cache_arguments,
        "-v", f"{agent_run_settings.repository_root}:{CONTAINER_REPOSITORY_MOUNT}",
        "-w", CONTAINER_REPOSITORY_MOUNT,
        agent_run_settings.agent_image,
        *command_in_container,
    ]  # fmt: skip


def split_agent_leading_flags(extra_arguments: Sequence[str]) -> tuple[bool, bool, list[str]]:
    """
    Take leading `--keep-cache` and `--rebuild-image` off, in any order. Return
    (keep_pip_cache, rebuild_agent_image, command). Flags after the command are the command's.
    """
    remaining_arguments = list(extra_arguments)
    keep_pip_cache = False
    rebuild_agent_image = False
    while remaining_arguments and remaining_arguments[0] in (
        KEEP_PIP_CACHE_FLAG,
        REBUILD_AGENT_IMAGE_FLAG,
    ):
        leading_flag = remaining_arguments.pop(0)
        if leading_flag == KEEP_PIP_CACHE_FLAG:
            keep_pip_cache = True
        else:
            rebuild_agent_image = True
    return keep_pip_cache, rebuild_agent_image, remaining_arguments
