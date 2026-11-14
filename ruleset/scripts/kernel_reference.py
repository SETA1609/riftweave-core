#!/usr/bin/env python3
"""Python reference for the kernel functions the golden vectors check.

One runner per kernel function, keyed by name in RUNNERS. Coefficients are read from
ruleset/data/, so a data change that breaks a rule fails a vector. The input and result
of each runner are listed in docs/core/vectors.md.
"""

import json
import math
from pathlib import Path

from casting_reference import run_casting, run_spellcraft
from hazard_reference import run_hazard
from loot_reference import run_loot

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


def operation_of(modifier):
    """A modifier's operation (docs/core/modifier.md)."""
    return modifier["operation"]


def strongest(group):
    """The strongest bonus and the strongest penalty of one group, in list order.

    A group is (index, modifier) pairs with one operation. `set` is not ranked: the latest wins.
    For add and resist the neutral value is 0; for multiply it is 1.
    """
    if operation_of(group[0][1]) == "set":
        return [group[-1]]
    neutral = 1 if operation_of(group[0][1]) == "multiply" else 0
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
    """Steps 2 and 3: same origin, source, target, and operation keeps the strongest; then, for a
    `highest` origin, same origin, target, and operation keeps the strongest bonus and penalty."""
    policies = origin_policies()
    pairs = _keep_strongest(
        list(enumerate(modifiers)),
        lambda m: (m.get("origin"), m["source"], m["target"], operation_of(m)),
    )
    highest = [pair for pair in pairs if "origin" in pair[1] and policies[pair[1]["origin"]] == "highest"]
    rest = [pair for pair in pairs if pair not in highest]
    highest = _keep_strongest(highest, lambda m: (m["origin"], m["target"], operation_of(m)))
    return [m for _, m in sorted(rest + highest, key=lambda pair: pair[0])]


def apply_modifiers(value, modifiers, elapsed_turns=0, removed_sources=()):
    """Steps 1 to 6 for one numeric target: expire, dedupe, origin, add, multiply (floored), set.

    The caller passes the modifiers of one target; grant and resist carry no number here.
    """
    numeric = [
        m
        for m in dedupe_modifiers(live_modifiers(modifiers, elapsed_turns, removed_sources))
        if operation_of(m) in ("add", "multiply", "set")
    ]
    value += sum(m["value"] for m in numeric if operation_of(m) == "add")
    factor = 1
    for m in numeric:
        if operation_of(m) == "multiply":
            factor *= m["value"]
    if factor != 1:
        value = math.floor(value * factor)
    sets = [m["value"] for m in numeric if operation_of(m) == "set"]
    if sets:
        value = sets[-1]
    return value


def granted_flags(modifiers, elapsed_turns=0, removed_sources=()):
    """Flags that at least one live grant names."""
    live = live_modifiers(modifiers, elapsed_turns, removed_sources)
    return sorted({m["target"] for m in live if operation_of(m) == "grant"})


def apply_resist(amount, tag, modifiers, elapsed_turns=0, removed_sources=()):
    """A packet's amount after every live resist on its tag, never below zero."""
    live = dedupe_modifiers(live_modifiers(modifiers, elapsed_turns, removed_sources))
    resist = sum(m["value"] for m in live if operation_of(m) == "resist" and m["target"] == tag)
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
    bonus = apply_modifiers(0, [m for m in vector.get("modifiers", []) if operation_of(m) == "add"])
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
    growth = actor.get("growth", {})
    result = {}
    for stat in load_data("derived/core.json")["derived"]:
        attribute = attributes.get(stat["attribute"], 0)
        attribute_term = (attribute * stat["per_point"]) // stat.get("per_point_divisor", 1)
        value = (
            stat["base"]
            + attribute_term
            + level * (attribute * stat["per_level_attribute"] + stat["per_level_flat"])
            + growth.get(stat["key"], 0)
        )
        own = [m for m in modifiers if m.get("target") == stat["key"]]
        result[stat["key"]] = apply_modifiers(value, own)
    return result


def evaluate_formula(node, attributes, variables=None):
    """Evaluate a formula tree: literal, attribute, variable, and the operations add, subtract,
    multiply, divide (floor)."""
    if "literal" in node:
        return node["literal"]
    if "attribute" in node:
        return attributes[node["attribute"]]
    if "variable" in node:
        return (variables or {})[node["variable"]]
    operation = node["operation"]
    arguments = [evaluate_formula(argument, attributes, variables) for argument in node["arguments"]]
    if operation == "add":
        return sum(arguments)
    if operation == "subtract":
        first, *rest = arguments
        return first - sum(rest)
    if operation == "multiply":
        product = 1
        for argument in arguments:
            product *= argument
        return product
    if operation == "divide":
        numerator, denominator = arguments
        return numerator // denominator
    raise NotImplementedError(f"formula operation {operation!r} has no reference")


def run_formula(vector):
    if "expression" in vector:
        expression = vector["expression"]
    else:
        expression = load_data(f"formulas/{vector['formula']}.json")["expression"]
    attributes = vector.get("actor", {}).get("attributes", {})
    return {"value": evaluate_formula(expression, attributes, vector.get("variables"))}


def run_actions(vector):
    data = load_data("actions/core.json")
    pool_rules = data["pool"]
    costs = {action["key"]: action["cost"] for action in data["actions"]}
    attribute = vector["actor"]["attributes"][pool_rules["attribute"]]
    pool = pool_rules["base"] + (attribute * pool_rules["per_point"]) // pool_rules["per_point_divisor"]
    modifiers = vector.get("modifiers", [])
    cost = sum(
        max(apply_modifiers(costs[key], [m for m in modifiers if m["target"] == key]), 0)
        for key in vector.get("actions", [])
    )
    # One cost on two clocks: a real-time client pays AP times stamina_per_ap (docs/core/actions.md).
    return {"pool": pool, "cost": cost, "stamina": cost * data["stamina_per_ap"]}


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
        elif resource == "toxicity":
            if mode == "empty" and "toxicity" in result:
                result["toxicity"] = 0
        elif resource == "food":
            continue  # read below, after the pools
        elif mode == "full":
            result[resource] = maxima[resource]
        elif mode == "fraction":
            gain = maxima[resource] * row["numerator"] // row["denominator"]
            result[resource] = min(maxima[resource], result[resource] + gain)
    result["conditions"] = [c["key"] for c in conditions]
    # Food at rest (docs/core/rest.md#food-at-rest): the occasion's food row, then the meal.
    running = vector.get("running", {})
    ticks, held = list(running.get("ticks", [])), list(running.get("modifiers", []))
    food = next((row["mode"] for row in occasion["recovery"] if row["resource"] == "food"), "none")
    if food == "clear":
        ticks = [t for t in ticks if t["origin"] != FOOD_ORIGIN]
        held = [m for m in held if m["origin"] != FOOD_ORIGIN]
    refused = []
    if "meal" in vector:
        items = {row["key"]: row for row in load_data("equipment/consumables.json")["equipment"]}
        meal = items.get(vector["meal"])
        if meal is None or meal["type"] != "food":
            refused.append(vector["meal"])
        else:
            pools = ("hp", "mp", "stamina")
            eaten = run_consume({"current": {p: result[p] for p in pools}, "max": {p: maxima[p] for p in pools},
                                 "events": [{"item": vector["meal"]}]})
            result.update({p: eaten[p] for p in pools})
            ticks = [t for t in ticks if t["origin"] != FOOD_ORIGIN] + eaten["ticks"]
            held = [m for m in held if m["origin"] != FOOD_ORIGIN] + eaten["modifiers"]
    result["running"] = {"ticks": ticks, "modifiers": held}
    result["refused"] = refused
    return result


FOOD_ORIGIN = "core:origin/food"


def to_copper(economy, cost):
    value = {d["unit"]: d["value"] for d in economy["currency"]["denominations"]}
    return cost["quantity"] * value[cost["unit"]]


def run_economy(vector):
    """`convert` a cost to copper, or price one `trade` (buy or sell) with an optional barter check."""
    economy = load_data("economy/core.json")
    op = vector["operation"]
    if op == "convert":
        return {"cp": to_copper(economy, vector["cost"])}
    if op != "trade":
        raise ValueError(f"economy operation {op!r} is not defined")
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
        load_data("formulas/price.json")["expression"],
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
            modifiers.append({"target": "attack", "operation": "add", "value": called["attack"], "source": f"called_shot:{called['key']}"})
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
        "counter": False,
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
    # The defender's perks: adds on the reaction targets and shield_dr, and grant flags.
    defender_mods = defender.get("modifiers", [])
    perk_flags = granted_flags(defender_mods)
    if kind == "block" and not physical and "block_magic" not in perk_flags:
        kind = None
    if kind is not None:
        # The reaction is a d100 check like the attack: same fumble and crit window.
        skill = defender["evasion_skill"] if kind == "evade" else reaction["skill"]
        target = "evasion" if kind == "evade" else "block"
        check = run_resolution(
            {
                "actor": {"skills": {kind: skill}, "attributes": defender.get("attributes", {})},
                "modifiers": [m for m in defender_mods if m["target"] == target],
                "roll": reaction["roll"],
            }
        )
        result["reaction_tags"] = check["tags"]
        if check["success"] and check["margin"] >= attack["margin"]:
            if kind == "block":
                result["blocked"] = True
            else:
                result["evaded"] = True
                result["counter"] = delivery == "single" and "crit" in check["tags"] and "riposte" in perk_flags
                if delivery == "single":
                    result.update(outcome="evaded", damage=0, hp=defender["hp"])
                    return result
    result["outcome"] = "hit"

    amount = damage["roll"]
    if "ability" in damage:
        attribute = vector["attacker"].get("attributes", {}).get(damage["ability"], 0)
        modifier = evaluate_formula(load_data("formulas/attribute_modifier.json")["expression"], {}, {"attribute": attribute})
        amount = max(amount + modifier, 0)
    if called is not None:
        amount = math.floor(amount * called["damage_multiplier"])
    if crit:
        amount = evaluate_formula(pipeline["crit_damage"], {}, {"damage": amount})
    if result["evaded"]:
        amount = evaluate_formula(pipeline["area_evade_damage"], {}, {"damage": amount})
    if result["blocked"]:
        shield_dr = apply_modifiers(reaction["shield_dr"], [m for m in defender_mods if m["target"] == "shield_dr"])
        if not physical:
            shield_dr //= 2
        amount = max(amount - shield_dr, 0)
        if "bulwark" in perk_flags and "crit" in result["reaction_tags"]:
            amount = 0
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
    op = vector["operation"]
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
        ceiling = attributes[rules["luck_roll"]["max_attribute"]]
        if not rules["luck_roll"]["min"] <= luck_roll <= ceiling:
            return {"error": "luck_roll out of range"}
        highest = evaluate_formula(rules["pool_roll_max"], attributes)
        flat = {stat["key"]: stat["per_level_flat"] for stat in load_data("derived/core.json")["derived"]}
        gain = {}
        for pool in ("hp", "mp", "sp"):
            roll = vector[f"{pool}_roll"]
            if not 0 <= roll <= highest:
                return {"error": f"{pool}_roll out of range"}
            gain[pool] = flat[pool] + roll
        points = evaluate_formula(rules["skill_points"], attributes, {"luck_roll": luck_roll})
        return {"skill_points": points, "perks": rules["perks_per_level"], "gain": gain}
    if op == "spend":
        cost = rules["skill_cost"]["tag" if vector["tagged"] else "normal"]
        if "group" in vector:
            group = next((g for g in load_data("skill_groups/core.json")["skill_groups"] if g["key"] == vector["group"]), None)
            if group is None:
                return {"error": f'skill group "{vector["group"]}" is not defined'}
            if not group["is_tracked"]:
                return {"error": f'skill group "{vector["group"]}" is not tracked'}
            step = rules["skill_group_cost"]["members_per_step"]
            cost *= -(-len(group["skills"]) // step)
        room = rules["skill_cap"] - vector["skill"]
        ranks = min(vector["points"] // cost, max(room, 0))
        return {"ranks": ranks, "spent": ranks * cost, "left": vector["points"] - ranks * cost}
    if op == "read_skill":
        return read_skill(vector["actor"], vector["read"], vector.get("groups"))
    if op == "take_perk":
        return take_perk(vector["actor"], vector.get("perks") or {}, vector["perk"])
    raise ValueError(f"progression operation {op!r} is not defined")


def take_perk(actor, held, key):
    """One perk checked against a sheet (docs/core/progression.md, rule 5): rank, level,
    attributes, skills, and perks, in that order; the next rank is taken only when none fails."""
    features = {row["key"]: row for row in load_data("features/core.json")["features"]}
    if key not in features:
        return {"error": f'perk "{key}" is not defined'}
    feature = features[key]
    if feature["type"] != "perk":
        return {"error": f'feature "{key}" is not a perk'}
    ranks = feature.get("ranks") or 1
    rank = held.get(key, 0)
    need = feature.get("prerequisite") or {}
    unmet = []
    if rank >= ranks:
        unmet.append({"needs": "rank", "max": ranks, "has": rank})
    level = actor.get("level", 1)
    if need.get("level", 0) > level:
        unmet.append({"needs": "level", "min": need["level"], "has": level})
    attributes = actor.get("attributes", {})
    for ability in sorted(need.get("abilities", {})):
        has = attributes.get(ability, 0)
        if has < need["abilities"][ability]:
            unmet.append({"needs": "attribute", "key": ability, "min": need["abilities"][ability], "has": has})
    for requirement in need.get("skills", []):
        refs = requirement["id"] if isinstance(requirement["id"], list) else [requirement["id"]]
        keys = [ref.rsplit("/", 1)[-1] for ref in refs]
        best = max(read_skill(actor, skill)["value"] for skill in keys)
        if best < requirement["min"]:
            unmet.append({"needs": "skill", "keys": keys, "min": requirement["min"], "has": best})
    for ref in need.get("perks", []):
        if held.get(ref.rsplit("/", 1)[-1], 0) < 1:
            unmet.append({"needs": "perk", "key": ref.rsplit("/", 1)[-1]})
    if unmet:
        return {"taken": False, "perk": key, "rank": rank, "unmet": unmet}
    return {"taken": True, "perk": key, "rank": rank + 1, "unmet": []}


def read_skill(actor, key, groups=None):
    """One skill off a sheet (docs/core/progression.md#two-sheets-one-rule-set). A full sheet reads the
    written value, else the skill's start; a grouped sheet (`groups` given) reads the start plus the
    member's group ranks, or the start alone for an untracked group."""
    governing = {s["key"]: s["associatedAbility"] for s in load_data("skills/core.json")["skills"]}
    if key not in governing:
        return {"error": f'skill "{key}" is not defined'}
    rule = load_data("progression/core.json")["progression"]["skill_start"]
    start = evaluate_formula(rule, {}, {"attribute": actor.get("attributes", {}).get(governing[key], 0)})
    if groups is None:
        written = actor.get("skills", {})
        if key in written:
            return {"value": written[key], "read_from": "sheet"}
        return {"value": start, "read_from": "start"}
    group = next(g for g in load_data("skill_groups/core.json")["skill_groups"] if f"core:skill/{key}" in g["skills"])
    if not group["is_tracked"]:
        return {"value": start, "read_from": "start"}
    return {"value": start + groups.get(group["key"], 0), "read_from": "group"}


def run_monster_sheet(vector):
    """A monster's stats and skills by the player rules (docs/core/actor-sheet.md)."""
    row = vector["monster"]
    attributes = row.get("attributes", {})
    result = run_derived({"actor": {"attributes": attributes, "level": row.get("level", 1)}})
    if "hitPoints" in row:
        result["hp"] = row["hitPoints"]
    progression = load_data("progression/core.json")["progression"]
    rows = {s["key"]: s for s in load_data("skills/core.json")["skills"]}
    growth = progression.get("monster_skill_growth", {"per_level": 0, "categories": []})
    level = row.get("level", 1)
    written = row.get("skills", {})
    skills = {}
    for key in vector.get("read_skills", []):
        if key in written:
            skills[key] = written[key]
        elif key in rows:
            start = evaluate_formula(progression["skill_start"], {}, {"attribute": attributes.get(rows[key]["associatedAbility"], 0)})
            if rows[key]["category"] in growth["categories"]:
                start = min(progression["skill_cap"], start + growth["per_level"] * (level - 1))
            skills[key] = start
        else:
            return {"error": f'skill "{key}" is not defined'}
    result["skills"] = skills
    return result


POOL_OF_PARAMETER = {"health": "hp", "mana": "mp", "stamina": "stamina"}
USED_BY_CONSUMING = ("drink", "eat")


def _origin_of(item, effect):
    """The effect's own origin, else the default for food or for a consumable."""
    if "origin" in effect:
        return effect["origin"]
    emitter = "food" if item["type"] == "food" else "consumable"
    row = next(o for o in load_data("origins/core.json")["origins"] if emitter in o.get("defaultFor", []))
    return f"core:origin/{row['key']}"


def run_consume(vector):
    """Use items and let time pass (docs/core/modifier.md#effects-over-time)."""
    items = {row["key"]: row for row in load_data("equipment/consumables.json")["equipment"]}
    policies = origin_policies()
    actions = load_data("actions/core.json")
    seconds_per_turn = actions["seconds_per_turn"]
    # At a table a stamina restore is AP that arrives next turn (docs/core/actions.md#one-rule-set-two-clocks).
    at_table = vector.get("at_table", False)
    stamina_per_ap, ap_cap = actions["stamina_per_ap"], actions["pool"]["cap"]
    table = {"ap": vector["current"].get("ap", 0), "pending_ap": 0, "stamina": 0}
    decay = load_data("rest/core.json")["toxicity_decay"]["per_second"]
    pools = dict(vector["current"])
    toxicity = pools.pop("toxicity", 0)
    pools.pop("ap", None)

    def restore(pool, amount):
        if at_table and pool == "stamina":
            table["stamina"] += amount
            table["pending_ap"] += table["stamina"] // stamina_per_ap
            table["stamina"] %= stamina_per_ap
        else:
            pools[pool] = min(pools[pool] + amount, maximum[pool])
    maximum = vector["max"]
    ticks, held, refused = [], [], []

    def advance(seconds):
        nonlocal toxicity
        toxicity = max(0, toxicity - decay * seconds)
        for tick in ticks:
            pool = tick["pool"]
            restore(pool, tick["magnitude"] * min(seconds, tick["seconds_left"]))
            tick["seconds_left"] -= seconds
        for row in held:
            row["seconds_left"] -= seconds
        ticks[:] = [tick for tick in ticks if tick["seconds_left"] > 0]
        held[:] = [row for row in held if row["seconds_left"] > 0]

    def consume(event):
        item = items[event["item"]]
        food = item["type"] == "food"
        if item["consumable"]["delivery"] not in USED_BY_CONSUMING or (food and event.get("in_combat")):
            refused.append(event["item"])
            return
        nonlocal toxicity
        toxicity += item["consumable"].get("toxicity", 0)
        source = f"{'food' if food else 'item'}:{event['item']}"
        effects = item["consumable"]["effects"]
        gone = {_origin_of(item, e) for e in effects if policies[_origin_of(item, e)] == "replace"}
        ticks[:] = [t for t in ticks if t["origin"] not in gone and t["source"] != source]
        held[:] = [m for m in held if m["origin"] not in gone and m["source"] != source]
        for effect in effects:
            key = effect["effect"].rsplit("/", 1)[-1]
            duration = effect.get("duration", 1)
            origin = _origin_of(item, effect)
            if key == "restore_resource":
                pool = POOL_OF_PARAMETER[effect["parameter"]]
                if duration == 1:
                    restore(pool, effect["magnitude"])
                else:
                    ticks.append({"source": source, "origin": origin, "pool": pool,
                                  "magnitude": effect["magnitude"], "seconds_left": duration})
            elif duration > 1:
                operation = "resist" if key == "resist" else "add" if "magnitude" in effect else "grant"
                held.append({"source": source, "origin": origin, "target": effect.get("parameter", key), "operation": operation,
                             "value": effect.get("magnitude", 1), "seconds_left": duration})

    for event in vector["events"]:
        if "item" in event:
            consume(event)
        elif "turns" in event:
            for _ in range(event["turns"]):
                # The AP restored last turn arrives at the start of this one, up to the pool cap.
                table["ap"] = min(table["ap"] + table["pending_ap"], ap_cap)
                table["pending_ap"] = 0
                advance(seconds_per_turn)
        else:
            advance(event["advance"])
    # Overdosed while toxicity stays above the limit (docs/core/modifier.md#toxicity).
    limit = maximum.get("toxicity")
    conditions = ["core:condition/overdosed"] if limit is not None and toxicity > limit else []
    result = {**pools, "toxicity": toxicity, "conditions": conditions, "ticks": ticks, "modifiers": held, "refused": refused}
    if at_table:
        result.update(ap=table["ap"], pending_ap=table["pending_ap"])
    return result
def run_stored_actor(vector):
    """A stored actor comes back unchanged, if it is consistent (docs/core/state.md)."""
    actor = vector["actor"]
    seen = set()
    for item in actor["inventory"]:
        if item["id"] in seen:
            return {"error": f"inventory: id {item['id']} is used twice"}
        seen.add(item["id"])
    for slot, item_id in sorted(actor["equipped"].items()):
        if item_id not in seen:
            return {"error": f"equipped {slot}: no inventory item {item_id}"}
    return {"actor": actor}


RUNNERS = {
    "monster_sheet": run_monster_sheet,
    "resolution": run_resolution,
    "derived": run_derived,
    "formula": run_formula,
    "modifiers": run_modifiers,
    "conditions": run_conditions,
    "actions": run_actions,
    "rest": run_rest,
    "consume": run_consume,
    "stored_actor": run_stored_actor,
    "economy": run_economy,
    "combat": run_combat,
    "loot": run_loot,
    "hazard": run_hazard,
    "casting": run_casting,
    "spellcraft": run_spellcraft,
    "progression": run_progression,
}


def run_vector(function, vector):
    return RUNNERS[function](vector)
