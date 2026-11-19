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


def summoning_modifiers(modifiers, active):
    """Resolve `while_summoning` for a holder whose summon state is known: with an active summon
    they hold while their source does (while_applied), without one they are dropped. A consumer
    that knows the holder's summons calls this once; live_modifiers drops unresolved ones."""
    resolved = []
    for modifier in modifiers:
        if modifier.get("duration") == "while_summoning":
            if not active:
                continue
            modifier = {**modifier, "duration": "while_applied"}
        resolved.append(modifier)
    return resolved


def live_modifiers(modifiers, elapsed_turns=0, removed_sources=()):
    """Step 1 of docs/core/modifier.md: drop removed sources, elapsed `turns` modifiers, and
    `while_summoning` modifiers that summoning_modifiers did not resolve."""
    live = []
    for modifier in modifiers:
        if modifier["source"] in removed_sources:
            continue
        if modifier.get("duration") == "while_summoning":
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
    # The one while_summoning rule, in the shared live step (docs/core/modifier.md).
    modifiers = summoning_modifiers(vector.get("modifiers", []), bool(vector.get("summoning")))
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
    # A row's `adds`: the attacker's modifiers on that target join the attack while the row applies.
    for row in applied:
        if "adds" in row:
            modifiers.extend({**m, "target": "attack_roll"} for m in vector.get("modifiers", []) if m["target"] == row["adds"])
    called = None
    if "called_shot" in vector:
        called = next(c for c in data["called_shots"] if c["key"] == vector["called_shot"])
        if called["attack"]:
            modifiers.append({"target": "attack_roll", "operation": "add", "value": called["attack"], "source": f"called_shot:{called['key']}"})
    return modifiers, called


def maneuver_of(vector):
    """The maneuver the attack is performed with (docs/core/actions.md#maneuvers), checked
    against the attack's range and the weapon's properties. None without one; a maneuver that
    does not fit raises ValueError."""
    if "maneuver" not in vector:
        return None
    rows = {row["key"]: row for row in load_data("maneuvers/core.json")["maneuvers"]}
    key = vector["maneuver"]
    if key not in rows:
        raise ValueError(f'maneuver "{key}" is not defined')
    row = rows[key]
    range_ = vector.get("range", "melee")
    if range_ not in row.get("ranges", [range_]):
        raise ValueError(f"{key} needs " + " or ".join(f"a {r} attack" for r in row["ranges"]))
    properties = vector["attacker"].get("weapon_properties", [])
    wanted = row.get("weapon_properties")
    if wanted and not set(wanted) & set(properties):
        raise ValueError(f"{key} needs " + " or ".join(f"a {p} weapon" for p in wanted))
    return row


def holds_two_weapons(loadout):
    """A loadout with a weapon in each hand (docs/core/combat.md#attacks-by-loadout)."""
    return loadout is not None and loadout["main"] in ("light", "one_handed") and loadout["off"] in ("light_weapon", "weapon")


def maneuver_cost(maneuver, loadout=None):
    """The AP of an action performed as the maneuver, at least 1, and its stamina in real time.
    With a loadout the attack costs the loadout's attack AP (attack_sequences/core.json)."""
    data = load_data("actions/core.json")
    base = next(a["cost"] for a in data["actions"] if a["key"] == maneuver["modifies"])
    if loadout is not None and maneuver["modifies"] == "attack":
        base = attack_sequence_of(loadout)["ap"]
    ap = max(base + maneuver["ap"], 1)
    return ap, ap * data["stamina_per_ap"]


def run_combat(vector):
    """One attack, in the order of docs/core/combat.md. Dice and rolls are inputs."""
    # The attacker's while_summoning modifiers, resolved once; beside_summon means a summon is out.
    summoning = bool(vector.get("summoning")) or "beside_summon" in vector.get("situations", [])
    vector = {**vector, "modifiers": summoning_modifiers(vector.get("modifiers", []), summoning)}
    pipeline = load_data("combat/pipeline.json")["pipeline"]
    types = {t["key"]: t for t in load_data("combat/damage_types.json")["damage_types"]}
    attacker, defender, damage = vector["attacker"], vector["defender"], vector["damage"]
    tags = types[damage["type"]]["tags"]

    delivery = vector.get("delivery", "single")
    try:
        situation_mods, called = situation_modifiers(vector, delivery)
        maneuver = maneuver_of(vector)
    except ValueError as error:
        return {"error": str(error)}
    loadout = vector.get("loadout")
    if loadout is not None and attack_sequence_of(loadout) is None:
        return {"error": no_sequence_error(loadout)}
    if maneuver is not None and maneuver.get("attack", 0):
        situation_mods.append({"target": "attack_roll", "operation": "add", "value": maneuver["attack"], "source": f"maneuver:{maneuver['key']}"})

    if delivery not in ("single", "area", "instant"):
        raise ValueError(f"delivery {delivery!r} is not defined")
    physical = pipeline["dr_applies_tag"] in tags
    # The defender's perks: adds on evasion, shields, and shield_dr, and grant flags.
    defender_mods = defender.get("modifiers", [])
    perk_flags = granted_flags(defender_mods)
    stance = defender.get("stance")
    if stance not in (None, "dodge", "block"):
        return {"error": f'stance "{stance}" is not defined'}
    if stance == "block" and ("shield_dr" not in defender or "shields_skill" not in defender):
        return {"error": "a block needs shield_dr and shields_skill"}
    on = lambda target: [m for m in defender_mods if m["target"] == target]
    evasion_rate = evaluate_formula(pipeline["evasion_rate"], {}, {"skill": apply_modifiers(defender.get("evasion_skill", 0), on("evasion"))})
    if stance == "dodge":
        evasion_rate = evaluate_formula(pipeline["dodge_evasion_rate"], {}, {"rate": evasion_rate})
    if holds_two_weapons(defender.get("loadout")):
        evasion_rate = max(apply_modifiers(evasion_rate, on("two_weapon_evasion_band")), 0)
    block_rate = 0
    if stance == "block" and (physical or "block_magic" in perk_flags):
        block_rate = evaluate_formula(pipeline["block_rate"], {}, {"skill": apply_modifiers(defender["shields_skill"], on("shields"))})
        if not physical:
            block_rate = evaluate_formula(pipeline["magic_block_rate"], {}, {"rate": block_rate})
    if delivery == "instant":
        evasion_rate = block_rate = 0
    rules = load_data("resolution/core.json")["resolution"]
    defender_window = rules["crit_window"] + defender.get("attributes", {}).get("lck", 0) // rules["crit_window_per_two_luck"]

    # A situation's `adds` target joins the attack only through its row (situation_modifiers).
    adds = {row["adds"] for row in load_data("combat/situations.json")["situations"] if "adds" in row}
    attack = run_resolution(
        {
            "actor": {"skills": {"attack_roll": attacker["skill"]}, "attributes": attacker.get("attributes", {})},
            "modifiers": [m for m in vector.get("modifiers", []) if m["target"] not in adds] + situation_mods,
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
        "evasion_rate": evasion_rate,
        "block_rate": block_rate,
        "counter": False,
        "stagger_ap": 0,
    }
    if called is not None:
        result["riders"] = []
    if attacker.get("weapon_effects") is not None:
        result["weapon_effects"] = []
    if maneuver is not None:
        result["ap"], result["stamina"] = maneuver_cost(maneuver, vector.get("loadout"))
        result["effects"] = []
    if not attack["success"]:
        result.update(outcome="fumble" if "fumble" in attack["tags"] else "miss", damage=0, hp=defender["hp"])
        return result

    # One roll, read in bands (docs/core/combat.md step 2): the evasion band, then a block
    # stance's band under it. A crit always hits; instant meets no band.
    margin, window, rate, block = attack["margin"], defender_window, evasion_rate, block_rate
    if not crit and margin < rate:
        result["evaded"] = True
        result["counter"] = stance == "dodge" and delivery == "single" and "riposte" in perk_flags and margin >= rate - window
        if delivery == "single":
            result.update(outcome="evaded", damage=0, hp=defender["hp"])
            return result
    elif not crit and margin < rate + block:
        result.update(outcome="blocked", blocked=True, damage=0, hp=defender["hp"])
        if "bulwark" in perk_flags and margin >= rate + block - window:
            result["stagger_ap"] = pipeline["bulwark_stagger_ap"]
        return result
    result["outcome"] = "hit"

    amount = damage["roll"]
    if "ability" in damage:
        attribute = vector["attacker"].get("attributes", {}).get(damage["ability"], 0)
        modifier = evaluate_formula(load_data("formulas/attribute_modifier.json")["expression"], {}, {"attribute": attribute})
        amount = max(amount + modifier, 0)
    if called is not None:
        amount = math.floor(amount * called["damage_multiplier"])
    if maneuver is not None:
        amount = math.floor(amount * maneuver.get("damage_multiplier", 1))
    # The weapon's effects (docs/core/combat.md#weapon-effects).
    weapon_effects = attacker.get("weapon_effects")
    for applied in weapon_effects or []:
        # A bane needs a parameter, as in Go: no parameter matches no creature type, not a defender without one.
        if applied["effect"] == "core:effect/bane" and applied.get("parameter") and applied["parameter"] == defender.get("creature_type"):
            amount += applied["magnitude"]
    if crit:
        amount = evaluate_formula(pipeline["crit_damage"], {}, {"damage": amount})
    if result["evaded"]:
        amount = evaluate_formula(pipeline["area_evade_damage"], {}, {"damage": amount})
    if physical:
        dr = defender.get("armor_dr", 0)
        if "shield_dr" in defender:
            dr += apply_modifiers(defender["shield_dr"], on("shield_dr"))
        if pipeline["pierce_tag"] in tags:
            dr = evaluate_formula(pipeline["pierce_dr"], {}, {"dr": dr})
        pierce = min(sum(a["magnitude"] for a in weapon_effects or [] if a["effect"] == "core:effect/armor_pierce"), 100)
        dr -= dr * pierce // 100
        amount = max(amount - dr, 0)
    amount = max(amount - defender.get("resist", {}).get(damage["type"], 0), 0)
    result.update(damage=amount, hp=max(defender["hp"] - amount, 0))
    if called is not None and not result["evaded"]:
        result["riders"] = [dict(rider) for rider in called.get("riders", [])]
    if maneuver is not None and not result["evaded"]:
        result["effects"] = [dict(applied) for applied in maneuver["effects"]]
    if weapon_effects is not None and not result["evaded"]:
        result["weapon_effects"] = landed_weapon_effects(weapon_effects, damage["type"], defender)
    return result


def attack_sequence_of(loadout):
    """The attack_sequences row that covers a loadout (docs/core/combat.md#attacks-by-loadout),
    or None when no row does."""
    data = load_data("attack_sequences/core.json")
    for row in data["attack_sequences"]:
        if loadout["main"] in row["main"] and loadout["off"] in row["off"]:
            return row
    return None


def no_sequence_error(loadout):
    return 'no attack sequence fits main hand "%s" and off hand "%s"' % (loadout["main"], loadout["off"])


def soften(value, softening):
    """A negative strike value after mainhand_penalty or offhand_penalty, never past 0."""
    return min(value + softening, 0) if value < 0 else value


def plan_strikes(vector, row, flags, on):
    """The strikes of the action in order, as (hand, attack value), and the AP it spends."""
    data = load_data("attack_sequences/core.json")
    action = vector["action"]
    soft = {"main": apply_modifiers(0, on("mainhand_penalty")), "off": apply_modifiers(0, on("offhand_penalty"))}
    table = [(strike["hand"], soften(strike["attack"], soft[strike["hand"]])) for strike in row["strikes"]]
    left = vector.get("ap")
    if action in ("attack", "jab"):
        if action == "jab":
            if "jab" not in row:
                raise ValueError(f"a jab needs a weapon in the off hand ({row['key']} has none)")
            table, cost = [("off", soften(row["jab"]["attack"], soft["off"]))], row["jab"]["ap"]
        else:
            cost = row["ap"]
        if left is not None and left < cost:
            raise ValueError(f"{action} needs {cost} AP")
        return table, cost
    need = row["ap"] + data["full_attack"]["min_extra_ap"]
    if left is None or left < need:
        raise ValueError(f"{action} needs {need} AP")
    if action == "circle_cut":
        if "can_circle_cut" not in flags:
            raise ValueError("circle_cut needs a perk that grants can_circle_cut")
        if vector["loadout"]["main"] not in data["circle_cut"]["main"]:
            raise ValueError("circle_cut needs " + " or ".join(f"a {m} main hand" for m in data["circle_cut"]["main"]))
        return [("main", data["circle_cut"]["attack"])] * len(vector["foes"]), left
    extra = data["full_attack"]["extra_main_attacks"]
    count = max(min(apply_modifiers(0, on("full_attack_strikes")), len(extra)), 0)
    strikes = table + [("main", value) for value in extra[:count]]
    off = [value for hand, value in table if hand == "off"]
    if off and apply_modifiers(0, on("full_attack_off_strikes")) >= 1:
        strikes.append(("off", off[0] + data["full_attack"]["extra_off_attack"]))
    return strikes, left


def run_attack_sequence(vector):
    """One attack action by loadout (docs/core/combat.md#attacks-by-loadout): the loadout's
    strikes, each its own roll through run_combat, against the foes in reach. Rolls are inputs."""
    action = vector["action"]
    if action not in ("attack", "jab", "full_attack", "circle_cut"):
        return {"error": f'action "{action}" is not defined'}
    row = attack_sequence_of(vector["loadout"])
    if row is None:
        return {"error": no_sequence_error(vector["loadout"])}
    attacker, foes = vector["attacker"], vector["foes"]
    modifiers = attacker.get("modifiers", [])
    flags = granted_flags(modifiers)
    on = lambda target: [m for m in modifiers if m["target"] == target]
    try:
        plan, ap = plan_strikes(vector, row, flags, on)
    except ValueError as error:
        return {"error": str(error)}
    if action in ("full_attack", "circle_cut"):
        aims = [0] * len(plan) if action == "full_attack" else list(range(len(foes)))
    else:
        aims = vector.get("targets", [0] * len(plan))
        if len(aims) != len(plan):
            return {"error": f"targets names {len(aims)} foes for {len(plan)} strikes"}
    for aim in aims:
        if aim < 0 or aim >= len(foes):
            return {"error": f"target {aim} is not a foe in reach"}
        if foes[aim]["hp"] == 0:
            return {"error": f"foe {aim} is already down"}
    hp = [foe["hp"] for foe in foes]
    # A circle cut strikes each foe once, so nothing carries (its own rule, [A]).
    carries = 0 if action == "circle_cut" else math.inf if "reaping_arc" in flags else 1 if "carry_through" in flags else 0
    rolls = list(vector["rolls"])
    queue = [{"hand": hand, "attack": value, "foe": aim, "free": False} for (hand, value), aim in zip(plan, aims)]
    struck_by_main, scissor_left = set(), action == "full_attack" and "scissor_cut" in flags
    strikes, lost, index = [], 0, 0
    while index < len(queue):
        strike = queue[index]
        index += 1
        carried = False
        if hp[strike["foe"]] == 0:
            standing = [i for i, left in enumerate(hp) if left > 0]
            if carries > 0 and standing:
                carries -= 1
                carried = True
                if action == "full_attack":
                    for later in queue[index - 1:]:
                        later["foe"] = standing[0]
                strike["foe"] = standing[0]
            else:
                lost += 1
                continue
        if not rolls:
            return {"error": f"the {action} needs a roll for strike {len(strikes) + 1}"}
        roll = rolls.pop(0)
        foe, hand = strike["foe"], strike["hand"]
        scissor = scissor_left and hand == "off" and foe in struck_by_main
        damage = dict(vector["off_damage"] if hand == "off" and "off_damage" in vector else vector["damage"])
        damage["roll"] = roll["damage"] * (2 if scissor else 1)
        skill = attacker.get("off_skill", attacker["skill"]) if hand == "off" else attacker["skill"]
        # An off-hand weapon has only its own properties and effects, never the main hand's.
        prefix = "off_" if hand == "off" else ""
        properties, effects = attacker.get(f"{prefix}weapon_properties", []), attacker.get(f"{prefix}weapon_effects", [])
        strike_modifiers = list(vector.get("modifiers", []))
        if strike["attack"]:
            strike_modifiers.append({"target": "attack_roll", "operation": "add", "value": strike["attack"], "source": f"strike:{hand}"})
        outcome = run_combat(
            {
                "attacker": {
                    "skill": skill,
                    "roll": roll["attack"],
                    "attributes": attacker.get("attributes", {}),
                    "flags": attacker.get("flags", []),
                    "weapon_properties": properties,
                    "weapon_effects": effects,
                },
                "defender": {**foes[foe], "hp": hp[foe]},
                "damage": damage,
                "modifiers": strike_modifiers,
                "situations": vector.get("situations", []),
                "summoning": vector.get("summoning", False),
            }
        )
        if "error" in outcome:
            return outcome
        hit = outcome["outcome"] == "hit"
        if scissor and hit:
            scissor_left = False
        elif scissor:
            scissor = False
        if hit and hand == "main":
            struck_by_main.add(foe)
        dropped = hp[foe] > 0 and outcome["hp"] == 0
        hp[foe] = outcome["hp"]
        strikes.append(
            {
                "hand": hand,
                "foe": foe,
                "attack": strike["attack"],
                "outcome": outcome["outcome"],
                "target": outcome["target"],
                "margin": outcome["margin"],
                "crit": outcome["crit"],
                "damage": outcome["damage"],
                "hp": outcome["hp"],
                "carried": carried,
                "free": strike["free"],
                "scissor": scissor and hit,
                # The landed effects, listed only when the strike's hand has any.
                **({"weapon_effects": outcome.get("weapon_effects", [])} if effects else {}),
            }
        )
        if dropped and index == len(queue) and carries > 0 and any(left > 0 for left in hp):
            queue.append({"hand": hand, "attack": strike["attack"], "foe": foe, "free": True})
    stamina = ap * load_data("actions/core.json")["stamina_per_ap"]
    return {"action": action, "sequence": row["key"], "ap": ap, "stamina": stamina, "strikes": strikes, "hp": hp, "lost": lost}


WEAPON_STEP_EFFECTS = ("core:effect/armor_pierce", "core:effect/bane")


def landed_weapon_effects(weapon_effects, damage_type, defender):
    """The weapon effects a hit lists for the caller (docs/core/combat.md#weapon-effects):
    lingering damage of the attack's type, a banish only on a summoned defender, and any other
    effect as written. Armor pierce and bane already acted on the damage."""
    landed = []
    for applied in weapon_effects:
        if applied["effect"] in WEAPON_STEP_EFFECTS:
            continue
        if applied["effect"] == "core:effect/banish_summoned" and "summoned" not in defender.get("flags", []):
            continue
        effect = dict(applied)
        if effect["effect"] == "core:effect/lingering_damage":
            effect = {"effect": effect["effect"], "parameter": damage_type, **{k: v for k, v in effect.items() if k not in ("effect", "parameter")}}
        landed.append(effect)
    return landed


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
        keys = [ref.rsplit("/", 1)[-1] for ref in requirement["any_of"]]
        best = max(read_skill(actor, skill)["value"] for skill in keys)
        if best < requirement["min"]:
            unmet.append({"needs": "skill", "keys": keys, "min": requirement["min"], "has": best})
    for item in need.get("perks", []):
        # A held perk { perk, rank } needs that rank, 1 when left out (M3-14, M5-70).
        ref, wanted = item["perk"], item.get("rank", 1)
        has = held.get(ref.rsplit("/", 1)[-1], 0)
        if has < wanted:
            entry = {"needs": "perk", "key": ref.rsplit("/", 1)[-1]}
            if wanted > 1:
                entry.update({"min": wanted, "has": has})
            unmet.append(entry)
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


def emit_perk(feature, rank):
    """A perk's modifiers at a rank (docs/modules/progression.md#perks-and-traits): an add at
    value x rank, a multiply of m at 1 + (m - 1) x rank, a grant or a set once."""
    emitted = []
    for modifier in feature.get("modifiers", []):
        modifier = dict(modifier)
        if operation_of(modifier) == "add":
            modifier["value"] = modifier["value"] * rank
        elif operation_of(modifier) == "multiply":
            modifier["value"] = 1 + (modifier["value"] - 1) * rank
        emitted.append(modifier)
    return emitted


def held_perk_modifiers(perks):
    """The modifiers of perks held as {perk, rank} (M3-15), in list order."""
    features = {row["key"]: row for row in load_data("features/core.json")["features"]}
    modifiers = []
    for held in perks:
        modifiers += emit_perk(features[held["perk"].rsplit("/", 1)[-1]], held.get("rank", 1))
    return modifiers


def run_monster_sheet(vector):
    """A monster's stats and skills by the player rules (docs/core/actor-sheet.md)."""
    row = vector["monster"]
    attributes = row.get("attributes", {})
    perk_modifiers = held_perk_modifiers(row.get("perks", []))
    result = run_derived({"actor": {"attributes": attributes, "level": row.get("level", 1)}, "modifiers": perk_modifiers})
    if "hitPoints" in row:
        result["hp"] = apply_modifiers(row["hitPoints"], [m for m in perk_modifiers if m["target"] == "hp"])
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
        if key in skills:
            skills[key] = apply_modifiers(skills[key], [m for m in perk_modifiers if m["target"] == key])
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


def run_share(vector):
    """What one ally gets from the sharers near it (docs/core/share.md, M3-14)."""
    features = {row["key"]: row for row in load_data("features/core.json")["features"]}
    groups = {row["key"]: row for row in load_data("summons/bonuses.json")["summon_bonuses"]}
    modifier_of = load_data("formulas/attribute_modifier.json")["expression"]
    elapsed = vector.get("elapsed_turns", 0)
    emitted = []
    for sharer in vector["sharers"]:
        perk = features.get(sharer["perk"])
        if perk is None:
            return {"error": f'feature "{sharer["perk"]}" is not defined'}
        share = perk.get("share")
        if share is None:
            return {"error": "share needs a perk with a share object"}
        rank = sharer.get("rank", 1)
        group = groups[share["bonus"].rsplit("/", 1)[-1]]
        rank_share = share.get("rank_share", group["rank_share"])
        if not 1 <= rank <= len(rank_share) or rank > perk.get("ranks", 1):
            return {"error": f'share rank "{rank}" is not defined'}
        if sharer["distance"] > share["range"]:
            continue
        attribute = sharer["attributes"].get(group["attribute"], 0)
        modifier = evaluate_formula(modifier_of, {}, {"attribute": attribute})
        value = math.floor(max(modifier, 0) * rank_share[rank - 1])
        if value <= 0:
            continue
        for target in group["attributes"]:
            emitted.append({"target": target, "operation": "add", "value": value,
                            "source": f"share:{perk['key']}:{sharer['name']}", "origin": "core:origin/morale",
                            "duration": "turns", "turns": share["turns"]})
    live = live_modifiers(emitted, elapsed)
    bonus = {}
    for modifier in live:
        if modifier["target"] not in bonus:
            bonus[modifier["target"]] = apply_modifiers(0, [m for m in live if m["target"] == modifier["target"]])
    return {"modifiers": live, "bonus": bonus}


def run_command_summons(vector):
    """What each of a commander's summons gets from an order and a held command
    (docs/core/commanding.md, GD_S06_03)."""
    features = {row["key"]: row for row in load_data("features/core.json")["features"]}
    commander = vector["commander"]
    summons = vector["summons"]
    granted = {name: [] for name in summons}
    ap = 0
    if "order" in vector:
        perk = features.get(commander.get("perk", ""))
        command = perk.get("command") if perk else None
        if command is None:
            return {"error": "an order needs a perk with a command object"}
        targets = vector["order"]["targets"]
        for name in targets:
            if name not in granted:
                return {"error": f"{name} is not an active summon"}
        if command["reach"] == "one" and len(targets) > 1:
            return {"error": f"an order with {perk['key']} reaches 1 summon, not {len(targets)}"}
        base = next(a["cost"] for a in load_data("actions/core.json")["actions"] if a["key"] == "command")
        ap = max(apply_modifiers(base, [m for m in commander.get("modifiers", []) if m["target"] == "command"]), 0)
        for name in targets:
            granted[name].append({"target": "ap", "operation": "add", "value": command["ap"],
                                  "source": f"command:{perk['key']}:{commander['name']}", "origin": "core:origin/morale",
                                  "duration": "turns", "turns": command["turns"]})
    held = commander.get("concentration")
    if held is not None:
        from casting_reference import load_data as load_casting_data  # spell bases resolved

        spells = {row["key"]: row for row in load_casting_data("spells/core.json")["spells"]}
        if held not in spells:
            return {"error": f"{held} is not a spell"}
        for applied in spells[held]["effects"]:
            if applied["effect"] != "core:effect/command_summons":
                continue
            for name in summons:
                granted[name].append({"target": "attack_roll", "operation": "add", "value": applied["magnitude"],
                                      "source": f"command:{held}:{commander['name']}", "origin": "core:origin/morale",
                                      "duration": "while_applied"})
    return {"ap": ap, "modifiers": granted}


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
    "attack_sequence": run_attack_sequence,
    "loot": run_loot,
    "hazard": run_hazard,
    "casting": run_casting,
    "spellcraft": run_spellcraft,
    "progression": run_progression,
    "share": run_share,
    "command_summons": run_command_summons,
}


def run_vector(function, vector):
    return RUNNERS[function](vector)
