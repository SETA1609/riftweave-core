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

`drops` gives each monster (`core:monster/<key>`) a `loot` list: the rows it rolls when defeated, in order. Every core monster has exactly one drop row. A monster's own row does not name its loot, so a module can give a monster other loot, or add a row, without touching the bestiary. [A]

A loot row is a table or a single item:

| field | on | meaning | default |
| --- | --- | --- | --- |
| `table` | table row | key of a table in this file | — |
| `rolls` | table row | how many times this monster rolls the table | the table's `rolls` |
| `shift` | table row | added to every roll on the table, capped at 100 | 0 |
| `item`, `quantity` | item row | one namespaced item | quantity 1 |
| `chance` | either | a d100 roll-under gate: the row yields only if a roll is at most `chance` | 100, no roll |
| `material` | either | the material every weapon and armor the row yields is made of, if the item is found in it (below) | the item's own |
| `size` | either | `undersized`, `normal`, or `oversized`: the size of every weapon the row yields | `normal` |

So monsters can share a table and still differ:

- **Rates:** two rolls instead of one, or a `chance` below 100.
- **Extras:** a second table row, or an item row (a skeleton's spear, a fire elemental's fire resistance potion).
- **Better results:** a `shift`. Entries in a table go from the least valuable to the most, so +15 skips the bottom 15 results: a dire wolf rolls the beast den with +15, a grove stalker without. A test checks the order. [V 2026-11-06]
- **Better versions:** a `material`. A stone golem's hoard chainmail is steel, worth twice the iron one. Only weapons and armor take a material. [A]
- **Bigger or smaller:** a `size`. A giant's hoard weapons are oversized, a goblin's undersized. Only weapons take a size. [V 2026-11-08]

## What lands on an item

A loot row composes the items it yields; sized and material variants are not rows ([weapons.md](../modules/weapons.md#rows-and-composed-items)). [V 2026-11-08]

- **Material.** Each equipment base lists its `default_materials`, the materials it is found and sold in (a row may list its own). A row's material lands on an item whose defaults include it. Otherwise the item drops in its first default material: the steel hoard yields a steel breastplate and an elven wood longbow, never a steel bow. Any other material is player-crafted only. A row with no material leaves the item as listed, at 100 percent. An item with no defaults keeps the row's material.
- **Size.** A weapon from a row with a `size` moves its dice one rung of `combat/dice_ladder.json` (down for undersized, up for oversized) and its attack by the size's `attack` (+10, −10). The result names the `size`, the sized `dice`, and the `attack`. Armor and everything else ignore the size.
- **Value.** `value_percent = material percent × size percent / 100`, rounded down: the material's `valueFactor` times 100 (100 with none) and the size's `value_percent` (75, 100, 150). An oversized elven wood longbow is 400 × 150 / 100 = 600 percent.

**Roll order.** The rolls are one list, used in row order: for each row, first its `chance` roll (if `chance` is below 100), then, if it yields, the table's rolls. A failed chance uses no table rolls. Too few or too many rolls is an error. [V 2026-11-06]

## Looter

Who loots matters too. The looter's shift is one number, added to every table roll on top of the row's own `shift`, and to every `chance`:

```
looter_shift = value of target `loot_roll`, starting at 0, after the looter's modifiers
luck         = add loot_roll ((LCK - 5) * 2), source attr:lck, origin innate      formulas/loot_luck.json
```

- **Luck** emits one modifier, the way it widens the crit window in [resolution.md](resolution.md). LCK 5 is neutral, LCK 10 is +10, LCK 1 is −8. An average looter gets the average the budget is checked against. [A]
- **Perks and everything else** emit `add` on `loot_roll`: the Scavenger perk gives +10 per rank. They stack by the kernel rules ([modifier.md](modifier.md)): two `highest` origins of one kind keep the strongest, a potion of fortune does not stack with another potion. [V 2026-11-06]
- **Table rolls:** roll + row shift + looter shift, kept between 1 and 100. A negative looter shift (a curse, bad luck) moves toward the bottom of the table.
- **Chance:** a row yields if its roll is at most `chance` + looter shift, kept between 1 and 99, so luck never makes a drop certain and bad luck never makes it impossible. A row with no `chance` always yields and is not affected. [A]
- A table rolled on its own (a treasure cache) takes the looter shift the same way. With no `actor` and no `modifiers` in the input, the looter shift is 0.
- The budget check uses a looter shift of 0. A lucky looter beats the budget on average; that is the point of luck. [A]

## Value and budget

```
value of coins   = quantity * the unit's copper value          economy.md
value of an item = price(base_cp, value_percent, percent = 100) * quantity
                   value_percent = material percent * size percent / 100, rounded down
budget_cp(level) = 500 * level * level                          formula tree
```

- An item is valued at its list price: its `cost` (from its row, its overrides, or its base) through the price formula with no merchant spread, and with no material unless its loot row names one. What a merchant pays for it is the economy's business. [A]
- **Consistency:** a table's expected value, the mean over its d100 ranges times `rolls`, must not exceed `budget_cp` at the table's level. A monster's expected value, the sum of its loot rows (each row's mean with its shift and material, times its rolls, times `chance / 100`), must not exceed `budget_cp` at the monster's level. A test checks every table and every monster. A single lucky roll may exceed the budget; the average may not. [V 2026-11-06]
- A material sets `value_percent` to its `valueFactor` times 100, as in [economy.md](economy.md). [A]
- The budget is per encounter. It grows with the square of the level: 5 gp at level 1, 125 gp at level 5, 500 gp at level 10, 2000 gp at level 20, against the 100 gp a character starts with ([character creation](../modules/character-creation.md)). [A]

## Not covered

Unique and quest items, magic item generation from enchantments, a per-level wealth curve for whole campaigns, and loot shared between party members. [U]
