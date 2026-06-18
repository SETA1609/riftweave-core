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


def load_store():
    store = {}
    for f in sorted(SCHEMA_DIR.glob("*.schema.json")):
        store[f.as_uri()] = json.loads(f.read_text())
    return store


def collect_data_files():
    for dirpath, _, filenames in os.walk(DATA_DIR):
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


def build_id_index(data_files, module_data_files=None):
    """Build a map of collection name -> set of integer ids for referential integrity."""
    id_index = {}
    for fp in data_files:
        try:
            data = json.loads(fp.read_text())
            for k, v in data.items():
                if k.startswith("$") or not isinstance(v, list):
                    continue
                id_index.setdefault(k, set())
                for item in v:
                    if isinstance(item, dict):
                        iid = item.get("id")
                        if isinstance(iid, int):
                            id_index[k].add(iid)
        except Exception:
            pass  # ignore malformed for index
    # Also index module data files
    if module_data_files:
        for fp, _ in module_data_files:
            try:
                data = json.loads(fp.read_text())
                for k, v in data.items():
                    if k.startswith("$") or not isinstance(v, list):
                        continue
                    id_index.setdefault(k, set())
                    for item in v:
                        if isinstance(item, dict):
                            iid = item.get("id")
                            if isinstance(iid, int):
                                id_index[k].add(iid)
            except Exception:
                pass
    return id_index


def check_references(data, coll_name, id_index):
    """Return list of referential integrity error messages."""
    errors = []
    if coll_name == "races":
        for r in data.get("races", []):
            if not isinstance(r, dict):
                continue
            lineage = r.get("lineage", {})
            if "parentRace" in lineage:
                pid = lineage["parentRace"]
                if not isinstance(pid, int) or pid not in id_index.get("races", set()):
                    errors.append(f"parentRace {pid} does not exist in races")
            for t in r.get("traits", []):
                if isinstance(t, dict):
                    tid = t.get("id")
                    if isinstance(tid, int) and tid not in id_index.get(
                        "traits", set()
                    ):
                        errors.append(f"trait {tid} does not exist in traits")

    elif coll_name == "features":
        for f in data.get("features", []):
            if not isinstance(f, dict):
                continue
            pr = f.get("prerequisite", {})
            for pid in pr.get("perks", []):
                if isinstance(pid, int) and pid not in id_index.get("features", set()):
                    errors.append(f"prereq perk {pid} does not exist")
            for s in pr.get("skills", []):
                if isinstance(s, dict):
                    sid = s.get("id")
                    if isinstance(sid, int) and sid not in id_index.get(
                        "skills", set()
                    ):
                        errors.append(f"prereq skill {sid} does not exist")

    elif coll_name == "spells":
        for s in data.get("spells", []):
            if not isinstance(s, dict):
                continue
            for e in s.get("effects", []):
                if isinstance(e, dict):
                    eid = e.get("effect")
                    if isinstance(eid, int) and eid not in id_index.get(
                        "effects", set()
                    ):
                        errors.append(f"effect {eid} does not exist")
            for rid in (s.get("cost", {}) or {}).get("reagents", []) or []:
                if isinstance(rid, int) and rid not in id_index.get(
                    "ingredients", set()
                ):
                    errors.append(f"reagent {rid} does not exist")

    elif coll_name == "backgrounds":
        for b in data.get("backgrounds", []):
            if not isinstance(b, dict):
                continue
            for sb in b.get("skill_bonuses", []) or []:
                if isinstance(sb, dict):
                    sid = sb.get("skill")
                    if isinstance(sid, int) and sid not in id_index.get(
                        "skills", set()
                    ):
                        errors.append(f"skill bonus {sid} does not exist")
            for ts in b.get("suggested_tag_skills", []) or []:
                if isinstance(ts, int) and ts not in id_index.get("skills", set()):
                    errors.append(f"suggested tag skill {ts} does not exist")
            for ss in b.get("starting_spells", []) or []:
                if isinstance(ss, int) and ss not in id_index.get("spells", set()):
                    errors.append(f"starting spell {ss} does not exist")
            for se in b.get("starting_equipment", []) or []:
                if isinstance(se, dict):
                    iid = se.get("item")
                    if isinstance(iid, int) and iid not in id_index.get(
                        "equipment", set()
                    ):
                        errors.append(f"starting equipment {iid} does not exist")

    elif coll_name == "monsters":
        for m in data.get("monsters", []):
            if not isinstance(m, dict):
                continue
            for a in m.get("abilities", []) or []:
                if isinstance(a, dict):
                    eid = a.get("effect")
                    if isinstance(eid, int) and eid not in id_index.get(
                        "effects", set()
                    ):
                        errors.append(f"ability effect {eid} does not exist")

    elif coll_name == "ingredients":
        for i in data.get("ingredients", []):
            if not isinstance(i, dict):
                continue
            for eid in i.get("effects", []) or []:
                if isinstance(eid, int) and eid not in id_index.get("effects", set()):
                    errors.append(f"effect {eid} does not exist")

    elif coll_name == "conditions":
        for c in data.get("conditions", []):
            if not isinstance(c, dict):
                continue
            for eid in c.get("appliedBy", []) or []:
                if isinstance(eid, int) and eid not in id_index.get("effects", set()):
                    errors.append(f"appliedBy effect {eid} does not exist in effects")
            for eid in c.get("removedBy", []) or []:
                if isinstance(eid, int) and eid not in id_index.get("effects", set()):
                    errors.append(f"removedBy effect {eid} does not exist in effects")

    elif coll_name == "equipment":
        for e in data.get("equipment", []) or []:
            if not isinstance(e, dict):
                continue
            cons = e.get("consumable") or {}
            for eff in cons.get("effects", []) or []:
                if isinstance(eff, dict):
                    eid = eff.get("effect")
                    if isinstance(eid, int) and eid not in id_index.get(
                        "effects", set()
                    ):
                        errors.append(f"effect {eid} does not exist")

    return errors


def validate_modules(store, id_index):
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

            try:
                data = json.loads(fp.read_text())
            except json.JSONDecodeError as e:
                rel = fp.relative_to(ROOT.parent)
                print(f"  \u2717  {rel}")
                print(f"       JSON {e}")
                failed += 1
                continue

            schema_ref = data.get("$schema")
            if not schema_ref:
                rel = fp.relative_to(ROOT.parent)
                print(f"  \u26a0  {rel}: no $schema field, skipping")
                # Not counting as failed, but not passing either
                total -= 1
                continue

            schema_path = (fp.parent / schema_ref).resolve()
            schema_uri = schema_path.as_uri()

            if schema_uri not in store:
                rel = fp.relative_to(ROOT.parent)
                print(f"  \u2717  {rel}")
                print(f"       SCHEMA '{schema_path.name}' not in store")
                failed += 1
                continue

            schema = store[schema_uri]

            resolver = RefResolver(
                base_uri=schema_uri,
                referrer=schema,
                store=store,
            )

            file_errors = []
            try:
                validate(data, schema, resolver=resolver)
            except ValidationError as e:
                path = (
                    "/" + "/".join(str(p) for p in e.absolute_path)
                    if e.absolute_path
                    else "/"
                )
                file_errors.append(f"SCHEMA {path} {e.message}")

            # Determine collection name for ref checks
            coll_name = None
            for k, v in data.items():
                if not k.startswith("$") and isinstance(v, list):
                    coll_name = k
                    break

            if coll_name:
                ref_errors = check_references(data, coll_name, id_index)
                file_errors.extend(f"REF: {e}" for e in ref_errors)

            rel = fp.relative_to(ROOT.parent)
            if not file_errors:
                print(f"  \u2713  {rel}")
                passed += 1
            else:
                print(f"  \u2717  {rel}")
                for e in file_errors:
                    print(f"       {e}")
                failed += 1

    return passed, failed, total, conflicts


def main():
    store = load_store()
    data_files = list(collect_data_files())

    # Collect module data files for ID index build
    module_data_files = list(collect_module_data_files())

    # Build ID index from core + module data for referential integrity
    id_index = build_id_index(data_files, module_data_files)

    total = passed = failed = 0

    for fp in data_files:
        data = json.loads(fp.read_text())
        schema_ref = data.get("$schema")
        if not schema_ref:
            rel = fp.relative_to(ROOT)
            print(f"  \u26a0  {rel}: no $schema field, skipping")
            continue

        schema_path = (fp.parent / schema_ref).resolve()
        schema_uri = schema_path.as_uri()

        if schema_uri not in store:
            rel = fp.relative_to(ROOT)
            print(f"  \u2717  {rel}: schema '{schema_path.name}' not in store")
            failed += 1
            continue

        total += 1
        schema = store[schema_uri]

        resolver = RefResolver(
            base_uri=schema_uri,
            referrer=schema,
            store=store,
        )

        errors = []
        try:
            validate(data, schema, resolver=resolver)
        except ValidationError as e:
            rel = fp.relative_to(ROOT)
            path = (
                "/" + "/".join(str(p) for p in e.absolute_path)
                if e.absolute_path
                else "/"
            )
            errors.append(f"SCHEMA {path} {e.message}")

        # Determine collection name (top-level array key)
        coll_name = None
        for k, v in data.items():
            if not k.startswith("$") and isinstance(v, list):
                coll_name = k
                break

        if coll_name:
            ref_errors = check_references(data, coll_name, id_index)
            errors.extend(f"REF: {e}" for e in ref_errors)

        rel = fp.relative_to(ROOT)
        if not errors:
            print(f"  \u2713  {rel}")
            passed += 1
        else:
            print(f"  \u2717  {rel}")
            for e in errors:
                print(f"       {e}")
            failed += 1

    # Module validation pass
    mod_passed, mod_failed, mod_total, mod_conflicts = validate_modules(store, id_index)

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
