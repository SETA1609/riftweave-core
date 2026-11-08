#!/usr/bin/env python3
"""Python reference for loot tables (docs/core/loot.md).

Rolls are inputs. Items are valued at list price with the price formula of
docs/core/economy.md, so treasure and prices are one scale.
"""

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"
EQUIPMENT = {
    "weapon": ("weapons.json", "weapons"),
    "armor": ("armor.json", "armor"),
    "consumable": ("consumables.json", None),
    "accessory": ("accessories.json", "accessories"),
    "gear": ("gear.json", None),
}


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def evaluate(node, variables):
    """The formula tree for the vars-only trees loot reads (lit, var, add, sub, mul, div)."""
    if "lit" in node:
        return node["lit"]
    if "var" in node:
        return variables[node["var"]]
    args = [evaluate(arg, variables) for arg in node["args"]]
    op = node["op"]
    if op == "add":
        return sum(args)
    if op == "sub":
        return args[0] - sum(args[1:])
    if op == "mul":
        product = 1
        for arg in args:
            product *= arg
        return product
    if op == "div":
        return args[0] // args[1]
    raise NotImplementedError(f"formula op {op!r} has no reference")


def to_copper(cost):
    values = {d["unit"]: d["value"] for d in load_data("economy/core.json")["currency"]["denominations"]}
    return cost["quantity"] * values[cost["unit"]]


def item_field(ref, field):
    """An equipment item's field: its own, its overrides', or its base's (None if none)."""
    category, key = ref.split(":", 1)[1].split("/", 1)
    file_name, base_dir = EQUIPMENT[category]
    collection = load_data(f"equipment/{file_name}")
    rows = next(value for value in collection.values() if isinstance(value, list))
    row = next(r for r in rows if r["key"] == key)
    value = row.get(field, row.get("overrides", {}).get(field))
    if value is None and "base" in row:
        value = load_data(f"equipment/bases/{base_dir}/{row['base']}.json").get(field)
    return value


def item_cost(ref):
    """An equipment item's cost: its own, its overrides', or its base's."""
    return item_field(ref, "cost")


def landed_material(ref, material):
    """The material that lands on an item (docs/core/loot.md#what-lands-on-an-item): the row's,
    if the item's default_materials include it, else the first default. No row material, or
    an item that takes none: None."""
    if material is None or not takes_material(ref):
        return None
    defaults = item_field(ref, "default_materials") or []
    key = material.rsplit("/", 1)[-1]
    if not defaults or key in defaults:
        return material
    return f"core:material/{defaults[0]}"


def is_weapon(ref):
    return ref.split(":", 1)[1].split("/", 1)[0] == "weapon"


def item_size(ref, size):
    """A loot row's size on an item: weapons only; normal or none is no size."""
    return size if size not in (None, "normal") and is_weapon(ref) else None


def sized_weapon(ref, size):
    """The sized dice and the attack change of a weapon at a size (combat/dice_ladder.json)."""
    ladder = load_data("combat/dice_ladder.json")
    step = ladder["sizes"][size]
    dice = item_field(ref, "weapon")["damage"]["dice"]
    rung = next(i for i, rung in enumerate(ladder["dice_ladder"]) if dice in rung)
    return ladder["dice_ladder"][rung + step["dice_step"]][0], step["attack"]


MATERIAL_CATEGORIES = ("weapon", "armor")


def takes_material(ref):
    """Only weapons and armor take a loot row's material."""
    return ref.split(":", 1)[1].split("/", 1)[0] in MATERIAL_CATEGORIES


def value_percent(material):
    """100, or the material's valueFactor times 100 (docs/core/economy.md)."""
    if material is None:
        return 100
    key = material.rsplit("/", 1)[-1]
    row = next(m for m in load_data("materials/core.json")["materials"] if m["key"] == key)
    return round(row["valueFactor"] * 100)


def item_value_cp(ref, material=None, size=None):
    """List price at value_percent = material percent x size percent / 100, rounded down. The
    material and size are the ones that landed (landed_material, item_size)."""
    price = load_data("formulas/price.json")["expr"]
    percent = value_percent(material)
    if size is not None:
        percent = percent * load_data("combat/dice_ladder.json")["sizes"][size]["value_percent"] // 100
    return evaluate(price, {"base_cp": to_copper(item_cost(ref)), "value_percent": percent, "percent": 100})


def entry_value_cp(entry, material=None, size=None):
    if entry["kind"] == "coins":
        return to_copper(entry["coins"])
    if entry["kind"] == "item":
        ref = entry["item"]
        return item_value_cp(ref, landed_material(ref, material), item_size(ref, size)) * entry["quantity"]
    return 0


def budget_cp(loot, level):
    return evaluate(loot["budget_cp"], {"level": level})


def table_by_key(loot, key):
    return next(t for t in loot["tables"] if t["key"] == key)


def entry_for(table, roll, shift=0):
    """The entry a roll lands on after its shift, kept between 1 and 100."""
    roll = max(1, min(roll + shift, 100))
    return next(e for e in table["entries"] if e["min"] <= roll <= e["max"])


def monster_level(ref):
    """A core monster's level: its overrides', or its base's."""
    key = ref.rsplit("/", 1)[-1]
    row = next(m for m in load_data("monsters/core.json")["monsters"] if m["key"] == key)
    if "level" in row.get("overrides", {}):
        return row["overrides"]["level"]
    if "level" in row:
        return row["level"]
    return load_data(f"monsters/bases/{row['base']}.json")["level"]


def expected_value_cp(table, shift=0, material=None, rolls=None, size=None):
    """Mean value over the d100 rolls after the shift, times the number of rolls."""
    total = sum(entry_value_cp(entry_for(table, roll, shift), material, size) for roll in range(1, 101))
    return total * (table["rolls"] if rolls is None else rolls) / 100


def row_expected_value_cp(loot, row):
    """One loot row's mean value, times chance / 100."""
    chance = row.get("chance", 100) / 100
    material = row.get("material")
    size = row.get("size")
    if "item" in row:
        ref = row["item"]
        return item_value_cp(ref, landed_material(ref, material), item_size(ref, size)) * row.get("quantity", 1) * chance
    table = table_by_key(loot, row["table"])
    return expected_value_cp(table, row.get("shift", 0), material, row.get("rolls"), size) * chance


def monster_expected_value_cp(loot, drop):
    return sum(row_expected_value_cp(loot, row) for row in drop["loot"])


def _add(items, ref, quantity, material, size=None):
    key = (ref, landed_material(ref, material), item_size(ref, size))
    items[key] = items.get(key, 0) + quantity


def _result(loot, items, coins_cp, level):
    value = coins_cp + sum(item_value_cp(ref, material, size) * q for (ref, material, size), q in items.items())
    listed = []
    for (ref, material, size), quantity in sorted(items.items(), key=lambda kv: (kv[0][0], kv[0][1] or "", kv[0][2] or "")):
        item = {"item": ref, "quantity": quantity}
        if material is not None:
            item["material"] = material
        if size is not None:
            dice, attack = sized_weapon(ref, size)
            item.update(size=size, dice=dice, attack=attack)
        listed.append(item)
    return {"items": listed, "coins_cp": coins_cp, "value_cp": value, "budget_cp": budget_cp(loot, level)}


def looter_shift(vector):
    """Luck plus the looter's loot_roll modifiers, after kernel stacking (docs/core/loot.md#looter).

    Luck emits add loot_roll from formulas/loot_luck.json, source attr:lck, origin innate.
    No actor and no modifiers: 0.
    """
    from kernel_reference import apply_modifiers, evaluate_formula  # kernel_reference imports this module

    modifiers = [m for m in vector.get("modifiers", []) if m["target"] == "loot_roll"]
    actor = vector.get("actor")
    if actor is not None:
        expr = load_data("formulas/loot_luck.json")["expr"]
        luck = evaluate_formula(expr, {"lck": actor.get("attributes", {}).get("lck", 0)})
        modifiers.insert(0, {"target": "loot_roll", "op": "add", "value": luck, "source": "attr:lck", "origin": "core:origin/innate"})
    return apply_modifiers(0, modifiers)


def chance_target(chance, shift):
    """A row's chance after the looter shift, kept between 1 and 99."""
    return max(1, min(chance + shift, 99))


def roll_table(table, rolls, shift, material, items, size=None):
    """Roll one table; return the copper it yields and add its items."""
    coins_cp = 0
    for roll in rolls:
        entry = entry_for(table, roll, shift)
        if entry["kind"] == "coins":
            coins_cp += to_copper(entry["coins"])
        elif entry["kind"] == "item":
            _add(items, entry["item"], entry["quantity"], material, size)
    return coins_cp


def run_loot(vector):
    loot = load_data("loot/core.json")
    rolls = vector["rolls"]
    for roll in rolls:
        if not 1 <= roll <= 100:
            return {"error": f"roll {roll} is not 1 to 100"}
    items = {}
    looter = looter_shift(vector)
    if "table" in vector:
        table = table_by_key(loot, vector["table"])
        if len(rolls) != table["rolls"]:
            return {"error": f"{table['key']} takes {table['rolls']} rolls, got {len(rolls)}"}
        coins_cp = roll_table(table, rolls, looter, None, items)
        return {"table": table["key"], **_result(loot, items, coins_cp, table["level"]), "looter_shift": looter}

    monster = vector["monster"]
    drop = next(d for d in loot["drops"] if d["monster"] == monster)
    queue, used, coins_cp, tables = list(rolls), 0, 0, []

    def take(n):
        nonlocal used
        if len(queue) < n:
            raise IndexError
        used += n
        return [queue.pop(0) for _ in range(n)]

    try:
        for row in drop["loot"]:
            if "chance" in row and take(1)[0] > chance_target(row["chance"], looter):
                continue
            material = row.get("material")
            if "item" in row:
                _add(items, row["item"], row.get("quantity", 1), material, row.get("size"))
                continue
            table = table_by_key(loot, row["table"])
            tables.append(table["key"])
            coins_cp += roll_table(table, take(row.get("rolls", table["rolls"])), row.get("shift", 0) + looter, material, items, row.get("size"))
    except IndexError:
        return {"error": f"{monster}: ran out of rolls after {used}"}
    if queue:
        return {"error": f"{monster}: {len(rolls)} rolls given, {used} used"}
    return {"tables": tables, **_result(loot, items, coins_cp, monster_level(monster)), "looter_shift": looter}
