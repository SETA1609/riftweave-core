#!/usr/bin/env python3
"""Validate all JSON data files against their $schema references.

If modules exist under ruleset/modules/, each module's manifest and data files
are also validated. Module data files share the same schema store as core data.
"""

import json
import os
import sys
import warnings
from pathlib import Path

from jsonschema import validate, ValidationError, RefResolver

ROOT = Path(__file__).resolve().parent.parent
DATA_DIR = ROOT / "data"
SCHEMA_DIR = ROOT / "schemas"
MODULES_DIR = ROOT / "modules"
MONSTERS_BASES_DIR = DATA_DIR / "monsters" / "bases"

# Maps equipment source filenames to namespaced reference categories.
# Cross-references use core:<category>/<key> with singular category names
# (e.g. core:weapon/shortsword, core:armor/leather_jerkin, core:consumable/potion_minor_healing).
EQUIPMENT_FILE_CATEGORIES = {
    "weapons.json": "weapon",
    "armor.json": "armor",
    "consumables.json": "consumable",
}

# Maps collection names to their singular category for namespaced refs.
COLLECTION_CATEGORIES = {
    "effects": "effect",
    "spells": "spell",
    "ingredients": "ingredient",
    "skills": "skill",
    "features": "feature",
    "races": "race",
    "traits": "trait",
    "conditions": "condition",
    "materials": "material",
    "gems": "gem",
    "monsters": "monster",
    "backgrounds": "background",
    "recipes": "recipe",
    "crafted_items": "crafted_item",
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
        equipment = data.get("equipment")
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


def check_references(data, coll_name, equipment_index=None, namespaced_ref_index=None):
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
                        errors.append(f"trait {tid} does not exist in traits")

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
            for eid in i.get("effects", []) or []:
                if isinstance(eid, int):
                    errors.append(
                        f"effect {eid}: numeric refs not allowed (use core:effect/<key>)"
                    )
                elif isinstance(eid, str) and not resolve_namespaced_ref(
                    eid, namespaced_ref_index
                ):
                    errors.append(f"effect {eid!r} does not resolve")

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

    elif coll_name == "crafted_items":
        for ci in data.get("crafted_items", []) or []:
            if not isinstance(ci, dict):
                continue
            ckey = ci.get("key", "?")
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
      attr = round(base_attr + perLevel * (level - base_level))

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


def _resolve_monster_bases(data, errors):
    """Mutate data in-place: replace base+override entries with merged flat entries.

    Supports:
      - $extend convention for array fields (abilities, tags)
      - Growth-based attribute and HP scaling from base templates
      - Full override of any field (replaces base)
    """
    global _BASE_CACHE
    if _BASE_CACHE is None:
        _BASE_CACHE = _base_cache()

    monsters = data.get("monsters")
    if not isinstance(monsters, list):
        return

    for i, entry in enumerate(monsters):
        if not isinstance(entry, dict):
            continue
        base_key = entry.get("base")
        if not base_key:
            continue
        overrides = entry.get("overrides")
        if not isinstance(overrides, dict):
            errors.append(f"monsters[{i}]: base '{base_key}' has no overrides object")
            continue
        if base_key not in _BASE_CACHE:
            errors.append(f"monsters[{i}]: base '{base_key}' not found in bases/")
            continue

        base_data = _BASE_CACHE[base_key]
        merged = dict(base_data)

        # Remove growth blueprint from merged entry (it's not a valid monster field)
        merged.pop("growth", None)

        # Apply growth-based attribute scaling
        level = overrides.get("level") or base_data.get("level", 1)
        computed_attrs = _apply_growth(base_data, level)
        if computed_attrs is not None:
            merged["attributes"] = computed_attrs

        # Apply growth-based HP scaling
        computed_hp = _apply_growth_hp(base_data, level)
        if computed_hp is not None:
            merged["hitPoints"] = computed_hp

        # Merge override fields on top (individual attribute overrides merge
        # into the computed set rather than replacing the whole object)
        for k, v in overrides.items():
            if (
                k == "attributes"
                and isinstance(v, dict)
                and isinstance(merged.get("attributes"), dict)
            ):
                merged["attributes"].update(v)
            else:
                merged[k] = _merge_field(merged.get(k), v)

        # Carry over id and key from the entry wrapper
        merged["id"] = entry.get("id")
        merged["key"] = entry.get("key")
        if entry.get("source"):
            merged["source"] = entry["source"]
        # Replace the entry in-place for validation
        monsters[i] = merged


def _validate_single_data_file(
    fp, store, rel_base, equipment_index=None, namespaced_ref_index=None
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

    # Resolve monster base+override entries before validation.
    # This merges base templates with override data and replaces
    # base+override entries with flat monster entries in-place.
    errors = []
    _resolve_monster_bases(data, errors)
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
            equipment_index=equipment_index,
            namespaced_ref_index=namespaced_ref_index,
        )
        errors.extend(f"REF: {e}" for e in ref_errors)

    if errors:
        return False, errors, rel
    return True, [], rel


def validate_modules(store, equipment_index=None):
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
                fp, store, ROOT.parent, equipment_index=equipment_index
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


def main():
    store = load_store()
    data_files = list(collect_data_files())

    # Collect module data files for index builds
    module_data_files = list(collect_module_data_files())

    equipment_index = build_equipment_index(
        data_files, module_data_files, rel_base=ROOT
    )
    namespaced_ref_index = build_namespaced_ref_index(data_files, module_data_files)

    total = passed = failed = 0

    equip_errors = check_equipment_uniqueness(
        data_files, module_data_files, rel_base=ROOT
    )
    if equip_errors:
        print("  \u2717  equipment (cross-file id/key uniqueness)")
        for e in equip_errors:
            print(f"       {e}")
        failed += 1

    for fp in data_files:
        p, errs, rel = _validate_single_data_file(
            fp,
            store,
            ROOT,
            equipment_index=equipment_index,
            namespaced_ref_index=namespaced_ref_index,
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
        store, equipment_index=equipment_index
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
