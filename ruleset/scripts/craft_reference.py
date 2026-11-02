#!/usr/bin/env python3
"""Crafting module reference (docs/modules/crafting.md, "The module slice").

Reads module data and an actor snapshot; returns the outcome of one craft attempt. It stores
nothing. The check is the kernel's resolution function; the grade and the hours are formula
trees in data/quality.json.

Usage:
    python ruleset/scripts/craft_reference.py   # runs ruleset/modules/crafting/tests/vectors.json
"""

import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from kernel_reference import evaluate_formula, run_resolution  # noqa: E402

RULESET = Path(__file__).resolve().parent.parent
MODULE_DIR = RULESET / "modules" / "crafting"
VECTORS = MODULE_DIR / "tests" / "vectors.json"
STATION_PREFIX = "module:crafting/station/"
ITEM_FILES = {
    "weapon": ("equipment/weapons.json", "weapons"),
    "armor": ("equipment/armor.json", "armor"),
    "accessory": ("equipment/accessories.json", "accessories"),
    "consumable": ("equipment/consumables.json", None),
}
COLLECTION_FILES = {
    "material": ("materials/core.json", "materials"),
    "ingredient": ("ingredients/core.json", "ingredients"),
    "gem": ("gems/core.json", "gems"),
    "skill": ("skills/core.json", "skills"),
}


def load_module():
    merged = {}
    for name in ("stations", "quality", "recipes"):
        data = json.loads((MODULE_DIR / "data" / f"{name}.json").read_text())
        merged.update({k: v for k, v in data.items() if not k.startswith("$")})
    return merged


def serves(have, need_ref, stations):
    """True if the station `have` is, or counts as, the station `need_ref`."""
    if have == need_ref:
        return True
    return need_ref in stations[have]["counts_as"]


def craft(module, recipe_key, actor, station_key, roll):
    quality = module["quality"]
    ladder = [g["grade"] for g in quality["grades"]]
    stations = {STATION_PREFIX + s["key"]: s for s in module["stations"]}
    recipe = next(r for r in module["crafting_recipes"] if r["key"] == recipe_key)
    if station_key not in stations or not serves(station_key, recipe["station"], stations):
        return {"outcome": "no_station", "consumed_percent": 0}
    station = stations[station_key]

    check = run_resolution(
        {
            "actor": {
                "skills": {"craft": actor["skills"][recipe["skill"]]},
                "attributes": actor.get("attributes", {}),
            },
            "modifiers": [{"op": "add", "value": station["skill_bonus"] - recipe["difficulty"]}],
            "roll": roll,
        }
    )
    if "fumble" in check["tags"]:
        return {"outcome": "ruined", "target": check["target"], "consumed_percent": 100}
    if not check["success"]:
        return {"outcome": "failed", "target": check["target"], "consumed_percent": quality["fail_loss_percent"]}

    variables = {
        "base": ladder.index(quality["base_grade"]),
        "margin": check["margin"],
        "crit_bonus": 1 if "crit" in check["tags"] else 0,
    }
    index = evaluate_formula(quality["grade_index"], {}, variables)
    index = max(0, min(index, ladder.index(station["max_grade"])))
    hours = evaluate_formula(
        quality["hours"], {}, {"hours": recipe["time"], "time_percent": station["time_percent"]}
    )
    return {
        "outcome": "made",
        "target": check["target"],
        "margin": check["margin"],
        "grade_index": index,
        "grade": ladder[index],
        "value_percent": quality["grades"][index]["value_percent"],
        "hours": max(quality["min_hours"], hours),
        "consumed_percent": 100,
        "output": recipe["output"],
    }


def run_vector(case):
    return craft(load_module(), case["recipe"], case["actor"], case["station"], case["roll"])


def _core_keys(rel_path, array):
    data = json.loads((RULESET / "data" / rel_path).read_text())
    return {e["key"]: e for e in data[array]}


def _equipment_keys():
    keys = {}
    for category, (rel_path, folder) in ITEM_FILES.items():
        keys[category] = _core_keys(rel_path, "equipment")
    return keys


def _valid_materials(category, entry):
    """valid_materials of the entry's base template, or None when it has no base."""
    folder = ITEM_FILES[category][1]
    base = entry.get("base")
    path = RULESET / "data" / "equipment" / "bases" / (folder or "") / f"{base}.json"
    if not base or not folder or not path.exists():
        return None
    return json.loads(path.read_text()).get("valid_materials")


def check_module():
    """Return a list of problems the schema cannot express."""
    module = load_module()
    problems = []
    schema = json.loads((RULESET / "schemas" / "schema.json").read_text())
    enum = schema["definitions"]["qualityGrade"]["enum"]
    ladder = [g["grade"] for g in module["quality"]["grades"]]
    if ladder != enum:
        problems.append(f"quality ladder {ladder} is not the shared ladder {enum}")
    if module["quality"]["base_grade"] not in ladder:
        problems.append("base_grade is not on the ladder")

    stations = {STATION_PREFIX + s["key"]: s for s in module["stations"]}
    if len(stations) != len(module["stations"]):
        problems.append("duplicate station key")
    for ref, s in stations.items():
        if s["max_grade"] not in ladder:
            problems.append(f"{ref}: max_grade is not on the ladder")
        for other in s["counts_as"]:
            if other not in stations or other == ref:
                problems.append(f"{ref}: bad counts_as {other}")

    equipment = _equipment_keys()
    collections = {c: _core_keys(*files) for c, files in COLLECTION_FILES.items()}
    recipes = module["crafting_recipes"]
    for field in ("id", "key"):
        values = [r[field] for r in recipes]
        if len(values) != len(set(values)):
            problems.append(f"duplicate recipe {field}")

    def resolve(ref):
        _, rest = ref.split(":", 1)
        category, key = rest.split("/", 1)
        pool = equipment.get(category) or collections.get(category) or {}
        return category, pool.get(key)

    for r in recipes:
        where = f"recipe {r['key']}"
        if r["station"] not in stations:
            problems.append(f"{where}: unknown station {r['station']}")
        if resolve(r["skill"])[1] is None:
            problems.append(f"{where}: unknown skill {r['skill']}")
        for entry in r["inputs"]:
            if resolve(entry["item"])[1] is None:
                problems.append(f"{where}: unknown input {entry['item']}")
        category, item = resolve(r["output"]["item"])
        if item is None:
            problems.append(f"{where}: unknown output {r['output']['item']}")
            continue
        material = r["output"].get("material")
        if material:
            if resolve(material)[1] is None:
                problems.append(f"{where}: unknown material {material}")
            valid = _valid_materials(category, item)
            if valid is not None and material.split("/")[-1] not in valid:
                problems.append(f"{where}: {material} is not a valid material for {r['output']['item']}")
        elif category in ("weapon", "armor", "accessory"):
            problems.append(f"{where}: gear needs an output material")
    return problems


def main():
    failures = 0
    for case in json.loads(VECTORS.read_text()):
        result = run_vector(case)
        wrong = {k: (v, result.get(k)) for k, v in case["expect"].items() if result.get(k) != v}
        if wrong:
            failures += 1
            print(f"FAIL {case['name']}: {wrong}")
    problems = check_module()
    for p in problems:
        print(f"DATA {p}")
    print(f"{failures} vector failure(s), {len(problems)} data problem(s)")
    return 1 if failures or problems else 0


if __name__ == "__main__":
    sys.exit(main())
