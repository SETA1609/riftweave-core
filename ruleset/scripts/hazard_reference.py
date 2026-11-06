#!/usr/bin/env python3
"""Python reference for hazards (docs/core/hazards.md).

Detect, disarm, and avoid are checks: one call of the resolution function through a check
row, with the hazard's difficulty as an add on the target. A trigger emits one packet and the
effect list; the consumer applies DR, resist, and conditions. Rolls are inputs.
"""

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def hazard_row(key):
    return next(h for h in load_data("hazards/core.json")["hazards"] if h["key"] == key)


def check_skill(ref):
    """The skill key a check row rolls (core:check/<key> -> its core:skill/<key>)."""
    key = ref.rsplit("/", 1)[-1]
    row = next(c for c in load_data("checks/core.json")["checks"] if c["key"] == key)
    return row["skill"].rsplit("/", 1)[-1]


def run_check(spec, vector):
    """One check through the resolution function; the difficulty is an add on the target."""
    from kernel_reference import run_resolution  # kernel_reference imports this module

    skill = check_skill(spec["check"])
    actor = vector["actor"]
    modifiers = [m for m in vector.get("modifiers", []) if m["target"] == skill]
    if spec.get("difficulty"):
        modifiers.append({"target": skill, "op": "add", "value": spec["difficulty"], "source": f"hazard:{vector['hazard']}"})
    return run_resolution(
        {
            "actor": {"skills": {skill: actor["skills"][skill]}, "attributes": actor.get("attributes", {})},
            "skill": skill,
            "modifiers": modifiers,
            "roll": vector["roll"],
        }
    )


def run_hazard(vector):
    hazard = hazard_row(vector["hazard"])
    op = vector["op"]
    if op == "detect":
        return run_check(hazard["detect"], vector)
    if op == "disarm":
        if "disarm" not in hazard:
            return {"error": f"{hazard['key']} cannot be disarmed"}
        result = run_check(hazard["disarm"], vector)
        return {**result, "triggered": "fumble" in result["tags"]}
    if op != "trigger":
        return {"error": f"op {op!r} is not defined"}

    avoided, half = False, False
    if "avoid" in hazard and "roll" in vector:
        avoided = run_check(hazard["avoid"], vector)["success"]
        half = avoided and hazard["avoid"]["on_avoid"] == "half"
    packet = None
    if "damage" in hazard and (not avoided or half):
        amount = vector.get("damage_roll", 0)
        packet = {"amount": amount // 2 if half else amount, "type": hazard["damage"]["type"]}
    effects = [] if avoided else list(hazard.get("effects", []))
    return {"avoided": avoided, "packet": packet, "effects": effects}
