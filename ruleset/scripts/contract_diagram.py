#!/usr/bin/env python3
"""Draw the command contract as Mermaid diagrams, read from backend/contract/ (APP-36).

Each command file (`<command>.json`) gives a class diagram: its input and its result as classes
with their fields, linked to the shared types of common.json, which are drawn once. The session
command gets a flowchart of what each step reads from the stored actor and writes back, read from
the table in docs/backend/sessions.md, and the test checks that table against the commands
`sessionAdapters` in backend/core/session.go adapts. Nothing is drawn by hand:
test_contract_diagram.py fails until the page is rewritten.

Run with:
    python ruleset/scripts/contract_diagram.py           print the page
    python ruleset/scripts/contract_diagram.py --write   rewrite docs/backend/contract-diagram.md
"""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
CONTRACT = ROOT / "backend" / "contract"
SESSIONS = ROOT / "docs" / "backend" / "sessions.md"
SESSION_GO = ROOT / "backend" / "core" / "session.go"
PAGE = ROOT / "docs" / "backend" / "contract-diagram.md"
# The page's Modified line (M5-53): bump it with the contract change that changes the page.
MODIFIED = "2026-11-18"
SHARED = "common.json"


def commands():
    """{name: contract} for every command file, sorted by name."""
    found = {}
    for path in sorted(CONTRACT.glob("*.json")):
        content = json.loads(path.read_text())
        if content.get("$schema") == "./command.schema.json":
            found[content["name"]] = content
    return found


def shared_types():
    return json.loads((CONTRACT / SHARED).read_text())["definitions"]


def object_types():
    """The shared types that are objects: the ones drawn as classes and linked to."""
    return {name for name, schema in shared_types().items() if schema.get("type") == "object" or "properties" in schema}


OBJECTS = None


def ref_name(ref):
    """(name, is a shared object type) for a $ref: common.json's own, or another schema's."""
    global OBJECTS
    if OBJECTS is None:
        OBJECTS = object_types()
    file, _, pointer = ref.partition("#")
    name = pointer.rsplit("/", 1)[-1]
    return name, file in (SHARED, "") and pointer.startswith("/definitions/") and name in OBJECTS


def type_of(node):
    """A field's type as Mermaid shows it, and the shared types it names."""
    if not isinstance(node, dict):
        return "any", []
    if "$ref" in node:
        name, shared = ref_name(node["$ref"])
        return name, [name] if shared else []
    for keyword in ("anyOf", "oneOf"):
        if keyword in node:
            parts = [type_of(option) for option in node[keyword]]
            return "_or_".join(dict.fromkeys(name for name, _ in parts)), [n for _, names in parts for n in names]
    if node.get("type") == "array":
        name, names = type_of(node.get("items", {}))
        return f"List~{name}~", names
    if "enum" in node:
        return "enum", []
    kind = node.get("type", "any")
    return ("_or_".join(kind) if isinstance(kind, list) else kind), []


def members(schema):
    """[(field, type)] and the shared types the fields name, for an object schema."""
    fields, named = [], []
    for field, node in sorted((schema.get("properties") or {}).items()):
        kind, names = type_of(node)
        fields.append((field, kind))
        named += [(field, name) for name in names]
    return fields, named


def class_block(name, schema):
    fields, named = members(schema)
    lines = [f"    class {name} {{"]
    lines += [f"        {kind} {field}" for field, kind in fields]
    lines.append("    }")
    return lines, named


def command_diagram(name, contract):
    """The command's input and result, each linked to the shared types its fields name."""
    lines = ["```mermaid", "classDiagram"]
    links, shown = [], set()
    for part, key in (("input", "input_schema"), ("result", "result_schema")):
        block, named = class_block(f"{name}_{part}", contract.get(key, {}))
        lines += block
        for field, shared in named:
            links.append(f"    {name}_{part} --> {shared} : {field}")
            shown.add(shared)
    lines += [f"    class {shared}" for shared in sorted(shown)]
    lines += sorted(set(links))
    lines.append("```")
    return "\n".join(lines)


def shared_diagram(types):
    """Every object type of common.json, once, with the shared types it names."""
    lines = ["```mermaid", "classDiagram"]
    links = []
    for name, schema in sorted(types.items()):
        if schema.get("type") != "object" and "properties" not in schema:
            continue
        block, named = class_block(name, schema)
        lines += block
        links += [f"    {name} --> {shared} : {field}" for field, shared in named if shared != name]
    lines += sorted(set(links))
    lines.append("```")
    return "\n".join(lines)


def session_table():
    """[(command, filled from the actor, written back)] from the table in sessions.md."""
    text = SESSIONS.read_text()
    section = text.split("## What a step reads and writes", 1)[1].split("\n### ", 1)[0]
    rows = []
    for line in section.splitlines():
        cells = [cell.strip() for cell in line.strip().strip("|").split("|")]
        if len(cells) == 3 and cells[0].startswith("`"):
            rows.append(tuple(cells))
    return [(command.strip("`"), filled, written) for command, filled, written in rows]


# A table row the session does not adapt: dev runs only with --dev (docs/backend/dev-mode.md).
NOT_ADAPTED = {"dev"}


def session_adapters():
    """The commands sessionAdapters in backend/core/session.go fills and writes back."""
    block = SESSION_GO.read_text().split("sessionAdapters", 1)[1].split("\n}", 1)[0]
    return re.findall(r'^\s+"([a-z_]+)":\s+\{', block, re.M)


def label(text):
    """Table text as a Mermaid edge label: no Markdown links or code marks, no quotes."""
    text = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", text)
    return text.replace("`", "").replace('"', "'")


def session_flowchart(rows):
    lines = ["```mermaid", "flowchart LR", '    actor[("stored actor")]']
    for command, filled, written in rows:
        lines.append(f'    {command}["{command}"]')
        lines.append(f'    actor -- "fills: {label(filled)}" --> {command}')
        lines.append(f'    {command} -- "writes back: {label(written)}" --> actor')
    lines.append("```")
    return "\n".join(lines)


def page():
    found = commands()
    out = [
        "# Command contract diagram",
        "",
        f"Modified: {MODIFIED}",
        "",
        "Generated by `ruleset/scripts/contract_diagram.py` from `backend/contract/`; do not edit by hand.",
        "Run `python ruleset/scripts/contract_diagram.py --write` after a contract change, and",
        "`test_contract_diagram.py` fails until the page matches. What the contract is:",
        "[README.md](README.md#commands) and [contract.md](../core/contract.md).",
        "",
        f"{len(found)} commands. Each is a class diagram of its input and its result; a field whose type is",
        "a shared type of `common.json` links to it, and the shared types are drawn once, below.",
        "",
        "## Shared types",
        "",
        "The object types of `common.json`, and the shared types their fields name.",
        "",
        shared_diagram(shared_types()),
        "",
        "## Session steps",
        "",
        "What each session step reads from the stored actor and writes back, from the table in",
        "[sessions.md](sessions.md#what-a-step-reads-and-writes).",
        "",
        session_flowchart(session_table()),
    ]
    for name, contract in found.items():
        out += ["", f"## {name}", "", contract["summary"], "", command_diagram(name, contract)]
    return "\n".join(out) + "\n"


def main(argv):
    text = page()
    if argv[1:] == ["--write"]:
        PAGE.write_text(text)
        print(f"wrote {PAGE.relative_to(ROOT)}")
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main(sys.argv)
