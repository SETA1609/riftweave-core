from pathlib import Path

import pytest

from devops.constants import (
    PIP_CACHE_VOLUME_NAME,
    PYTHON_VERSION_FILE_NAME,
    toolchain_image_tag,
)
from devops.container import (
    build_agent_container_run_arguments,
    build_agent_image_build_arguments,
    build_container_run_arguments,
    build_image_build_arguments,
    build_pip_cache_volume_chown_arguments,
    check_python_version_matches_dockerfile,
    detect_container_binary,
    find_repository_root,
    split_agent_leading_flags,
    split_keep_pip_cache_flag,
)
from devops.exceptions import ContainerSetupError
from devops.models import AgentContainerRunSettings, ContainerRunSettings


def write_fake_repository(
    root: Path,
    pinned_version: str,
    dockerfile_version: str | None,
    devcontainer_version: str | None = None,
) -> Path:
    (root / "container").mkdir(parents=True)
    (root / PYTHON_VERSION_FILE_NAME).write_text(f"{pinned_version}\n")
    dockerfile_lines = ["FROM python:3.13-slim-bookworm"]
    if dockerfile_version is not None:
        dockerfile_lines.append(f"ARG PYTHON_VERSION={dockerfile_version}")
        dockerfile_lines.append("ARG PYTHON_VERSION=9.9.9")
    (root / "container" / "Dockerfile").write_text("\n".join(dockerfile_lines) + "\n")
    if dockerfile_version is not None or devcontainer_version is not None:
        version_in_devcontainer = (
            devcontainer_version if devcontainer_version is not None else dockerfile_version
        )
        devcontainer_directory = root / ".devcontainer"
        devcontainer_directory.mkdir()
        (devcontainer_directory / "devcontainer.json").write_text(
            "{\n"
            '  "build": {\n'
            '    "args": {\n'
            f'      "BASE_IMAGE": "riftweave:python-{version_in_devcontainer}"\n'
            "    }\n"
            "  }\n"
            "}\n"
        )
    return root


def test_toolchain_image_tag_uses_the_pinned_version():
    assert toolchain_image_tag("3.13.18") == "riftweave:python-3.13.18"


def test_check_python_version_matches_dockerfile_returns_the_pinned_version(tmp_path):
    root = write_fake_repository(tmp_path, "3.13.18", "3.13.18")
    assert check_python_version_matches_dockerfile(root) == "3.13.18"


def test_check_python_version_matches_dockerfile_rejects_a_mismatch(tmp_path):
    root = write_fake_repository(tmp_path, "3.13.18", "3.12.0")
    expected_message = r"\.toolchain-python-version is 3\.13\.18 but .* pins 3\.12\.0"
    with pytest.raises(ContainerSetupError, match=expected_message):
        check_python_version_matches_dockerfile(root)


def test_check_python_version_matches_dockerfile_rejects_a_dockerfile_without_the_arg(tmp_path):
    root = write_fake_repository(tmp_path, "3.13.18", None)
    with pytest.raises(ContainerSetupError, match="no ARG PYTHON_VERSION="):
        check_python_version_matches_dockerfile(root)


def test_check_python_version_matches_dockerfile_rejects_a_devcontainer_mismatch(tmp_path):
    root = write_fake_repository(tmp_path, "3.13.18", "3.13.18", devcontainer_version="3.12.0")
    with pytest.raises(ContainerSetupError, match=r"devcontainer\.json"):
        check_python_version_matches_dockerfile(root)


def test_find_repository_root_walks_up_to_the_directory_with_the_toolchain_pin(tmp_path):
    root = write_fake_repository(tmp_path, "3.13.18", "3.13.18")
    nested_directory = root / "tools" / "devops" / "src" / "devops"
    nested_directory.mkdir(parents=True)
    (nested_directory.parents[1] / ".python-version").write_text("3.13\n")
    assert find_repository_root(nested_directory) == root


def test_find_repository_root_fails_outside_the_repository(tmp_path):
    with pytest.raises(ContainerSetupError, match="repository root"):
        find_repository_root(tmp_path)


def test_detect_container_binary_prefers_docker():
    assert detect_container_binary(lambda binary: True) == "docker"


def test_detect_container_binary_falls_back_to_podman():
    assert detect_container_binary(lambda binary: binary == "podman") == "podman"


def test_detect_container_binary_fails_when_no_daemon_is_reachable():
    with pytest.raises(ContainerSetupError, match="docker or podman"):
        detect_container_binary(lambda binary: False)


def test_build_container_run_arguments_mounts_the_repository_as_the_host_user():
    arguments = build_container_run_arguments(
        ContainerRunSettings(
            container_binary="docker",
            repository_root=Path("/home/user/project"),
            image="riftweave:python-3.13.18",
            user_and_group_ids=(1000, 1000),
        ),
        ["python", "ruleset/scripts/validate.py"],
    )
    assert arguments == [
        "docker", "run", "--rm",
        "--user", "1000:1000",
        "-e", "HOME=/tmp/python-home",
        "-e", "PIP_CACHE_DIR=/tmp/python-home/pip-cache",
        "-v", "/home/user/project:/src",
        "-w", "/src",
        "riftweave:python-3.13.18",
        "python", "ruleset/scripts/validate.py",
    ]  # fmt: skip


def test_build_container_run_arguments_runs_as_root_without_user_ids():
    arguments = build_container_run_arguments(
        ContainerRunSettings(
            container_binary="podman",
            repository_root=Path("/repo"),
            image="image:tag",
            user_and_group_ids=None,
        ),
        ["python", "--version"],
    )
    assert "--user" not in arguments
    assert arguments[:3] == ["podman", "run", "--rm"]


def test_build_image_build_arguments_passes_the_pinned_version():
    arguments = build_image_build_arguments(
        container_binary="docker",
        repository_root=Path("/repo"),
        image="riftweave:python-3.13.18",
        python_version="3.13.18",
    )
    assert arguments == [
        "docker", "build",
        "--file", "/repo/container/Dockerfile",
        "--build-arg", "PYTHON_VERSION=3.13.18",
        "--tag", "riftweave:python-3.13.18",
        "/repo",
    ]  # fmt: skip


def test_build_container_run_arguments_keeps_the_pip_cache_in_the_named_volume():
    arguments = build_container_run_arguments(
        ContainerRunSettings(
            container_binary="docker",
            repository_root=Path("/repo"),
            image="image:tag",
            user_and_group_ids=(1000, 1000),
            keep_pip_cache=True,
        ),
        ["python", "ruleset/scripts/validate.py"],
    )
    assert arguments == [
        "docker", "run", "--rm",
        "--user", "1000:1000",
        "-e", "HOME=/tmp/python-home",
        "-e", "PIP_CACHE_DIR=/pip-cache",
        "-v", "riftweave-pip-cache:/pip-cache",
        "-v", "/repo:/src",
        "-w", "/src",
        "image:tag",
        "python", "ruleset/scripts/validate.py",
    ]  # fmt: skip


def test_build_container_run_arguments_mounts_no_volume_without_keep_pip_cache():
    arguments = build_container_run_arguments(
        ContainerRunSettings(
            container_binary="docker",
            repository_root=Path("/repo"),
            image="image:tag",
            user_and_group_ids=None,
        ),
        ["python", "ruleset/scripts/validate.py"],
    )
    assert not any(PIP_CACHE_VOLUME_NAME in argument for argument in arguments)


@pytest.mark.parametrize(
    ("allocate_terminal", "expected_terminal_arguments"),
    [(True, ["-i", "-t"]), (False, ["-i"])],
)
def test_build_container_run_arguments_keeps_stdin_open_for_the_shell(
    allocate_terminal, expected_terminal_arguments
):
    arguments = build_container_run_arguments(
        ContainerRunSettings(
            container_binary="docker",
            repository_root=Path("/repo"),
            image="image:tag",
            user_and_group_ids=None,
        ),
        ["bash"],
        keep_stdin_open=True,
        allocate_terminal=allocate_terminal,
    )
    expected_prefix = ["docker", "run", "--rm", *expected_terminal_arguments, "-e"]
    assert arguments[: len(expected_prefix)] == expected_prefix


@pytest.mark.parametrize(
    ("extra_arguments", "expected_keep", "expected_rest"),
    [
        ([], False, []),
        (["--keep-cache"], True, []),
        (
            ["--keep-cache", "ruleset/scripts/test_validate.py"],
            True,
            ["ruleset/scripts/test_validate.py"],
        ),
        (["--keep-cache", "--keep-cache", "-v"], True, ["-v"]),
        (["script.py", "--", "--keep-cache"], False, ["script.py", "--", "--keep-cache"]),
    ],
)
def test_split_keep_pip_cache_flag_takes_only_a_leading_flag(
    extra_arguments, expected_keep, expected_rest
):
    assert split_keep_pip_cache_flag(extra_arguments) == (expected_keep, expected_rest)


def test_build_pip_cache_volume_chown_arguments_runs_as_root_and_hands_over_the_volume():
    arguments = build_pip_cache_volume_chown_arguments("docker", "image:tag", (1000, 1001))
    assert arguments == [
        "docker", "run", "--rm",
        "--user", "0:0",
        "-v", "riftweave-pip-cache:/pip-cache",
        "--entrypoint", "chown",
        "image:tag",
        "1000:1001", "/pip-cache",
    ]  # fmt: skip


def build_agent_run_settings(**overrides) -> AgentContainerRunSettings:
    settings_fields = {
        "container_binary": "docker",
        "repository_root": Path("/home/user/project"),
        "agent_image": "riftweave-agent:python-3.13.18",
        "home_volume_name": "personal-agent-home",
    }
    settings_fields.update(overrides)
    return AgentContainerRunSettings(**settings_fields)


def test_build_agent_image_build_arguments_bases_on_the_toolchain_image_with_host_ids():
    arguments = build_agent_image_build_arguments(
        "docker",
        Path("/home/user/project"),
        "riftweave-agent:python-3.13.18",
        "riftweave:python-3.13.18",
        (1000, 1001),
    )
    assert arguments == [
        "docker", "build",
        "--file", "/home/user/project/container/agent.Dockerfile",
        "--build-arg", "BASE_IMAGE=riftweave:python-3.13.18",
        "--build-arg", "AGENT_USER_ID=1000",
        "--build-arg", "AGENT_GROUP_ID=1001",
        "--tag", "riftweave-agent:python-3.13.18",
        "/home/user/project",
    ]  # fmt: skip


def test_build_agent_container_run_arguments_mounts_the_home_volume_and_not_the_host_home():
    arguments = build_agent_container_run_arguments(build_agent_run_settings(), ["claude"])
    assert arguments == [
        "docker", "run", "--rm", "-i",
        "-v", "personal-agent-home:/home/agent",
        "-e", "PIP_CACHE_DIR=/tmp/pip-cache",
        "-v", "/home/user/project:/src",
        "-w", "/src",
        "riftweave-agent:python-3.13.18",
        "claude",
    ]  # fmt: skip
    assert not any(
        ".claude" in argument or argument.startswith("/home/user/.") for argument in arguments
    )


def test_build_agent_container_run_arguments_keeps_the_pip_cache_and_maps_podman_users():
    arguments = build_agent_container_run_arguments(
        build_agent_run_settings(container_binary="podman", keep_pip_cache=True),
        ["bash"],
        allocate_terminal=True,
    )
    assert arguments[:6] == ["podman", "run", "--rm", "-i", "-t", "--userns=keep-id"]
    assert "PIP_CACHE_DIR=/pip-cache" in arguments
    assert f"{PIP_CACHE_VOLUME_NAME}:/pip-cache" in arguments


@pytest.mark.parametrize(
    ("extra_arguments", "expected_split"),
    [
        ([], (False, False, [])),
        (["claude"], (False, False, ["claude"])),
        (["--rebuild-image", "--keep-cache", "bash"], (True, True, ["bash"])),
        (["claude", "--keep-cache"], (False, False, ["claude", "--keep-cache"])),
    ],
)
def test_split_agent_leading_flags_takes_only_leading_flags(extra_arguments, expected_split):
    assert split_agent_leading_flags(extra_arguments) == expected_split
