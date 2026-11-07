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
    step = ladder["size_steps"][weapon_row.get("size", "normal")]
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
        band = bands["hits_to_kill"]["standard"]
        if not _within(hits, band):
            yield _finding("monster_hits_to_kill", row["key"], hits, band, f"level {row['level']}, {row['hitPoints']} HP")


CHECKS = {
    "weapon_dice": _weapon_dice,
    "armor_piece_dr": _armor_piece_dr,
    "worn_set_dr": _worn_set_dr,
    "monster_damage": _monster_damage,
    "monster_hits_to_kill": _monster_hits_to_kill,
}


def report(bands, weapons, armor, monsters):
    """Every finding, with its check's mode and whether the owner accepted it."""
    accepted = {(a["check"], a["subject"]) for a in bands["accepted"]}
    findings = []
    for check in bands["checks"]:
        for finding in CHECKS[check["key"]](bands, weapons, armor, monsters):
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
