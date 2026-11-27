# GD_S17_02_ammunition-items-and-reloading Ammunition items, the reload action, and recovering spent ammunition

Date: 2026-11-23

- **Decision:** ammunition is spent, reloaded, and recovered by rule ([combat.md#ammunition](../core/combat.md#ammunition)).
  - **Kinds:** `ammoKind` in `schema.json`: `arrow`, `bolt`, `bullet`, `scatter`, `stone`. Each base with the `ammunition` property names its kind in `weapon.ammo` (bows `arrow`, crossbows `bolt`, pistol and arquebus `bullet`, blunderbuss `scatter`, sling `stone`). [A]
  - **Rows:** equipment items of type `ammunition` in `equipment/ammunition.json`, refs `core:ammunition/<key>` (the issue left the place open): arrows 20, crossbow bolts 20, bullets 10, scatter charges 10, sling stones 20, with an optional `damage_type` and `effects` for GD_S17_04 and GD_S17_11. Prices and weights are [A].
  - **Spending:** `combat` takes an optional `ammunition` (`rounds`, `loaded`): a weapon with `ammunition` or `thrown` spends one round; none left, or a `loading` weapon not loaded, is no attack. Leaving the input out is a table that does not count plain rounds. [V 2026-11-23, combat vectors in Python and Go]
  - **Reload:** a `reload` action, 2 AP. Fast Reload now emits `add reload -1`, as Sidestep emits `add dodge -1`, so the `actions` command applies it with no new rule; the pending `reload_cost` modifier target is removed. [V 2026-11-23, actions vectors]
  - **Recovery:** `combat/ammunition.json` holds the share per kind (arrow and bolt 0.5, bullet, scatter, and stone 0, thrown 1); the new `ammunition` command's `recover` rounds each down. [V 2026-11-23, ammunition vectors]
- **Why a command for recovery:** the share is data GD_S17_10's perks change, and Python and Go must agree on the rounding; a command with vectors pins both. The `modifiers` input is accepted for those perks and not read yet. [A]
- **Repeating crossbow:** no `loading`, so its magazine feeds each attack and its size is not counted. [A]
- **Differences from the issue:** `fast_reload` targets `reload`, not `reload_cost` (breaking for a caller that emitted `reload_cost`); the recovery is a command, not prose.
- **Vector corrected after red:** the two shot vectors first expected 6 damage and 24 HP; piercing damage meets DR 2 for 7 under the existing combat rule, so their `damage` and `hp` were corrected to 7 and 23 in the data step. The ammunition fields the vectors test were not changed.
- **Not done here:** `attack_sequence` with a ranged loadout spending a round per strike waits for GD_S17_05's ranged rows; the catalog's Compendium does not list ammunition yet.
