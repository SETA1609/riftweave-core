#### GD_S05_01_loadout-attacks-melee-ladders Attacks by loadout, the jab, the full attack, and the melee perk ladders
- Status: done
- Merged: #253, 2026-11-19
- Priority: medium
- Needs: M3-27, M5-67
- Was: M3-28
- Modified: 2026-11-19
- Do: Owner request, 2026-11-16 and 2026-11-17: the melee builds (M5-47) need perks for carrying a blow into the next foe, for fighting with two weapons, and for extra strikes. AP is `2 + floor(AGI / 3)` (`derived/core.json`, 2 to 5 AP); today every attack costs 3 AP.
  - **The attack strikes with everything held** (owner 2026-11-17): one `attack` action strikes once with the main hand and once with an off-hand weapon, if there is one. Its AP cost depends on the loadout, read from the weapons' `light` and `two_handed` properties. The strikes of one attack may go to different foes in reach. Each strike is its own roll through the whole pipeline (docs/core/combat.md).
  - **The jab** is a new action: one strike with the off-hand weapon alone, at the off-hand weapon's own cost.
  - **The full attack** is a new action: it spends **all** remaining AP, at least the loadout's attack cost + 1 [A], against **one** target. It makes the loadout's strikes plus the strikes perks add (below), one after the other. If the target drops, the strikes left are lost unless a perk carries them on (Carry Through).
  - **How strikes stack is a table, not prose:** a new `attack_sequences` collection (schema and data) holds, per loadout, the attack's AP cost and its strikes in order, each with its hand and attack modifier. Rows [A]:

    | Loadout | Attack AP | Strikes | Jab AP (off hand only) |
    | --- | --- | --- | --- |
    | one light weapon, off hand empty, a shield, or a buckler | 2 | main 0 | - |
    | one one-handed weapon, off hand empty, a shield, or a buckler | 3 | main 0 | - |
    | one two-handed weapon (or a versatile one in both hands) | 3 | main 0 | - |
    | two light weapons | 3 | main -10, off -20 | 2, at -10 |
    | one-handed main, light off hand (sword and dagger) | 4 | main -10, off -20 | 2, at -10 |
    | two one-handed weapons | 5 | main -10, off -30 | 3, at -20 |

    A two-handed attack stays 3 AP, because an AGI 3 character has only 3 AP; a two-handed weapon pays for itself by leaving no hand for a shield [A]. Perks add strikes or soften the penalties through modifier targets in `modifier_targets/core.json` (`full_attack_strikes`, `offhand_penalty`, `mainhand_penalty`) [A], so the table stays the one place that says how attacks stack. The numbers are [A] until the M5-18 playtest driver reruns a dual-wield build against a one-weapon build of the same level.
  - **Perk names are original** (owner 2026-11-16): no feat or perk name taken from another game. M5-67 renames the other perks that break this rule first, so both issues do not edit the same rows at once. `power_attack` becomes `full_weight` and `cleave` becomes `carry_through`; every ref (features, examples, vectors, docs) moves with them.
  - **Carry Through ladder** (power before carry, owner 2026-11-16):
    - `full_weight` (was `power_attack`): the same bonus, now open to Blades, Blunt, or Piercing 40.
    - `carry_through` (was `cleave`, Lv6): needs `full_weight`. When a strike drops a foe, the next strike of the same action (or one free strike, after an attack) goes to a foe in reach.
    - `reaping_arc` (Lv12, needs `carry_through`): the carry repeats for every foe a strike drops, with no limit but the foes in reach.
    - `circle_cut` (Lv18, two-handed weapon, needs `reaping_arc`): an action that spends all AP and strikes every foe in reach once, at -10. It is the one many-target full attack.
  - **Extra strikes ladder** (any weapon, melee or ranged, owner 2026-11-18; the ranged rows and the reload between strikes are `GD_S17_05_ranged-full-attack`'s):
    - `follow_up` (Lv6): a full attack gains one main-hand strike at -20.
    - `press_on` (Lv18, needs `follow_up`): one more main-hand strike at -30.
  - **Two-weapon ladder:**
    - `off_hand_drill` (Lv4): both two-weapon penalties are 10 smaller.
    - `crossed_guard` (Lv8, needs `off_hand_drill`): while holding two weapons, the evasion band (M3-27) grows by 5 [A].
    - `twin_rhythm` (Lv14, needs `off_hand_drill`): a full attack with two weapons gains a second off-hand strike at the off value -10.
    - `scissor_cut` (Lv20, needs `twin_rhythm`): when a main-hand and an off-hand strike of one full attack both hit the same foe, add the off-hand weapon's damage dice once more.
- Keys: the `attack` cost by loadout; the `jab` and `full_attack` actions; the `attack_sequences` collection; `full_attack_strikes`, `offhand_penalty`, `mainhand_penalty` modifier targets; `full_weight`, `carry_through`, `reaping_arc`, `circle_cut`, `follow_up`, `press_on`, `off_hand_drill`, `crossed_guard`, `twin_rhythm`, `scissor_cut`
- Files: ~14 (docs/core/combat.md, docs/core/actions.md, docs/modules/combat.md, ruleset/schemas/attack_sequence.schema.json, ruleset/data/attack_sequences/core.json, ruleset/data/actions/core.json, ruleset/data/modifier_targets/core.json, ruleset/data/features/core.json, ruleset/data/examples/characters.json, ruleset/scripts/validate.py, ruleset/tests/vectors/combat.json, ruleset/scripts/test_perks.py, docs/decisions/GD_S05_01_loadout-attacks-melee-ladders.md, changes/GD_S05_01_loadout-attacks-melee-ladders.md)
- See: docs/modules/combat.md (two weapons, parry), docs/core/actions.md, ruleset/data/derived/core.json (`ap`), docs/issues/done/M3-27.md, docs/issues/done/M5-47.md
- Done when:
  - A vector runs an attack for each loadout row at the table's AP cost and modifiers, a jab at the off-hand cost, and a full attack whose strikes stop when the target drops and carry on with `carry_through`.
  - A vector shows `follow_up` and `twin_rhythm` adding strikes and `off_hand_drill` softening both penalties.
  - `power_attack` and `cleave` no longer exist as keys; every ref resolves to the new names.
  - `validate.py` and the tests pass.
