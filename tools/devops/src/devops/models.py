"""
Custom types shared by the devops CLI modules: dataclasses and type aliases.
"""

from __future__ import annotations

from collections.abc import Callable, Sequence
from dataclasses import dataclass
from pathlib import Path

from devops.run_log import RunLog

SubcommandFunction = Callable[[Path, Sequence[str], RunLog], int]


@dataclass(frozen=True)
class ContainerRunSettings:
    """
    What every `<binary> run` needs, whatever command it runs.
    `user_and_group_ids` is None to run as root. With `keep_pip_cache`, the pip cache
    lives in the named volume instead of the throwaway home.
    """

    container_binary: str
    repository_root: Path
    image: str
    user_and_group_ids: tuple[int, int] | None
    keep_pip_cache: bool = False


@dataclass(frozen=True)
class AgentContainerRunSettings:
    """
    What `devops agent` needs to run the agent image.
    The image's own user has the host's uid and gid, so no `--user` is passed.
    """

    container_binary: str
    repository_root: Path
    agent_image: str
    home_volume_name: str
    keep_pip_cache: bool = False
