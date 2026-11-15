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


def spell_cost(spell):
    """A spell's mana cost from its effects alone (docs/modules/magic.md#cost-formula):
    per effect magnitude_factor x baseCost x magnitude x duration, plus area_factor x baseCost x
    radius for an area, times the target multiplier; the sum rounded up. No caster input."""
    coefficients = load_data("casting/core.json")["casting"]["spell_cost"]
    effects = {e["key"]: e for e in load_data("effects/core.json")["effects"]}
    total = 0.0
    for applied in spell["effects"]:
        effect = effects[applied["effect"].rsplit("/", 1)[-1]]
        magnitude = 1 if effect.get("magnitudeUnit") in UNPRICED_MAGNITUDE_UNITS else applied.get("magnitude", 1)
        cost = coefficients["magnitude_factor"] * effect["baseCost"] * magnitude * applied.get("duration", 1)
        if applied["target"] == "area":
            cost += coefficients["area_factor"] * effect["baseCost"] * applied.get("radius", 0)
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
        "save": None,
        "summons": list(caster.get("summons", [])),
        "dismissed": [],
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
    if rules["armor_penalty"][armor]:
        modifiers.append({"target": skill, "operation": "add", "value": rules["armor_penalty"][armor], "source": f"armor:{armor}"})
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
    if spell.get("delivery") == "instant" and "target" in vector:
        result["save"], result["effects"] = roll_save(spell, result["effects"], vector["target"], rules["save"])
    if spell.get("concentration"):
        result["concentration"] = spell["key"]
        if held is not None and held != spell["key"]:
            result["ended"] = held
    summoned = [e["parameter"] for e in spell["effects"] if e["effect"] == SUMMON_EFFECT for _ in range(int(e.get("magnitude", 1)))]
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
