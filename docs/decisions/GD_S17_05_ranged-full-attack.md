# GD_S17_05_ranged-full-attack The full attack and the extra-strike perks with ranged weapons

Date: 2026-11-23

- **Decision:** the attack by loadout covers ranged weapons ([combat.md#ranged-loadouts](../core/combat.md#ranged-loadouts)).
  - **Vocabulary:** `mainHand` gains `bow`, `hand_gun` (a hand crossbow or a pistol), `long_gun` (a two-handed crossbow, the arquebus, the blunderbuss), and `light_thrown`; `offHand` gains `hand_gun` and `light_thrown`. They are `ranged_hands` in `attack_sequences/core.json`, so which kinds strike at range is data. [A]
  - **Rows:** the issue's six, as `bow`, `hand_gun`, `long_gun`, `two_hand_guns`, `two_light_thrown`, and `mixed_hand` (a light or one-handed melee main hand with an off-hand `hand_gun` or `light_thrown`). [V 2026-11-23, attack_sequence vectors in Python and Go]
  - **Reach:** a foe may be `in_reach: false`, in range only. A melee strike may not aim at it or carry to it; a ranged strike may, so Carry Through and Reaping Arc read "in range" at range. A ranged strike passes `range: ranged` to `combat`, so only it reads a situation's ranged row (shooting into melee). [V 2026-11-23]
  - **Rounds:** `ammunition` per hand, as `combat`'s; a counted ranged hand spends a round a strike. A `loading` hand fires once and reloads before its next strike. In a full attack the reload spends the budget over the row's attack AP; its cost is the `reload` action with its modifiers (Fast Reload `add reload -1`) plus `full_attack.one_handed_reload_extra` (1) when the other hand holds anything. A reload that does not fit ends the full attack, and the strikes left are `stopped`. In an attack or a jab, an unloaded or empty counted hand is an error. [V 2026-11-23]
  - **Opportunity:** a ranged main hand makes no opportunity attack (`reason: ranged`). [V 2026-11-23]
- **Differences from the issue:**
  - **Jabs:** the three rows with an off-hand ranged weapon have a jab (2 AP, -10, as two light weapons): the existing gate requires one for every off-hand weapon, and an off-hand shot alone is plausible. [A]
  - **One-handed reload everywhere:** the issue names the extra AP for the mixed hand; it applies to any reload with the other hand full (two pistols too), since the reason is the same. [A]
  - **A circle cut** strikes only the foes in reach.
- **Vector corrected after red:** the new vectors' piercing strikes first expected DR 2 to take 2; piercing meets DR 2 for 1 under the existing combat rule, so their `damage` and `hp` were taken from the reference. Rounds, reloads, stops, carries, and aims were not changed.
- **Not done here:**
  - The mixed hand's opportunity attack (its melee hand alone) is [U] until GD_S17_13.
  - The `actions` command does not know the loadout, so its `reload` row does not add the one-handed AP; a caller adds `add reload 1`.
  - The aimed shot is GD_S17_06's held breath.
