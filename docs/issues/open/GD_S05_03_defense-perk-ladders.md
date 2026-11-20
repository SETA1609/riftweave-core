#### GD_S05_03_defense-perk-ladders Defense perk ladders: shields, shield bash, evasion, no armor, and armor on casting
- Status: review
- Priority: medium
- Needs: M3-27, GD_S05_02_fighting-stances, M5-47
- Was: M3-30
- Modified: 2026-11-19
- Do: Owner request, 2026-11-17: the evasion and shield perks stop at level 12, a few no longer fit M3-27's bands, and the melee builds (M5-47) need defense choices up to level 30. Rework them as tiered perks, add a shield bash, no-armor, and guardian ladder, and make heavier armor cost more on casting.
  - **Tiers 3 to 6 levels apart** (owner 2026-11-17): a tiered perk whose first tier has no level gate takes its next tiers at about Lv6, 12, 18, 24; one whose first tier is gated at level L takes them at about L+6, L+12. The feature schema has `ranks` but no level gate per rank: add one (for example `rank_levels: [1, 6, 12]`) and have `validate.py` check it.
  - **Spells you can block and spells you save against** are two different defenses (docs/core/casting.md#delivery-and-saves), and every text in this issue names which one it means. A **blockable** spell has `single` or `area` delivery: it is rolled against the defender's bands (M3-27), so it can land in the block band. An **instant** spell has nothing to dodge or block: it takes no band, and its defense is the **save** the spell names (an attribute or a skill). Spell Ward works only on blockable spells; it never changes a save.
  - **Shields:**
    - `shield_training`, tiers no gate / Lv6 / Lv12 (Shields 30, END 5): a carried shield's DR +2 / +3 / +4. Tier 1 stays today's +2.
    - `shield_wall`, tiers no gate / Lv6 / Lv12 (Shields 40, END 6): while blocking, the block band is 3 / 6 / 9 wider. It replaces today's `block_reduction` x1.15, which has nothing to reduce once a block stops all damage.
    - `spell_ward`, tiers Lv6 / Lv12 / Lv18 (Shields 50, WIL 6, needs `shield_training`): it no longer stops a spell. While blocking, when a **blockable** spell's roll lands in the block band, the defender resists `END x 2 x tier` percent of its damage, at most 50%. A blockable spell outside the band, and every instant spell, deals its full result. At a table, read the percentage from this table instead of working it out; the data holds it as a lookup row the formula generates and a vector checks:

      | END | Tier 1 | Tier 2 | Tier 3 |
      | --- | --- | --- | --- |
      | 1 | 2% | 4% | 6% |
      | 2 | 4% | 8% | 12% |
      | 3 | 6% | 12% | 18% |
      | 4 | 8% | 16% | 24% |
      | 5 | 10% | 20% | 30% |
      | 6 | 12% | 24% | 36% |
      | 7 | 14% | 28% | 42% |
      | 8 | 16% | 32% | 48% |
      | 9 | 18% | 36% | 50% |
      | 10 | 20% | 40% | 50% |

    - `bulwark`, Lv12 (Shields 70, needs `shield_training` tier 2): unchanged from M3-27 (a block in the band's lowest rolls costs the attacker 2 AP); it no longer needs `spell_ward`, since it is a physical perk.
    - `guardian`, tiers Lv6 / Lv12 / Lv18 (Shields 50, any shield): while blocking, one adjacent ally's evasion band +10 / the block band also covers attacks on that ally / it covers every adjacent ally. The tower shield's built-in "half cover to adjacent allies" is removed from its base; covering others is this perk, with any shield.
  - **Shield bash** (`boss_strike`, tiers no gate / Lv6 / Lv12 / Lv18; Shields 30, STR 5): 1: a bash is a jab (M3-28) with the shield, 2 AP, blunt damage = the shield's DR + 1d4 [A]. 2: the `attack` action also strikes with the shield as a light off-hand weapon (main -10, off -20, M3-28's table). 3: a bash that hits costs the target 2 AP on its next turn. 4: right after a block in your band, a bash costs 1 AP.
  - **Evasion:**
    - `sidestep` unchanged (dodge costs 1 AP less).
    - `armored_evasion`, tiers Lv6 / Lv12 / Lv18 (Evasion 50, needs `sidestep`): +5 / +10 / +15 Evasion.
    - `riposte` unchanged (M3-27).
    - `acrobat` is **dropped from core** (owner 2026-11-17): its stamina discount buys nothing at a table (25% of a 15-stamina dodge is under 1 AP), so it is a perk for a real-time engine, not this ruleset's iteration. Remove the row, its `dodge_stamina` modifier target, and its case in `ruleset/scripts/test_modifier_units.py`; rerun `playtest_driver.py` so the build that took it (`ranged_bows`) picks another perk.
  - **No armor:** `unarmored_grace`, tiers no gate / Lv6 / Lv12 / Lv18 (AGI 6, Evasion 30): with no body armor (clothes and robes allowed), the evasion band +5 / +10 / +15 / +20. `flowing_step`, Lv12 (needs `unarmored_grace` tier 2): with no body armor, the first move each turn draws no opportunity attack.
  - **Armor on casting:** the heavier the gear, the larger the penalty on the casting skill (`casting/core.json` `armor_penalty`, today 0 / 0 / -10 / -25) [A]: none 0, light -5, medium -15, heavy -30; a shield adds its own: buckler 0, standard -5, tower -15. `mailed_casting` replaces `battle_mage` (an Elder Scrolls class name), tiers Lv6 / Lv12 / Lv18 (INT 6, END 6): the armor and shield penalty x0.75 / x0.5 / x0.25.
  - **The melee builds move to the new perks:** M5-47's three melee builds (`ruleset/data/builds/core.json`, `docs/modules/builds/melee.md`) were written with the perks of 2026-11-16. Rework them so each takes the ladders of M3-28 to M3-30 where they fit, with every tier taken at or after its gate, and close the gaps M5-47 recorded in `docs/decisions/M5-47.md`.
  - **Names are original** (M5-67): `boss_strike`, `guardian`, `unarmored_grace`, `flowing_step`, `mailed_casting` are new; `battle_mage` moves with every ref.
- Keys: `rank_levels` on features; the spell ward lookup row; `shield_training`, `shield_wall`, `spell_ward`, `bulwark`, `guardian`, `boss_strike`, `armored_evasion`, `unarmored_grace`, `flowing_step`, `mailed_casting`; `armor_penalty` and a shield penalty in `casting/core.json`
- Files: ~20 (docs/core/combat.md, docs/core/casting.md, docs/modules/combat.md, docs/modules/armor.md, ruleset/schemas/feature.schema.json, ruleset/schemas/casting.schema.json, ruleset/data/features/core.json, ruleset/data/casting/core.json, ruleset/data/modifier_targets/core.json, ruleset/data/equipment/bases/armor/tower_shield.json, ruleset/scripts/validate.py, ruleset/tests/vectors/combat.json, ruleset/tests/vectors/casting.json, ruleset/scripts/test_perks.py, ruleset/scripts/test_modifier_units.py, docs/playtests/report.json, ruleset/data/builds/core.json, docs/modules/builds/melee.md, docs/decisions/GD_S05_03_defense-perk-ladders.md, changes/GD_S05_03_defense-perk-ladders.md)
- See: docs/issues/done/M3-27.md (bands, block and dodge stances, Riposte and Bulwark), docs/issues/open/GD_S05_01_loadout-attacks-melee-ladders.md (the jab and `attack_sequences`), docs/issues/open/GD_S05_02_fighting-stances.md (`buckler_play`), docs/core/casting.md#delivery-and-saves
- Done when:
  - `validate.py` rejects a tier taken before its `rank_levels` gate.
  - A vector shows Spell Ward resisting a blockable spell in the block band at the table's percentage, capped at 50%, and doing nothing outside the band or against an instant spell.
  - Vectors cover a shield bash as a jab and as part of an attack, `unarmored_grace` with and without body armor, and the casting penalty for each armor weight with and without `mailed_casting`.
  - The three melee builds take the new ladders and still pass M5-47's legality test.
  - `battle_mage` and `acrobat` no longer exist as keys, nor `dodge_stamina` as a target; every ref resolves.
  - `validate.py` and the tests pass.
