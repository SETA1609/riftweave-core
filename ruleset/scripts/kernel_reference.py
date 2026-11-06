#!/usr/bin/env python3
"""Python reference for the kernel functions the golden vectors check.

One runner per kernel function, keyed by name in RUNNERS. Coefficients are read from
ruleset/data/, so a data change that breaks a rule fails a vector. The input and result
of each runner are listed in docs/core/vectors.md.
"""

import json
import math
from pathlib import Path

from loot_reference import run_loot
from wuxing_reference import get_interaction

RULESET = Path(__file__).resolve().parent.parent
DATA_DIR = RULESET / "data"


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def live_modifiers(modifiers, elapsed_turns=0, removed_sources=()):
    """Step 1 of docs/core/modifier.md: drop removed sources and elapsed `turns` modifiers."""
    live = []
    for modifier in modifiers:
        if modifier["source"] in removed_sources:
            continue
        if modifier.get("duration") == "turns" and elapsed_turns >= modifier["turns"]:
            continue
        live.append(modifier)
    return live


def origin_policies():
    """Stacking policy by origin ref, from ruleset/data/origins/core.json."""
    return {f"core:origin/{row['key']}": row["stack"] for row in load_data("origins/core.json")["origins"]}


def strongest(group):
    """The strongest bonus and the strongest penalty of one group, in list order.

    A group is (index, modifier) pairs with one op. `set` is not ranked: the latest wins.
    For add and resist the neutral value is 0; for mul it is 1.
    """
    if group[0][1]["op"] == "set":
        return [group[-1]]
    neutral = 1 if group[0][1]["op"] == "mul" else 0
    bonuses = [pair for pair in group if pair[1]["value"] > neutral]
    penalties = [pair for pair in group if pair[1]["value"] < neutral]
    kept = []
    if bonuses:
        kept.append(max(bonuses, key=lambda pair: pair[1]["value"]))
    if penalties:
        kept.append(min(penalties, key=lambda pair: pair[1]["value"]))
    return kept


def _keep_strongest(pairs, key):
    groups = {}
    for pair in pairs:
        groups.setdefault(key(pair[1]), []).append(pair)
    return sorted((kept for group in groups.values() for kept in strongest(group)), key=lambda pair: pair[0])


def dedupe_modifiers(modifiers):
    """Steps 2 and 3: same origin, source, target, and op keeps the strongest; then, for a
    `highest` origin, same origin, target, and op keeps the strongest bonus and penalty."""
    policies = origin_policies()
    pairs = _keep_strongest(
        list(enumerate(modifiers)),
        lambda m: (m.get("origin"), m["source"], m["target"], m["op"]),
    )
    highest = [pair for pair in pairs if "origin" in pair[1] and policies[pair[1]["origin"]] == "highest"]
    rest = [pair for pair in pairs if pair not in highest]
    highest = _keep_strongest(highest, lambda m: (m["origin"], m["target"], m["op"]))
    return [m for _, m in sorted(rest + highest, key=lambda pair: pair[0])]


def apply_modifiers(value, modifiers, elapsed_turns=0, removed_sources=()):
    """Steps 1 to 6 for one numeric target: expire, dedupe, origin, add, mul (floored), set.

    The caller passes the modifiers of one target; grant and resist carry no number here.
    """
    numeric = [
        m
        for m in dedupe_modifiers(live_modifiers(modifiers, elapsed_turns, removed_sources))
        if m["op"] in ("add", "mul", "set")
    ]
    value += sum(m["value"] for m in numeric if m["op"] == "add")
    factor = 1
    for m in numeric:
        if m["op"] == "mul":
            factor *= m["value"]
    if factor != 1:
        value = math.floor(value * factor)
    sets = [m["value"] for m in numeric if m["op"] == "set"]
    if sets:
        value = sets[-1]
    return value


def granted_flags(modifiers, elapsed_turns=0, removed_sources=()):
    """Flags that at least one live grant names."""
    live = live_modifiers(modifiers, elapsed_turns, removed_sources)
    return sorted({m["target"] for m in live if m["op"] == "grant"})


def apply_resist(amount, tag, modifiers, elapsed_turns=0, removed_sources=()):
    """A packet's amount after every live resist on its tag, never below zero."""
    live = dedupe_modifiers(live_modifiers(modifiers, elapsed_turns, removed_sources))
    resist = sum(m["value"] for m in live if m["op"] == "resist" and m["target"] == tag)
    return max(amount - resist, 0)


def run_modifiers(vector):
    modifiers = vector.get("modifiers", [])
    elapsed = vector.get("elapsed_turns", 0)
    removed = vector.get("removed_sources", [])
    target = vector.get("target")
    own = [m for m in modifiers if target is None or m["target"] == target]
    result = {
        "value": apply_modifiers(vector["value"], own, elapsed, removed),
        "flags": granted_flags(modifiers, elapsed, removed),
    }
    if "packet" in vector:
        packet = vector["packet"]
        result["packet"] = apply_resist(packet["amount"], packet["tag"], modifiers, elapsed, removed)
    return result


def run_conditions(vector):
    """Apply effect and remove events to an actor's conditions, then read its values and flags."""
    rows = {row["key"]: row for row in load_data("conditions/core.json")["conditions"]}
    active = dict(vector.get("conditions", {}))
    for event in vector.get("events", []):
        if "remove" in event:
            active.pop(event["remove"].rsplit("/", 1)[-1], None)
            continue
        effect = event["effect"]
        for key, row in rows.items():
            if effect in row.get("removedBy", []):
                active.pop(key, None)
        for key, row in rows.items():
            if effect in row.get("appliedBy", []):
                cap = row.get("maxStacks", 1) if row.get("stacking") else 1
                active[key] = min(active.get(key, 0) + 1, cap)
    modifiers = list(vector.get("modifiers", []))
    for key in active:
        modifiers.extend(rows[key]["modifiers"])
    values = {
        target: apply_modifiers(base, [m for m in modifiers if m["target"] == target])
        for target, base in vector.get("base", {}).items()
    }
    return {"conditions": dict(sorted(active.items())), "flags": granted_flags(modifiers), "values": values}


def run_resolution(vector):
    rules = load_data("resolution/core.json")["resolution"]
    actor = vector["actor"]
    skills = actor["skills"]
    skill = vector.get("skill")
    if skill is None:
        if len(skills) != 1:
            raise ValueError("actor has several skills; name one with `skill`")
        (skill,) = skills
    bonus = apply_modifiers(0, [m for m in vector.get("modifiers", []) if m["op"] == "add"])
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


def run_wuxing(vector):
    cycles = load_data("wuxing/core.json")["cycles"]
    cycle_id, interaction, multiplier, affects = get_interaction(
        cycles,
        vector["actor_phase"],
        vector["recipient_phase"],
        vector.get("actor_magnitude"),
        vector.get("recipient_magnitude"),
    )
    return {"cycle": cycle_id, "interaction": interaction, "multiplier": multiplier, "affects": affects}


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


def to_copper(economy, cost):
    value = {d["unit"]: d["value"] for d in economy["currency"]["denominations"]}
    return cost["quantity"] * value[cost["unit"]]


def run_economy(vector):
    """`convert` a cost to copper, or price one `trade` (buy or sell) with an optional barter check."""
    economy = load_data("economy/core.json")
    op = vector["op"]
    if op == "convert":
        return {"cp": to_copper(economy, vector["cost"])}
    if op != "trade":
        raise ValueError(f"economy op {op!r} is not defined")
    rules = economy["barter"]
    merchant = next(m for m in economy["merchants"] if m["key"] == vector["merchant"])
    buying = vector["side"] == "buy"
    shift = 0
    if "barter" in vector:
        check = run_resolution(
            {
                "actor": {"skills": {rules["skill"]: vector["barter"]["skill_value"]}},
                "roll": vector["barter"]["roll"],
            }
        )
        if "fumble" in check["tags"]:
            shift = -rules["fumble_penalty"]
        elif check["success"]:
            shift = min(rules["max_shift"], check["margin"] // rules["margin_divisor"])
    percent = merchant["buy_percent"] - shift if buying else merchant["sell_percent"] + shift
    base_cp = to_copper(economy, vector["item_cost"])
    price = evaluate_formula(
        load_data("formulas/price.json")["expr"],
        {},
        {"base_cp": base_cp, "value_percent": vector.get("value_percent", 100), "percent": percent},
    )
    return {"base_cp": base_cp, "shift": shift, "percent": percent, "price_cp": price}


def situation_modifiers(vector, delivery):
    """Step 0 of docs/core/combat.md: the attack modifiers of every situation that applies,
    and the called-shot row. A situation that forbids the range raises ValueError."""
    data = load_data("combat/situations.json")
    rows = {row["key"]: row for row in data["situations"]}
    range_ = vector.get("range", "melee")
    flags = {
        "attacker": set(vector["attacker"].get("flags", [])),
        "defender": set(vector["defender"].get("flags", [])),
    }
    applied = [rows[key] for key in vector.get("situations", [])]
    for row in data["situations"]:
        side, flag = next(iter(row.get("when", {}).items()), (None, None))
        if side is not None and flag in flags[side] and row not in applied:
            applied.append(row)
    modifiers = []
    for row in applied:
        if range_ in row.get("forbids", []):
            raise ValueError(f"{row['key']} forbids a {range_} attack")
        if delivery in row.get("deliveries", ["single", "area", "instant"]):
            modifiers.extend(row.get(range_, []))
    called = None
    if "called_shot" in vector:
        called = next(c for c in data["called_shots"] if c["key"] == vector["called_shot"])
        if called["attack"]:
            modifiers.append({"target": "attack", "op": "add", "value": called["attack"], "source": f"called_shot:{called['key']}"})
    return modifiers, called


def run_combat(vector):
    """One attack, in the order of docs/core/combat.md. Dice and rolls are inputs."""
    pipeline = load_data("combat/pipeline.json")["pipeline"]
    types = {t["key"]: t for t in load_data("combat/damage_types.json")["damage_types"]}
    attacker, defender, damage = vector["attacker"], vector["defender"], vector["damage"]
    tags = types[damage["type"]]["tags"]

    delivery = vector.get("delivery", "single")
    try:
        situation_mods, called = situation_modifiers(vector, delivery)
    except ValueError as error:
        return {"error": str(error)}

    attack = run_resolution(
        {
            "actor": {"skills": {"attack": attacker["skill"]}, "attributes": attacker.get("attributes", {})},
            "modifiers": vector.get("modifiers", []) + situation_mods,
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
    if called is not None:
        result["riders"] = []
    if not attack["success"]:
        result.update(outcome="fumble" if "fumble" in attack["tags"] else "miss", damage=0, hp=defender["hp"])
        return result

    # One reaction per attack, rolled only on a hit. Delivery decides what it can do:
    # single is evaded whole, area is halved by an evade, instant takes no reaction.
    physical = pipeline["dr_applies_tag"] in tags
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
    if called is not None:
        amount = math.floor(amount * called["damage_mul"])
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
    if called is not None:
        result["riders"] = [dict(rider) for rider in called.get("riders", [])]
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
    "conditions": run_conditions,
    "actions": run_actions,
    "wuxing": run_wuxing,
    "rest": run_rest,
    "economy": run_economy,
    "combat": run_combat,
    "loot": run_loot,
    "progression": run_progression,
}


def run_vector(function, vector):
    return RUNNERS[function](vector)
