#!/usr/bin/env python3
"""Balance report (docs/core/balance.md): monsters and gear outside the guideline bands.

The bands and the checks, each a warning or a gate, are in ruleset/data/balance/core.json.
`--gate` exits non-zero on a finding of a gate check that is not accepted.
"""

import argparse
import json
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import validate  # noqa: E402  (base + override resolution of equipment and monsters)

DATA_DIR = Path(__file__).resolve().parent.parent / "data"
CATEGORIES = ("light", "medium", "heavy")
DICE = re.compile(r"^(\d+)d(\d+)([+-]\d+)?$")


def dice_mean(dice):
    match = DICE.match(dice)
    count, sides, bonus = int(match[1]), int(match[2]), int(match[3] or 0)
    return count * (sides + 1) / 2 + bonus


def sized_dice(weapon_row, ladder=None):
    """The dice after the row's size step (combat/dice_ladder.json); normal or no size keeps them."""
    if ladder is None:
        ladder = json.loads((DATA_DIR / "combat/dice_ladder.json").read_text())
    dice = weapon_row["weapon"]["damage"]["dice"]
    step = ladder["sizes"][weapon_row.get("size", "normal")]["dice_step"]
    if step == 0:
        return dice
    rung = next(i for i, rung in enumerate(ladder["dice_ladder"]) if dice in rung)
    return ladder["dice_ladder"][rung + step][0]


def _within(value, band):
    return band[0] <= value <= band[1]


def _tier(bands, level):
    return next(t for t in bands["tiers"] if _within(level, t["levels"]))


def _finding(check, subject, value, band, detail):
    return {"check": check, "subject": subject, "value": value, "band": list(band), "detail": detail}


def _weapon_dice(bands, weapons, _armor, _monsters):
    for row in weapons:
        dice = sized_dice(row)
        mean = dice_mean(dice)
        if not _within(mean, bands["weapon_dice"]):
            yield _finding("weapon_dice", row["key"], mean, bands["weapon_dice"], f"{dice} has mean {mean}")


def _armor_piece_dr(bands, _weapons, armor, _monsters):
    for row in armor:
        category = row.get("armor", {}).get("category")
        if category not in CATEGORIES:
            continue
        top = max(t["worn_dr"][category][1] for t in bands["tiers"])
        if row["armor"]["drBase"] > top:
            yield _finding("armor_piece_dr", row["key"], row["armor"]["drBase"], [0, top], f"{category} piece over the highest {category} set")


def _worn_set_dr(bands, _weapons, armor, _monsters):
    for index, category in enumerate(CATEGORIES):
        allowed = CATEGORIES[: index + 1]
        best = {}
        for row in armor:
            piece = row.get("armor", {})
            if piece.get("category") in allowed and piece.get("slot") != "offhand":
                best[piece["slot"]] = max(best.get(piece["slot"], 0), piece["drBase"])
        total = sum(best.values())
        if not any(_within(total, t["worn_dr"][category]) for t in bands["tiers"]):
            band = [min(t["worn_dr"][category][0] for t in bands["tiers"]), max(t["worn_dr"][category][1] for t in bands["tiers"])]
            yield _finding("worn_set_dr", category, total, band, f"best {category} set: " + ", ".join(f"{s} {v}" for s, v in sorted(best.items())))


def _damage(monster):
    hits = [a["magnitude"] for a in monster.get("abilities", []) if a["effect"].rsplit("/", 1)[-1].startswith("damage_")]
    hits += [a["damage"]["amount"] for a in monster.get("attacks", [])]
    return max(hits) if hits else None


def _monster_damage(bands, _weapons, _armor, monsters):
    for row in monsters:
        damage = _damage(row)
        band = _tier(bands, row["level"])["expected_damage"]
        if damage is not None and not _within(damage, band):
            yield _finding("monster_damage", row["key"], damage, band, f"level {row['level']}, strongest damage {damage}")


def _monster_hits_to_kill(bands, _weapons, _armor, monsters):
    for row in monsters:
        damage = _tier(bands, row["level"])["expected_damage"]
        hits = round(row["hitPoints"] / ((damage[0] + damage[1]) / 2), 1)
        role = row.get("role", "standard")
        band = bands["hits_to_kill"][role]
        if not _within(hits, band):
            yield _finding("monster_hits_to_kill", row["key"], hits, band, f"level {row['level']}, {role}, {row['hitPoints']} HP")


def _json(rel):
    return json.loads((DATA_DIR / rel).read_text())


RESTORE = "core:effect/restore_resource"


def _over_time(effect):
    return effect["effect"] == RESTORE and effect.get("duration", 1) > 1


def toxicity_of(effects):
    """formulas/toxicity.json on a drink's effects: the spell cost on the drinker of its instant
    effects and held buffs, plus half the spell cost of its over-time pool restores."""
    import kernel_reference
    from casting_reference import spell_cost

    def cost(part):
        return spell_cost({"effects": [{**e, "target": "self"} for e in part]}) if part else 0

    variables = {"instant_cost": cost([e for e in effects if not _over_time(e)]),
                 "over_time_cost": cost([e for e in effects if _over_time(e)])}
    return kernel_reference.evaluate_formula(_json("formulas/toxicity.json")["expr"], {}, variables)


def _drinks(consumables):
    return [row for row in consumables if row.get("consumable", {}).get("delivery") == "drink"]


def _toxicity_formula(bands, _weapons, _armor, _monsters, consumables, _spells):
    tolerance = bands["toxicity"]["tolerance"]
    for row in _drinks(consumables):
        static, formula = row["consumable"]["toxicity"], toxicity_of(row["consumable"]["effects"])
        if abs(static - formula) > tolerance:
            yield _finding("toxicity_formula", row["key"], static, [formula - tolerance, formula + tolerance],
                           f"formula gives {formula}")


def _instant_toxicity_above_over_time(_bands, _weapons, _armor, _monsters, _consumables, _spells):
    import kernel_reference

    expr = _json("formulas/toxicity.json")["expr"]
    instant = kernel_reference.evaluate_formula(expr, {}, {"instant_cost": 20, "over_time_cost": 0})
    over_time = kernel_reference.evaluate_formula(expr, {}, {"instant_cost": 0, "over_time_cost": 20})
    if instant <= over_time:
        yield _finding("instant_toxicity_above_over_time", "formulas/toxicity.json", instant, [over_time + 1, instant],
                       f"20 at once gives {instant}, over time {over_time}")


def _restored(effects):
    """Each pool an effect list restores, and its total (magnitude times duration)."""
    totals = {}
    for e in effects:
        if e["effect"] == RESTORE:
            totals[e["parameter"]] = totals.get(e["parameter"], 0) + e["magnitude"] * e.get("duration", 1)
    return totals


def _stock_potion_below_magic(bands, _weapons, _armor, _monsters, consumables, spells):
    from casting_reference import paid_cost

    rules = _json("casting/core.json")["casting"]
    skill = bands["magic_skill"]
    mana_per_toxicity = max((_restored(r["consumable"]["effects"]).get("mana", 0) / r["consumable"]["toxicity"]
                             for r in _drinks(consumables) if r["consumable"]["toxicity"] > 0), default=0)
    best = {}
    for spell in spells:
        mana = paid_cost(spell["cost"]["mana"], skill, [], rules)
        for pool, total in _restored(spell["effects"]).items():
            best[pool] = max(best.get(pool, 0), total / mana)
    for row in _drinks(consumables):
        toxicity = row["consumable"]["toxicity"]
        for pool, total in _restored(row["consumable"]["effects"]).items():
            if pool not in best or toxicity == 0:
                continue
            route = round(best[pool] * mana_per_toxicity, 2)
            rate = round(total / toxicity, 2)
            if rate >= route:
                yield _finding("stock_potion_below_magic", row["key"], rate, [0, route],
                               f"{pool} per toxicity; a Potion of Mana on the best spell gives {route} at skill {skill}")


def _scroll_price(bands, _weapons, _armor, _monsters, consumables, spells):
    import kernel_reference

    economy = _json("economy/core.json")
    gold = kernel_reference.to_copper(economy, {"quantity": 1, "unit": "gp"})
    by_ref = {f"core:spell/{s['key']}": s for s in spells}
    for row in consumables:
        spell = by_ref.get(row.get("consumable", {}).get("spell"))
        if row.get("type") != "scroll" or spell is None:
            continue
        least = bands["scroll_gp_per_mana"] * spell["cost"]["mana"]
        price = kernel_reference.to_copper(economy, row["cost"]) / gold
        if price < least:
            yield _finding("scroll_price", row["key"], price, [least, None], f"{spell['key']} costs {spell['cost']['mana']} mana")


CHECKS = {
    "toxicity_formula": _toxicity_formula,
    "instant_toxicity_above_over_time": _instant_toxicity_above_over_time,
    "stock_potion_below_magic": _stock_potion_below_magic,
    "scroll_price": _scroll_price,
    "weapon_dice": _weapon_dice,
    "armor_piece_dr": _armor_piece_dr,
    "worn_set_dr": _worn_set_dr,
    "monster_damage": _monster_damage,
    "monster_hits_to_kill": _monster_hits_to_kill,
}


CONSUMABLE_CHECKS = {"toxicity_formula", "instant_toxicity_above_over_time", "stock_potion_below_magic", "scroll_price"}


def report(bands, weapons, armor, monsters, consumables=None, spells=None):
    """Every finding, with its check's mode and whether the owner accepted it. The consumable
    checks read consumables and spells, from the data unless given."""
    accepted = {(a["check"], a["subject"]) for a in bands["accepted"]}
    if consumables is None:
        consumables = _json("equipment/consumables.json")["equipment"]
    if spells is None:
        spells = _json("spells/core.json")["spells"]
    findings = []
    for check in bands["checks"]:
        if check["key"] in CONSUMABLE_CHECKS:
            found = CHECKS[check["key"]](bands, weapons, armor, monsters, consumables, spells)
        else:
            found = CHECKS[check["key"]](bands, weapons, armor, monsters)
        for finding in found:
            finding.update(mode=check["mode"], accepted=(finding["check"], finding["subject"]) in accepted)
            findings.append(finding)
    return findings


def gate_failures(findings):
    return [f for f in findings if f["mode"] == "gate" and not f["accepted"]]


def _load(rel, resolve):
    path = DATA_DIR / rel
    data = json.loads(path.read_text())
    errors = []
    resolve(data, errors, path)
    if errors:
        raise ValueError(f"{rel}: {errors}")
    return data


def load_all():
    bands = json.loads((DATA_DIR / "balance/core.json").read_text())
    weapons = _load("equipment/weapons.json", validate._resolve_equipment_bases)["equipment"]
    armor = _load("equipment/armor.json", validate._resolve_equipment_bases)["equipment"]
    monsters = _load("monsters/core.json", lambda data, errors, _path: validate._resolve_monster_bases(data, errors))["monsters"]
    return bands, weapons, armor, monsters


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--gate", action="store_true", help="exit 1 on a finding of a gate check that is not accepted")
    args = parser.parse_args(argv)
    findings = report(*load_all())
    for f in findings:
        mark = "accepted" if f["accepted"] else f["mode"]
        print(f"{mark:8}  {f['check']:20}  {f['subject']:24}  {f['value']} not in {f['band']}  ({f['detail']})")
    failures = gate_failures(findings)
    print(f"{len(findings)} finding(s), {len(failures)} failing a gate.")
    return 1 if args.gate and failures else 0


if __name__ == "__main__":
    sys.exit(main())
