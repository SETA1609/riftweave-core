#!/usr/bin/env python3
"""Python reference for casting (docs/core/casting.md).

One cast: the gate, the mana and reagent cost, one call of the resolution function on the
spell's <color>_magic skill with the armor penalty, then the effects, concentration, and
dispel. The roll is an input.
"""

import json
import math
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"


def load_data(rel_path):
    if rel_path == "spells/core.json":
        import validate  # spell bases resolve before anything reads a spell (docs/core/composition.md)

        return validate.load_spells(DATA_DIR / rel_path)
    return json.loads((DATA_DIR / rel_path).read_text())


SUMMON_EFFECT = "core:effect/summon"

# A magnitude in these units is not a strength: the effect counts it as 1 (magic.md#cost-formula).
UNPRICED_MAGNITUDE_UNITS = (None, "seconds")


def mean_count(count):
    """A summon block's count as a number: itself, or the mean of its dice (1d6 is 3.5)."""
    if isinstance(count, int):
        return count
    dice, _, bonus = count.replace("-", "+-").partition("+")
    number, sides = map(int, dice.split("d"))
    return number * (sides + 1) / 2 + (int(bonus) if bonus else 0)


def dice_range(count):
    """The totals a dice count can make, low and high."""
    dice, _, bonus = count.replace("-", "+-").partition("+")
    number, sides = map(int, dice.split("d"))
    extra = int(bonus) if bonus else 0
    return number + extra, number * sides + extra


def summon_effects(spell):
    """A summon block priced as the summon effect it replaces (docs/core/casting.md#summons):
    its group, its count (the mean of a minion count's dice), its duration, target ranged."""
    block = spell.get("summon")
    if block is None:
        return []
    return [{"effect": SUMMON_EFFECT, "parameter": block["group"], "magnitude": mean_count(block.get("count", 1)),
             "duration": block["duration"], "target": "ranged"}]


def spell_cost(spell):
    """A spell's mana cost from its effects alone (docs/modules/magic.md#cost-formula):
    per effect magnitude_factor x base_cost x magnitude x duration, plus area_factor x base_cost x
    radius for an area, times the target multiplier; the sum rounded up. No caster input."""
    coefficients = load_data("casting/core.json")["casting"]["spell_cost"]
    effects = {e["key"]: e for e in load_data("effects/core.json")["effects"]}
    total = 0.0
    for applied in spell["effects"] + summon_effects(spell):
        effect = effects[applied["effect"].rsplit("/", 1)[-1]]
        magnitude = 1 if effect.get("magnitude_unit") in UNPRICED_MAGNITUDE_UNITS else applied.get("magnitude", 1)
        cost = coefficients["magnitude_factor"] * effect["base_cost"] * magnitude * applied.get("duration", 1)
        if applied["target"] == "area":
            cost += coefficients["area_factor"] * effect["base_cost"] * applied.get("radius", 0)
        total += cost * coefficients["target_multiplier"][applied["target"]]
    return math.ceil(round(total, 6))


def paid_cost(authored, skill, modifiers, rules):
    """What the caster pays (docs/core/casting.md#paid-cost): the skill discount, rounded up,
    then the modifiers on spell_cost by the kernel rules, at least 1."""
    from kernel_reference import apply_modifiers  # kernel_reference imports this module

    discount = min(skill / rules["skill_discount"]["divisor"], rules["skill_discount"]["cap"])
    base = math.ceil(round(authored * (1 - discount), 6))
    own = [m for m in modifiers if m["target"] == "spell_cost"]
    return max(rules["skill_discount"]["minimum_paid"], apply_modifiers(base, own))


def spell_row(spells, key):
    for spell in spells:
        if spell["key"] == key:
            return spell
    raise LookupError(f'spell "{key}" is not defined')


def roll_save(spell, effects, defender, rules):
    """An instant spell's save and the effects that land (docs/core/casting.md#delivery-and-saves)."""
    from kernel_reference import run_resolution  # kernel_reference imports this module

    attributes = defender.get("attributes", {})
    named = spell["save"]
    if "skill" in named:
        value = defender.get("skills", {}).get(named["skill"].rsplit("/", 1)[-1], 0)
    else:
        value = attributes.get(named["attribute"], 0) * rules["attribute_multiplier"]
    check = run_resolution(
        {"actor": {"skills": {"save": value}, "attributes": attributes}, "skill": "save", "modifiers": [], "roll": defender["roll"]}
    )
    if not check["success"]:
        return {**check, "result": "failed"}, effects
    if "crit" in check["tags"] or check["margin"] >= rules["full_margin"]:
        return {**check, "result": "full"}, []
    scaled = []
    for effect in effects:
        effect = dict(effect)
        for field in ("magnitude", "duration"):
            if field in effect:
                effect[field] = max(1, math.floor(effect[field] * rules["partial_percent"] / 100))
        scaled.append(effect)
    return {**check, "result": "partial"}, scaled


DAMAGE_PREFIX = "core:effect/damage_"
# The combat defender fields a casting target may carry (docs/core/casting.md#damage-to-a-target).
DEFENDER_FIELDS = ("hp", "armor_dr", "shield_dr", "shields_skill", "resist", "evasion_skill", "stance", "flags", "modifiers", "attributes")


def apply_damage(spell, effects, check, vector, caster):
    """Each damage packet through combat steps 2 to 8, the cast check as the attack roll."""
    from kernel_reference import apply_modifiers, run_combat  # kernel_reference imports this module

    types = {row["key"] for row in load_data("combat/damage_types.json")["damage_types"]}
    templates = {f"core:monster_template/{row['key']}": row for row in load_data("monsters/templates.json")["monster_templates"]}
    elements = {templates[ref].get("damage_type") for ref in caster.get("summon_templates", []) if ref in templates}
    bonded = apply_modifiers(0, [m for m in vector.get("modifiers", []) if m["target"] == "bonded_element_damage"])
    target = vector["target"]
    hp = target["hp"]
    packets = []
    for effect in effects:
        kind = effect["effect"].removeprefix(DAMAGE_PREFIX)
        if not effect["effect"].startswith(DAMAGE_PREFIX) or kind not in types or effect.get("duration", 1) > 1:
            continue
        amount = int(effect.get("magnitude", 0))
        if kind in elements:
            amount += bonded
        defender = {field: target[field] for field in DEFENDER_FIELDS if field in target}
        hit = run_combat({
            "attacker": {"skill": check["target"], "roll": vector["roll"], "attributes": caster.get("attributes", {})},
            "defender": {**defender, "hp": hp},
            "damage": {"roll": amount, "type": kind},
            "delivery": spell.get("delivery", "single"),
        })
        if "error" in hit:
            return None, None, hit["error"]
        hp = hit["hp"]
        packets.append({"effect": effect["effect"], "type": kind, "amount": amount, "damage": hit["damage"],
                        "evaded": hit["evaded"], "blocked": hit["blocked"]})
    return packets, hp, None


BOUND_WEAPON_EFFECT = "core:effect/bound_weapon"
DAMAGE_PREFIX = "core:effect/damage_"
LINGERING = "core:effect/lingering_damage"


def bound_weapon(spell, vector, rules):
    """The weapon a bound weapon spell conjures (docs/core/casting.md#bound-weapons), or None
    when the spell conjures none. A coating or an enchantment that does not fit raises
    ValueError."""
    applied = next((e for e in spell["effects"] if e["effect"] == BOUND_WEAPON_EFFECT), None)
    if applied is None:
        return None
    from monster_reference import weapon_row  # the composed weapon row

    weapon = weapon_row(applied["parameter"])
    if weapon is None:
        raise ValueError(f"{applied['parameter']} is not a weapon")
    features = {f"core:feature/{row['key']}": row for row in load_data("features/core.json")["features"]}
    perks = [features[ref] for ref in vector["caster"].get("perks", []) if ref in features]
    material = rules["bound_weapon"]["material"]
    materials = {f"core:material/{row['key']}": row for row in load_data("materials/core.json")["materials"]}
    effects = []
    # The material of the unlocking perk highest on the ladder (its prerequisite level), then the
    # higher weapon.attack, then the key: never the order the caller lists perks in.
    best = None
    for perk in perks:
        for ref in perk.get("unlocks", []):
            if ref.startswith("core:material/"):
                rank = (perk.get("prerequisite", {}).get("level", 0), materials[ref].get("weapon", {}).get("attack", 0), ref)
                if best is None or rank > best:
                    best = rank
        effects += [dict(effect) for effect in perk.get("bound_weapon_effects", [])]
    if best is not None:
        material = best[2]
    registry = {f"core:effect/{row['key']}": row for row in load_data("effects/core.json")["effects"]}
    damage_type, phase = rules["bound_weapon"]["damage_type"], None
    for field, channel in (("coating", "coating"), ("enchantment", "enchantment")):
        ref = vector.get(field)
        if ref is None:
            continue
        row = registry.get(ref)
        if row is None or not ref.startswith(DAMAGE_PREFIX):
            raise ValueError(f"{ref} is not an elemental damage effect")
        if channel not in row.get("channels", []):
            raise ValueError(f"{ref} is not a{'n' if channel[0] in 'aeiou' else ''} {channel}")
        damage_type, phase = ref.removeprefix(DAMAGE_PREFIX), row.get("phase")
    for effect in effects:
        if effect["effect"] == LINGERING and "parameter" not in effect:
            effect["parameter"] = damage_type
    return {
        "weapon": applied["parameter"],
        "skill": weapon["weapon"]["skill"].removeprefix("core:skill/"),
        "dice": weapon["weapon"]["damage"]["dice"],
        "material": material,
        "attack": materials[material].get("weapon", {}).get("attack", 0),
        "damage_type": damage_type,
        "phase": phase,
        "effects": effects,
    }


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
    from kernel_reference import summoning_modifiers  # kernel_reference imports this module

    # A while_summoning modifier holds only while a summon is out (docs/core/casting.md#the-bonded-summoner).
    vector = {**vector, "modifiers": summoning_modifiers(vector.get("modifiers", []), bool(caster.get("summons")))}
    skill = f"{spell['color']}_magic"
    if skill not in caster["skills"]:
        return {"error": f"casting {spell['key']} needs {skill}"}

    target = vector.get("target")
    if spell.get("delivery") == "instant" and target is not None and "roll" not in target:
        return {"error": "an instant spell's save needs the target's roll"}

    block = spell.get("summon")
    count = None
    if block is not None and block.get("minion"):
        count = block.get("count", 1)
        if not isinstance(count, int):
            if "count_roll" not in vector:
                return {"error": f"a minion count of {count} needs count_roll"}
            low, high = dice_range(count)
            if not low <= vector["count_roll"] <= high:
                return {"error": f"count_roll {vector['count_roll']} is not a roll of {count}"}
            count = vector["count_roll"]

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
        "save": None,
        "summons": list(caster.get("summons", [])),
        "dismissed": [],
        "minions": list(caster.get("minions", [])),
        "damage": [],
        "hp": target["hp"] if target is not None and "hp" in target else None,
    }
    cost = spell.get("cost", {})
    reagents = list(caster.get("reagents", []))
    if any(flag in rules["prevented_by"] for flag in caster.get("flags", [])):
        return {**result, "outcome": "prevented"}
    paid = paid_cost(cost.get("mana", 0), caster["skills"][skill], vector.get("modifiers", []), rules)
    if mana < paid:
        return {**result, "outcome": "no_mana"}
    for ref in cost.get("reagents", []):
        if ref not in reagents:
            return {**result, "outcome": "no_reagents"}
        reagents.remove(ref)
    result["mana"] = mana - paid
    result["consumed"] = list(cost.get("reagents", []))

    modifiers = [m for m in vector.get("modifiers", []) if m["target"] == skill]
    armor = caster.get("armor", "none")
    if armor not in rules["armor_penalty"]:
        return {"error": f'armor category "{armor}" is not defined'}
    from kernel_reference import apply_modifiers  # kernel_reference imports this module

    shield = caster.get("shield", "none")
    if shield not in rules["shield_penalty"]:
        return {"error": f'shield "{shield}" is not defined'}
    # Mailed Casting multiplies each penalty, floored (GD_S05_03).
    scale = [m for m in vector.get("modifiers", []) if m["target"] == "armor_cast_penalty"]
    for source, penalty in ((f"armor:{armor}", rules["armor_penalty"][armor]), (f"shield:{shield}", rules["shield_penalty"][shield])):
        penalty = apply_modifiers(penalty, scale)
        if penalty:
            modifiers.append({"target": skill, "operation": "add", "value": penalty, "source": source})
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
    try:
        conjured = bound_weapon(spell, vector, rules)
    except ValueError as error:
        return {"error": str(error)}
    if conjured is not None:
        result["bound_weapon"] = conjured
    if spell.get("delivery") == "instant" and "target" in vector:
        result["save"], result["effects"] = roll_save(spell, result["effects"], vector["target"], rules["save"])
    if target is not None and "hp" in target:
        result["damage"], result["hp"], error = apply_damage(spell, result["effects"], check, vector, caster)
        if error:
            return {"error": error}
    if spell.get("concentration"):
        result["concentration"] = spell["key"]
        if held is not None and held != spell["key"]:
            result["ended"] = held
    if count is not None:
        # Minions join the caster's minions, outside the summon limit (docs/core/casting.md#minions).
        result["minions"] += [block["group"]] * count
    summoned = [block["group"]] if block is not None and count is None else []
    if summoned:
        # The newest join; past the limit the oldest are dismissed (docs/core/casting.md#summons).
        from kernel_reference import apply_modifiers  # kernel_reference imports this module

        limit = apply_modifiers(rules["summon_limit"], [m for m in vector.get("modifiers", []) if m["target"] == "summon_limit"])
        summons = result["summons"] + summoned
        result["dismissed"] = summons[: max(len(summons) - limit, 0)]
        result["summons"] = summons[len(result["dismissed"]):]
    for effect in spell["effects"]:
        if effect["effect"] == "core:effect/dispel":
            for row in standing:
                if row["level"] <= effect["magnitude"] and row["key"] not in result["dispelled"]:
                    result["dispelled"].append(row["key"])
    return result


# Spellcraft (docs/core/casting.md#spellcraft, docs/modules/magic.md#spell-knowledge).

def round_half_up(value):
    """Round to the nearest integer, a half away from zero, as Go's math.Round does."""
    return math.floor(round(value, 6) + 0.5)


def craft_reason(effect, applied, color, known):
    """Why one applied effect cannot go into a crafted spell of `color`, or None."""
    if "spell" not in effect["channels"]:
        return "channel"
    if effect.get("craftable") is False:
        return "uncraftable"
    if applied.get("parameter") in effect.get("uncraftable_parameters", []):
        return "excluded_parameter"
    if applied["effect"] not in known:
        return "unknown"
    if effect.get("color") not in (None, color):
        return "off_color"
    return None


def effect_threshold(effect_ref):
    """An effect's craft threshold: that of the cheapest spell carrying it (magnitude 1, instant, touch)."""
    cost = spell_cost({"effects": [{"effect": effect_ref, "magnitude": 1, "target": "touch"}]})
    return round_half_up(cost * 1.25)


def craftable(effect):
    return "spell" in effect["channels"] and effect.get("craftable") is not False


def run_spellcraft(vector):
    effects = {f"core:effect/{e['key']}": e for e in load_data("effects/core.json")["effects"]}
    caster = vector["caster"]
    known = caster.get("known", [])
    op = vector["operation"]
    if op == "craft":
        spell = vector["spell"]
        rejected = []
        for applied in spell["effects"]:
            if applied["effect"] not in effects:
                return {"error": f'effect "{applied["effect"]}" is not defined'}
            reason = craft_reason(effects[applied["effect"]], applied, spell["color"], known)
            if reason:
                rejected.append({"effect": applied["effect"], "reason": reason})
        if rejected:
            return {"valid": False, "rejected": rejected, "mana": None, "cast_threshold": None,
                    "craft_threshold": None, "skill_ok": False}
        mana = spell_cost(spell)
        craft_threshold = round_half_up(mana * 1.25)
        skill_ok = caster.get("skills", {}).get(f"{spell['color']}_magic", 0) >= craft_threshold
        return {"valid": skill_ok, "rejected": [], "mana": mana, "cast_threshold": mana,
                "craft_threshold": craft_threshold, "skill_ok": skill_ok}
    if op == "learn":
        try:
            spell = spell_row(load_data("spells/core.json")["spells"], vector["spell"])
        except LookupError as error:
            return {"error": str(error)}
        taught = []
        for applied in spell["effects"]:
            ref = applied["effect"]
            if craftable(effects[ref]) and ref not in known and ref not in taught:
                taught.append(ref)
        return {"taught": taught, "known": sorted(set(known) | set(taught))}
    if op == "research":
        from kernel_reference import run_resolution  # kernel_reference imports this module

        ref = vector["effect"]
        if ref not in effects:
            return {"error": f'effect "{ref}" is not defined'}
        effect = effects[ref]
        refused = {"outcome": "refused", "check": None, "known": sorted(known), "days": 0, "cost": None}
        if not craftable(effect):
            return {**refused, "reason": "uncraftable"}
        if ref in known:
            return {**refused, "reason": "known"}
        color = effect.get("color") or vector.get("color")
        if not color:
            return {**refused, "reason": "no_color"}
        skill = f"{color}_magic"
        value = caster.get("skills", {}).get(skill, 0)
        if value < effect_threshold(ref):
            return {**refused, "reason": "skill"}
        check = run_resolution({
            "actor": {"skills": {skill: value}, "attributes": caster.get("attributes", {})},
            "skill": skill,
            "modifiers": [m for m in vector.get("modifiers", []) if m["target"] == skill],
            "roll": vector["roll"],
        })
        research = load_data("casting/core.json")["casting"]["research"]
        learned = check["success"]
        return {"outcome": "learned" if learned else "failed", "reason": None, "check": check,
                "known": sorted(set(known) | ({ref} if learned else set())),
                "days": research["days"], "cost": dict(research["cost"])}
    return {"error": f'spellcraft op "{op}" is not defined'}
