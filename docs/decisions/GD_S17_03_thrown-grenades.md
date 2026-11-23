# GD_S17_03_thrown-grenades Grenades: thrown flasks with a fixed area damage

Date: 2026-11-23

- **Decision:** a grenade is a consumable thrown as an area attack ([combat.md#grenades](../core/combat.md#grenades)).
  - **Rows:** type `grenade` (a new item type), delivery `throw` (already in the schema for bombs), in `equipment/consumables.json`, so a ref is `core:consumable/<key>`. A thrown consumable must name its `radius`, `range_normal` 6, and `range_long` 18, and may name `damage` (`dice`, `type`); no other delivery may. Prices and weights are [A].
  - **The five:** Fire, Frost, and Acid Flasks (2d6, radius 1), the Thunder Flask (2d6 thunder, radius 2), and the Smoke Pot (no damage, radius 2), as the issue's table. [V 2026-11-23, test_equipment]
  - **The throw:** `combat` takes `grenade` and `burst_distance`. Delivery is area, the damage is the grenade's dice and type with no ability, and a defender past the radius is `outside`: no roll, no damage. A roll off the dice, another type, a single delivery, an ability, or a maneuver, called shot, or ammunition with it is refused. [V 2026-11-23, combat vectors in Python and Go]
  - **Riders:** the grenade's effects land on a hit that is not evaded, as a maneuver's do, and come back in the result's `effects`. Each is an effect with the new `grenade` channel, which `CHANNEL_REQUIREMENTS` and `check_grenades` gate. [V 2026-11-23, test_equipment]
- **Effects:** `burn`, `slow`, `stagger`, and `sunder_armor` take the `grenade` channel; `sunder_armor` is no longer reserved, as the Acid Flask uses it. `smoke_screen` is new: it blocks sight in the area for 15 seconds, 3 turns. Its rule (what a creature in the smoke can see) is [U] until a vision rule exists.
- **The Smoke Pot is no attack:** with no damage there is nothing to roll, so `combat` refuses it; the table places the smoke. [A]
- **Riders on an evaded burst:** halved damage still lands, but the riders do not, the same rule as a maneuver's effects. [A]
- **Found:** `equipment.schema.json` holds the item type list twice (`equipmentData` and `equipment`); both took `grenade`. Folding the second into a `$ref` is left for a normal-form issue.
- **Not done here:** the throw's AP is the `attack` row's by rule and needs no action change; merchant stock by level, crafting (GD_S15_02), and the thrower perks (GD_S17_09).
