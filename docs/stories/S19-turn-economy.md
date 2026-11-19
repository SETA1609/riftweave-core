# S19 Turn-based action economy

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S19 Turn-based action economy

**Goal:** A turn at the table is a set of small, priced choices, as in Fallout and Divinity: Original Sin 2, not one attack or one spell. A move goes a set distance; drawing, swapping, and readying a weapon each have a cost; unspent AP either guards you or carries a little, never into a burst; a fight starts on a partial pool; perks, traits, and conditions can change the pool; a minor spell costs less AP than a major one; and initiative ties break one way everywhere.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S19_01_file-turn-economy-issues` | PO | File the turn-based action economy issues | — | — |
| `GD_S19_02_movement-distance-per-ap` | GD | How far one move goes: squares per AP from the race's speed | — | — |
| `GD_S19_03_swap-draw-costs` | GD | Drawing, sheathing, and swapping a loadout cost AP | — | GD_S05_02_fighting-stances |
| `GD_S19_04_ready-and-opportunity-attacks` | GD | Ready an attack on reserved AP, and opportunity attacks in the pipeline | — | GD_S19_02_movement-distance-per-ap |
| `GD_S19_05_ap-bank-cap-and-guard` | GD | Carry a little AP, and turn the rest into guard | — | — |
| `GD_S19_06_combat-start-ap` | GD | A fight starts on a partial pool | — | GD_S19_05_ap-bank-cap-and-guard |
| `GD_S19_07_ap-perks-traits-haste-slow` | GD | Perks, creation traits, and conditions that change the AP pool | — | GD_S19_05_ap-bank-cap-and-guard |
| `GD_S19_08_spell-ap-by-level` | GD | A minor spell costs less AP than a major one | — | — |
| `GD_S19_09_initiative-ties` | GD | One rule for initiative ties | — | — |
| `GD_S19_10_side-initiative-table-module` | GD | Side initiative for the table: fast foes, players, slow foes | — | GD_S19_09_initiative-ties, GD_S16_01_optional-defence-roll |

**Done when:** every issue above is done, `docs/core/actions.md` prices every turn choice above with vectors in Python and Go, and a playtest turn at pool 3 and pool 4 shows at least two choices each turn for every loadout of `attack_sequences/core.json`.
