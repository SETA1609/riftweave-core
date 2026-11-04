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


def item_cost(ref):
    """An equipment item's cost: its own, its overrides', or its base's."""
    category, key = ref.split(":", 1)[1].split("/", 1)
    file_name, base_dir = EQUIPMENT[category]
    collection = load_data(f"equipment/{file_name}")
    rows = next(value for value in collection.values() if isinstance(value, list))
    row = next(r for r in rows if r["key"] == key)
    cost = row.get("cost") or row.get("overrides", {}).get("cost")
    if cost is None:
        cost = load_data(f"equipment/bases/{base_dir}/{row['base']}.json")["cost"]
    return cost


def item_value_cp(ref):
    price = load_data("formulas/price.json")["expr"]
    return evaluate(price, {"base_cp": to_copper(item_cost(ref)), "value_percent": 100, "percent": 100})


def entry_value_cp(entry):
    if entry["kind"] == "coins":
        return to_copper(entry["coins"])
    if entry["kind"] == "item":
        return item_value_cp(entry["item"]) * entry["quantity"]
    return 0


def budget_cp(loot, level):
    return evaluate(loot["budget_cp"], {"level": level})


def find_table(loot, vector):
    key = vector.get("table")
    if key is None:
        key = next(d["table"] for d in loot["drops"] if d["monster"] == vector["monster"])
    return next(t for t in loot["tables"] if t["key"] == key)


def expected_value_cp(table):
    """Mean value over the d100 ranges, times the number of rolls."""
    total = sum(entry_value_cp(e) * (e["max"] - e["min"] + 1) for e in table["entries"])
    return total * table["rolls"] / 100


def run_loot(vector):
    loot = load_data("loot/core.json")
    table = find_table(loot, vector)
    rolls = vector["rolls"]
    if len(rolls) != table["rolls"]:
        return {"error": f"{table['key']} takes {table['rolls']} rolls, got {len(rolls)}"}
    coins_cp, items = 0, {}
    for roll in rolls:
        if not 1 <= roll <= 100:
            return {"error": f"roll {roll} is not 1 to 100"}
        entry = next(e for e in table["entries"] if e["min"] <= roll <= e["max"])
        if entry["kind"] == "coins":
            coins_cp += to_copper(entry["coins"])
        elif entry["kind"] == "item":
            items[entry["item"]] = items.get(entry["item"], 0) + entry["quantity"]
    value = coins_cp + sum(item_value_cp(ref) * quantity for ref, quantity in items.items())
    return {
        "table": table["key"],
        "items": [{"item": ref, "quantity": quantity} for ref, quantity in sorted(items.items())],
        "coins_cp": coins_cp,
        "value_cp": value,
        "budget_cp": budget_cp(loot, table["level"]),
    }
