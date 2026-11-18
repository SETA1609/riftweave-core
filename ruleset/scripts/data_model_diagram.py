#!/usr/bin/env python3
"""Draw the data model as Mermaid diagrams, read from the schemas (docs/core/data-model.md).

An entity is a collection of rows: a top-level array of objects in a ruleset/schemas file. A
relationship is a field whose schema is a reference, a string pattern that names
`core:<category>/<key>` (most are the `*Ref` definitions in schema.json). Nothing is drawn by
hand, so the diagram cannot drift from the schemas: test_data_model_diagram.py fails until the
page is rewritten.

A second page draws composition as Mermaid class diagrams (docs/core/composition.md, M5-64):
every collection whose rows may name a `base` and `overrides`, what the base may be, and, from
the data, which rows build on which base.

Run with:
    python ruleset/scripts/data_model_diagram.py           print the data model page
    python ruleset/scripts/data_model_diagram.py --write   rewrite both pages
"""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
SCHEMAS = ROOT / "ruleset" / "schemas"
PAGE = ROOT / "docs" / "core" / "data-model-diagram.md"
COMPOSITION_PAGE = ROOT / "docs" / "core" / "composition-diagram.md"
DATA = ROOT / "ruleset" / "data"
# The composition page's Modified line: bump it with the schema or data change that changes it.
COMPOSITION_MODIFIED = "2026-11-18"
# The most bases an instance diagram draws, and the most rows under each, so it stays readable.
INSTANCE_BASES = 10
INSTANCE_ROWS = 3
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


# --- Composition (M5-64) -------------------------------------------------------------------

# A category in a base reference's pattern: core:(weapon|armor)_base/ or core:creature_type/.
CATEGORY_PATTERN = re.compile(r"\^\(core:(\(?[a-z_|]+\)?)(_base)?/")


def class_name(name):
    """A Mermaid class name: letters, digits, and underscores, starting with a letter."""
    return re.sub(r"[^A-Za-z0-9_]", "_", name)


def base_categories(node, current, schemas):
    """The categories a `base` field may name, in schema order: weapon_base, creature_type, ..."""
    if "$ref" in node:
        node, current = resolve(node["$ref"], current, schemas)
    found = []
    for option in node.get("oneOf", []):
        found += [c for c in base_categories(option, current, schemas) if c not in found]
    match = CATEGORY_PATTERN.match(node.get("pattern") or "")
    if match:
        found += [part + (match.group(2) or "") for part in match.group(1).strip("()").split("|")]
    return found


def families(schemas):
    """[(collection, schema file, full row definition, composed row definition, base categories)]:
    each collection whose items are one of a full row or a row with `base` and `overrides`."""
    found = []
    for name, schema in schemas.items():
        if name in SKIPPED_SCHEMAS:
            continue
        for collection, node in (schema.get("properties") or {}).items():
            items = node.get("items", {}) if isinstance(node, dict) else {}
            options = [resolve(option["$ref"], name, schemas)[0] for option in items.get("oneOf", []) if "$ref" in option]
            composed = [option for option in options if {"base", "overrides"} <= set(option.get("properties", {}))]
            full = [option for option in options if option not in composed]
            if composed and full:
                found.append((collection, name, fields_of(full[0], name, schemas), fields_of(composed[0], name, schemas),
                              base_categories(composed[0]["properties"]["base"], name, schemas)))
    return found


def fields_of(definition, current, schemas, seen=frozenset()):
    """The top-level fields a row definition allows, through its allOf, oneOf, anyOf, and $ref."""
    fields = set(definition.get("properties", {}))
    if "$ref" in definition and definition["$ref"] not in seen:
        resolved, name = resolve(definition["$ref"], current, schemas)
        fields |= set(fields_of(resolved, name, schemas, seen | {definition["$ref"]}))
    for keyword in ("allOf", "oneOf", "anyOf"):
        for option in definition.get(keyword, []):
            fields |= set(fields_of(option, current, schemas, seen))
    return sorted(fields)


def schema_diagram(collection, full, composed, categories):
    """The family's shape: the row as written, each kind of base it may build on, and the
    composed row that names one, with its own fields."""
    row = class_name(collection.upper())
    lines = ["```mermaid", "classDiagram"]
    lines.append(f"    class {row} {{")
    lines += [f"        {field}" for field in full]
    lines.append("    }")
    for category in categories:
        lines.append(f"    class {class_name(category)}")
    composed_name = f"{row}_WITH_BASE"
    lines.append(f"    class {composed_name} {{")
    lines += [f"        {field}" for field in composed]
    lines.append("    }")
    for category in categories:
        lines.append(f"    {class_name(category)} <|-- {composed_name} : base")
    lines.append(f"    {composed_name} ..> {row} : composes to")
    lines.append("```")
    return "\n".join(lines)


def data_rows(collection):
    """The rows of a collection as written in ruleset/data (bases not yet merged)."""
    rows = []
    for path in sorted(DATA.rglob("*.json")):
        if "bases" in path.parts:
            continue
        content = json.loads(path.read_text())
        if isinstance(content.get(collection), list):
            rows += [row for row in content[collection] if isinstance(row, dict)]
    return rows


def instance_diagram(collection):
    """Which rows build on which base, from the data: the most-used bases first, each with up to
    INSTANCE_ROWS of its rows; the rest are counted."""
    by_base = {}
    for row in data_rows(collection):
        if isinstance(row.get("base"), str):
            by_base.setdefault(row["base"], []).append(row["key"])
    ranked = sorted(by_base.items(), key=lambda item: (-len(item[1]), item[0]))
    lines = ["```mermaid", "classDiagram"]
    for base, keys in ranked[:INSTANCE_BASES]:
        name = class_name(base.removeprefix("core:").replace("/", "__"))
        lines.append(f'    class {name}["{base.removeprefix("core:")}"]')
        for key in sorted(keys)[:INSTANCE_ROWS]:
            lines.append(f"    {name} <|-- {class_name(collection)}__{class_name(key)}")
        if len(keys) > INSTANCE_ROWS:
            lines.append(f'    note for {name} "{len(keys) - INSTANCE_ROWS} more"')
    lines.append("```")
    rows = sum(len(keys) for keys in by_base.values())
    shown = min(len(ranked), INSTANCE_BASES)
    return rows, len(ranked), shown, "\n".join(lines)


def template_diagram(schemas):
    """The monster templates: a layer between a monster's base and its overrides, one class per
    slot, the slot enum of monster_template.schema.json."""
    definition = schemas["monster_template.schema.json"]["definitions"]
    template = next(node for node in definition.values() if "slot" in node.get("properties", {}))
    slots = template["properties"]["slot"]["enum"]
    lines = ["```mermaid", "classDiagram", "    class MONSTER_TEMPLATES {"]
    lines += [f"        {field}" for field in fields_of(template, "monster_template.schema.json", schemas)]
    lines.append("    }")
    for slot in slots:
        lines.append(f"    MONSTER_TEMPLATES <|-- {class_name(slot.upper())}_TEMPLATE : slot {slot}")
    lines.append("    MONSTERS_WITH_BASE ..> MONSTER_TEMPLATES : templates, the layers")
    lines.append("```")
    return slots, "\n".join(lines)


def perk_kinds_pending(schemas):
    """True while the features schema has no base perk with kinds (M5-58)."""
    return "basePerk" not in schemas["feature.schema.json"].get("definitions", {})


def composition_page(schemas):
    found = families(schemas)
    out = [
        "# Composition diagram",
        "",
        f"Modified: {COMPOSITION_MODIFIED}",
        "",
        "Generated by `ruleset/scripts/data_model_diagram.py` from `ruleset/schemas/` and `ruleset/data/`;",
        "do not edit by hand. Run `python ruleset/scripts/data_model_diagram.py --write` after a change,",
        "and `test_data_model_diagram.py` fails until the page matches. The rules:",
        "[composition.md](composition.md).",
        "",
        "The [data model diagram](data-model-diagram.md) shows references; this page shows what a row is",
        "built on. A row may be written out whole, or name a `base` and list its `overrides`; the second",
        "composes to the first, so a consumer reads one flat row. An arrow `<|--` runs from a base to what",
        "builds on it.",
        "",
        f"{len(found)} collections compose rows from a base: {', '.join(f'`{c}`' for c, *_ in found)}.",
    ]
    for collection, name, full, composed, categories in found:
        rows, bases, shown, instances = instance_diagram(collection)
        out += ["", f"## {collection.capitalize()}", "",
                f"From `{name}`: a row as written, or a row on a base ({', '.join(f'`{c}`' for c in categories)}).", "",
                schema_diagram(collection, full, composed, categories), "",
                f"In the data, {rows} rows build on {bases} bases"
                + (f"; the {shown} most used are drawn" if shown < bases else "; all are drawn")
                + f", with up to {INSTANCE_ROWS} rows each." if bases else f"In the data, no {collection} row builds on a base.", ""]
        if bases:
            out.append(instances)
    slots, templates = template_diagram(schemas)
    out += ["", "## Monster templates", "",
            "A template is a layer: it changes a monster between its base and its overrides. Each fills one",
            f"slot ({', '.join(f'`{s}`' for s in slots)}), from `monster_template.schema.json`.", "", templates]
    out += ["", "## Perk kinds", ""]
    if perk_kinds_pending(schemas):
        out.append("Pending M5-58: the features schema has no base perk yet, so the perk kinds are not drawn.")
    return "\n".join(out) + "\n"


def main(argv):
    schemas = load_schemas()
    text = page(collections(schemas))
    if argv[1:] == ["--write"]:
        PAGE.write_text(text)
        COMPOSITION_PAGE.write_text(composition_page(schemas))
        print(f"wrote {PAGE.relative_to(ROOT)} and {COMPOSITION_PAGE.relative_to(ROOT)}")
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main(sys.argv)
