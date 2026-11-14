#!/usr/bin/env python3
"""Validate all JSON data files against their $schema references.

If modules exist under ruleset/modules/, each module's manifest and data files
are also validated. Module data files share the same schema store as core data.
"""

import json
import os
import re
from functools import partial
import sys
import warnings
from pathlib import Path

from jsonschema import validate, ValidationError, RefResolver

ROOT = Path(__file__).resolve().parent.parent
DATA_DIR = ROOT / "data"
SCHEMA_DIR = ROOT / "schemas"
MODULES_DIR = ROOT / "modules"
MONSTERS_BASES_DIR = DATA_DIR / "monsters" / "bases"


# Mapping: delivery context name -> required channel name.
# Extend this dict to add enforcement for new delivery channels.
CHANNEL_REQUIREMENTS = {
    "ingredient": "ingredient",
    "enchantment": "enchantment",
}


def _format_channel_violation(
    effect_ref, required_channel, declared_channels, location
):
    declared = sorted(declared_channels) if declared_channels else []
    effect_key = effect_ref.split("/")[-1] if "/" in effect_ref else effect_ref
    return (
        f"Validation Error:\n"
        f"  Effect '{effect_key}' is used as '{required_channel}' in {location},\n"
        f"  but only declares these channels: {declared}"
    )


def validate_effect_channel(effect_ref, used_in, effect_channels):
    """Return True if the effect is permitted for the given delivery context.

    The required channel is looked up from CHANNEL_REQUIREMENTS using `used_in`.
    Module refs (module:) are accepted without channel data.
    Non-core refs pass silently (existence errors are caught elsewhere).
    """
    if not isinstance(effect_ref, str) or effect_channels is None:
        return True
    if effect_ref.startswith("module:"):
        return True
    if not effect_ref.startswith("core:"):
        return True
    required_channel = CHANNEL_REQUIREMENTS.get(used_in)
    if required_channel is None:
        return True
    declared = effect_channels.get(effect_ref, set())
    return required_channel in declared


def build_effect_channels(data_files, module_data_files=None):
    """Build 'core:effect/<key>' -> set(channels) map for enforcement.

    Collects from core and module data files (later entries override,
    supporting module overrides of core effects).
    """
    channels = {}
    fps = list(data_files or [])
    if module_data_files:
        fps.extend(fp for fp, _ in module_data_files)
    for fp in fps:
        try:
            data = json.loads(fp.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        effs = data.get("effects")
        if not isinstance(effs, list):
            continue
        for item in effs:
            if isinstance(item, dict):
                key = item.get("key")
                chlist = item.get("channels")
                if isinstance(key, str) and isinstance(chlist, list):
                    channels[f"core:effect/{key}"] = set(chlist)
    return channels


# Maps equipment source filenames to namespaced reference categories.
# Cross-references use core:<category>/<key> with singular category names
# (e.g. core:weapon/shortsword, core:armor/leather_jerkin, core:consumable/potion_minor_healing).
EQUIPMENT_FILE_CATEGORIES = {
    "weapons.json": "weapon",
    "armor.json": "armor",
    "consumables.json": "consumable",
    "accessories.json": "accessory",
    "gear.json": "gear",
}

# Maps collection names to their singular category for namespaced refs.
COLLECTION_CATEGORIES = {
    "effects": "effect",
    "spells": "spell",
    "ingredients": "ingredient",
    "skills": "skill",
    "features": "feature",
    "races": "race",
    "conditions": "condition",
    "materials": "material",
    "gems": "gem",
    "monsters": "monster",
    "summons": "summon",
    "backgrounds": "background",
    "recipes": "recipe",
    "crafted_items": "crafted_item",
    "origins": "origin",
    "hazards": "hazard",
    "checks": "check",
    "natural_weapons": "natural_weapon",
    "monster_templates": "monster_template",
    "skill_groups": "skill_group",
    "summon_bonuses": "summon_bonus",
}


def load_store():
    store = {}
    for f in sorted(SCHEMA_DIR.glob("*.schema.json")):
        store[f.as_uri()] = json.loads(f.read_text())
    return store


def collect_data_files():
    for dirpath, dirnames, filenames in os.walk(DATA_DIR):
        # Skip bases/ directories (they are resolved via base+override merging,
        # not validated as standalone data files)
        dirnames[:] = [d for d in dirnames if d != "bases"]
        for f in filenames:
            if f.endswith(".json"):
                yield Path(dirpath) / f


def collect_module_data_files():
    """Yield (file_path, module_dir) for each data file declared in a module manifest."""
    if not MODULES_DIR.exists():
        return
    for entry in sorted(MODULES_DIR.iterdir()):
        if not entry.is_dir() or entry.name.startswith("."):
            continue
        manifest_path = entry / "manifest.json"
        if not manifest_path.exists():
            continue
        try:
            manifest = json.loads(manifest_path.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        for rel in manifest.get("data", []):
            fp = (entry / rel).resolve()
            if fp.exists():
                yield fp, entry


def _equipment_category_for_file(fp):
    """Return the core:<category> prefix for an equipment data file, if known."""
    return EQUIPMENT_FILE_CATEGORIES.get(fp.name)


def _iter_equipment_entries(data_files, module_data_files=None, rel_base=ROOT):
    """Yield (rel_path, category, item) for each equipment entry."""
    file_paths = list(data_files)
    if module_data_files:
        file_paths.extend(fp for fp, _ in module_data_files)

    for fp in file_paths:
        try:
            data = json.loads(fp.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        equipment = equipment_rows(fp, equipment=data.get("equipment"))
        if not isinstance(equipment, list):
            continue
        rel = fp.relative_to(rel_base)
        category = _equipment_category_for_file(fp)
        for item in equipment:
            if isinstance(item, dict):
                yield rel, category, item


def build_equipment_index(data_files, module_data_files=None, rel_base=ROOT):
    """Build lookup sets for equipment referential integrity checks."""
    return {
        "namespaced_keys": _collect_equipment_namespaced_keys(
            data_files, module_data_files, rel_base
        ),
        "bare_keys": _collect_equipment_bare_keys(
            data_files, module_data_files, rel_base
        ),
    }


def build_namespaced_ref_index(data_files, module_data_files=None):
    """Build a set of valid namespaced cross-references across all data files.

    Indexes core:category/key refs only. Module-style refs are accepted at
    validation time without being pre-indexed.
    """
    refs = set()
    for fp in data_files:
        try:
            data = json.loads(fp.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        for k, v in data.items():
            if k.startswith("$") or not isinstance(v, list):
                continue
            cat = COLLECTION_CATEGORIES.get(k)
            if not cat:
                continue
            for item in v:
                if isinstance(item, dict):
                    key = item.get("key")
                    if isinstance(key, str):
                        refs.add(f"core:{cat}/{key}")
    if module_data_files:
        for fp, _ in module_data_files:
            try:
                data = json.loads(fp.read_text())
            except (json.JSONDecodeError, OSError):
                continue
            for k, v in data.items():
                if k.startswith("$") or not isinstance(v, list):
                    continue
                cat = COLLECTION_CATEGORIES.get(k)
                if not cat:
                    continue
                for item in v:
                    if isinstance(item, dict):
                        key = item.get("key")
                        if isinstance(key, str):
                            refs.add(f"core:{cat}/{key}")
    return refs


def build_bare_key_index(data_files, module_data_files=None):
    """Build a dict mapping collection name -> set of bare keys for crafting validation.

    Indexes bare keys for materials, gems, ingredients, and crafted_items
    so recipe/crafted_item cross-references can be validated.
    """
    index = {}
    for fp in data_files:
        try:
            data = json.loads(fp.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        for k, v in data.items():
            if k.startswith("$") or not isinstance(v, list):
                continue
            if k not in COLLECTION_CATEGORIES:
                continue
            keys = set()
            for item in v:
                if isinstance(item, dict):
                    key = item.get("key")
                    if isinstance(key, str):
                        keys.add(key)
            if keys:
                index.setdefault(k, set()).update(keys)
    if module_data_files:
        for fp, _ in module_data_files:
            try:
                data = json.loads(fp.read_text())
            except (json.JSONDecodeError, OSError):
                continue
            for k, v in data.items():
                if k.startswith("$") or not isinstance(v, list):
                    continue
                if k not in COLLECTION_CATEGORIES:
                    continue
                keys = set()
                for item in v:
                    if isinstance(item, dict):
                        key = item.get("key")
                        if isinstance(key, str):
                            keys.add(key)
                if keys:
                    index.setdefault(k, set()).update(keys)
    return index


def resolve_namespaced_ref(ref, namespaced_ref_index):
    """Check if a namespaced string ref exists in the index.

    Only accepts core:category/key and module: prefixed refs.
    Bare keys are rejected. Returns False for non-strings and unresolvable refs.
    """
    if not isinstance(ref, str):
        return False
    if ref.startswith("module:"):
        return True
    if not ref.startswith("core:"):
        return False
    return ref in namespaced_ref_index


def _collect_equipment_namespaced_keys(
    data_files, module_data_files=None, rel_base=ROOT
):
    keys = set()
    for _rel, category, item in _iter_equipment_entries(
        data_files, module_data_files, rel_base
    ):
        key = item.get("key")
        if category and isinstance(key, str):
            keys.add(f"core:{category}/{key}")
    return keys


def _collect_equipment_bare_keys(data_files, module_data_files=None, rel_base=ROOT):
    keys = set()
    for _rel, _category, item in _iter_equipment_entries(
        data_files, module_data_files, rel_base
    ):
        key = item.get("key")
        if isinstance(key, str):
            keys.add(key)
    return keys


def equipment_ref_valid(ref, id_index, equipment_index):
    """Return True if a namespaced equipment reference resolves.
    Only accepts core:category/key and module: prefixed refs. Bare keys rejected.
    """
    if isinstance(ref, str):
        if ref.startswith("core:"):
            return ref in equipment_index["namespaced_keys"]
        if ref.startswith("module:"):
            return True
    return False


def check_equipment_uniqueness(data_files, module_data_files=None, rel_base=ROOT):
    """Ensure equipment ids and keys are unique across split equipment files."""
    seen_ids = {}  # id -> source
    seen_keys = {}  # key -> source
    errors = []

    for rel, _category, item in _iter_equipment_entries(
        data_files, module_data_files, rel_base
    ):
        key = item.get("key", "?")
        source = f"{rel}:{key}"

        eid = item.get("id")
        if isinstance(eid, int):
            if eid in seen_ids:
                errors.append(
                    f"DUPLICATE equipment id {eid}: {seen_ids[eid]} vs {source}"
                )
            else:
                seen_ids[eid] = source

        if isinstance(key, str):
            if key in seen_keys:
                errors.append(
                    f"DUPLICATE equipment key '{key}': {seen_keys[key]} vs {source}"
                )
            else:
                seen_keys[key] = source

    return errors


DAMAGE_TYPE_ALIASES = {
    "cold": "frost",
    "lightning": "shock",
}


def load_canonical_damage_types(data_dir=DATA_DIR):
    """Return the damage-type keys from the combat registry. One vocabulary."""
    path = data_dir / "combat" / "damage_types.json"
    data = json.loads(path.read_text())
    keys = set()
    for item in data.get("damage_types", []):
        if isinstance(item, dict) and isinstance(item.get("key"), str):
            keys.add(item["key"])
    return keys


def _damage_type_error(found, source, canonical):
    alias = DAMAGE_TYPE_ALIASES.get(found)
    if alias:
        return (
            f"damage type '{found}' in {source} is not canonical; use '{alias}'"
        )
    return (
        f"damage type '{found}' in {source} is not in damage_types.json "
        f"(known: {', '.join(sorted(canonical))})"
    )


def check_damage_type_vocabulary(data_files, canonical, rel_base=ROOT):
    """Reject weapon damage types and armor resistance keys outside the registry.

    cold/lightning are aliases, not keys. Bases are included because validate.py
    does not schema-check bases/ on its own.
    """
    errors = []
    if not canonical:
        return ["damage_types.json has no keys; cannot check vocabulary"]

    def consider(found, source):
        if not isinstance(found, str) or found in canonical:
            return
        errors.append(_damage_type_error(found, source, canonical))

    for rel, _category, item in _iter_equipment_entries(data_files, rel_base=rel_base):
        key = item.get("key", "?")
        weapon = (item.get("overrides") or {}).get("weapon") or item.get("weapon") or {}
        damage = weapon.get("damage") or {}
        consider(damage.get("type"), f"{rel}:{key}")
        armor = (item.get("overrides") or {}).get("armor") or item.get("armor") or {}
        resistances = armor.get("resistances") or {}
        if isinstance(resistances, dict):
            for tag in resistances:
                consider(tag, f"{rel}:{key} resistances")

    bases = rel_base / "data" / "equipment" / "bases"
    if bases.exists():
        for fp in sorted(bases.rglob("*.json")):
            try:
                item = json.loads(fp.read_text())
            except (json.JSONDecodeError, OSError):
                continue
            if not isinstance(item, dict):
                continue
            rel = fp.relative_to(rel_base)
            key = item.get("key", fp.stem)
            weapon = item.get("weapon") or {}
            consider((weapon.get("damage") or {}).get("type"), f"{rel}:{key}")
            resistances = (item.get("armor") or {}).get("resistances") or {}
            if isinstance(resistances, dict):
                for tag in resistances:
                    consider(tag, f"{rel}:{key} resistances")
    return errors


NAMESPACED_REF = re.compile(r"^core:[a-z_]+/[a-z0-9_]+$")


def _walk_strings(node):
    if isinstance(node, str):
        yield node
    elif isinstance(node, dict):
        for value in node.values():
            yield from _walk_strings(value)
    elif isinstance(node, list):
        for value in node:
            yield from _walk_strings(value)


def _all_files(data_files, module_data_files=None):
    files = list(data_files)
    if module_data_files:
        files.extend(fp for fp, _ in module_data_files)
    return files


def _load_json_file(fp):
    try:
        return json.loads(fp.read_text())
    except (json.JSONDecodeError, OSError):
        return None


def _rel(fp, rel_base):
    try:
        return fp.relative_to(rel_base).as_posix()
    except ValueError:
        return fp.name


def _base_directories():
    """Each directory of base templates and the category that names its files."""
    return {
        MONSTERS_BASES_DIR: "creature_type",
        DATA_DIR / "equipment" / "bases" / "weapons": "weapon_base",
        DATA_DIR / "equipment" / "bases" / "armor": "armor_base",
        DATA_DIR / "equipment" / "bases" / "accessories": "accessory_base",
    }


def build_base_refs():
    """Every composition base as a reference (docs/core/keys.md): a base file is a row."""
    return {
        f"core:{category}/{path.stem}"
        for directory, category in _base_directories().items()
        if directory.exists()
        for path in directory.glob("*.json")
    }


def _base_files():
    return [path for directory in _base_directories() if directory.exists() for path in sorted(directory.glob("*.json"))]


def check_all_references(data_files, module_data_files=None, rel_base=ROOT):
    """Gate 1 (docs/core/integrity.md): every core:<category>/<key> string in any data or module
    file names an entry that exists. Walks every string, so no field escapes the check."""
    known = build_namespaced_ref_index(data_files, module_data_files)
    known |= build_equipment_index(data_files, module_data_files, rel_base=rel_base)["namespaced_keys"]
    # The base templates are the shipped ones, so a check of another tree leaves them out.
    base_files = _base_files() if rel_base == ROOT else []
    if base_files:
        known |= build_base_refs()
    errors = []
    for fp in [*_all_files(data_files, module_data_files), *base_files]:
        data = _load_json_file(fp)
        if data is None:
            continue
        missing = sorted({s for s in _walk_strings(data) if NAMESPACED_REF.match(s) and s not in known})
        for ref in missing:
            errors.append(f"{_rel(fp, rel_base)}: reference {ref} does not resolve")
    return errors


def _is_effect_registry(data):
    entries = data.get("effects") if isinstance(data, dict) else None
    return (
        isinstance(entries, list)
        and bool(entries)
        and all(isinstance(e, dict) and "channels" in e and "key" in e for e in entries)
    )


def check_effect_parameters(data_files, module_data_files=None):
    """Gate 2 (docs/core/integrity.md): where an effect names a parameter_category, every use of
    that effect has a parameter that references that category."""
    wanted = {}
    files = list(_all_files(data_files, module_data_files))
    loaded = [(fp, _load_json_file(fp)) for fp in files]
    for _, data in loaded:
        if data is not None and _is_effect_registry(data):
            for row in data["effects"]:
                if row.get("parameter_category"):
                    wanted[f"core:effect/{row['key']}"] = row["parameter_category"]
    errors = []

    def walk(node, fp):
        if isinstance(node, dict):
            category = wanted.get(node.get("effect"))
            if category and "parameter" in node:
                value = node["parameter"]
                core = isinstance(value, str) and value.startswith(f"core:{category}/")
                module = isinstance(value, str) and re.match(rf"^module:[a-z0-9_]+/{category}/", value)
                if not core and not module:
                    errors.append(f"{fp}: {node['effect']} needs a {category} reference, got {value!r}")
            for child in node.values():
                walk(child, fp)
        elif isinstance(node, list):
            for child in node:
                walk(child, fp)

    for fp, data in loaded:
        if data is not None:
            walk(data, fp)
    return errors


def check_skill_groups(data_files):
    """The grouped sheet (docs/core/progression.md#the-grouped-sheet-variant): every core skill is a
    member of exactly one core skill group."""
    skills, membership, errors = set(), {}, []
    for fp in data_files:
        data = _load_json_file(fp)
        if not isinstance(data, dict):
            continue
        for row in data.get("skills", []) if isinstance(data.get("skills"), list) else []:
            if isinstance(row, dict) and "key" in row:
                skills.add(row["key"])
        for group in data.get("skill_groups", []) or []:
            for ref in group.get("skills", []):
                key = ref.rsplit("/", 1)[-1]
                if key in membership:
                    errors.append(f"skill {key} is in skill groups {membership[key]} and {group['key']}")
                membership.setdefault(key, group["key"])
    if membership:
        for key in sorted(skills - set(membership)):
            errors.append(f"skill {key} is in no skill group")
    return errors


def check_effect_usage(data_files, module_data_files=None):
    """Gate 3 (docs/core/integrity.md): every core registry effect is referenced from a file outside the
    registry, or is marked reserved; a reserved effect must not be referenced."""
    registry = []
    referenced = set()
    core_files = {fp for fp in data_files}
    for fp in _all_files(data_files, module_data_files):
        data = _load_json_file(fp)
        if data is None:
            continue
        if _is_effect_registry(data):
            if fp in core_files:  # a module's own effects are its author's business
                registry.extend(data["effects"])
        else:
            referenced.update(s for s in _walk_strings(data) if NAMESPACED_REF.match(s))
    errors = []
    for effect in registry:
        ref = f"core:effect/{effect['key']}"
        if effect.get("reserved") and ref in referenced:
            errors.append(f"effect {effect['key']} is marked reserved but is referenced; remove the flag")
        elif not effect.get("reserved") and ref not in referenced:
            errors.append(f"effect {effect['key']} is referenced nowhere; use it or mark it reserved")
    return errors


# Gate 4 (docs/core/integrity.md, docs/core/modifier.md#target): where a modifier target's row may
# live, keyed by collection. A table's rows are keyed by "key", the abilities by "id".
MODIFIER_TARGET_TABLES = {
    "abilities": ("id", "number"),
    "skills": ("key", "number"),
    "derived": ("key", "number"),
    "actions": ("key", "number"),
    "conditions": ("key", "flag"),
    "damage_types": ("key", "damage_type"),
}
MODIFIER_OPERATIONS = {"add": "number", "multiply": "number", "set": "number", "grant": "flag", "resist": "damage_type"}
def modifier_operation(node):
    """A modifier's operation (docs/core/modifier.md)."""
    return node.get("operation")


def collect_vector_files():
    """The golden vectors: ruleset/tests/vectors/ and each module's tests/."""
    files = sorted((ROOT / "tests" / "vectors").glob("*.json"))
    if MODULES_DIR.exists():
        files.extend(sorted(MODULES_DIR.glob("*/tests/*.json")))
    return files


def _walk_modifiers(node):
    """Yield every modifier: an object with a known operation and a string target."""
    if isinstance(node, dict):
        if modifier_operation(node) in MODIFIER_OPERATIONS and isinstance(node.get("target"), str):
            yield node
        for value in node.values():
            yield from _walk_modifiers(value)
    elif isinstance(node, list):
        for value in node:
            yield from _walk_modifiers(value)


def check_modifier_targets(data_files, module_data_files=None, vector_files=()):
    """Gate 4: every modifier target names one row of the kind its op needs; every modifier_targets
    row is named by a modifier, repeats no other table, and is pending exactly when no vector uses it."""
    errors = []
    loaded = [(fp, _load_json_file(fp)) for fp in _all_files(data_files, module_data_files)]
    meanings = {}  # key -> [(kind, where)]
    target_rows = []  # (row, where)
    for fp, data in loaded:
        if not isinstance(data, dict):
            continue
        for collection, (identity, kind) in MODIFIER_TARGET_TABLES.items():
            for row in data.get(collection) or []:
                if isinstance(row, dict) and isinstance(row.get(identity), str):
                    meanings.setdefault(row[identity], []).append((kind, collection))
        for row in data.get("modifier_targets") or []:
            if isinstance(row, dict) and isinstance(row.get("key"), str):
                target_rows.append((row, _rel(fp, ROOT)))

    rows_by_key = {}
    for row, where in target_rows:
        key = row["key"]
        if key in meanings:
            tables = ", ".join(sorted({collection for _, collection in meanings[key]}))
            errors.append(f"{where}: modifier target {key} repeats a key of {tables}; a modifier names that row")
        elif key in rows_by_key:
            errors.append(f"{where}: modifier target {key} is defined twice")
        else:
            rows_by_key[key] = row
    for key, row in rows_by_key.items():
        meanings.setdefault(key, []).append((row.get("kind"), "modifier_targets"))

    named = set()
    for fp, data in loaded:
        if data is None:
            continue
        for mod in _walk_modifiers(data):
            target, op = mod["target"], modifier_operation(mod)
            named.add(target)
            if target not in meanings:
                errors.append(f"{_rel(fp, ROOT)}: modifier target {target} names no row")
                continue
            fitting = sorted({collection for kind, collection in meanings[target] if kind == MODIFIER_OPERATIONS[op]})
            if not fitting:
                kinds = ", ".join(sorted({kind for kind, _ in meanings[target]}))
                errors.append(f"{_rel(fp, ROOT)}: {op} on {target} needs a {MODIFIER_OPERATIONS[op]}, but {target} is a {kinds}")
            elif len(fitting) > 1:
                errors.append(f"{_rel(fp, ROOT)}: {op} on {target} is ambiguous: it is a key of {', '.join(fitting)}")

    in_vectors = set()
    for fp in vector_files:
        data = _load_json_file(fp)
        if data is not None:
            in_vectors.update(mod["target"] for mod in _walk_modifiers(data))
    for key, row in rows_by_key.items():
        if key not in named:
            errors.append(f"modifier target {key} is named by no modifier; use it or remove the row")
        if row.get("pending") and key in in_vectors:
            errors.append(f"modifier target {key} is pending but a golden vector uses it; remove the flag")
        elif not row.get("pending") and key not in in_vectors:
            errors.append(f"modifier target {key} is used by no golden vector; mark it pending")
    return errors


def check_references(
    data,
    coll_name,
    rel_path="",
    equipment_index=None,
    namespaced_ref_index=None,
    effect_channels=None,
):
    """Return list of referential integrity error messages."""
    errors = []

    if coll_name == "races":
        for r in data.get("races", []):
            if not isinstance(r, dict):
                continue
            lineage = r.get("lineage", {})
            if "parentRace" in lineage:
                pid = lineage["parentRace"]
                if isinstance(pid, str) and not resolve_namespaced_ref(
                    pid, namespaced_ref_index
                ):
                    errors.append(f"parentRace {pid} does not exist in races")
            for t in r.get("traits", []):
                if isinstance(t, dict):
                    tid = t.get("id")
                    if isinstance(tid, str) and not resolve_namespaced_ref(
                        tid, namespaced_ref_index
                    ):
                        errors.append(f"trait {tid} does not exist in features")

    elif coll_name == "features":
        for f in data.get("features", []):
            if not isinstance(f, dict):
                continue
            pr = f.get("prerequisite", {})
            for pid in pr.get("perks", []):
                if isinstance(pid, str) and not resolve_namespaced_ref(
                    pid, namespaced_ref_index
                ):
                    errors.append(f"prereq perk {pid} does not exist")
            for s in pr.get("skills", []):
                if isinstance(s, dict):
                    sid = s.get("id")
                    if isinstance(sid, str) and not resolve_namespaced_ref(
                        sid, namespaced_ref_index
                    ):
                        errors.append(f"prereq skill {sid} does not exist")
                    elif isinstance(sid, list):
                        for sss in sid:
                            if isinstance(sss, str) and not resolve_namespaced_ref(
                                sss, namespaced_ref_index
                            ):
                                errors.append(f"prereq skill {sss} does not exist")

    elif coll_name == "effects":
        for e in data.get("effects", []):
            if not isinstance(e, dict):
                continue
            ekey = e.get("key", "?")
            for sub in e.get("sub_effects", []) or []:
                if isinstance(sub, str) and not resolve_namespaced_ref(
                    sub, namespaced_ref_index
                ):
                    errors.append(
                        f"effect '{ekey}': sub_effect '{sub}' does not resolve"
                    )

    elif coll_name == "spells":
        for s in data.get("spells", []):
            if not isinstance(s, dict):
                continue
            for e in s.get("effects", []):
                if isinstance(e, dict):
                    eid = e.get("effect")
                    if isinstance(eid, int):
                        errors.append(
                            f"effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                        )
                    elif isinstance(eid, str) and not resolve_namespaced_ref(
                        eid, namespaced_ref_index
                    ):
                        errors.append(f"effect {eid!r} does not resolve")
            for rid in (s.get("cost", {}) or {}).get("reagents", []) or []:
                if isinstance(rid, int):
                    errors.append(
                        f"reagent {rid}: numeric refs not allowed (use core:ingredient/<key>)"
                    )
                elif isinstance(rid, str) and not resolve_namespaced_ref(
                    rid, namespaced_ref_index
                ):
                    errors.append(f"reagent {rid!r} does not resolve")
            # Validate spell effect parameter refs (when they are namespaced)
            for eff in s.get("effects", []):
                if isinstance(eff, dict):
                    param = eff.get("parameter")
                    if isinstance(param, str) and (
                        param.startswith("core:") or param.startswith("module:")
                    ):
                        eq_idx = equipment_index or {"namespaced_keys": set(), "bare_keys": set()}
                        if not resolve_namespaced_ref(param, namespaced_ref_index) and not equipment_ref_valid(
                            param, {}, eq_idx
                        ):
                            errors.append(
                                f"spell '{s.get('key', '?')}': effect parameter '{param}' does not resolve"
                            )

    elif coll_name == "backgrounds":
        for b in data.get("backgrounds", []):
            if not isinstance(b, dict):
                continue
            for sb in b.get("skill_bonuses", []) or []:
                if isinstance(sb, dict):
                    sid = sb.get("skill")
                    if isinstance(sid, int):
                        errors.append(
                            f"skill bonus {sid}: numeric refs not allowed (use core:skill/<key>)"
                        )
                    elif isinstance(sid, str) and (
                        namespaced_ref_index is None
                        or not resolve_namespaced_ref(sid, namespaced_ref_index)
                    ):
                        errors.append(f"skill bonus {sid!r} does not resolve")
            for ts in b.get("suggested_tag_skills", []) or []:
                if isinstance(ts, int):
                    errors.append(
                        f"suggested tag skill {ts}: numeric refs not allowed (use core:skill/<key>)"
                    )
                elif (
                    isinstance(ts, str)
                    and namespaced_ref_index
                    and not resolve_namespaced_ref(ts, namespaced_ref_index)
                ):
                    errors.append(f"suggested tag skill {ts!r} does not resolve")
            for ss in b.get("starting_spells", []) or []:
                if isinstance(ss, int):
                    errors.append(
                        f"starting spell {ss}: numeric refs not allowed (use core:spell/<key>)"
                    )
                elif (
                    isinstance(ss, str)
                    and namespaced_ref_index
                    and not resolve_namespaced_ref(ss, namespaced_ref_index)
                ):
                    errors.append(f"starting spell {ss!r} does not resolve")
            for se in b.get("starting_equipment", []) or []:
                if isinstance(se, dict):
                    item_ref = se.get("item")
                    eq_idx = equipment_index or {
                        "namespaced_keys": set(),
                        "bare_keys": set(),
                    }
                    if not equipment_ref_valid(item_ref, {}, eq_idx):
                        errors.append(f"starting equipment {item_ref!r} does not exist")

    elif coll_name == "monsters":
        for m in data.get("monsters", []):
            if not isinstance(m, dict):
                continue
            for a in m.get("abilities", []) or []:
                if isinstance(a, dict):
                    eid = a.get("effect")
                    if isinstance(eid, int):
                        errors.append(
                            f"ability effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                        )
                    elif isinstance(eid, str) and not resolve_namespaced_ref(
                        eid, namespaced_ref_index
                    ):
                        errors.append(f"ability effect {eid!r} does not resolve")

    elif coll_name == "ingredients":
        for i in data.get("ingredients", []):
            if not isinstance(i, dict):
                continue
            ikey = i.get("key", "?")
            for eid in i.get("effects", []) or []:
                if isinstance(eid, int):
                    errors.append(
                        f"effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                    )
                elif isinstance(eid, str) and not resolve_namespaced_ref(
                    eid, namespaced_ref_index
                ):
                    errors.append(f"effect {eid!r} does not resolve")
                if isinstance(eid, str) and not validate_effect_channel(
                    eid, "ingredient", effect_channels
                ):
                    declared = (
                        effect_channels.get(eid, set()) if effect_channels else set()
                    )
                    errors.append(
                        _format_channel_violation(
                            eid,
                            CHANNEL_REQUIREMENTS["ingredient"],
                            declared,
                            f"ingredient '{ikey}' ({rel_path})",
                        )
                    )

    elif coll_name == "conditions":
        for c in data.get("conditions", []):
            if not isinstance(c, dict):
                continue
            ckey = c.get("key", "?")
            for eid in c.get("appliedBy", []) or []:
                if isinstance(eid, int):
                    errors.append(
                        f"condition '{ckey}' appliedBy effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                    )
                elif isinstance(eid, str) and not resolve_namespaced_ref(
                    eid, namespaced_ref_index
                ):
                    errors.append(
                        f"condition '{ckey}' appliedBy effect {eid!r} does not resolve"
                    )
            for eid in c.get("removedBy", []) or []:
                if isinstance(eid, int):
                    errors.append(
                        f"condition '{ckey}' removedBy effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                    )
                elif isinstance(eid, str) and not resolve_namespaced_ref(
                    eid, namespaced_ref_index
                ):
                    errors.append(
                        f"condition '{ckey}' removedBy effect {eid!r} does not resolve"
                    )
            # Validate stacking consistency
            stacking = c.get("stacking", False)
            max_stacks = c.get("maxStacks")
            if stacking and max_stacks is None:
                errors.append(f"condition '{ckey}' has stacking:true but no maxStacks")
            if not stacking and max_stacks is not None:
                errors.append(
                    f"condition '{ckey}' has maxStacks but stacking is not true"
                )
            # Validate that prone has empty removedBy (special case)
            if ckey == "prone" and c.get("removedBy"):
                errors.append(
                    f"condition 'prone' should have empty removedBy (stand action, not effect)"
                )
            # Validate that cure is not in prone's removedBy
            if ckey == "prone" and "core:effect/cure" in c.get("removedBy", []):
                errors.append(
                    f"condition 'prone' must not have cure (core:effect/cure) in removedBy"
                )
            # Check for duplicate effect refs in appliedBy and removedBy
            ab = c.get("appliedBy", [])
            if len(ab) != len(set(ab)):
                errors.append(
                    f"condition '{ckey}' has duplicate effect refs in appliedBy"
                )
            rb = c.get("removedBy", [])
            if len(rb) != len(set(rb)):
                errors.append(
                    f"condition '{ckey}' has duplicate effect refs in removedBy"
                )

    elif coll_name == "equipment":
        for e in data.get("equipment", []) or []:
            if not isinstance(e, dict):
                continue
            ekey = e.get("key", "?")
            cons = e.get("consumable") or {}
            for eff in cons.get("effects", []) or []:
                if isinstance(eff, dict):
                    eid = eff.get("effect")
                    if isinstance(eid, int):
                        errors.append(
                            f"effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                        )
                    elif isinstance(eid, str) and not resolve_namespaced_ref(
                        eid, namespaced_ref_index
                    ):
                        errors.append(f"effect {eid!r} does not resolve")
            # Validate weapon skill reference (after base resolution)
            weapon = e.get("weapon") or {}
            skill = weapon.get("skill")
            if isinstance(skill, str) and not resolve_namespaced_ref(
                skill, namespaced_ref_index
            ):
                errors.append(
                    f"equipment '{ekey}': weapon skill '{skill}' does not resolve"
                )

    elif coll_name == "crafted_items":
        for ci in data.get("crafted_items", []) or []:
            if not isinstance(ci, dict):
                continue
            ckey = ci.get("key", "?")
            bk = ci.get("base_key")
            mk = ci.get("material_key")


            # Check base_key resolves as namespaced equipment ref
            if isinstance(bk, str) and not equipment_ref_valid(
                bk,
                {},
                equipment_index or {"namespaced_keys": set(), "bare_keys": set()},
            ):
                errors.append(
                    f"crafted item '{ckey}': base_key '{bk}' does not resolve to any equipment"
                )

            # Check material_key resolves as namespaced material ref
            if isinstance(mk, str) and not resolve_namespaced_ref(
                mk, namespaced_ref_index
            ):
                errors.append(
                    f"crafted item '{ckey}': material_key '{mk}' does not resolve to any material"
                )

            # Any material may be crafted onto any base (a granite bow is fine);
            # default_materials only names what a base is found or sold in.
            gj = ci.get("engraving_jewel")
            if isinstance(gj, dict):
                gk = gj.get("gem_key")
                if isinstance(gk, str) and not resolve_namespaced_ref(
                    gk, namespaced_ref_index
                ):
                    errors.append(
                        f"crafted item '{ckey}': engraving gem_key '{gk}' does not resolve to any gem"
                    )
            for ench in ci.get("enchantments", []) or []:
                if isinstance(ench, dict):
                    eid = ench.get("effect")
                    if isinstance(eid, int):
                        errors.append(
                            f"crafted item '{ckey}' enchantment effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                        )
                    elif isinstance(eid, str) and not resolve_namespaced_ref(
                        eid, namespaced_ref_index
                    ):
                        errors.append(
                            f"crafted item '{ckey}' enchantment effect {eid!r} does not resolve"
                        )
                    if isinstance(eid, str) and not validate_effect_channel(
                        eid, "enchantment", effect_channels
                    ):
                        declared = (
                            effect_channels.get(eid, set())
                            if effect_channels
                            else set()
                        )
                        errors.append(
                            _format_channel_violation(
                                eid,
                                CHANNEL_REQUIREMENTS["enchantment"],
                                declared,
                                f"enchantment on crafted item '{ckey}' ({rel_path})",
                            )
                        )
            for coat in ci.get("coatings", []) or []:
                if isinstance(coat, dict):
                    eid = coat.get("effect")
                    if isinstance(eid, int):
                        errors.append(
                            f"crafted item '{ckey}' coating effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                        )
                    elif isinstance(eid, str) and not resolve_namespaced_ref(
                        eid, namespaced_ref_index
                    ):
                        errors.append(
                            f"crafted item '{ckey}' coating effect {eid!r} does not resolve"
                        )

    elif coll_name == "recipes":
        eq = equipment_index or {"namespaced_keys": set(), "bare_keys": set()}
        for r in data.get("recipes", []) or []:
            if not isinstance(r, dict):
                continue
            rkey = r.get("key", "?")
            for inp in r.get("inputs", []) or []:
                if not isinstance(inp, dict):
                    continue
                itype = inp.get("type")
                ikey = inp.get("key")
                if not isinstance(ikey, str):
                    continue
                if itype == "base_item":
                    if not equipment_ref_valid(ikey, {}, eq):
                        errors.append(
                            f"recipe '{rkey}': input base_item '{ikey}' does not resolve to any equipment"
                        )
                elif itype == "material":
                    if not resolve_namespaced_ref(ikey, namespaced_ref_index):
                        errors.append(
                            f"recipe '{rkey}': input material '{ikey}' does not resolve to any material"
                        )
                elif itype == "gem":
                    if not resolve_namespaced_ref(ikey, namespaced_ref_index):
                        errors.append(
                            f"recipe '{rkey}': input gem '{ikey}' does not resolve to any gem"
                        )
                elif itype == "ingredient":
                    if not resolve_namespaced_ref(ikey, namespaced_ref_index):
                        errors.append(
                            f"recipe '{rkey}': input ingredient '{ikey}' does not resolve to any ingredient"
                        )
                elif itype == "reagent":
                    if not resolve_namespaced_ref(ikey, namespaced_ref_index):
                        errors.append(
                            f"recipe '{rkey}': input reagent '{ikey}' does not resolve to any crafted item"
                        )
            output = r.get("output", {})
            if isinstance(output, dict):
                okey = output.get("crafted_item_key")
                if isinstance(okey, str) and not resolve_namespaced_ref(
                    okey, namespaced_ref_index
                ):
                    errors.append(
                        f"recipe '{rkey}': output crafted_item_key '{okey}' does not resolve to any crafted item"
                    )
            # Validate recipe skill reference
            skill = r.get("skill")
            if isinstance(skill, str) and not resolve_namespaced_ref(
                skill, namespaced_ref_index
            ):
                errors.append(f"recipe '{rkey}': skill '{skill}' does not resolve")

    return errors


def _base_cache():
    """Lazy-load and cache monster base templates.
    Returns dict mapping base key -> merged base data dict.
    """
    cache = {}
    if not MONSTERS_BASES_DIR.exists():
        return cache
    for f in sorted(MONSTERS_BASES_DIR.glob("*.json")):
        try:
            base_key = f.stem
            cache[base_key] = json.loads(f.read_text())
        except (json.JSONDecodeError, OSError):
            pass
    return cache


_BASE_CACHE = None


def _apply_growth(base, level):
    """Compute expected attribute values for a monster at a given level using
    the base template's growth rules.

    Growth is defined per ability as a perLevel rate. The formula:
      attr = round(base_attr + perLevel * (level - base_level)), at least 1 and with no maximum
      (docs/modules/monsters.md#attribute-growth--level-scaling)

    Returns a dict of ability -> int, or None if base has no growth section.
    """
    growth = base.get("growth")
    if not growth:
        return None
    attr_growth = growth.get("attributes", {})
    if not attr_growth:
        return None
    base_level = base.get("level", 1)
    base_attrs = base.get("attributes", {})
    delta = level - base_level
    computed = {}
    for ability in attr_growth:
        base_val = base_attrs.get(ability, 1)
        per = attr_growth[ability].get("perLevel", 0)
        computed[ability] = max(1, round(base_val + per * delta))
    return computed


def _apply_growth_hp(base, level):
    """Compute expected hit points using hitPointsPerLevel growth."""
    growth = base.get("growth")
    if not growth:
        return None
    hp_per_level = growth.get("hitPointsPerLevel")
    if hp_per_level is None:
        return None
    base_level = base.get("level", 1)
    base_hp = base.get("hitPoints", 10)
    delta = level - base_level
    return max(1, round(base_hp + hp_per_level * delta))


def _merge_field(base_val, override_val):
    """Merge a single field from override into base.

    Supports $extend convention: if override_val is a dict with an
    $extend key, the $extend array is appended to the base array.
    Otherwise override replaces base (or is added if base lacks the field).
    """
    if isinstance(override_val, dict) and "$extend" in override_val:
        items = override_val["$extend"]
        if isinstance(base_val, list) and isinstance(items, list):
            return base_val + items
        if isinstance(items, list):
            return list(items)
        return base_val
    return override_val


MONSTER_REF_PREFIX = "core:monster/"
CREATURE_TYPE_REF_PREFIX = "core:creature_type/"
MONSTER_TEMPLATES_FILE = DATA_DIR / "monsters" / "templates.json"
_MONSTER_TEMPLATES = None
# The order the slots apply in, whatever order a row lists its templates (docs/modules/monsters.md#templates).
TEMPLATE_SLOTS = ("element", "power", "nature")
# The damage effects an element template turns into its own element.
ELEMENT_DAMAGE_EFFECTS = {f"core:effect/damage_{kind}" for kind in ("fire", "frost", "shock", "acid", "poison")}


def _monster_templates():
    """Monster templates by key (monsters/templates.json), loaded once."""
    global _MONSTER_TEMPLATES
    if _MONSTER_TEMPLATES is None:
        _MONSTER_TEMPLATES = {}
        if MONSTER_TEMPLATES_FILE.exists():
            for template in json.loads(MONSTER_TEMPLATES_FILE.read_text()).get("monster_templates", []):
                _MONSTER_TEMPLATES[template["key"]] = template
    return _MONSTER_TEMPLATES


def _hp_per_end():
    """HP per point of END, the hp row's per_point (derived/core.json)."""
    rows = json.loads((DATA_DIR / "derived" / "core.json").read_text())["derived"]
    return next(row["per_point"] for row in rows if row["key"] == "hp")


def _grow_damage(merged, step):
    """Add a damage growth step to every attack's bonus and every damage_* ability's magnitude."""
    if not step:
        return
    attacks = []
    for attack in merged.get("attacks", []) or []:
        attack = json.loads(json.dumps(attack))
        damage = attack.setdefault("damage", {})
        damage["bonus"] = damage.get("bonus", 0) + step
        attacks.append(attack)
    if attacks:
        merged["attacks"] = attacks
    abilities = []
    for ability in merged.get("abilities", []) or []:
        ability = dict(ability)
        if str(ability.get("effect", "")).rsplit("/", 1)[-1].startswith("damage_") and "magnitude" in ability:
            ability["magnitude"] += step
        abilities.append(ability)
    if abilities:
        merged["abilities"] = abilities


def compose_row(base, overrides, layers=(), entry=None):
    """One row from a base, layers, and overrides (docs/core/composition.md)."""
    merged = json.loads(json.dumps(base))
    for apply_layer in layers:
        apply_layer(merged)
    for field, value in overrides.items():
        if field == "attributes" and isinstance(value, dict) and isinstance(merged.get(field), dict):
            merged[field] = {**merged[field], **value}
        else:
            merged[field] = _merge_field(merged.get(field), value)
    if entry is not None:
        merged.pop("id", None)
        if entry.get("id") is not None:
            merged["id"] = entry["id"]
        merged["key"] = entry.get("key")
        if entry.get("source"):
            merged["source"] = entry["source"]
    return merged


SPELL_REF_PREFIX = "core:spell/"
SPELL_IDENTITY = ("label", "description", "source")  # key comes from compose_row


def resolve_spell_bases(data, errors):
    """Mutate data in place: a spell with `base` and `overrides` becomes its flat row
    (docs/core/composition.md#a-spell-built-on-a-spell). An object field of the overrides merges
    per field into the base's; the row keeps its own key, label, description, and source. One
    level: a base may not itself name a base."""
    spells = data.get("spells")
    if not isinstance(spells, list):
        return
    rows = {row.get("key"): row for row in spells if isinstance(row, dict)}
    for index, entry in enumerate(spells):
        if not isinstance(entry, dict) or "base" not in entry:
            continue
        key, base_ref = entry.get("key"), entry["base"]
        base = rows.get(base_ref.removeprefix(SPELL_REF_PREFIX)) if base_ref.startswith(SPELL_REF_PREFIX) else None
        if base is None:
            errors.append(f"spell '{key}': base spell '{base_ref}' does not exist")
            continue
        if "base" in base:
            errors.append(f"spell '{key}': base spell '{base_ref}' names a base itself; a spell composes one level")
            continue
        overrides = {}
        for field, value in (entry.get("overrides") or {}).items():
            if isinstance(value, dict) and isinstance(base.get(field), dict):
                value = {**base[field], **value}
            overrides[field] = value
        merged = compose_row(base, overrides, entry=entry)
        for field in SPELL_IDENTITY:
            merged.pop(field, None)
            if field in entry:
                merged[field] = entry[field]
        spells[index] = merged


def load_spells(path):
    """spells/core.json (or a module's spell file) with its spell bases resolved."""
    data = json.loads(Path(path).read_text())
    errors = []
    resolve_spell_bases(data, errors)
    if errors:
        raise ValueError(f"{path}: {errors}")
    return data


def _apply_monster_template(merged, template):
    """Apply one template (docs/modules/monsters.md#templates) to a merged entry in place."""
    if "phase" in template:
        merged["phase"] = template["phase"]
    attacks = []
    for attack in merged.get("attacks", []) or []:
        attack = json.loads(json.dumps(attack))
        if "damage_type" in template and "type" not in attack.get("damage", {}):
            attack.setdefault("damage", {})["type"] = template["damage_type"]
        if "phase" in template:
            attack["phase"] = template["phase"]
        attacks.append(attack)
    if attacks:
        merged["attacks"] = attacks
    if "damage_type" in template:
        element = f"core:effect/damage_{template['damage_type']}"
        abilities = []
        for ability in merged.get("abilities", []) or []:
            if ability.get("effect") in ELEMENT_DAMAGE_EFFECTS:
                ability = {**ability, "effect": element}
            abilities.append(ability)
        if abilities:
            merged["abilities"] = abilities
    if template.get("attributes"):
        attributes = dict(merged.get("attributes") or {})
        for ability, add in template["attributes"].items():
            attributes[ability] = max(1, attributes.get(ability, 0) + add)
        merged["attributes"] = attributes
        if "end" in template["attributes"] and "hitPoints" in merged:
            merged["hitPoints"] = max(1, merged["hitPoints"] + template["attributes"]["end"] * _hp_per_end())
    if template.get("resist"):
        resist = dict(merged.get("resist") or {})
        for kind, amount in template["resist"].items():
            resist[kind] = resist.get(kind, 0) + amount
        merged["resist"] = resist
    if template.get("spells"):
        spells = list(merged.get("spells") or [])
        merged["spells"] = spells + [spell for spell in template["spells"] if spell not in spells]
    if template.get("loot"):
        merged["loot"] = list(merged.get("loot") or []) + template["loot"]
    tags = list(merged.get("tags") or [])
    tags += [tag for tag in template.get("tags", []) if tag not in tags]
    if tags:
        merged["tags"] = tags
    if merged.get("label") and template.get("label_prefix"):  # a summon bonus has no prefix
        merged["label"] = f"{template['label_prefix']} {merged['label']}"


def _template_layers(key, template_keys, filled, errors):
    """The layers for a row's templates in slot order, and the slots they fill; None after an error.

    `filled` maps each slot the row's base monster already fills to its template key."""
    templates = _monster_templates()
    by_slot = {}
    for template_key in template_keys or []:
        if template_key not in templates:
            errors.append(f"monster '{key}': template '{template_key}' not found in monsters/templates.json")
            return None
        slot = templates[template_key].get("slot", "element")
        if slot in by_slot or slot in filled:
            other = by_slot.get(slot) or filled[slot]
            errors.append(f"monster '{key}': templates '{other}' and '{template_key}' are both in the {slot} slot")
            return None
        by_slot[slot] = template_key
    layers = [partial(_apply_monster_template, template=templates[by_slot[slot]]) for slot in TEMPLATE_SLOTS if slot in by_slot]
    return layers, {**filled, **by_slot}


def _resolve_monster_bases(data, errors):
    """Mutate data in-place: replace base+override entries with merged flat entries.

    The base is a creature type (bases/<name>.json) or another monster in the same
    file (core:monster/<key>), resolved first (docs/modules/monsters.md#derived-monsters).
    Then the row's templates apply in order, then its overrides:
      - $extend convention for array fields (abilities, tags)
      - growth-based attribute and HP scaling from the creature type, on the type's
        level for a type base and on the level difference for a base monster
      - full override of any field (replaces base)
    """
    global _BASE_CACHE
    if _BASE_CACHE is None:
        _BASE_CACHE = _base_cache()

    monsters = data.get("monsters")
    if not isinstance(monsters, list):
        return

    rows = {entry.get("key"): entry for entry in monsters if isinstance(entry, dict)}
    resolved = {}

    def resolve(entry, chain):
        """Return (merged entry, creature type data) or None after recording an error."""
        key = entry.get("key")
        if key in resolved:
            return resolved[key]
        base_key = entry.get("base")
        if not base_key:
            resolved[key] = (entry, None, {})
            return resolved[key]
        overrides = entry.get("overrides")
        if not isinstance(overrides, dict):
            errors.append(f"monster '{key}': base '{base_key}' has no overrides object")
            return None

        if base_key.startswith(MONSTER_REF_PREFIX):
            parent_key = base_key[len(MONSTER_REF_PREFIX):]
            if parent_key in chain:
                errors.append(f"monster '{key}': base monster cycle {' -> '.join(chain + [parent_key])}")
                return None
            if parent_key not in rows:
                errors.append(f"monster '{key}': base monster '{base_key}' does not exist")
                return None
            if "level" not in overrides:
                errors.append(f"monster '{key}': a derived monster states its level in overrides")
                return None
            parent = resolve(rows[parent_key], chain + [parent_key])
            if parent is None:
                return None
            parent_entry, type_data, filled = parent
            merged = json.loads(json.dumps(parent_entry))
            for field in ("id", "key", "source"):
                merged.pop(field, None)
            growth = (type_data or {}).get("growth") or {}
            delta = overrides["level"] - parent_entry.get("level", overrides["level"])
            if delta and growth:
                attributes = dict(merged.get("attributes") or {})
                for ability, rule in (growth.get("attributes") or {}).items():
                    if ability in attributes:
                        attributes[ability] = max(1, round(attributes[ability] + rule.get("perLevel", 0) * delta))
                merged["attributes"] = attributes
                if "hitPointsPerLevel" in growth and "hitPoints" in merged:
                    merged["hitPoints"] = max(1, round(merged["hitPoints"] + growth["hitPointsPerLevel"] * delta))
                if delta > 0:
                    _grow_damage(merged, round(growth.get("damagePerLevel", 0) * delta))
        else:
            type_name = base_key.removeprefix(CREATURE_TYPE_REF_PREFIX)
            if not base_key.startswith(CREATURE_TYPE_REF_PREFIX) or type_name not in _BASE_CACHE:
                errors.append(f"monster '{key}': base '{base_key}' not found in bases/")
                return None
            type_data = _BASE_CACHE[type_name]
            filled = {}
            merged = dict(type_data)
            merged.pop("growth", None)
            level = overrides.get("level") or type_data.get("level", 1)
            computed_attrs = _apply_growth(type_data, level)
            if computed_attrs is not None:
                merged["attributes"] = computed_attrs
            computed_hp = _apply_growth_hp(type_data, level)
            if computed_hp is not None:
                merged["hitPoints"] = computed_hp

        planned = _template_layers(key, entry.get("templates"), filled, errors)
        if planned is None:
            return None
        layers, filled = planned

        merged = compose_row(merged, overrides, layers=layers, entry=entry)
        resolved[key] = (merged, type_data, filled)
        return resolved[key]

    for i, entry in enumerate(monsters):
        if not isinstance(entry, dict) or not entry.get("base"):
            continue
        result = resolve(entry, [entry.get("key")])
        if result is not None:
            monsters[i] = result[0]


# The reference category of each equipment file's bases (docs/core/keys.md).
EQUIPMENT_BASE_CATEGORIES = {"weapons.json": "weapon_base", "armor.json": "armor_base", "accessories.json": "accessory_base"}

# Maps equipment data files to their bases/ subdirectories.
EQUIPMENT_BASES_DIRS = {
    "weapons.json": DATA_DIR / "equipment" / "bases" / "weapons",
    "armor.json": DATA_DIR / "equipment" / "bases" / "armor",
    "accessories.json": DATA_DIR / "equipment" / "bases" / "accessories",
}


def equipment_rows(fp, errors=None, equipment=None):
    """An equipment file's rows: each base beside it with `is_item: true` as a row of its own key
    with no overrides, in file-name order, then the file's own rows
    (docs/core/composition.md#a-base-that-is-an-item).

    `equipment` is the file's row list when the caller has read it already. A base that does not
    state `is_item`, or a row that repeats an item base's key, is reported in `errors`.
    """
    if equipment is None:
        equipment = json.loads(fp.read_text()).get("equipment")
    if not isinstance(equipment, list):
        return equipment
    category = EQUIPMENT_BASE_CATEGORIES.get(fp.name)
    base_dir = fp.parent / "bases" / fp.stem
    if not category or not base_dir.is_dir():
        return equipment
    report = errors.append if errors is not None else (lambda _message: None)
    items = []
    for path in sorted(base_dir.glob("*.json")):
        try:
            base = json.loads(path.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        is_item = base.get("is_item") if isinstance(base, dict) else None
        if not isinstance(is_item, bool):
            report(f"bases/{fp.stem}/{path.name}: is_item must be true or false")
        elif is_item:
            items.append({"key": path.stem, "base": f"core:{category}/{path.stem}", "overrides": {}})
    item_keys = {row["key"] for row in items}
    for row in equipment:
        if isinstance(row, dict) and row.get("key") in item_keys:
            report(f"{fp.name}: '{row['key']}' is a duplicate: bases/{fp.stem}/{row['key']}.json is already that item")
    return items + equipment


def _resolve_equipment_bases(data, errors, fp):
    """Mutate data in-place: replace base+override equipment entries with merged flat entries.

    Supports $extend convention for array fields, similar to monster bases.
    Base templates live in equipment/bases/<category>/<key>.json.
    """
    equipment = data.get("equipment")
    if not isinstance(equipment, list):
        return

    base_dir = EQUIPMENT_BASES_DIRS.get(fp.name)
    if not base_dir or not base_dir.exists():
        return

    bases = {}
    for f in sorted(base_dir.glob("*.json")):
        try:
            bases[f.stem] = json.loads(f.read_text())
        except (json.JSONDecodeError, OSError):
            pass

    for i, entry in enumerate(equipment):
        if not isinstance(entry, dict):
            continue
        base_key = entry.get("base")
        if not base_key:
            continue
        overrides = entry.get("overrides")
        if not isinstance(overrides, dict):
            errors.append(f"equipment[{i}]: base '{base_key}' has no overrides object")
            continue
        base_name = base_key.rsplit("/", 1)[-1]
        if not base_key.startswith(f"core:{EQUIPMENT_BASE_CATEGORIES[fp.name]}/") or base_name not in bases:
            cat = fp.name.replace(".json", "")
            errors.append(
                f"equipment[{i}]: base '{base_key}' not found in bases/{cat}/"
            )
            continue

        merged = compose_row(bases[base_name], overrides, entry=entry)
        # default_materials names what a base (or a row's overrides) is found and sold in
        # (docs/core/loot.md#what-lands-on-an-item); loot reads it, the item schema does not.
        merged.pop("default_materials", None)
        # is_item says whether the base is also an item (docs/core/composition.md#a-base-that-is-an-item).
        merged.pop("is_item", None)

        # Validate merged enchantment_slots is an integer
        es = merged.get("enchantment_slots")
        if es is not None and not isinstance(es, int):
            errors.append(
                f"equipment[{i}]: enchantment_slots must be an integer, got {type(es).__name__}"
            )

        # Replace the entry in-place for validation
        equipment[i] = merged


def _validate_single_data_file(
    fp,
    store,
    rel_base,
    equipment_index=None,
    namespaced_ref_index=None,
    effect_channels=None,
):
    """Validate one data file against its $schema and check cross-references.

    Returns (passed, errors, rel_path) where:
      passed  — True on success, False on failure, None if skipped (no $schema)
      errors  — list of error message strings
      rel_path — file path relative to rel_base for display
    """
    rel = fp.relative_to(rel_base)

    try:
        data = json.loads(fp.read_text())
    except json.JSONDecodeError as e:
        return False, [f"JSON {e}"], rel

    # Resolve monster and equipment base+override entries before validation.
    # This merges base templates with override data and replaces
    # base+override entries with flat entries in-place.
    errors = []
    if isinstance(data, dict) and isinstance(data.get("equipment"), list):
        data["equipment"] = equipment_rows(fp, errors, data["equipment"])
    _resolve_monster_bases(data, errors)
    _resolve_equipment_bases(data, errors, fp)
    resolve_spell_bases(data, errors)
    if errors:
        return False, errors, rel

    schema_ref = data.get("$schema")
    if not schema_ref:
        return None, [], rel

    schema_path = (fp.parent / schema_ref).resolve()
    schema_uri = schema_path.as_uri()

    if schema_uri not in store:
        return False, [f"SCHEMA '{schema_path.name}' not in store"], rel

    schema = store[schema_uri]
    resolver = RefResolver(
        base_uri=schema_uri,
        referrer=schema,
        store=store,
    )

    try:
        validate(data, schema, resolver=resolver)
    except ValidationError as e:
        path = (
            "/" + "/".join(str(p) for p in e.absolute_path) if e.absolute_path else "/"
        )
        errors.append(f"SCHEMA {path} {e.message}")

    coll_name = None
    for k, v in data.items():
        if not k.startswith("$") and isinstance(v, list):
            coll_name = k
            break

    if coll_name:
        ref_errors = check_references(
            data,
            coll_name,
            rel_path=str(rel),
            equipment_index=equipment_index,
            namespaced_ref_index=namespaced_ref_index,
            effect_channels=effect_channels,
        )
        errors.extend(f"REF: {e}" for e in ref_errors)

    if errors:
        return False, errors, rel
    return True, [], rel


def validate_modules(store, equipment_index=None, effect_channels=None):
    """Discover and validate modules.
    Returns (module_passed, module_failed, module_total, module_conflicts).
    """
    try:
        from module_loader import (
            discover_modules,
            load_module,
            load_manifest_schema,
            ModuleLoadError,
        )
    except ImportError:
        return 0, 0, 0, []

    if not MODULES_DIR.exists():
        return 0, 0, 0, []

    manifest_schema = load_manifest_schema()
    discovered = discover_modules()
    if not discovered:
        return 0, 0, 0, []

    total = passed = failed = 0
    conflicts = []

    for module_dir, manifest in discovered:
        try:
            mod = load_module(module_dir, manifest, manifest_schema, store)
        except ModuleLoadError as e:
            rel = module_dir.relative_to(ROOT.parent)
            print(f"  \u2717  {rel}/manifest.json")
            print(f"       MANIFEST {e}")
            failed += 1
            # Count each data file as failed too
            for rel_path in manifest.get("data", []):
                fp = (module_dir / rel_path).resolve()
                if fp.exists():
                    rel_fp = fp.relative_to(ROOT.parent)
                    print(f"  \u2717  {rel_fp}")
                    print(f"       MANIFEST (parent module failed)")
                    failed += 1
            continue

        # Validate each data file in the module
        for rel_path in manifest.get("data", []):
            total += 1
            fp = (module_dir / rel_path).resolve()
            if not fp.exists():
                rel = fp.relative_to(ROOT.parent)
                print(f"  \u2717  {rel}")
                print(f"       DATA FILE NOT FOUND: {rel_path}")
                failed += 1
                continue

            p, errs, rel = _validate_single_data_file(
                fp,
                store,
                ROOT.parent,
                equipment_index=equipment_index,
                effect_channels=effect_channels,
            )
            if p is None:
                print(f"  \u26a0  {rel}: no $schema field, skipping")
                total -= 1
                continue
            if p:
                print(f"  \u2713  {rel}")
                passed += 1
            else:
                print(f"  \u2717  {rel}")
                for e in errs:
                    print(f"       {e}")
                failed += 1

    return passed, failed, total, conflicts


def check_racial_trait_refs(data_files):
    """Racial traits live only in features, as type racial_trait (M1-06).

    A race's traits[].id must be a core:feature/<key> ref (module: refs pass)
    to a racial_trait feature, and every core racial_trait must be listed by
    at least one race.
    """
    feature_types = {}
    race_refs = []
    for fp in data_files:
        try:
            data = json.loads(fp.read_text())
        except (json.JSONDecodeError, OSError):
            continue
        if not isinstance(data, dict):
            continue
        for f in data.get("features", []):
            if isinstance(f, dict) and isinstance(f.get("key"), str):
                feature_types[f"core:feature/{f['key']}"] = f.get("type")
        for r in data.get("races", []):
            if not isinstance(r, dict):
                continue
            for t in r.get("traits", []):
                if isinstance(t, dict) and isinstance(t.get("id"), str):
                    race_refs.append((r.get("key", "?"), t["id"]))

    errors = []
    used = set()
    for race_key, ref in race_refs:
        if ref.startswith("module:"):
            continue
        if not ref.startswith("core:feature/"):
            errors.append(
                f"race '{race_key}': trait '{ref}' must be a core:feature/<key> ref"
            )
            continue
        if ref not in feature_types:
            errors.append(f"race '{race_key}': trait '{ref}' does not exist in features")
            continue
        if feature_types[ref] != "racial_trait":
            errors.append(
                f"race '{race_key}': trait '{ref}' has type "
                f"'{feature_types[ref]}', not 'racial_trait'"
            )
            continue
        used.add(ref)
    for ref, ftype in sorted(feature_types.items()):
        if ftype == "racial_trait" and ref not in used:
            errors.append(f"racial_trait '{ref}' is not listed by any race")
    return errors


def main():
    store = load_store()
    data_files = list(collect_data_files())

    # Collect module data files for index builds
    module_data_files = list(collect_module_data_files())

    equipment_index = build_equipment_index(
        data_files, module_data_files, rel_base=ROOT
    )
    namespaced_ref_index = build_namespaced_ref_index(data_files, module_data_files)
    effect_channels = build_effect_channels(data_files, module_data_files)
    total = passed = failed = 0

    equip_errors = check_equipment_uniqueness(
        data_files, module_data_files, rel_base=ROOT
    )
    if equip_errors:
        print("  \u2717  equipment (cross-file id/key uniqueness)")
        for e in equip_errors:
            print(f"       {e}")
        failed += 1

    damage_errors = check_damage_type_vocabulary(
        data_files, load_canonical_damage_types(), rel_base=ROOT
    )
    if damage_errors:
        print("  ✗  damage type vocabulary")
        for e in damage_errors:
            print(f"       {e}")
        failed += 1

    for label, errors in (
        ("references (every core: ref resolves)", check_all_references(data_files, module_data_files)),
        ("effects (used or reserved)", check_effect_usage(data_files, module_data_files)),
        ("skill groups (every skill in one)", check_skill_groups(data_files)),
        ("effect parameters", check_effect_parameters(data_files, module_data_files)),
        ("modifier targets (each names one row)", check_modifier_targets(data_files, module_data_files, collect_vector_files())),
    ):
        if errors:
            print(f"  \u2717  {label}")
            for e in errors:
                print(f"       {e}")
            failed += 1

    trait_errors = check_racial_trait_refs(data_files)
    if trait_errors:
        print("  \u2717  racial traits (features of type racial_trait)")
        for e in trait_errors:
            print(f"       {e}")
        failed += 1

    for fp in data_files:
        p, errs, rel = _validate_single_data_file(
            fp,
            store,
            ROOT,
            equipment_index=equipment_index,
            namespaced_ref_index=namespaced_ref_index,
            effect_channels=effect_channels,
        )
        if p is None:
            print(f"  \u26a0  {rel}: no $schema field, skipping")
            continue
        if not p and errs and "not in store" in errs[0]:
            # Schema not found — fail but don't count in total
            print(f"  \u2717  {rel}: {errs[0]}")
            failed += 1
            continue
        total += 1
        if p:
            print(f"  \u2713  {rel}")
            passed += 1
        else:
            print(f"  \u2717  {rel}")
            for e in errs:
                print(f"       {e}")
            failed += 1

    # Module validation pass
    mod_passed, mod_failed, mod_total, mod_conflicts = validate_modules(
        store, equipment_index=equipment_index, effect_channels=effect_channels
    )

    combined_passed = passed + mod_passed
    combined_failed = failed + mod_failed
    combined_total = total + mod_total

    print()
    if mod_total > 0:
        print(
            f"  Modules: {mod_passed} passed, {mod_failed} failed (of {mod_total} module data files)"
        )
    print(
        f"{combined_total} file(s): {combined_passed} passed, {combined_failed} failed"
    )

    return 1 if combined_failed else 0


if __name__ == "__main__":
    with warnings.catch_warnings():
        warnings.filterwarnings("ignore", category=DeprecationWarning)
        sys.exit(main())
