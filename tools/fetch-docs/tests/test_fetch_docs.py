"""Offline tests for fetch_docs.py: every download goes through a fake."""

import importlib.util
import json
import pathlib
import sys

import pytest

SCRIPT_PATH = pathlib.Path(__file__).resolve().parents[1] / "fetch_docs.py"
_module_spec = importlib.util.spec_from_file_location("fetch_docs", SCRIPT_PATH)
fetch_docs = importlib.util.module_from_spec(_module_spec)
# dataclasses looks the module up by name, so register it before running it.
sys.modules["fetch_docs"] = fetch_docs
_module_spec.loader.exec_module(fetch_docs)

COMMIT_SHA = "70304594e73a0123456789abcdef0123456789ab"
SCHEMA_SPEC_MAIN_COMMIT_URL = (
    "https://api.github.com/repos/json-schema-org/json-schema-spec/commits/main"
)


class FakeDownloader:
    """Serves fixed bytes per URL and records every URL asked for."""

    def __init__(self, bytes_by_url: dict[str, bytes]) -> None:
        self.bytes_by_url = bytes_by_url
        self.requested_urls: list[str] = []

    def __call__(self, url: str) -> bytes:
        self.requested_urls.append(url)
        if url not in self.bytes_by_url:
            raise fetch_docs.FetchDocsError("download_failed", f"not found: {url}")
        return self.bytes_by_url[url]


def github_commit_response(commit_sha: str = COMMIT_SHA) -> bytes:
    return json.dumps({"sha": commit_sha}).encode()


def test_repository_root_is_two_levels_above_the_tool_directory():
    assert (fetch_docs.find_repository_root_path() / "tools" / "fetch-docs").is_dir()


def test_default_cache_directory_is_under_the_gitignored_cache_directory():
    expected = fetch_docs.find_repository_root_path() / "cache" / "docs"
    assert fetch_docs.find_default_cache_directory_path() == expected


def test_safe_relative_path_keeps_a_normal_relative_path():
    safe_path = fetch_docs.build_safe_relative_path("ruleset/schemas/schema.json")
    assert str(safe_path) == "ruleset/schemas/schema.json"


@pytest.mark.parametrize("untrusted_path", ["../escape.txt", "chapters/../../x", "/etc/passwd", ""])
def test_safe_relative_path_rejects_escapes_and_empty_paths(untrusted_path):
    with pytest.raises(fetch_docs.FetchDocsError) as error:
        fetch_docs.build_safe_relative_path(untrusted_path)
    assert error.value.error_code == "invalid_path"


def test_url_cache_file_path_mirrors_host_and_path(tmp_path):
    cache_file_path = fetch_docs.build_url_cache_file_path(
        tmp_path, "https://wiki.libsdl.org/SDL3/SDL_CreateWindow"
    )
    assert cache_file_path == tmp_path / "url" / "wiki.libsdl.org" / "SDL3" / "SDL_CreateWindow"


def test_url_cache_file_path_saves_a_directory_url_as_index_html(tmp_path):
    cache_file_path = fetch_docs.build_url_cache_file_path(tmp_path, "https://ziglang.org/learn/")
    assert cache_file_path == tmp_path / "url" / "ziglang.org" / "learn" / "index.html"


def test_download_refuses_plain_http_before_any_network_access():
    with pytest.raises(fetch_docs.FetchDocsError) as error:
        fetch_docs.download_url_bytes("http://example.com/page")
    assert error.value.error_code == "https_required"


def test_github_raw_file_url_points_at_the_exact_commit():
    raw_url = fetch_docs.build_github_raw_file_url(
        "json-schema-org/json-schema-spec", COMMIT_SHA, "jsonschema-validation.xml"
    )
    assert raw_url == (
        f"https://raw.githubusercontent.com/json-schema-org/json-schema-spec/{COMMIT_SHA}/jsonschema-validation.xml"
    )


def test_resolving_a_reference_asks_the_github_commits_api():
    downloader = FakeDownloader({SCHEMA_SPEC_MAIN_COMMIT_URL: github_commit_response()})
    commit_sha = fetch_docs.resolve_github_reference_to_commit_sha(
        "json-schema-org/json-schema-spec", "main", downloader
    )
    assert commit_sha == COMMIT_SHA


@pytest.mark.parametrize("owner_and_repository", ["no-slash", "too/many/slashes", "../x", "a/"])
def test_resolving_rejects_a_malformed_repository_name(owner_and_repository):
    with pytest.raises(fetch_docs.FetchDocsError) as error:
        fetch_docs.resolve_github_reference_to_commit_sha(
            owner_and_repository, "main", FakeDownloader({})
        )
    assert error.value.error_code == "invalid_repository"


def test_resolving_rejects_a_response_without_a_commit_sha():
    downloader = FakeDownloader(
        {"https://api.github.com/repos/owner/repository/commits/main": b"{}"}
    )
    with pytest.raises(fetch_docs.FetchDocsError) as error:
        fetch_docs.resolve_github_reference_to_commit_sha("owner/repository", "main", downloader)
    assert error.value.error_code == "invalid_github_response"


def test_fetching_a_url_saves_the_bytes_and_a_source_file(tmp_path):
    url = "https://wiki.libsdl.org/SDL3/SDL_CreateWindow"
    downloader = FakeDownloader({url: b"page text"})
    document = fetch_docs.fetch_url_document(url, tmp_path, downloader)
    assert document.source_url == url
    assert document.byte_count == len(b"page text")
    assert document.cache_file_path.read_bytes() == b"page text"
    source_file_path = document.cache_file_path.with_name(
        document.cache_file_path.name + ".source.json"
    )
    source_metadata = json.loads(source_file_path.read_text())
    assert source_metadata["url"] == url
    assert "fetched_at_utc" in source_metadata


def test_fetching_a_plain_http_url_is_refused(tmp_path):
    with pytest.raises(fetch_docs.FetchDocsError) as error:
        fetch_docs.fetch_url_document("http://example.com/", tmp_path, FakeDownloader({}))
    assert error.value.error_code == "https_required"


def test_fetching_github_files_saves_them_under_the_commit_with_source_metadata(tmp_path):
    raw_url = f"https://raw.githubusercontent.com/owner/repository/{COMMIT_SHA}/docs/guide.md"
    downloader = FakeDownloader(
        {
            "https://api.github.com/repos/owner/repository/commits/v1.0": github_commit_response(),
            raw_url: b"# Guide",
        }
    )
    summary = fetch_docs.fetch_github_documents(
        "owner/repository", ["docs/guide.md"], "v1.0", tmp_path, downloader
    )
    commit_directory_path = tmp_path / "github" / "owner" / "repository" / COMMIT_SHA[:12]
    assert (commit_directory_path / "docs" / "guide.md").read_bytes() == b"# Guide"
    source_metadata = json.loads((commit_directory_path / "SOURCE.json").read_text())
    assert source_metadata["repository"] == "owner/repository"
    assert source_metadata["reference"] == "v1.0"
    assert source_metadata["commit"] == COMMIT_SHA
    assert source_metadata["files"][0]["path"] == "docs/guide.md"
    assert source_metadata["files"][0]["url"] == raw_url
    assert summary["commit"] == COMMIT_SHA
    assert summary["files"][0]["cache_file_path"] == str(
        commit_directory_path / "docs" / "guide.md"
    )


def test_listing_the_cache_returns_every_source_file_sorted_by_path(tmp_path):
    url = "https://wiki.libsdl.org/SDL3/SDL_CreateWindow"
    raw_url = f"https://raw.githubusercontent.com/owner/repository/{COMMIT_SHA}/a.md"
    downloader = FakeDownloader(
        {
            url: b"page",
            "https://api.github.com/repos/owner/repository/commits/main": github_commit_response(),
            raw_url: b"a",
        }
    )
    fetch_docs.fetch_url_document(url, tmp_path, downloader)
    fetch_docs.fetch_github_documents("owner/repository", ["a.md"], "main", tmp_path, downloader)
    cached_documents = fetch_docs.list_cached_documents(tmp_path)
    assert [entry.get("commit") for entry in cached_documents] == [COMMIT_SHA, None]
    assert cached_documents[1]["url"] == url
    assert [entry["source_file_path"] for entry in cached_documents] == sorted(
        entry["source_file_path"] for entry in cached_documents
    )


def test_listing_an_empty_cache_returns_nothing(tmp_path):
    assert fetch_docs.list_cached_documents(tmp_path / "missing") == []


def test_command_github_without_a_reference_uses_the_default_branch(tmp_path, capsys):
    raw_url = f"https://raw.githubusercontent.com/owner/repository/{COMMIT_SHA}/src/header.h"
    downloader = FakeDownloader(
        {
            # A repository whose default branch is master: only HEAD resolves without asking.
            "https://api.github.com/repos/owner/repository/commits/HEAD": github_commit_response(),
            raw_url: b"// header",
        }
    )
    exit_code = fetch_docs.main(
        ["github", "owner/repository", "src/header.h", "--cache-directory", str(tmp_path)],
        downloader,
    )
    assert exit_code == 0
    printed_summary = json.loads(capsys.readouterr().out)
    assert printed_summary["reference"] == "HEAD"
    assert printed_summary["commit"] == COMMIT_SHA


def test_command_url_prints_the_saved_file(tmp_path, capsys):
    url = "https://ziglang.org/learn/"
    exit_code = fetch_docs.main(
        ["url", url, "--cache-directory", str(tmp_path)], FakeDownloader({url: b"<html>"})
    )
    assert exit_code == 0
    printed_summary = json.loads(capsys.readouterr().out)
    assert printed_summary["url"] == url
    assert pathlib.Path(printed_summary["cache_file_path"]).read_bytes() == b"<html>"


def test_command_list_prints_the_cached_documents(tmp_path, capsys):
    url = "https://ziglang.org/learn/"
    fetch_docs.fetch_url_document(url, tmp_path, FakeDownloader({url: b"<html>"}))
    exit_code = fetch_docs.main(["list", "--cache-directory", str(tmp_path)], FakeDownloader({}))
    assert exit_code == 0
    printed_summary = json.loads(capsys.readouterr().out)
    assert printed_summary["documents"][0]["url"] == url


def test_command_errors_print_json_with_a_code_and_exit_1(tmp_path, capsys):
    exit_code = fetch_docs.main(
        ["github", "no-slash", "README.md", "--cache-directory", str(tmp_path)], FakeDownloader({})
    )
    assert exit_code == 1
    captured = capsys.readouterr()
    assert captured.out == ""
    assert json.loads(captured.err)["error"]["code"] == "invalid_repository"
