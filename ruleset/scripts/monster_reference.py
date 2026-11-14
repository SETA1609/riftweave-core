#!/usr/bin/env python3
"""Monster reference (docs/modules/monsters.md): attacks, spell-like abilities, power and nature,
generated monsters, and the example rows' HP.

Composition itself is validate.py's (`_resolve_monster_bases`); this module reads the composed rows.
"""

import json
import math
import sys
from functools import lru_cache
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import validate  # noqa: E402  (base + override resolution)

DATA_DIR = Path(__file__).resolve().parent.parent / "data"
NATURAL_PREFIX = "core:natural_weapon/"
WEAPON_PREFIX = "core:weapon/"
FEATURE_PREFIX = "core:feature/"
TEMPLATE_PREFIX = "core:monster_template/"
MONSTER_PREFIX = "core:monster/"


def _load(rel):
    return json.loads((DATA_DIR / rel).read_text())


@lru_cache(maxsize=None)
def _natural_file():
    return _load("natural_weapons/core.json")


def natural_weapons():
    """Natural weapons by key (natural_weapons/core.json)."""
    return {row["key"]: row for row in _natural_file()["natural_weapons"]}


@lru_cache(maxsize=None)
def _weapons():
    path = DATA_DIR / "equipment" / "weapons.json"
    data = {"equipment": validate.equipment_rows(path)}
    errors = []
    validate._resolve_equipment_bases(data, errors, path)
    return {row["key"]: row for row in data["equipment"]}


def weapon_row(ref):
    """The carried weapon a `core:weapon/<key>` ref names, or None."""
    return _weapons().get(ref.removeprefix(WEAPON_PREFIX))


@lru_cache(maxsize=None)
def _ladder():
    return _load("combat/dice_ladder.json")["dice_ladder"]


def _step(dice, steps):
    ladder = _ladder()
    rung = next(index for index, rungs in enumerate(ladder) if dice in rungs)
    return ladder[max(0, min(len(ladder) - 1, rung + steps))][0]


def attack_dice(monster, attack):
    """The dice an attack rolls: a natural weapon's moved by creature size, or a carried weapon's
    by the weapon's own size (docs/modules/monsters.md#natural-weapons)."""
    weapon = attack["weapon"]
    if weapon.startswith(NATURAL_PREFIX):
        dice = natural_weapons()[weapon.removeprefix(NATURAL_PREFIX)]["dice"]
        return _step(dice, _natural_file()["size_steps"][monster.get("size", "medium")])
    import balance_report  # balance_report imports validate; imported here to keep the module light

    return balance_report.sized_dice(weapon_row(weapon))


def attack_skill(attack):
    """The skill an attack rolls to hit with: `unarmed` for a natural weapon (the file's `skill`),
    else the carried weapon's own skill (docs/modules/monsters.md#natural-weapons)."""
    weapon = attack["weapon"]
    ref = _natural_file()["skill"] if weapon.startswith(NATURAL_PREFIX) else weapon_row(weapon)["weapon"]["skill"]
    return ref.rsplit("/", 1)[-1]


UNARMED_TARGETS = ("unarmed_damage", "unarmed_strike_damage")


def attack_damage(actor, attack, roll, modifiers):
    """One hit's damage before the combat pipeline: the dice roll, the attribute modifier, and the
    growth bonus; on an unarmed attack the unarmed perks then add (`unarmed_damage`) and multiply
    (`unarmed_strike_damage`), floored, through the kernel's modifier steps."""
    from kernel_reference import apply_modifiers

    damage = roll + attack_modifier(actor, attack) + attack.get("damage", {}).get("bonus", 0)
    if attack_skill(attack) == "unarmed":
        for target in UNARMED_TARGETS:
            damage = apply_modifiers(damage, [m for m in modifiers if m["target"] == target])
    return damage


SLAM = {"key": "slam", "label": "Slam", "weapon": NATURAL_PREFIX + "slam", "ability": "str"}


def monster_attacks(monster):
    """A row's attacks; a row with none and no damaging ability slams."""
    if monster.get("attacks"):
        return monster["attacks"]
    return [] if damaging_abilities(monster) else [dict(SLAM)]


def feature_attack_dice(feature, size):
    """A feature's natural weapon dice, moved by the actor's size as a creature's are."""
    return attack_dice({"size": size}, {"weapon": feature["natural_weapon"]})


def attack_type(attack):
    """The attack's damage type: its own, else its weapon's."""
    named = attack.get("damage", {}).get("type")
    if named:
        return named
    weapon = attack["weapon"]
    if weapon.startswith(NATURAL_PREFIX):
        return natural_weapons()[weapon.removeprefix(NATURAL_PREFIX)]["damage_type"]
    return weapon_row(weapon)["weapon"]["damage"]["type"]


@lru_cache(maxsize=None)
def _modifier_rule():
    return _load("formulas/attribute_modifier.json")["expression"]


def attack_modifier(monster, attack):
    """The attribute modifier (formulas/attribute_modifier.json) of the attack's `ability`, str by
    default; a missing attribute reads 0. No upper limit."""
    from kernel_reference import evaluate_formula

    value = monster.get("attributes", {}).get(attack.get("ability", "str"), 0)
    return evaluate_formula(_modifier_rule(), {}, {"attribute": value})


def attack_mean(monster, attack):
    """The attack's average damage: dice mean, plus the modifier, plus the growth bonus."""
    import balance_report

    bonus = attack.get("damage", {}).get("bonus", 0)
    return balance_report.dice_mean(attack_dice(monster, attack)) + attack_modifier(monster, attack) + bonus


def damaging_abilities(monster):
    """The spell-like abilities that deal damage: abilities whose effect is a damage_* effect."""
    return [a for a in monster.get("abilities", []) or [] if a["effect"].rsplit("/", 1)[-1].startswith("damage_")]


@lru_cache(maxsize=None)
def _effects():
    return {row["key"]: row for row in _load("effects/core.json")["effects"]}


def effect_phase(ref):
    return _effects()[ref.rsplit("/", 1)[-1]].get("phase")


def ability_cost(ability):
    """Mana for one use: the spell cost formula on the ability's effect (magic.md#cost-formula)."""
    from casting_reference import spell_cost

    return spell_cost({"effects": [ability]})


@lru_cache(maxsize=None)
def _cast_ap():
    return next(row["cost"] for row in _load("actions/core.json")["actions"] if row["key"] == "cast")


def use_ability(monster, ability, mana):
    """Use a spell-like ability: spend its mana and the cast AP, or refuse when the mana is short.
    A creature with no attacks, whose abilities are its only attack, pays no mana."""
    cost = 0 if not monster.get("attacks") else ability_cost(ability)
    if mana < cost:
        return {"refused": "mana"}
    return {"mp": mana - cost, "ap": _cast_ap()}


def raw_monsters():
    return _load("monsters/core.json")["monsters"]


def _templates():
    return validate._monster_templates()


@lru_cache(maxsize=None)
def _features():
    return {row["key"]: row for row in _load("features/core.json")["features"]}


def held_templates(bought):
    """The power and nature a caster holds from the perks bought, in order: the latest per slot."""
    held = {"power": None, "nature": None}
    for ref in bought:
        for unlocked in _features().get(ref.removeprefix(FEATURE_PREFIX), {}).get("unlocks", []):
            if unlocked.startswith(TEMPLATE_PREFIX):
                key = unlocked.removeprefix(TEMPLATE_PREFIX)
                slot = _templates()[key]["slot"]
                if slot in held:
                    held[slot] = key
    return held


def _chain_templates(rows, key):
    """Every template the row and its base monsters apply."""
    found = []
    while key in rows:
        row = rows[key]
        found += row.get("templates", [])
        base = row.get("base", "")
        if not base.startswith(MONSTER_PREFIX):
            break
        key = base.removeprefix(MONSTER_PREFIX)
    return found


def creature_type(ref):
    """The creature type at the root of a row's chain (a template never changes it)."""
    rows = {row["key"]: row for row in raw_monsters()}
    key = ref.removeprefix(MONSTER_PREFIX)
    while rows[key]["base"].startswith(MONSTER_PREFIX):
        key = rows[key]["base"].removeprefix(MONSTER_PREFIX)
    return rows[key]["base"].removeprefix(validate.CREATURE_TYPE_REF_PREFIX)


def generate(ref, power=None, nature=None, level=None):
    """A generated monster: the row as a derived monster at its level or `level`, with the power and
    nature added in the slots it leaves empty (docs/modules/monsters.md#generated-monsters)."""
    raw = raw_monsters()
    rows = {row["key"]: row for row in raw}
    key = ref.removeprefix(MONSTER_PREFIX)
    data = {"monsters": json.loads(json.dumps(raw))}
    validate._resolve_monster_bases(data, [])
    plain = {row["key"]: row for row in data["monsters"]}[key]
    filled = {_templates()[t]["slot"] for t in _chain_templates(rows, key)}
    added = [t for t, slot in ((power, "power"), (nature, "nature")) if t and slot not in filled]
    generated = {"key": f"generated_{key}", "base": ref, "templates": added,
                 "overrides": {"level": level or plain["level"]}}
    data = {"monsters": json.loads(json.dumps(raw)) + [generated]}
    errors = []
    validate._resolve_monster_bases(data, errors)
    if errors:
        raise ValueError(errors)
    return data["monsters"][-1]


def summon(ref, caster):
    """A summoned creature (docs/modules/magic.md#summon-groups): the row at the caster's level with
    the caster's power and nature, then each summon bonus the caster's perks grant adds the
    caster's attribute modifier, never below 0, to its group, through a template's attribute merge."""
    from kernel_reference import evaluate_formula

    held = held_templates(caster.get("perks", []))
    creature = generate(ref, held["power"], held["nature"], caster["level"])
    bonuses = {row["key"]: row for row in _load("summons/bonuses.json")["summon_bonuses"]}
    rule = _modifier_rule()
    adds = {}
    for perk in caster.get("perks", []):
        granted = _features().get(perk.removeprefix(FEATURE_PREFIX), {}).get("summon_bonus")
        if granted is None:
            continue
        bonus = bonuses[granted.rsplit("/", 1)[-1]]
        modifier = evaluate_formula(rule, {}, {"attribute": caster["attributes"].get(bonus["attribute"], 0)})
        for attribute in bonus["attributes"]:
            adds[attribute] = adds.get(attribute, 0) + max(modifier, 0)
    adds = {attribute: add for attribute, add in adds.items() if add}
    if adds:
        validate._apply_monster_template(creature, {"attributes": adds})
    return creature


def example_hp(row):
    """An example row's HP by the player formulas at average rolls: the derived hp plus
    (level - 1) x floor(floor(LCK / 2) / 2) of growth (docs/modules/monsters.md#generated-monsters)."""
    from kernel_reference import run_derived

    derived = run_derived({"actor": {"attributes": row["attributes"], "level": row["level"]}})
    luck = row["attributes"].get("lck", 0)
    return derived["hp"] + (row["level"] - 1) * math.floor((luck // 2) / 2)
