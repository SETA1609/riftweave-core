# M2 Characters

Roadmap: [M2](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M2 Characters

**Goal:** a character can be built and advanced from level 1 to 30 with real choices.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** perk, skill, race, background, and equipment floors are met, and two multi-level characters are vectors.

Done: M2-01, M2-02, M2-03, M2-04, M2-05, M2-06 ([done.md](done.md#m2-characters))

#### M2-07 Effects over time and food
- Status: review
- Priority: medium
- Needs: M1-04, M1-05, M2-05
- Do: An applied effect's `duration` is seconds, and 1 (the default) is instant; its `magnitude` is per second, so a pool effect (restore, damage, drain) totals magnitude times duration and a stat effect holds its magnitude that long (`schema.json`, since M2-05). One effect serves both: no separate instant and over-time effects. Write the tick rule in the kernel, and fix the seconds per turn for turn-based play (`contract.md` and `modifier.md` say turns). `burn` and `poison_health_drain` are over-time effects with rules of their own; decide whether they become conditions that apply a ticking `damage_*`. Food (`type: food`) is always over time; a potion may be either. Food can be eaten at any time except in combat. Add a `food` origin with a new stacking policy `replace`: a new dish removes every modifier and running tick of the previous dish, even effects the new dish lacks (fries, then a hot dog: only the hot dog counts). Each dish is one source, `food:<key>`. A dish's effects are the combined effects of its ingredients, written on the row; cooking a dish from ingredients is crafting (POST-01).
- Keys: `core:origin/food`; origin `stack` value `replace`; the tick rule for `appliedEffect.duration`
- Files: ~10 (docs/core/modifier.md, docs/core/contract.md, ruleset/schemas/schema.json, ruleset/schemas/origin.schema.json, ruleset/data/origins/core.json, ruleset/data/equipment/consumables.json, ruleset/scripts/kernel_reference.py, backend/, ruleset/tests/vectors/)
- See: docs/modules/equipment.md (consumables), docs/decisions/M2-05.md
- Done when: vectors cover an instant restore, a restore over time and its total, a second dish replacing the first (including an effect only the first had), a potion and a dish stacking, and eating refused in combat; the Python reference and Go pass them.

#### M2-09 Potion toxicity limits repeat drinking
- Status: todo
- Priority: medium
- Needs: M2-07
- Do: Nothing limits how many potions a character drinks, so potions crowd out food. Each `delivery: drink` row gets a `toxicity` value (0 allowed; decide whether `antidote` and `potion_cure_disease` are 0). Drinking adds it to a stored `toxicity` on the actor, which falls per second by the tick rule from M2-07 and clears on a long rest. The limit is a derived row, `tox_cap`, on END (proposal `20 + END × 5` [A]), with stronger potions costing more of it (proposal minor 10, standard 20, greater 35 [A]). Going over the limit applies a condition instead of refusing the drink, so an overdose is a risk the player can choose (proposal [A]; decide its penalties). Food adds no toxicity. M2-07 keeps food out of combat, so food's value here is preparation before a fight and cheap recovery after one; decide whether a dish also speeds toxicity decay. A real-time client and a turn-based table read the same numbers.
- Keys: `toxicity` on consumable rows; `toxicity` on the actor; the `tox_cap` derived row; `core:condition/overdosed`
- Files: ~11 (docs/modules/equipment.md, docs/core/derived.md, ruleset/schemas/equipment.schema.json, ruleset/data/equipment/consumables.json, ruleset/data/derived/core.json, ruleset/data/conditions/core.json, ruleset/data/rest/core.json, ruleset/tests/vectors/, ruleset/scripts/kernel_reference.py, backend/internal/derived/, backend/contract/)
- See: docs/modules/equipment.md (consumables), docs/modules/conditions.md, docs/core/derived.md
- Done when: every `drink` row states `toxicity`; vectors cover a drink raising toxicity, decay over time, the limit at END 1 and END 10, a drink over the limit applying the condition, a long rest clearing toxicity, and a dish adding none; the Python reference and Go pass them.

#### M2-08 Luck growth of HP, mana, and stamina
- Status: review
- Priority: medium
- Needs: M1-03
- Do: Each level-up after the first adds a luck roll to each resource pool, on top of its fixed growth: HP `4 + hp_roll`, mana `2 + mp_roll`, stamina `2 + sp_roll`, each roll from 0 to `⌊LCK / 2⌋`, inclusive, one per pool. The rolls are inputs to the `level_up` op, as `luck_roll` is for skill points (`docs/core/progression.md`), and the result returns each pool's gain. A pool's maximum then depends on its rolls, so it is no longer a pure function of attributes and level: the actor stores its summed `growth` (`hp`, `mp`, `sp`), and the derived stat is `base + attr term + level × per_level_flat + growth`. Growth 0 gives today's values (mana `8 × INT + 2 × level`: 10 to 82 at level 1, 68 to 140 at level 30), and the best rolls add `29 × 5` = 145 at level 30 with LCK 10 (mana 285 at INT 10). Stamina becomes a derived row (`sp`, END, proposal `10 + END × 5 + level × 2` [A]), since the stamina potion and `restore_resource` already name it. Fix the docs that disagree with the data: `progression.md` grants `INT × 8 + 2` mana and `END × 8 + 4` HP per level, where the data adds only the flat part; its step list drops the skill points' luck roll that its table has.
- Keys: `hp_roll`, `mp_roll`, `sp_roll` on `level_up`; `growth` on the actor; the `sp` derived row; the roll cap `⌊LCK / 2⌋` as a formula in `progression/core.json`
- Files: ~13 (docs/core/progression.md, docs/core/derived.md, docs/modules/progression.md, docs/modules/combat.md, ruleset/data/derived/core.json, ruleset/data/progression/core.json, ruleset/schemas/derived.schema.json, ruleset/tests/vectors/progression.json, ruleset/tests/vectors/derived.json, ruleset/scripts/kernel_reference.py, backend/internal/progression/, backend/internal/derived/, backend/contract/)
- See: docs/core/progression.md, docs/core/derived.md, docs/modules/advancement.md (a breakthrough's second growth roll reads the same per-level growth)
- Done when: vectors show a level-up with rolls 0 (today's growth) and with the best rolls, a roll above `⌊LCK / 2⌋` refused, LCK 1 capping the roll at 0, and the derived HP, mana, and stamina reading stored growth; the docs give one per-level rule that matches the data; Python and Go pass.
