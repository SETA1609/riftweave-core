#!/usr/bin/env python3
"""Rebase open pull requests onto a target branch and push with --force-with-lease.

Runs from GitHub Actions on every push to main. A conflict in a CHANGELOG.md
where both sides only added lines keeps both: the rebased branch's entries come
first, then the target's, and the `Last updated:` line takes the newer
timestamp. Any other conflict, in a CHANGELOG.md or anywhere else, aborts that
branch and leaves its remote tip alone for a person to resolve.

Run locally with --branches, it refuses a dirty working tree and never moves a
local branch that has commits its remote does not.

Branches with no open pull request are left alone, so a squash-merged branch
that GitHub has not deleted is not replayed onto main.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request

CHANGELOG_MERGE_DRIVER = "changelog-entries"
ATTRIBUTES_BEGIN = "# rebase-open-branches changelog-entries begin"
ATTRIBUTES_END = "# rebase-open-branches changelog-entries end"
ATTRIBUTES_BLOCK = f"""{ATTRIBUTES_BEGIN}
CHANGELOG.md merge={CHANGELOG_MERGE_DRIVER}
{ATTRIBUTES_END}
"""
GITHUB_API_TIMEOUT_SECONDS = 30
BRANCH_NAME = re.compile(r"[A-Za-z0-9][A-Za-z0-9._/-]*")
LAST_UPDATED_LINE = re.compile(r"^Last updated: (\S+)$", re.MULTILINE)


def branch_name_is_safe(name: str) -> bool:
    """Reject refspecs and option-like names before they reach git."""
    if not name or name.startswith("-") or name.startswith("/") or name.endswith("/"):
        return False
    if ".." in name or "@{" in name or "//" in name or name.startswith("refs/"):
        return False
    return BRANCH_NAME.fullmatch(name) is not None


def pull_request_branch_names(pull_requests: list[dict], repository: str, target: str) -> list[str]:
    """Head branches of open pull requests in this repository, against target."""
    names: list[str] = []
    seen: set[str] = set()
    for pull_request in pull_requests:
        head = pull_request.get("head") or {}
        base = pull_request.get("base") or {}
        head_repository = (head.get("repo") or {}).get("full_name")
        name = head.get("ref") or ""
        if base.get("ref") != target or head_repository != repository:
            continue
        if name == target or not branch_name_is_safe(name) or name in seen:
            continue
        seen.add(name)
        names.append(name)
    return names


def run_git(repository: str, *args: str, check: bool = False) -> subprocess.CompletedProcess[str]:
    environment = os.environ.copy()
    environment["GIT_EDITOR"] = "true"
    return subprocess.run(
        ["git", "-C", repository, *args],
        check=check,
        capture_output=True,
        text=True,
        env=environment,
    )


def git_path(repository: str, path: str) -> str:
    completed = run_git(repository, "rev-parse", "--git-path", path, check=True)
    raw = completed.stdout.strip()
    if os.path.isabs(raw):
        return raw
    return os.path.join(repository, raw)


def merge_changelog_sides(base_text: str, target_text: str, branch_text: str) -> str | None:
    """Merge three versions of a CHANGELOG.md, or return None to leave the conflict.

    Every `Last updated: <UTC timestamp>` line takes the newest of the three
    timestamps. A conflict where the base side is empty (both sides only added
    lines, such as a new entry at the top of `## Unreleased`) keeps the branch's
    lines, then the target's, separated by a blank line. Any other conflict,
    where both sides changed the same existing lines, returns None.
    """
    newest_timestamp = find_newest_last_updated_timestamp(base_text, target_text, branch_text)
    if newest_timestamp is not None:
        stamped_line = f"Last updated: {newest_timestamp}"
        base_text, target_text, branch_text = (
            LAST_UPDATED_LINE.sub(stamped_line, text)
            for text in (base_text, target_text, branch_text)
        )
    with tempfile.TemporaryDirectory() as directory:
        paths = []
        for name, text in (("branch", branch_text), ("base", base_text), ("target", target_text)):
            path = os.path.join(directory, name)
            with open(path, "w", encoding="utf-8") as handle:
                handle.write(text)
            paths.append(path)
        completed = subprocess.run(
            [
                "git",
                "merge-file",
                "-p",
                "--diff3",
                "-L",
                "branch",
                "-L",
                "base",
                "-L",
                "target",
                *paths,
            ],
            check=False,
            capture_output=True,
            text=True,
        )
    # merge-file exits with the number of conflicts, or a negative (>127) code on error.
    if completed.returncode > 127:
        return None
    return resolve_added_lines_conflicts(completed.stdout)


def find_newest_last_updated_timestamp(*texts: str) -> str | None:
    """ISO 8601 UTC timestamps sort as text, so the newest is the largest."""
    timestamps = [match.group(1) for text in texts for match in LAST_UPDATED_LINE.finditer(text)]
    return max(timestamps) if timestamps else None


def resolve_added_lines_conflicts(merged_text: str) -> str | None:
    """Resolve diff3 conflicts whose base is empty: branch lines, then target lines."""
    resolved: list[str] = []
    section = None
    branch_lines: list[str] = []
    base_lines: list[str] = []
    target_lines: list[str] = []
    for line in merged_text.splitlines(keepends=True):
        if section is None and line == "<<<<<<< branch\n":
            section, branch_lines, base_lines, target_lines = "branch", [], [], []
        elif section == "branch" and line == "||||||| base\n":
            section = "base"
        elif section == "base" and line == "=======\n":
            section = "target"
        elif section == "target" and line == ">>>>>>> target\n":
            if "".join(base_lines).strip():
                return None
            resolved.extend(join_added_blocks(branch_lines, target_lines))
            section = None
        elif section == "branch":
            branch_lines.append(line)
        elif section == "base":
            base_lines.append(line)
        elif section == "target":
            target_lines.append(line)
        else:
            resolved.append(line)
    if section is not None:
        return None
    return "".join(resolved)


def join_added_blocks(branch_lines: list[str], target_lines: list[str]) -> list[str]:
    """Branch block, then target block, with one blank line between entries."""
    joined = list(branch_lines)
    if joined and not joined[-1].endswith("\n"):
        joined[-1] += "\n"
    if joined and target_lines and joined[-1].strip() and target_lines[0].strip():
        joined.append("\n")
    joined.extend(target_lines)
    return joined


def write_merged_changelog(base_path: str, target_path: str, branch_path: str) -> int:
    """Merge driver for CHANGELOG.md, called by git as `%O %A %B`.

    During a rebase, %A is the target branch being built on and %B is the commit
    being replayed. Writes the merge into target_path and exits 0, or exits 1
    and leaves the conflict for the caller to abort.
    """
    try:
        texts = []
        for path in (base_path, target_path, branch_path):
            with open(path, encoding="utf-8") as handle:
                texts.append(handle.read())
    except (OSError, UnicodeDecodeError):
        return 1
    merged = merge_changelog_sides(*texts)
    if merged is None:
        return 1
    with open(target_path, "w", encoding="utf-8") as handle:
        handle.write(merged)
    return 0


def install_changelog_merge_driver(repository: str) -> None:
    script = os.path.abspath(__file__)
    driver = f"{sys.executable} -I {script} merge-changelog %O %A %B"
    run_git(
        repository,
        "config",
        f"merge.{CHANGELOG_MERGE_DRIVER}.name",
        "Keep both sides' new CHANGELOG.md entries",
        check=True,
    )
    run_git(
        repository,
        "config",
        f"merge.{CHANGELOG_MERGE_DRIVER}.driver",
        driver,
        check=True,
    )
    attributes_path = git_path(repository, "info/attributes")
    os.makedirs(os.path.dirname(attributes_path), exist_ok=True)
    existing = ""
    if os.path.exists(attributes_path):
        with open(attributes_path, encoding="utf-8") as handle:
            existing = handle.read()
    if ATTRIBUTES_BEGIN not in existing:
        with open(attributes_path, "a", encoding="utf-8") as handle:
            if existing and not existing.endswith("\n"):
                handle.write("\n")
            handle.write(ATTRIBUTES_BLOCK)


def remove_changelog_merge_driver(repository: str) -> None:
    run_git(repository, "config", "--unset", f"merge.{CHANGELOG_MERGE_DRIVER}.name")
    run_git(repository, "config", "--unset", f"merge.{CHANGELOG_MERGE_DRIVER}.driver")
    attributes_path = git_path(repository, "info/attributes")
    if not os.path.exists(attributes_path):
        return
    with open(attributes_path, encoding="utf-8") as handle:
        existing = handle.read()
    begin = existing.find(ATTRIBUTES_BEGIN)
    end = existing.find(ATTRIBUTES_END)
    if begin == -1 or end == -1:
        return
    end = end + len(ATTRIBUTES_END)
    if end < len(existing) and existing[end] == "\n":
        end += 1
    updated = existing[:begin] + existing[end:]
    with open(attributes_path, "w", encoding="utf-8") as handle:
        handle.write(updated)


def current_branch_name(repository: str) -> str:
    completed = run_git(repository, "rev-parse", "--abbrev-ref", "HEAD", check=True)
    return completed.stdout.strip()


def working_tree_is_clean(repository: str) -> bool:
    """True when no tracked file has staged or unstaged changes."""
    completed = run_git(repository, "status", "--porcelain", "--untracked-files=no")
    return completed.returncode == 0 and not completed.stdout.strip()


def local_branch_can_follow_remote(repository: str, branch: str, previous_remote_tip: str) -> bool:
    """True when the local branch is missing or still at previous_remote_tip.

    A local branch with commits of its own is never moved, so a run with
    --no-push (or a local run that pushes) cannot drop unpushed work.
    """
    local = run_git(repository, "rev-parse", "--verify", "--quiet", f"refs/heads/{branch}")
    if local.returncode != 0:
        return True
    return local.stdout.strip() == previous_remote_tip


def head_has_conflict_markers(repository: str) -> bool:
    completed = run_git(repository, "grep", "-l", "-E", r"^<<<<<<<", "HEAD")
    return completed.returncode == 0


def rebase_one_branch(repository: str, branch: str, target: str, push: bool) -> bool:
    """Replay branch onto origin/target. Return False when it was left unpushed."""
    print(f"==> Rebase {branch} onto {target}")
    remote_reference = f"refs/remotes/origin/{branch}"
    previous_remote_tip = run_git(repository, "rev-parse", remote_reference).stdout.strip()
    checkout = run_git(repository, "checkout", "--detach", f"origin/{branch}")
    if checkout.returncode != 0:
        print(checkout.stderr.strip() or checkout.stdout.strip())
        print(f"  failed to check out origin/{branch}")
        return False
    rebase = run_git(repository, "rebase", f"origin/{target}")
    if rebase.returncode != 0:
        print(rebase.stderr.strip() or rebase.stdout.strip())
        run_git(repository, "rebase", "--abort")
        print(f"  {branch} has a conflict that needs a person — left unchanged")
        return False
    if head_has_conflict_markers(repository):
        print(f"  {branch} still has conflict markers — left unchanged")
        return False
    if push and not push_rebased_branch(repository, branch):
        return False
    if local_branch_can_follow_remote(repository, branch, previous_remote_tip):
        run_git(repository, "branch", "-f", branch, "HEAD")
    else:
        print(f"  local {branch} has commits its remote does not — left where it is")
    print(f"  {branch} rebased")
    return True


def push_rebased_branch(repository: str, branch: str) -> bool:
    expected = run_git(repository, "rev-parse", f"refs/remotes/origin/{branch}", check=True)
    lease = expected.stdout.strip()
    push = run_git(
        repository,
        "push",
        f"--force-with-lease=refs/heads/{branch}:{lease}",
        "origin",
        f"HEAD:refs/heads/{branch}",
    )
    if push.returncode != 0:
        print(push.stderr.strip() or push.stdout.strip())
        print(f"  {branch} was not pushed")
        return False
    if push.stderr.strip():
        print(push.stderr.strip())
    return True


def fetch_open_pull_requests(repository_name: str, token: str, target: str) -> list[dict]:
    """Every open pull request against target. Follows GitHub pagination."""
    url = (
        f"https://api.github.com/repos/{repository_name}/pulls"
        f"?state=open&base={target}&per_page=100"
    )
    pull_requests: list[dict] = []
    while url:
        request = urllib.request.Request(
            url,
            headers={
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {token}",
                "X-GitHub-Api-Version": "2022-11-28",
                "User-Agent": "rebase-open-branches",
            },
        )
        try:
            with urllib.request.urlopen(request, timeout=GITHUB_API_TIMEOUT_SECONDS) as response:
                page = json.load(response)
                link = response.headers.get("Link")
        except urllib.error.HTTPError as error:
            detail = error.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"GitHub pull request list failed: {error.code} {detail}") from error
        if not isinstance(page, list):
            raise RuntimeError("GitHub pull request list was not a JSON array")
        pull_requests.extend(page)
        url = next_link(link)
    return pull_requests


def next_link(link_header: str | None) -> str | None:
    if not link_header:
        return None
    for part in link_header.split(","):
        if 'rel="next"' not in part:
            continue
        start = part.find("<")
        end = part.find(">")
        if start != -1 and end != -1:
            return part[start + 1 : end]
    return None


def configure_actions_identity(repository: str) -> None:
    if os.environ.get("GITHUB_ACTIONS") != "true":
        return
    run_git(repository, "config", "user.name", "github-actions[bot]", check=True)
    run_git(
        repository,
        "config",
        "user.email",
        "41898282+github-actions[bot]@users.noreply.github.com",
        check=True,
    )


def restore_start_branch(repository: str, start: str) -> None:
    if run_git(repository, "rev-parse", "--verify", "REBASE_HEAD").returncode == 0:
        run_git(repository, "rebase", "--abort")
    if start == "HEAD":
        return
    run_git(repository, "checkout", "--quiet", start)


def rebase_branches(
    repository: str,
    target: str,
    branches: list[str],
    push: bool,
    fetch: bool,
) -> int:
    if not working_tree_is_clean(repository):
        print("The working tree has uncommitted changes to tracked files; commit or stash them.")
        return 1
    if fetch:
        fetched = run_git(repository, "fetch", "origin", "--prune")
        if fetched.returncode != 0:
            print(fetched.stderr.strip() or fetched.stdout.strip())
            return 1
    start = current_branch_name(repository)
    install_changelog_merge_driver(repository)
    failures = 0
    try:
        configure_actions_identity(repository)
        if not branches:
            print(f"No open pull requests against {target}.")
            return 0
        for branch in branches:
            if not rebase_one_branch(repository, branch, target, push):
                failures += 1
    finally:
        remove_changelog_merge_driver(repository)
        restore_start_branch(repository, start)
    if failures:
        print(f"{failures} branch(es) were not updated.")
        return 1
    return 0


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repository", default=".", help="Path to the git checkout")
    parser.add_argument("--target", default="main", help="Branch to rebase onto")
    parser.add_argument(
        "--branches",
        nargs="*",
        help="Branches to rebase. Omit to use open pull requests from the GitHub API.",
    )
    parser.add_argument("--no-push", action="store_true", help="Rebase locally and do not push")
    parser.add_argument(
        "--no-fetch",
        action="store_true",
        help="Use the refs already in the checkout",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    arguments = list(sys.argv[1:] if argv is None else argv)
    if arguments and arguments[0] == "merge-changelog":
        if len(arguments) != 4:
            print(
                "usage: rebase_open_branches.py merge-changelog BASE TARGET BRANCH",
                file=sys.stderr,
            )
            return 2
        return write_merged_changelog(arguments[1], arguments[2], arguments[3])

    args = parse_args(arguments)
    repository = os.path.abspath(args.repository)
    if not branch_name_is_safe(args.target):
        print(f"Unsafe target branch name: {args.target}", file=sys.stderr)
        return 2
    if args.branches is None:
        repository_name = os.environ.get("GITHUB_REPOSITORY", "")
        token = os.environ.get("GITHUB_TOKEN", "")
        if not repository_name or not token:
            print(
                "GITHUB_REPOSITORY and GITHUB_TOKEN are required to list pull requests.",
                file=sys.stderr,
            )
            return 2
        pull_requests = fetch_open_pull_requests(repository_name, token, args.target)
        branches = pull_request_branch_names(pull_requests, repository_name, args.target)
    else:
        branches = []
        for name in args.branches:
            if not branch_name_is_safe(name) or name == args.target:
                print(f"Skipping unsafe branch name: {name}")
                continue
            branches.append(name)
    return rebase_branches(
        repository,
        args.target,
        branches,
        push=not args.no_push,
        fetch=not args.no_fetch,
    )


if __name__ == "__main__":
    sys.exit(main())
