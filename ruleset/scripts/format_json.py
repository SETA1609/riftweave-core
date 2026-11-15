#!/usr/bin/env python3
"""Lay out JSON files in the one format of CONTRIBUTING.md#json-format (M5-51). Stdlib only.

- 2-space indent.
- An object or array whose one-line form fits in the width (100 columns), counting its indent,
  its key, and its comma, stays on one line; otherwise each member goes on its own line and the
  rule applies again inside.
- On one line, an object has a space inside its braces and an array none inside its brackets.
- Nothing is reordered or rewritten: key order as written, numbers as spelled, strings unwrapped,
  UTF-8 kept. One trailing newline.

Usage:
  format_json.py [--check] [--width N] [PATH ...]
With no PATH it reads the paths listed in ruleset/scripts/format_paths.txt. A PATH that is a
directory means every *.json file under it. --check writes nothing, lists the files it would
change, and exits 1 if there are any. A file that is not JSON exits 2.
"""
import argparse
import json
import sys
from pathlib import Path

SCRIPTS = Path(__file__).resolve().parent
ROOT = SCRIPTS.parent.parent
PATHS_FILE = SCRIPTS / "format_paths.txt"
WIDTH = 100
INDENT = "  "


class Number(str):
    """A number exactly as the file spells it, so formatting never rewrites 1.0 or 1e3."""


def _parse(text):
    return json.loads(text, parse_int=Number, parse_float=Number, parse_constant=Number)


def _scalar(value):
    if isinstance(value, Number):
        return str(value)
    if value is True:
        return "true"
    if value is False:
        return "false"
    if value is None:
        return "null"
    return json.dumps(value, ensure_ascii=False)


def _one_line(value):
    if isinstance(value, dict):
        if not value:
            return "{}"
        return "{ " + ", ".join(f"{_scalar(key)}: {_one_line(member)}" for key, member in value.items()) + " }"
    if isinstance(value, list):
        return "[" + ", ".join(_one_line(member) for member in value) + "]"
    return _scalar(value)


def _lines(value, depth, head, tail, width):
    """The lines of one value: `head` is what precedes it on its first line (a key), `tail` its comma."""
    indent = INDENT * depth
    one = _one_line(value)
    if not isinstance(value, (dict, list)) or not value or len(indent + head + one + tail) <= width:
        return [indent + head + one + tail]
    if isinstance(value, dict):
        opening, closing = "{", "}"
        members = [(f"{_scalar(key)}: ", member) for key, member in value.items()]
    else:
        opening, closing = "[", "]"
        members = [("", member) for member in value]
    lines = [indent + head + opening]
    for index, (key, member) in enumerate(members):
        lines.extend(_lines(member, depth + 1, key, "," if index < len(members) - 1 else "", width))
    lines.append(indent + closing + tail)
    return lines


def format_text(text, width=WIDTH):
    """The formatted text of one JSON document."""
    return "\n".join(_lines(_parse(text), 0, "", "", width)) + "\n"


def _files(paths):
    for path in paths:
        if path.is_dir():
            yield from sorted(path.rglob("*.json"))
        else:
            yield path


def _listed_paths():
    if not PATHS_FILE.exists():
        return []
    entries = [line.strip() for line in PATHS_FILE.read_text(encoding="utf-8").splitlines()]
    return [ROOT / entry for entry in entries if entry and not entry.startswith("#")]


def main(argv=None):
    parser = argparse.ArgumentParser(description="Lay out JSON files in the one format (CONTRIBUTING.md#json-format).")
    parser.add_argument("paths", nargs="*", type=Path, help="files or directories (default: ruleset/scripts/format_paths.txt)")
    parser.add_argument("--check", action="store_true", help="write nothing; list files that would change and exit 1 if any")
    parser.add_argument("--width", type=int, default=WIDTH, help=f"columns a one-line object or array may use (default {WIDTH})")
    args = parser.parse_args(argv)

    paths = args.paths or _listed_paths()
    changed, broken = [], []
    for path in _files(paths):
        text = path.read_text(encoding="utf-8")
        try:
            formatted = format_text(text, width=args.width)
        except json.JSONDecodeError as error:
            broken.append(f"{path}: {error}")
            continue
        if formatted != text:
            changed.append(path)
            if not args.check:
                path.write_text(formatted, encoding="utf-8")
    for line in broken:
        print(line, file=sys.stderr)
    for path in changed:
        print(f"{'would format' if args.check else 'formatted'} {path}")
    if broken:
        return 2
    return 1 if args.check and changed else 0


if __name__ == "__main__":
    sys.exit(main())
