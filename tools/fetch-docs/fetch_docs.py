#!/usr/bin/env python3
"""Fetch upstream docs and specs into the local, gitignored cache/docs/ directory.

Agents use this to read a spec, a library's source, or a reference page in
place, so a claim can be checked and quoted exactly. Standard library only, so
it runs anywhere Python 3 does, including the agent container, which has no
curl.

Subcommands print one JSON object to stdout. Errors print a JSON object with
an error code to stderr and exit 1. Downloaded files are untrusted data: they
are size-capped, written only under the cache directory, and never executed.

    fetch_docs.py url URL
    fetch_docs.py github OWNER/REPOSITORY PATH... [--reference REF]
    fetch_docs.py list

Every subcommand takes --cache-directory PATH (default: cache/docs/ in this
repository). The reference defaults to HEAD, which GitHub resolves to the
repository's default branch, whatever it is called (main, master, ...).

Error codes: https_required, invalid_path, invalid_repository, too_large,
download_failed, invalid_github_response.
"""

from __future__ import annotations

import argparse
import dataclasses
import datetime
import hashlib
import json
import pathlib
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from collections.abc import Callable

DEFAULT_MAXIMUM_BYTE_COUNT = 20 * 1024 * 1024
DEFAULT_TIMEOUT_SECONDS = 30
# GitHub resolves HEAD to the repository's default branch.
DEFAULT_GITHUB_REFERENCE = "HEAD"
SOURCE_METADATA_FILE_NAME = "SOURCE.json"
URL_SOURCE_METADATA_SUFFIX = ".source.json"
USER_AGENT = "riftweave-fetch-docs"
REPOSITORY_NAME_PATTERN = re.compile(r"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$")
COMMIT_SHA_PATTERN = re.compile(r"^[0-9a-f]{40}$")

# Downloads a URL and returns its bytes. Tests pass a fake one.
DownloadBytes = Callable[[str], bytes]


class FetchDocsError(Exception):
    """A failure with a stable code, reported to the caller as JSON."""

    def __init__(self, error_code: str, message: str) -> None:
        super().__init__(message)
        self.error_code = error_code
        self.message = message


@dataclasses.dataclass(frozen=True)
class FetchedDocument:
    source_url: str
    cache_file_path: pathlib.Path
    byte_count: int


def find_repository_root_path() -> pathlib.Path:
    """The repository root: this file is at tools/fetch-docs/."""
    return pathlib.Path(__file__).resolve().parents[2]


def find_default_cache_directory_path() -> pathlib.Path:
    """cache/docs/ under the repository root (cache/ is gitignored)."""
    return find_repository_root_path() / "cache" / "docs"


def download_url_bytes(
    url: str,
    maximum_byte_count: int = DEFAULT_MAXIMUM_BYTE_COUNT,
    timeout_seconds: int = DEFAULT_TIMEOUT_SECONDS,
) -> bytes:
    """Download over HTTPS only; fail instead of reading past maximum_byte_count."""
    require_https_url(url)
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            # A redirect must not downgrade to plain HTTP.
            require_https_url(response.geturl())
            downloaded_bytes = response.read(maximum_byte_count + 1)
    except urllib.error.HTTPError as error:
        raise FetchDocsError("download_failed", f"{url}: HTTP {error.code}") from error
    except (urllib.error.URLError, TimeoutError) as error:
        raise FetchDocsError("download_failed", f"{url}: {error}") from error
    if len(downloaded_bytes) > maximum_byte_count:
        raise FetchDocsError("too_large", f"{url}: more than {maximum_byte_count} bytes")
    return downloaded_bytes


def build_safe_relative_path(untrusted_path: str) -> pathlib.PurePosixPath:
    """A relative path with no '..', no absolute root, and no empty parts."""
    if not untrusted_path or untrusted_path.startswith("/") or "\\" in untrusted_path:
        raise FetchDocsError("invalid_path", f"not a safe relative path: {untrusted_path!r}")
    path_parts = untrusted_path.split("/")
    if any(part in ("", ".", "..") for part in path_parts):
        raise FetchDocsError("invalid_path", f"not a safe relative path: {untrusted_path!r}")
    return pathlib.PurePosixPath(*path_parts)


def build_url_cache_file_path(cache_directory_path: pathlib.Path, url: str) -> pathlib.Path:
    """cache/docs/url/<host>/<path>; a path ending in '/' is saved as index.html."""
    parsed_url = urllib.parse.urlsplit(url)
    url_path = parsed_url.path or "/"
    if url_path.endswith("/"):
        url_path += "index.html"
    if parsed_url.query:
        # Different queries must not overwrite each other.
        query_hash = hashlib.sha256(parsed_url.query.encode()).hexdigest()[:8]
        url_path += f"_{query_hash}"
    relative_path = build_safe_relative_path(f"{parsed_url.hostname}{url_path}")
    return cache_directory_path / "url" / relative_path


def build_github_raw_file_url(owner_and_repository: str, commit_sha: str, file_path: str) -> str:
    return f"https://raw.githubusercontent.com/{owner_and_repository}/{commit_sha}/{file_path}"


def resolve_github_reference_to_commit_sha(
    owner_and_repository: str, reference: str, download_bytes: DownloadBytes
) -> str:
    """Ask the GitHub API which commit a branch, tag, or SHA names right now."""
    require_valid_repository_name(owner_and_repository)
    quoted_reference = urllib.parse.quote(reference, safe="")
    api_url = f"https://api.github.com/repos/{owner_and_repository}/commits/{quoted_reference}"
    try:
        commit_sha = json.loads(download_bytes(api_url)).get("sha", "")
    except (json.JSONDecodeError, UnicodeDecodeError, AttributeError) as error:
        raise FetchDocsError("invalid_github_response", f"{api_url}: {error}") from error
    if not isinstance(commit_sha, str) or not COMMIT_SHA_PATTERN.match(commit_sha):
        raise FetchDocsError("invalid_github_response", f"{api_url}: no commit SHA in the reply")
    return commit_sha


def fetch_url_document(
    url: str, cache_directory_path: pathlib.Path, download_bytes: DownloadBytes
) -> FetchedDocument:
    """Save one URL and write <file>.source.json next to it."""
    require_https_url(url)
    cache_file_path = build_url_cache_file_path(cache_directory_path, url)
    downloaded_bytes = download_bytes(url)
    write_file_bytes(cache_file_path, downloaded_bytes)
    source_metadata = {
        "url": url,
        "fetched_at_utc": format_current_utc_time(),
        "byte_count": len(downloaded_bytes),
    }
    source_file_path = cache_file_path.with_name(cache_file_path.name + URL_SOURCE_METADATA_SUFFIX)
    write_json_file(source_file_path, source_metadata)
    return FetchedDocument(url, cache_file_path, len(downloaded_bytes))


def fetch_github_documents(
    owner_and_repository: str,
    file_paths: list[str],
    reference: str,
    cache_directory_path: pathlib.Path,
    download_bytes: DownloadBytes,
) -> dict:
    """Save files under cache/docs/github/<owner>/<repository>/<commit>/ plus SOURCE.json."""
    require_valid_repository_name(owner_and_repository)
    # Check every path before downloading anything.
    safe_file_paths = [build_safe_relative_path(file_path) for file_path in file_paths]
    commit_sha = resolve_github_reference_to_commit_sha(
        owner_and_repository, reference, download_bytes
    )
    commit_directory_path = (
        cache_directory_path
        / "github"
        / pathlib.PurePosixPath(owner_and_repository)
        / commit_sha[:12]
    )
    source_file_path = commit_directory_path / SOURCE_METADATA_FILE_NAME
    file_entries_by_path = {}
    if source_file_path.is_file():
        # Earlier fetches from the same commit stay listed.
        for file_entry in json.loads(source_file_path.read_text()).get("files", []):
            file_entries_by_path[file_entry["path"]] = file_entry
    fetched_files = []
    for safe_file_path in safe_file_paths:
        raw_url = build_github_raw_file_url(owner_and_repository, commit_sha, str(safe_file_path))
        downloaded_bytes = download_bytes(raw_url)
        cache_file_path = commit_directory_path / safe_file_path
        write_file_bytes(cache_file_path, downloaded_bytes)
        file_entry = {
            "path": str(safe_file_path),
            "url": raw_url,
            "byte_count": len(downloaded_bytes),
        }
        file_entries_by_path[file_entry["path"]] = file_entry
        fetched_files.append({**file_entry, "cache_file_path": str(cache_file_path)})
    source_metadata = {
        "repository": owner_and_repository,
        "reference": reference,
        "commit": commit_sha,
        "fetched_at_utc": format_current_utc_time(),
        "files": [file_entries_by_path[path] for path in sorted(file_entries_by_path)],
    }
    write_json_file(source_file_path, source_metadata)
    return {**source_metadata, "files": fetched_files}


def list_cached_documents(cache_directory_path: pathlib.Path) -> list[dict]:
    """Every SOURCE.json and *.source.json under the cache, sorted by path."""
    if not cache_directory_path.is_dir():
        return []
    source_file_paths = [
        *cache_directory_path.rglob(SOURCE_METADATA_FILE_NAME),
        *cache_directory_path.rglob("*" + URL_SOURCE_METADATA_SUFFIX),
    ]
    cached_documents = []
    for source_file_path in source_file_paths:
        source_metadata = json.loads(source_file_path.read_text())
        source_metadata["source_file_path"] = str(
            source_file_path.relative_to(cache_directory_path)
        )
        cached_documents.append(source_metadata)
    return sorted(cached_documents, key=lambda document: document["source_file_path"])


def build_argument_parser() -> argparse.ArgumentParser:
    shared_options = argparse.ArgumentParser(add_help=False)
    shared_options.add_argument(
        "--cache-directory",
        type=pathlib.Path,
        default=None,
        help="where to save files (default: cache/docs/ in this repository)",
    )
    argument_parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    subcommands = argument_parser.add_subparsers(dest="command", required=True)
    url_command = subcommands.add_parser("url", parents=[shared_options], help="fetch one URL")
    url_command.add_argument("url")
    github_command = subcommands.add_parser(
        "github", parents=[shared_options], help="fetch files from a GitHub repository"
    )
    github_command.add_argument("repository", help="OWNER/REPOSITORY")
    github_command.add_argument("paths", nargs="+", help="file paths inside the repository")
    github_command.add_argument(
        "--reference",
        default=DEFAULT_GITHUB_REFERENCE,
        help="branch, tag, or commit (default: HEAD)",
    )
    subcommands.add_parser("list", parents=[shared_options], help="list what is in the cache")
    return argument_parser


def main(
    argument_list: list[str] | None = None, download_bytes: DownloadBytes | None = None
) -> int:
    arguments = build_argument_parser().parse_args(argument_list)
    download = download_bytes or download_url_bytes
    cache_directory_path = arguments.cache_directory or find_default_cache_directory_path()
    try:
        if arguments.command == "url":
            document = fetch_url_document(arguments.url, cache_directory_path, download)
            result = {
                "url": document.source_url,
                "cache_file_path": str(document.cache_file_path),
                "byte_count": document.byte_count,
            }
        elif arguments.command == "github":
            result = fetch_github_documents(
                arguments.repository,
                arguments.paths,
                arguments.reference,
                cache_directory_path,
                download,
            )
        else:
            result = {"documents": list_cached_documents(cache_directory_path)}
    except FetchDocsError as error:
        error_report = {"error": {"code": error.error_code, "message": error.message}}
        print(json.dumps(error_report, sort_keys=True), file=sys.stderr)
        return 1
    print(json.dumps(result, indent=2, sort_keys=True))
    return 0


def require_https_url(url: str) -> None:
    if urllib.parse.urlsplit(url).scheme != "https":
        raise FetchDocsError("https_required", f"only https URLs are fetched: {url}")


def require_valid_repository_name(owner_and_repository: str) -> None:
    repository_parts = owner_and_repository.split("/")
    if not REPOSITORY_NAME_PATTERN.match(owner_and_repository) or any(
        part in (".", "..") for part in repository_parts
    ):
        raise FetchDocsError(
            "invalid_repository", f"expected OWNER/REPOSITORY, got {owner_and_repository!r}"
        )


def format_current_utc_time() -> str:
    return datetime.datetime.now(datetime.UTC).isoformat(timespec="seconds")


def write_file_bytes(file_path: pathlib.Path, file_bytes: bytes) -> None:
    file_path.parent.mkdir(parents=True, exist_ok=True)
    file_path.write_bytes(file_bytes)


def write_json_file(file_path: pathlib.Path, json_value: dict) -> None:
    file_path.parent.mkdir(parents=True, exist_ok=True)
    file_path.write_text(json.dumps(json_value, indent=2, sort_keys=True) + "\n")


if __name__ == "__main__":
    raise SystemExit(main())
