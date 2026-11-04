#!/usr/bin/env python3
"""List the keys removed and added in ruleset/data/ between two versions.

Usage:
    python ruleset/scripts/key_diff.py OLD NEW

OLD and NEW are git refs (tags, branches, commits) or WORKTREE for the files on disk.
Exit code 1 if any key was removed, 0 otherwise. See docs/core/versioning.md.
"""

import json
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
DATA_PREFIX = "ruleset/data/"
WORKTREE = "WORKTREE"


def _git(*args):
    return subprocess.run(
        ["git", *args], cwd=REPO_ROOT, check=True, capture_output=True, text=True
    ).stdout


def load_ref_files(ref):
    """Map each data file's path under ruleset/data/ to its parsed JSON, at a git ref or on disk."""
    files = {}
    if ref == WORKTREE:
        for path in sorted((REPO_ROOT / DATA_PREFIX).rglob("*.json")):
            files[path.relative_to(REPO_ROOT / DATA_PREFIX).as_posix()] = json.loads(path.read_text())
        return files
    listing = _git("ls-tree", "-r", "--name-only", ref, "--", DATA_PREFIX).splitlines()
    for name in listing:
        if name.endswith(".json"):
            files[name[len(DATA_PREFIX):]] = json.loads(_git("show", f"{ref}:{name}"))
    return files


def _entry_key(entry):
    if not isinstance(entry, dict):
        return None
    if isinstance(entry.get("key"), str):
        return entry["key"]
    if isinstance(entry.get("id"), str):
        return entry["id"]
    return None


def collect_keys(files):
    """Return the set of (collection, key). The collection is a wrapper array's name, or the folder
    name for a file that is a single entry. A numeric id is not a key."""
    keys = set()
    for path, data in files.items():
        if not isinstance(data, dict):
            continue
        own = _entry_key(data)
        if own is not None:
            keys.add((Path(path).parent.name, own))
        for name, value in data.items():
            if isinstance(value, list):
                for entry in value:
                    key = _entry_key(entry)
                    if key is not None:
                        keys.add((name, key))
    return keys


def diff_keys(old, new):
    """Return (removed, added) as sorted lists."""
    return sorted(old - new), sorted(new - old)


def main(argv=None, out=None, loader=load_ref_files):
    argv = sys.argv[1:] if argv is None else argv
    out = out or sys.stdout
    if len(argv) != 2:
        print(__doc__, file=out)
        return 2
    old = collect_keys(loader(argv[0]))
    new = collect_keys(loader(argv[1]))
    removed, added = diff_keys(old, new)
    print(f"removed (breaking): {len(removed)}", file=out)
    for collection, key in removed:
        print(f"  - {collection}/{key}", file=out)
    print(f"added: {len(added)}", file=out)
    for collection, key in added:
        print(f"  + {collection}/{key}", file=out)
    return 1 if removed else 0


if __name__ == "__main__":
    sys.exit(main())
