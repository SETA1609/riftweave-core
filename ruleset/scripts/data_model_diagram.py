#!/usr/bin/env python3
"""Draw the data model as Mermaid ER diagrams, read from the schemas (docs/core/data-model.md).

An entity is a collection of rows: a top-level array of objects in a ruleset/schemas file. A
relationship is a field whose schema is a reference, a string pattern that names
`core:<category>/<key>` (most are the `*Ref` definitions in schema.json). Nothing is drawn by
hand, so the diagram cannot drift from the schemas: test_data_model_diagram.py fails until the
page is rewritten.

Run with:
    python ruleset/scripts/data_model_diagram.py           print the page
    python ruleset/scripts/data_model_diagram.py --write   rewrite docs/core/data-model-diagram.md
"""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
SCHEMAS = ROOT / "ruleset" / "schemas"
PAGE = ROOT / "docs" / "core" / "data-model-diagram.md"
# The page's Modified line (M5-53): bump it with the schema change that changes the diagram.
MODIFIED = "2026-11-15"

# A reference pattern: ^(core:<category>/... or core:(a|b)_base/...
REF_PATTERN = re.compile(r"^\^\(core:(\(?[a-z_|]+\)?)(_base)?/")

# The collection that holds a category's rows, where it is not the category plus "s"
# (validate.py's COLLECTION_CATEGORIES, read the other way).
COLLECTION_OF = {
    "weapon": "equipment", "armor": "equipment", "consumable": "equipment", "accessory": "equipment",
    "gear": "equipment", "weapon_base": "equipment", "armor_base": "equipment", "accessory_base": "equipment",
    "creature_type": "monsters", "summon_bonus": "summon_bonuses", "crafted_item": "crafted_items",
    "monster_template": "monster_templates", "natural_weapon": "natural_weapons", "skill_group": "skill_groups",
}
SKIPPED_SCHEMAS = {"module.manifest.schema.json", "schema.json"}

# The areas each get a smaller diagram; a collection in none is only in the overview.
AREAS = [
    ("Characters", ["abilities", "skills", "skill_groups", "features", "races", "backgrounds", "origins", "characters"]),
    ("Magic", ["spells", "effects", "summons", "summon_bonuses", "conditions", "modifier_targets"]),
    ("Items and crafting", ["equipment", "materials", "gems", "tiers", "ingredients", "recipes", "crafted_items", "forms"]),
    ("Monsters and the world", ["monsters", "monster_templates", "natural_weapons", "hazards", "checks", "tables", "drops",
                                "merchants", "factions", "situations", "called_shots"]),
]


def load_schemas():
    return {path.name: json.loads(path.read_text()) for path in sorted(SCHEMAS.glob("*.json"))}


def resolve(ref, current, schemas):
    """The schema a $ref names, and the file it is in."""
    file, _, pointer = ref.partition("#")
    name = file or current
    node = schemas[name]
    for part in [p for p in pointer.split("/") if p]:
        node = node[part]
    return node, name


def target_of(node):
    """The collection a reference schema names, or None."""
    pattern = node.get("pattern") if isinstance(node, dict) else None
    match = REF_PATTERN.match(pattern or "")
    if not match:
        return None
    category = match.group(1).strip("()").split("|")[0] + (match.group(2) or "")
    return COLLECTION_OF.get(category, category + "s")


def references(node, current, schemas, path, seen):
    """(field path, target collection) for every reference under node."""
    if not isinstance(node, dict):
        return []
    target = target_of(node)
    if target:
        return [(path, target)]
    found = []
    if "$ref" in node:
        key = (node["$ref"] if "#" in node["$ref"] and node["$ref"].split("#")[0] else current + node["$ref"])
        if key not in seen:
            resolved, name = resolve(node["$ref"], current, schemas)
            found += references(resolved, name, schemas, path, seen | {key})
    for keyword in ("oneOf", "anyOf", "allOf"):
        for option in node.get(keyword, []):
            found += references(option, current, schemas, path, seen)
    for field, child in (node.get("properties") or {}).items():
        found += references(child, current, schemas, f"{path}.{field}" if path else field, seen)
    for keyword in ("items", "additionalProperties"):
        child = node.get(keyword)
        if isinstance(child, dict):
            found += references(child, current, schemas, path + ("[]" if keyword == "items" else "{}"), seen)
    return found


def collections(schemas):
    """{collection: [(field, target)]} for every collection of rows in the schemas.

    An override repeats its base's fields, so `overrides.` is dropped from a path. Where two
    schemas name a collection alike (`tiers`, `recipes`), the one whose file is not named after it
    is prefixed with its file's name (`balance_tiers`).
    """
    model = {}
    owners = {}
    for name, schema in schemas.items():
        if name in SKIPPED_SCHEMAS:
            continue
        for collection in (schema.get("properties") or {}):
            owners.setdefault(collection, []).append(name)
    for name, schema in schemas.items():
        if name in SKIPPED_SCHEMAS:
            continue
        for collection, node in (schema.get("properties") or {}).items():
            if collection.startswith("$") or not isinstance(node, dict) or node.get("type") != "array":
                continue
            items = node.get("items", {})
            if "$ref" in items:
                items, _ = resolve(items["$ref"], name, schemas)
            # Rows are objects; an array with no item schema at all (material_form's forms) is
            # drawn too, with nothing known about its fields.
            if items and items.get("type") != "object" and "oneOf" not in items and "properties" not in items:
                continue
            edges = {(field.removeprefix("overrides."), target) for field, target in references(node.get("items", {}), name, schemas, "", frozenset())}
            stem = name.removesuffix(".schema.json")
            if len(owners[collection]) > 1 and not collection.startswith(stem):
                collection = f"{stem}_{collection}"
            model.setdefault(collection, set()).update(edges)
    return {collection: sorted(edges) for collection, edges in sorted(model.items())}


def entity(name):
    return name.upper()


def attribute(field):
    """A field path as a Mermaid attribute name: snake_case, a dot as an underscore."""
    snake = re.sub(r"(?<=[a-z0-9])([A-Z])", r"_\1", field.replace("[]", "").replace("{}", "")).lower()
    return re.sub(r"[^a-z0-9_]+", "_", snake).strip("_")


def diagram(model, members=None):
    """The members with their references, and each collection they name as a bare entity."""
    lines = ["```mermaid", "erDiagram"]
    sources = set(members or model)
    shown = sources | {target for member in sources for _, target in model.get(member, [])}
    for collection in sorted(shown):
        edges = model.get(collection, []) if collection in sources else []
        lines.append(f"    {entity(collection)} {{")
        lines.append("        string key PK")
        for field in sorted({attribute(f) for f, _ in edges}):
            lines.append(f"        string {field} FK")
        lines.append("    }")
    for collection in sorted(sources):
        by_target = {}
        for field, target in model.get(collection, []):
            by_target.setdefault(target, []).append(field)
        for target, fields in sorted(by_target.items()):
            lines.append(f'    {entity(collection)} }}o--|| {entity(target)} : "{", ".join(fields)}"')
    lines.append("```")
    return "\n".join(lines)


def page(model):
    edges = sum(len({t for _, t in e}) for e in model.values())
    out = [
        "# Data model diagram",
        "",
        f"Modified: {MODIFIED}",
        "",
        "Generated by `ruleset/scripts/data_model_diagram.py` from `ruleset/schemas/`; do not edit by hand.",
        "Run `python ruleset/scripts/data_model_diagram.py --write` after a schema change, and",
        "`test_data_model_diagram.py` fails until the page matches. What the model is and why:",
        "[data-model.md](data-model.md).",
        "",
        "An entity is a collection of rows, its `key` the primary key. A line is a reference: the field",
        "(a dotted path, `[]` for each item of a list) holds `core:<category>/<key>` naming a row of the",
        "other collection, so many rows name one. Composition bases are references too: an equipment row's",
        "`base` names an equipment base, a monster's `base` a creature type or a monster.",
        "",
        f"{len(model)} collections; {edges} pairs of collections joined by a reference.",
        "",
        "## Overview",
        "",
        diagram(model),
    ]
    for title, members in AREAS:
        present = [m for m in members if m in model]
        out += ["", f"## {title}", "", f"The {', '.join(f'`{m}`' for m in present)} collections and what they name.", "",
                diagram(model, present)]
    return "\n".join(out) + "\n"


def main(argv):
    text = page(collections(load_schemas()))
    if argv[1:] == ["--write"]:
        PAGE.write_text(text)
        print(f"wrote {PAGE.relative_to(ROOT)}")
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main(sys.argv)
