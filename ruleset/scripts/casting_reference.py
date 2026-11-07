#!/usr/bin/env python3
"""Python reference for casting (docs/core/casting.md).

One cast: the gate, the mana and reagent cost, one call of the resolution function on the
spell's <color>_magic skill with the armor penalty, then the effects, concentration, and
dispel. The roll is an input.
"""

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def spell_row(spells, key):
    for spell in spells:
        if spell["key"] == key:
            return spell
    raise LookupError(f'spell "{key}" is not defined')


def run_casting(vector):
    from kernel_reference import run_resolution  # kernel_reference imports this module

    rules = load_data("casting/core.json")["casting"]
    spells = load_data("spells/core.json")["spells"]
    try:
        spell = spell_row(spells, vector["spell"])
        standing = [spell_row(spells, key) for key in vector.get("standing", [])]
    except LookupError as error:
        return {"error": str(error)}
    caster = vector["caster"]
    skill = f"{spell['color']}_magic"
    if skill not in caster["skills"]:
        return {"error": f"casting {spell['key']} needs {skill}"}

    mana = caster["mana"]
    held = caster.get("concentration")
    result = {
        "outcome": "",
        "mana": mana,
        "consumed": [],
        "check": None,
        "effects": [],
        "concentration": held,
        "ended": None,
        "dispelled": [],
    }
    cost = spell.get("cost", {})
    reagents = list(caster.get("reagents", []))
    if any(flag in rules["prevented_by"] for flag in caster.get("flags", [])):
        return {**result, "outcome": "prevented"}
    if mana < cost.get("mana", 0):
        return {**result, "outcome": "no_mana"}
    for ref in cost.get("reagents", []):
        if ref not in reagents:
            return {**result, "outcome": "no_reagents"}
        reagents.remove(ref)
    result["mana"] = mana - cost.get("mana", 0)
    result["consumed"] = list(cost.get("reagents", []))

    modifiers = [m for m in vector.get("modifiers", []) if m["target"] == skill]
    armor = caster.get("armor", "none")
    if armor not in rules["armor_penalty"]:
        return {"error": f'armor category "{armor}" is not defined'}
    if rules["armor_penalty"][armor]:
        modifiers.append({"target": skill, "op": "add", "value": rules["armor_penalty"][armor], "source": f"armor:{armor}"})
    check = run_resolution(
        {
            "actor": {"skills": {skill: caster["skills"][skill]}, "attributes": caster.get("attributes", {})},
            "skill": skill,
            "modifiers": modifiers,
            "roll": vector["roll"],
        }
    )
    result["check"] = check
    if not check["success"]:
        return {**result, "outcome": "failed"}

    result["outcome"] = "cast"
    result["effects"] = [dict(effect) for effect in spell["effects"]]
    if spell.get("concentration"):
        result["concentration"] = spell["key"]
        if held is not None and held != spell["key"]:
            result["ended"] = held
    for effect in spell["effects"]:
        if effect["effect"] == "core:effect/dispel":
            for row in standing:
                if row["level"] <= effect["magnitude"] and row["key"] not in result["dispelled"]:
                    result["dispelled"].append(row["key"])
    return result
