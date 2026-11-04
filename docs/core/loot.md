# Loot

What a defeated monster or a treasure cache leaves behind, as data. The rows are in `ruleset/data/loot/core.json`, the vectors in `ruleset/tests/vectors/loot.json`. Values use the price formula of [economy.md](economy.md), so treasure and prices are one scale. Nothing here needs another module: breakthrough cores (`coreDrop` on monster rows) belong to the breakthrough module and are not loot here. [A]

## Tables

A loot table has a `level` (the encounter it is for), a number of `rolls`, and `entries`. Each entry covers a range of a d100 roll (`min` to `max`, inclusive) and is one of:

| kind | yields |
| --- | --- |
| `coins` | `coins` (`quantity` and `unit`) |
| `item` | `quantity` of the namespaced `item` |
| `nothing` | nothing |

The ranges of one table cover 1 to 100 exactly once, so every roll lands on one entry. A table is rolled `rolls` times; each roll is independent, and repeats add up (two rolls on the same coins entry give twice the coins). The rolls are inputs, so a vector pins them; the reference does not roll. [V 2026-11-04]

## Drops

`drops` maps each monster (`core:monster/<key>`) to the key of one table. Every core monster has exactly one entry. A monster's own row does not name its table, so a module can give a monster other loot by adding a table and a drop row, without touching the bestiary. [A]

## Value and budget

```
value of coins   = quantity * the unit's copper value          economy.md
value of an item = price(base_cp, value_percent = 100, percent = 100) * quantity
budget_cp(level) = 500 * level * level                          formula tree
```

- An item is valued at its list price: its `cost` (from its row, its overrides, or its base) through the price formula with no material and no merchant spread. What a merchant pays for it is the economy's business. [A]
- **Consistency:** a table's expected value, the mean over its d100 ranges times `rolls`, must not exceed `budget_cp` at the table's level. A test checks every table. A single lucky roll may exceed the budget; the average may not. [V 2026-11-04]
- The budget is per encounter. It grows with the square of the level: 5 gp at level 1, 125 gp at level 5, 500 gp at level 10, 2000 gp at level 20, against the 100 gp a character starts with ([character creation](../modules/character-creation.md)). [A]

## Not covered

Unique and quest items, magic item generation from materials and enchantments, a per-level wealth curve for whole campaigns, and loot shared between party members. [U]
