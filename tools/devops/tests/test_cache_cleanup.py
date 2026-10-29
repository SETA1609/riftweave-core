from pathlib import Path

import pytest

from devops.cache_cleanup import (
    list_default_cache_paths,
    measure_directory_size_bytes,
    select_cache_paths_to_remove,
)
from devops.exceptions import CacheCleanupRefusedError


def write_file_of_size(file_path: Path, size_bytes: int) -> None:
    file_path.parent.mkdir(parents=True, exist_ok=True)
    file_path.write_bytes(b"x" * size_bytes)


def test_list_default_cache_paths_finds_python_caches_and_skips_logs(tmp_path):
    (tmp_path / "ruleset" / "scripts" / "__pycache__").mkdir(parents=True)
    (tmp_path / ".ruff_cache").mkdir()
    (tmp_path / ".venv").mkdir()
    (tmp_path / "cache" / "logs" / "__pycache__").mkdir(parents=True)
    (tmp_path / ".git" / "__pycache__").mkdir(parents=True)
    assert list_default_cache_paths(tmp_path) == [
        tmp_path / ".ruff_cache",
        tmp_path / ".venv",
        tmp_path / "ruleset" / "scripts" / "__pycache__",
    ]


def test_select_cache_paths_to_remove_returns_the_existing_default_paths(tmp_path):
    (tmp_path / ".pytest_cache").mkdir()
    (tmp_path / ".ruff_cache").mkdir()
    selected_paths = select_cache_paths_to_remove(tmp_path, list_default_cache_paths(tmp_path))
    assert selected_paths == [
        tmp_path.resolve() / ".pytest_cache",
        tmp_path.resolve() / ".ruff_cache",
    ]


def test_select_cache_paths_to_remove_skips_paths_that_do_not_exist(tmp_path):
    (tmp_path / ".ruff_cache").mkdir()
    selected_paths = select_cache_paths_to_remove(
        tmp_path, [tmp_path / ".pytest_cache", tmp_path / ".ruff_cache"]
    )
    assert selected_paths == [tmp_path.resolve() / ".ruff_cache"]


def test_select_cache_paths_to_remove_refuses_a_path_outside_the_repository(tmp_path):
    repository_root = tmp_path / "repo"
    repository_root.mkdir()
    (tmp_path / ".venv").mkdir()
    with pytest.raises(CacheCleanupRefusedError, match="not inside"):
        select_cache_paths_to_remove(repository_root, [repository_root / ".." / ".venv"])


def test_select_cache_paths_to_remove_refuses_the_repository_root(tmp_path):
    with pytest.raises(CacheCleanupRefusedError, match="not inside"):
        select_cache_paths_to_remove(tmp_path, [tmp_path / "sub" / ".."])


def test_select_cache_paths_to_remove_refuses_a_path_reached_through_a_symlinked_parent(tmp_path):
    repository_root = tmp_path / "repo"
    outside_directory = tmp_path / "outside"
    (outside_directory / ".venv").mkdir(parents=True)
    repository_root.mkdir()
    (repository_root / "linked").symlink_to(outside_directory)
    with pytest.raises(CacheCleanupRefusedError, match="not inside"):
        select_cache_paths_to_remove(repository_root, [repository_root / "linked" / ".venv"])


def test_select_cache_paths_to_remove_refuses_a_symlink(tmp_path):
    (tmp_path / "real-cache").mkdir()
    (tmp_path / ".venv").symlink_to(tmp_path / "real-cache")
    with pytest.raises(CacheCleanupRefusedError, match="symlink"):
        select_cache_paths_to_remove(tmp_path, [tmp_path / ".venv"])


@pytest.mark.parametrize("protected_path", ["cache/logs", "cache/logs/x"])
def test_select_cache_paths_to_remove_refuses_protected_directories(tmp_path, protected_path):
    (tmp_path / protected_path).mkdir(parents=True)
    with pytest.raises(CacheCleanupRefusedError, match="never cleaned"):
        select_cache_paths_to_remove(tmp_path, [tmp_path / protected_path])


def test_select_cache_paths_to_remove_refuses_the_whole_set_when_one_path_is_unsafe(tmp_path):
    (tmp_path / ".ruff_cache").mkdir()
    (tmp_path / "cache" / "logs").mkdir(parents=True)
    with pytest.raises(CacheCleanupRefusedError):
        select_cache_paths_to_remove(
            tmp_path, [tmp_path / ".ruff_cache", tmp_path / "cache" / "logs"]
        )


def test_measure_directory_size_bytes_sums_files_without_following_symlinks(tmp_path):
    write_file_of_size(tmp_path / "cache" / "a", 10)
    write_file_of_size(tmp_path / "cache" / "nested" / "b", 32)
    write_file_of_size(tmp_path / "big-outside", 4096)
    (tmp_path / "cache" / "link").symlink_to(tmp_path / "big-outside")
    link_size_bytes = (tmp_path / "cache" / "link").lstat().st_size
    assert measure_directory_size_bytes(tmp_path / "cache") == 42 + link_size_bytes
