# M2 Characters

Roadmap: [M2](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M2 Characters

**Goal:** a character can be built and advanced from level 1 to 30 with real choices.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** perk, skill, race, background, and equipment floors are met, and two multi-level characters are vectors.

Done: M2-01, M2-02, M2-03, M2-04, M2-05, M2-06 ([done.md](done.md#m2-characters))

#### M2-07 Effects over time and food
- Status: todo
- Priority: medium
- Needs: M1-04, M1-05, M2-05
- Do: An applied effect's `duration` is seconds, and 1 (the default) is instant; its `magnitude` is per second, so a pool effect (restore, damage, drain) totals magnitude times duration and a stat effect holds its magnitude that long (`schema.json`, since M2-05). One effect serves both: no separate instant and over-time effects. Write the tick rule in the kernel, and fix the seconds per turn for turn-based play (`contract.md` and `modifier.md` say turns). `burn` and `poison_health_drain` are over-time effects with rules of their own; decide whether they become conditions that apply a ticking `damage_*`. Food (`type: food`) is always over time; a potion may be either. Food can be eaten at any time except in combat. Add a `food` origin with a new stacking policy `replace`: a new dish removes every modifier and running tick of the previous dish, even effects the new dish lacks (fries, then a hot dog: only the hot dog counts). Each dish is one source, `food:<key>`. A dish's effects are the combined effects of its ingredients, written on the row; cooking a dish from ingredients is crafting (POST-01).
- Keys: `core:origin/food`; origin `stack` value `replace`; the tick rule for `appliedEffect.duration`
- Files: ~10 (docs/core/modifier.md, docs/core/contract.md, ruleset/schemas/schema.json, ruleset/schemas/origin.schema.json, ruleset/data/origins/core.json, ruleset/data/equipment/consumables.json, ruleset/scripts/kernel_reference.py, backend/, ruleset/tests/vectors/)
- See: docs/modules/equipment.md (consumables), docs/decisions/M2-05.md
- Done when: vectors cover an instant restore, a restore over time and its total, a second dish replacing the first (including an effect only the first had), a potion and a dish stacking, and eating refused in combat; the Python reference and Go pass them.
