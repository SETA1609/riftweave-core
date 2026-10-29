"""
End-of-work cleanup of Python caches inside the repo.
Only `devops clean` uses this, and only when the user asks for it or says the work is finished.
"""

from __future__ import annotations

import os
import shutil
from collections.abc import Sequence
from pathlib import Path

from devops.constants import (
    PROTECTED_RELATIVE_DIRECTORIES,
    PYTHON_CACHE_DIRECTORY_NAMES,
)
from devops.exceptions import CacheCleanupRefusedError


def _relative_to_repository(repository_root: Path, candidate_path: Path) -> Path | None:
    try:
        return candidate_path.relative_to(repository_root)
    except ValueError:
        return None


def _is_protected_relative_path(relative_path: Path) -> bool:
    return any(
        relative_path == protected_directory or relative_path.is_relative_to(protected_directory)
        for protected_directory in PROTECTED_RELATIVE_DIRECTORIES
    )


def list_default_cache_paths(repository_root: Path) -> list[Path]:
    """
    Cache directories under the repo: `__pycache__`, pytest, ruff, and virtualenvs.
    Does not follow symlinks, descend into `.git`, or enter `cache/logs/`.
    """
    found_paths: list[Path] = []
    for walk_root, directory_names, _ in os.walk(repository_root, followlinks=False):
        walk_path = Path(walk_root)
        directories_to_walk: list[str] = []
        for directory_name in directory_names:
            candidate_path = walk_path / directory_name
            relative_path = _relative_to_repository(repository_root, candidate_path)
            if relative_path is None or _is_protected_relative_path(relative_path):
                continue
            if directory_name == ".git":
                continue
            if directory_name in PYTHON_CACHE_DIRECTORY_NAMES:
                found_paths.append(candidate_path)
                continue
            directories_to_walk.append(directory_name)
        directory_names[:] = directories_to_walk
    found_paths.sort(key=lambda path: path.relative_to(repository_root).as_posix())
    return found_paths


def select_cache_paths_to_remove(
    repository_root: Path, candidate_paths: Sequence[Path]
) -> list[Path]:
    """
    Return the candidates that exist, in order. Refuse the whole set if any is unsafe.
    Unsafe: outside the repo (after resolving `..` and symlinks in parent directories), the
    repo root itself, a symlink, or inside `cache/logs/`.
    """
    resolved_root = repository_root.resolve()
    selected_paths: list[Path] = []
    for candidate_path in candidate_paths:
        # Resolve the parent but keep the last component, so a symlinked cache dir is caught
        # below instead of being followed. `.` and `..` have no last component to keep.
        if candidate_path.name in ("", ".", ".."):
            resolved_candidate = candidate_path.resolve()
        else:
            resolved_candidate = candidate_path.parent.resolve() / candidate_path.name
        if (
            not resolved_candidate.is_relative_to(resolved_root)
            or resolved_candidate == resolved_root
        ):
            raise CacheCleanupRefusedError(f"refusing {candidate_path}: not inside {resolved_root}")
        relative_candidate = resolved_candidate.relative_to(resolved_root)
        for protected_directory in PROTECTED_RELATIVE_DIRECTORIES:
            if relative_candidate.is_relative_to(protected_directory):
                raise CacheCleanupRefusedError(
                    f"refusing {candidate_path}: {protected_directory}/ is never cleaned"
                )
        if resolved_candidate.is_symlink():
            raise CacheCleanupRefusedError(f"refusing {candidate_path}: it is a symlink")
        if resolved_candidate.exists():
            selected_paths.append(resolved_candidate)
    return selected_paths


def measure_directory_size_bytes(directory_path: Path) -> int:
    """
    Sum of file sizes under `directory_path`, not following symlinks.
    """
    total_bytes = 0
    for walk_root, _, file_names in os.walk(directory_path):
        for file_name in file_names:
            total_bytes += (Path(walk_root) / file_name).lstat().st_size
    return total_bytes


def remove_directory_tree(directory_path: Path) -> None:
    shutil.rmtree(directory_path)
