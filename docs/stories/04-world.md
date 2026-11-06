# M4 World

Roadmap: [M4](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M4 World

**Goal:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** prices, treasure, and monster stats are consistent under checkable balance bands.

Done: M4-01 ([done.md](done.md#m4-world))

#### M4-02 Loot and treasure tables
- Status: review
- Priority: high
- Needs: M4-01
- Was: RW-21
- Do: No loot data exists. Monsters need drops, and levels need treasure budgets. Monsters that share a table must still differ: per-monster rolls, a `chance` gate, a `shift` toward better entries, extra table or item rows, and a `material` for better versions.
- Files: ~4 (ruleset/data/loot/, ruleset/schemas/loot.schema.json, ruleset/tests/vectors/loot.json)
- See: issues.md section 1
- Done when: `ruleset/data/loot/` with a schema; A treasure-by-level budget consistent with the price formula, checked per table and per monster; seeded roll vectors, including chance, shift, material, and several rows per monster.

#### M4-03 Bestiary to the content floor
- Status: todo
- Priority: high
- Needs: M1-02
- Was: RW-22
- Do: 11 monsters. An SRD needs a bestiary that spans the level range and the phases.
- Files: ~4 (ruleset/data/monsters/, ruleset/schemas/monster.schema.json)
- See: ruleset/data/monsters/core.json, docs/modules/monsters.md
- Done when: at least 40 monsters covering levels 1 to 30; every phase and damage type is used at least once as an attack and as a resistance; each monster inherits from a base plus overrides.

#### M4-04 Traps and environmental hazards
- Status: todo
- Priority: medium
- Needs: M1-02, M1-05
- Was: RW-23
- Do: Mentioned as missing in the backlog. Minimal core: a hazard is a trigger plus a modifier or damage packet, with a detect and a disarm check.
- Files: ~5 (docs/core/, ruleset/data/hazards/, ruleset/schemas/hazard.schema.json, ruleset/tests/vectors/hazards.json)
- See: issues.md section 8
- Done when: schema and at least 10 hazards; detect and disarm use the one resolution function.

#### M4-05 Encounter and balance validation
- Status: todo
- Priority: medium
- Needs: M1-02, M1-08
- Was: RW-24
- Do: `combat.md` has balance guidelines (damage, DR, HP by tier, hits-to-kill). Turn them into checks so data cannot drift.
- Files: ~4 (ruleset/data/balance/, ruleset/scripts/, .github/workflows/validate-ruleset.yml)
- See: docs/modules/combat.md
- Done when: A script reports monsters and gear outside the guideline bands; bands are data, not prose; report runs in CI as a warning, then as a gate.

#### M4-06 Looter shift: Luck and perks on loot rolls
- Status: todo
- Priority: medium
- Needs: —
- Do: Loot rolls read only the monster's rows; who loots does not matter. Add a looter shift: Luck emits `add loot_roll ((LCK - 5) * 2)` as a formula, perks and other sources emit `add loot_roll` modifiers that stack by the kernel rules, and the total moves every table roll and every `chance`. Add a Scavenger perk that uses it.
- Keys: modifier target `loot_roll`; formula `loot_luck`; perk `core:feature/scavenger`; loot input `actor`, `modifiers`; result `looter_shift`
- Files: ~10 (docs/core/loot.md, docs/core/vectors.md, ruleset/data/formulas/loot_luck.json, ruleset/data/features/core.json, ruleset/scripts/loot_reference.py, ruleset/tests/vectors/loot.json, backend/contract/loot.json)
- See: docs/core/loot.md, docs/core/resolution.md, docs/core/modifier.md
- Done when: vectors cover Luck high, neutral, and low, a perk, two modifiers of one `highest` origin, the 1-to-100 and 1-to-99 clamps, and no looter; the reference passes them; the budget check still uses a shift of 0.
