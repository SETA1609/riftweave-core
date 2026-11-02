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


def evaluate_formula(node, attributes, variables=None):
    """Evaluate a formula tree: lit, attr, var, and the ops add, sub, mul, div (floor)."""
    if "lit" in node:
        return node["lit"]
    if "attr" in node:
        return attributes[node["attr"]]
    if "var" in node:
        return (variables or {})[node["var"]]
    op = node["op"]
    args = [evaluate_formula(arg, attributes, variables) for arg in node["args"]]
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


def run_rest(vector):
    """Apply one rest occasion: refill or fraction per resource, then clear timed conditions."""
    occasion = next(o for o in load_data("rest/core.json")["rest"] if o["key"] == vector["occasion"])
    maxima = run_derived(vector)
    maxima["stamina"] = vector.get("max", {}).get("stamina", 0)
    result = {resource: amount for resource, amount in vector["current"].items()}
    conditions = vector.get("conditions", [])
    for row in occasion["recovery"]:
        resource, mode = row["resource"], row["mode"]
        if resource == "conditions":
            if mode == "clear":
                conditions = [c for c in conditions if c["duration"] not in row["durations"]]
        elif mode == "full":
            result[resource] = maxima[resource]
        elif mode == "fraction":
            gain = maxima[resource] * row["numerator"] // row["denominator"]
            result[resource] = min(maxima[resource], result[resource] + gain)
    result["conditions"] = [c["key"] for c in conditions]
    return result


def run_combat(vector):
    """One attack, in the order of docs/core/combat.md. Dice and rolls are inputs."""
    pipeline = load_data("combat/pipeline.json")["pipeline"]
    types = {t["key"]: t for t in load_data("combat/damage_types.json")["damage_types"]}
    attacker, defender, damage = vector["attacker"], vector["defender"], vector["damage"]
    tags = types[damage["type"]]["tags"]

    attack = run_resolution(
        {
            "actor": {"skills": {"attack": attacker["skill"]}, "attributes": attacker.get("attributes", {})},
            "modifiers": vector.get("modifiers", []),
            "roll": attacker["roll"],
        }
    )
    crit = "crit" in attack["tags"]
    result = {
        "target": attack["target"],
        "margin": attack["margin"],
        "crit": crit,
        "evaded": False,
        "blocked": False,
        "reaction_tags": [],
    }
    if not attack["success"]:
        result.update(outcome="fumble" if "fumble" in attack["tags"] else "miss", damage=0, hp=defender["hp"])
        return result

    # One reaction per attack, rolled only on a hit. Delivery decides what it can do:
    # single is evaded whole, area is halved by an evade, instant takes no reaction.
    physical = pipeline["dr_applies_tag"] in tags
    delivery = vector.get("delivery", "single")
    if delivery not in ("single", "area", "instant"):
        raise ValueError(f"delivery {delivery!r} is not defined")
    reaction = defender.get("reaction") if delivery != "instant" else None
    kind = reaction["kind"] if reaction else None
    if kind not in (None, "evade", "block"):
        raise ValueError(f"reaction kind {kind!r} is not defined")
    if kind == "block" and not physical:
        kind = None
    if kind is not None:
        # The reaction is a d100 check like the attack: same fumble and crit window.
        skill = defender["evasion_skill"] if kind == "evade" else reaction["skill"]
        check = run_resolution(
            {
                "actor": {"skills": {kind: skill}, "attributes": defender.get("attributes", {})},
                "roll": reaction["roll"],
            }
        )
        result["reaction_tags"] = check["tags"]
        if check["success"] and check["margin"] >= attack["margin"]:
            if kind == "block":
                result["blocked"] = True
            else:
                result["evaded"] = True
                if delivery == "single":
                    result.update(outcome="evaded", damage=0, hp=defender["hp"])
                    return result
    result["outcome"] = "hit"

    amount = damage["roll"]
    if crit:
        amount = evaluate_formula(pipeline["crit_damage"], {}, {"damage": amount})
    if result["evaded"]:
        amount = evaluate_formula(pipeline["area_evade_damage"], {}, {"damage": amount})
    if result["blocked"]:
        amount = max(amount - reaction["shield_dr"], 0)
    if physical:
        dr = defender.get("armor_dr", 0)
        if pipeline["pierce_tag"] in tags:
            dr = evaluate_formula(pipeline["pierce_dr"], {}, {"dr": dr})
        amount = max(amount - dr, 0)
    amount = max(amount - defender.get("resist", {}).get(damage["type"], 0), 0)
    result.update(damage=amount, hp=max(defender["hp"] - amount, 0))
    return result


def run_progression(vector):
    """XP and level, one level-up pool, and spending a pool on one skill; `op` picks which."""
    rules = load_data("progression/core.json")["progression"]
    op = vector["op"]
    if op == "xp":
        level = vector["level"]
        total = evaluate_formula(rules["xp_to_reach"], {}, {"level": level})
        step = None
        if level > rules["start_level"]:
            step = total - evaluate_formula(rules["xp_to_reach"], {}, {"level": level - 1})
        return {"xp_total": total, "xp_step": step}
    if op == "level":
        level = rules["start_level"]
        while level < rules["max_level"]:
            needed = evaluate_formula(rules["xp_to_reach"], {}, {"level": level + 1})
            if needed > vector["xp"]:
                break
            level += 1
        return {"level": level}
    if op == "level_up":
        attributes = vector["actor"]["attributes"]
        luck_roll = vector["luck_roll"]
        ceiling = attributes[rules["luck_roll"]["max_attr"]]
        if not rules["luck_roll"]["min"] <= luck_roll <= ceiling:
            return {"error": "luck_roll out of range"}
        points = evaluate_formula(rules["skill_points"], attributes, {"luck_roll": luck_roll})
        return {"skill_points": points, "perks": rules["perks_per_level"]}
    if op == "spend":
        cost = rules["skill_cost"]["tag" if vector["tagged"] else "normal"]
        room = rules["skill_cap"] - vector["skill"]
        ranks = min(vector["points"] // cost, max(room, 0))
        return {"ranks": ranks, "spent": ranks * cost, "left": vector["points"] - ranks * cost}
    raise ValueError(f"progression op {op!r} is not defined")


RUNNERS = {
    "resolution": run_resolution,
    "derived": run_derived,
    "formula": run_formula,
    "modifiers": run_modifiers,
    "actions": run_actions,
    "rest": run_rest,
    "combat": run_combat,
    "progression": run_progression,
}


def run_vector(function, vector):
    return RUNNERS[function](vector)
