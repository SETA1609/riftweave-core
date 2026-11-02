#!/usr/bin/env python3
"""Python reference for the kernel functions the golden vectors check.

One runner per kernel function, keyed by name in RUNNERS. Coefficients are read from
ruleset/data/, so a data change that breaks a rule fails a vector. The input and result
of each runner are listed in docs/core/vectors.md.
"""

import json
from pathlib import Path

RULESET = Path(__file__).resolve().parent.parent
DATA_DIR = RULESET / "data"


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def apply_modifiers(value, modifiers):
    """Contract stacking for add and mul: same source, latest wins; add, then mul."""
    latest_by_source = {}
    for modifier in modifiers:
        op = modifier["op"]
        if op not in ("add", "mul"):
            raise NotImplementedError(f"modifier op {op!r} has no reference yet (M1-05)")
        latest_by_source[(op, modifier["source"])] = modifier["value"]
    for (op, _), amount in latest_by_source.items():
        if op == "add":
            value += amount
    for (op, _), amount in latest_by_source.items():
        if op == "mul":
            value *= amount
    return value


def run_modifiers(vector):
    return {"value": apply_modifiers(vector["value"], vector.get("modifiers", []))}


def run_resolution(vector):
    rules = load_data("resolution/core.json")["resolution"]
    actor = vector["actor"]
    skills = actor["skills"]
    skill = vector.get("skill")
    if skill is None:
        if len(skills) != 1:
            raise ValueError("actor has several skills; name one with `skill`")
        (skill,) = skills
    bonus = sum(m["value"] for m in vector.get("modifiers", []) if m["op"] == "add")
    target = skills[skill] + bonus
    if rules["target_cap"] is not None:
        target = min(target, rules["target_cap"])
    roll = vector["roll"]
    fumble_at = rules["fumble_at"]
    success = roll <= target and roll < fumble_at
    luck = actor.get("attributes", {}).get("lck", 0)
    window = rules["crit_window"] + luck // rules["crit_window_per_two_luck"]
    tags = []
    if success and roll <= window:
        tags.append("crit")
    if roll >= fumble_at:
        tags.append("fumble")
    return {
        "target": target,
        "success": success,
        "margin": target - roll if success else None,
        "tags": tags,
    }


def run_derived(vector):
    actor = vector["actor"]
    attributes = actor.get("attributes", {})
    level = actor.get("level", 1)
    modifiers = vector.get("modifiers", [])
    result = {}
    for stat in load_data("derived/core.json")["derived"]:
        attr = attributes.get(stat["attr"], 0)
        attr_term = (attr * stat["per_point"]) // stat.get("per_point_divisor", 1)
        value = (
            stat["base"]
            + attr_term
            + level * (attr * stat["per_level_attr"] + stat["per_level_flat"])
        )
        own = [m for m in modifiers if m.get("target") == stat["key"]]
        result[stat["key"]] = apply_modifiers(value, own)
    return result


def evaluate_formula(node, attributes):
    """Evaluate a formula tree: lit, attr, and the ops add, sub, mul, div (floor)."""
    if "lit" in node:
        return node["lit"]
    if "attr" in node:
        return attributes[node["attr"]]
    op = node["op"]
    args = [evaluate_formula(arg, attributes) for arg in node["args"]]
    if op == "add":
        return sum(args)
    if op == "sub":
        first, *rest = args
        return first - sum(rest)
    if op == "mul":
        product = 1
        for arg in args:
            product *= arg
        return product
    if op == "div":
        numerator, denominator = args
        return numerator // denominator
    raise NotImplementedError(f"formula op {op!r} has no reference")


def run_formula(vector):
    if "expr" in vector:
        expr = vector["expr"]
    else:
        expr = load_data(f"formulas/{vector['formula']}.json")["expr"]
    attributes = vector.get("actor", {}).get("attributes", {})
    return {"value": evaluate_formula(expr, attributes)}


def run_actions(vector):
    data = load_data("actions/core.json")
    pool_rules = data["pool"]
    costs = {action["key"]: action["cost"] for action in data["actions"]}
    attr = vector["actor"]["attributes"][pool_rules["attr"]]
    pool = pool_rules["base"] + (attr * pool_rules["per_point"]) // pool_rules["per_point_divisor"]
    cost = sum(costs[key] for key in vector.get("actions", []))
    return {"pool": pool, "cost": cost}


RUNNERS = {
    "resolution": run_resolution,
    "derived": run_derived,
    "formula": run_formula,
    "modifiers": run_modifiers,
    "actions": run_actions,
}


def run_vector(function, vector):
    return RUNNERS[function](vector)
