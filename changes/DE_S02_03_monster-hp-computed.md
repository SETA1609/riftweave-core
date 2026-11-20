### DE_S02_03_monster-hp-computed Third normal form: monster hit points are computed
- Removed: `hitPoints` on monster rows, overrides and creature-type bases, `growth.hitPointsPerLevel`, and the minion template's `hit_points_multiplier`; `validate.py` fails any of them (`check_monster_hit_points`) **(breaking)**.
- Added: `monster_hp_multiplier` in `progression/core.json` (standard 0.4, boss 0.8, minion 0.2) and a template's `role`; the minion template sets `role: minion`.
- Changed: a monster's HP is the player formula before perks times its role's multiplier, rounded down and at least 1, then its perks' hp modifiers; the `monster_sheet` and `print` inputs take `role` instead of `hitPoints` **(breaking)**.
- Changed: END on 15 monster rows so their HP stays within 20% of the old authored value and inside the hits-to-kill band; the wolf example rows keep their attributes and move about 60% down (`docs/decisions/DE_S02_03_monster-hp-computed.md`).
