#!/usr/bin/env python3
"""Check a release tag before the release workflow publishes it (CONTRIBUTING.md#releasing, M5-62).

Usage:
    python ruleset/scripts/release_check.py TAG [--previous TAG] [--notes PATH]

- TAG must be `v` plus the version in VERSION.
- CHANGELOG.md must have a `## v<version>` section.
- Every key removed since the previous tag (key_diff.py) must be named in a bullet of that
  section marked **(breaking)**. Without --previous, the latest tag before TAG is used; with no
  earlier tag, there is nothing to compare.
- With --notes, the section's body is written there, for the release notes.

Exit code 1 and one line per problem when a check fails, 0 otherwise.
"""

import argparse
import re
import subprocess
import sys
from pathlib import Path

import key_diff

REPO_ROOT = Path(__file__).resolve().parent.parent.parent


def changelog_section(changelog_text, title):
    """The body of the `## <title>` section, or None when there is none."""
    match = re.search(rf"^## {re.escape(title)}[ \t]*$(.*?)(?=^## |\Z)", changelog_text, re.MULTILINE | re.DOTALL)
    return match.group(1).strip("\n") if match else None


def unmarked_removals(removed, section):
    """The removed (collection, key) pairs that no **(breaking)** bullet of the section names."""
    breaking = [line for line in section.splitlines() if line.lstrip().startswith("- ") and "(breaking)" in line]
    return [(collection, key) for collection, key in removed if not any(re.search(rf"\b{re.escape(key)}\b", line) for line in breaking)]


def previous_tag(tag):
    """The latest tag before TAG, or None for the first release."""
    completed = subprocess.run(["git", "describe", "--tags", "--abbrev=0", f"{tag}^"], cwd=REPO_ROOT, capture_output=True, text=True)
    return completed.stdout.strip() or None if completed.returncode == 0 else None


def check(tag, version, changelog_text, removed):
    """The problems with releasing TAG; an empty list when it may be released."""
    problems = []
    if tag != f"v{version}":
        problems.append(f"the tag {tag} is not v{version}, the version in VERSION")
    section = changelog_section(changelog_text, tag)
    if section is None:
        problems.append(f"CHANGELOG.md has no '## {tag}' section")
    else:
        for collection, key in unmarked_removals(removed, section):
            problems.append(f"{collection}/{key} was removed, but no (breaking) bullet of '## {tag}' names it")
    return problems


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    parser.add_argument("tag")
    parser.add_argument("--previous", help="The tag to diff keys against; default the latest before TAG.")
    parser.add_argument("--notes", type=Path, help="Write the changelog section here.")
    arguments = parser.parse_args(argv)
    version = (REPO_ROOT / "VERSION").read_text().strip()
    changelog_text = (REPO_ROOT / "CHANGELOG.md").read_text()
    previous = arguments.previous or previous_tag(arguments.tag)
    removed = []
    if previous:
        removed, _ = key_diff.diff_keys(
            key_diff.collect_keys(key_diff.load_ref_files(previous)),
            key_diff.collect_keys(key_diff.load_ref_files(arguments.tag)),
        )
    problems = check(arguments.tag, version, changelog_text, removed)
    for problem in problems:
        print(problem, file=sys.stderr)
    if problems:
        return 1
    if arguments.notes:
        arguments.notes.write_text(changelog_section(changelog_text, arguments.tag) + "\n")
    print(f"{arguments.tag}: ready to release" + (f" (keys checked against {previous})" if previous else " (first release)"))
    return 0


if __name__ == "__main__":
    sys.exit(main())
