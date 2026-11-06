# M2 Characters

Roadmap: [M2](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M2 Characters

**Goal:** a character can be built and advanced from level 1 to 30 with real choices.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** perk, skill, race, background, and equipment floors are met, and two multi-level characters are vectors.

Done: M2-02, M2-04 ([done.md](done.md#m2-characters))

#### M2-01 Perk catalogue to the content floor
- Status: doing
- Priority: high
- Needs: M1-05, M1-06
- Was: RW-11
- Do: 23 features exist (13 perks, 10 creation). A character gains a perk about every 3 levels, so 10 or more by level 30; real choice needs a wide pool.
- Files: ~3 (ruleset/data/features/core.json, docs/modules/progression.md)
- See: ruleset/data/features/core.json, docs/modules/progression.md
- Done when: at least 60 perks, spread across all 8 attributes and the combat, magic, and social skill groups; every perk emits modifiers; prerequisites are keys; no perk is dominated by another at the same level.

#### M2-03 Multi-level example characters as vectors
- Status: review
- Priority: high
- Needs: M1-03, M1-05
- Was: RW-13
- Do: Single-level creation examples exist. Walk two characters (one martial, one caster) from creation to level 10, with perks, tag skills, gear, and spells, and encode them as vectors.
- Files: ~4 (ruleset/tests/vectors/characters.json, docs/modules/character-creation.md)
- See: docs/modules/character-creation.md, docs/core/loop.md
- Done when: two characters, each with a level-by-level record; each level's derived stats match a vector; prose in `character-creation.md` is generated from or checked against the record.

#### M2-05 Equipment to the content floor
- Status: review
- Priority: medium
- Needs: M1-02
- Was: RW-15
- Do: 20 weapons, 24 armor, 3 accessories, 7 consumables. Accessories and consumables are far below playable.
- Files: ~6 (ruleset/data/equipment/, ruleset/schemas/equipment.schema.json)
- See: ruleset/data/equipment/
- Done when: each weapon skill has at least 3 weapons across tiers; at least 12 accessories and 20 consumables; adventuring gear and tools exist as rows.

#### M2-06 Gated block and dodge perks
- Status: review
- Priority: medium
- Needs: M1-02, M1-05
- Do: After M1-02 an evade reaches single and area magic and avoids all damage, while a block only subtracts shield DR against physical types. Balance them with gated perk ladders instead of pipeline changes. Block: Shield Training (shield DR +2), Spell Ward (block also works against single and area magic, at half shield DR), Bulwark (a crit block negates the hit). Dodge: a mirror ladder (cheaper dodges in armor; a counter-attack on a crit dodge). Instant spells stay unblockable and undodgeable, even with a perk: their defense is the save (M3-05). Block-side resistance scales with the block (margin or shield DR), not a general `resist`, so it does not stack with armor and the resist step.
- Keys: perk rows in `ruleset/data/features/core.json`; a `grant` modifier that widens the damage types block applies to (proposal)
- Files: ~5 (ruleset/data/features/core.json, docs/core/combat.md, docs/modules/combat.md, ruleset/tests/vectors/combat.json, ruleset/scripts/kernel_reference.py)
- See: docs/core/combat.md (reactions), docs/decisions.md (M1-02 balance)
- Done when: each perk is a row gated by skill thresholds and prerequisite perks; vectors show Spell Ward blocking a fire bolt at half shield DR, Bulwark negating a hit on a crit block, and an instant spell still unblocked with every perk.

#### M2-07 Effects over time and food
- Status: todo
- Priority: medium
- Needs: M1-04, M1-05, M2-05
- Do: An applied effect's `duration` is seconds, and 1 (the default) is instant; its `magnitude` is per second, so a pool effect (restore, damage, drain) totals magnitude times duration and a stat effect holds its magnitude that long (`schema.json`, since M2-05). One effect serves both: no separate instant and over-time effects. Write the tick rule in the kernel, and fix the seconds per turn for turn-based play (`contract.md` and `modifier.md` say turns). `burn` and `poison_health_drain` are over-time effects with rules of their own; decide whether they become conditions that apply a ticking `damage_*`. Food (`type: food`) is always over time; a potion may be either. Food can be eaten at any time except in combat. Add a `food` origin with a new stacking policy `replace`: a new dish removes every modifier and running tick of the previous dish, even effects the new dish lacks (fries, then a hot dog: only the hot dog counts). Each dish is one source, `food:<key>`. A dish's effects are the combined effects of its ingredients, written on the row; cooking a dish from ingredients is crafting (POST-01).
- Keys: `core:origin/food`; origin `stack` value `replace`; the tick rule for `appliedEffect.duration`
- Files: ~10 (docs/core/modifier.md, docs/core/contract.md, ruleset/schemas/schema.json, ruleset/schemas/origin.schema.json, ruleset/data/origins/core.json, ruleset/data/equipment/consumables.json, ruleset/scripts/kernel_reference.py, backend/, ruleset/tests/vectors/)
- See: docs/modules/equipment.md (consumables), docs/decisions/M2-05.md
- Done when: vectors cover an instant restore, a restore over time and its total, a second dish replacing the first (including an effect only the first had), a potion and a dish stacking, and eating refused in combat; the Python reference and Go pass them.
