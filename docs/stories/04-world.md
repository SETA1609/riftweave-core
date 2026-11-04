# M4 World

Roadmap: [M4](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M4 World

**Goal:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** prices, treasure, and monster stats are consistent under checkable balance bands.

Done: M4-01 ([done.md](done.md#m4-world))

#### M4-02 Loot and treasure tables
- Status: todo
- Priority: high
- Needs: M4-01
- Was: RW-21
- Do: No loot data exists. Monsters need drops, and levels need treasure budgets.
- Files: ~4 (ruleset/data/loot/, ruleset/schemas/loot.schema.json, ruleset/tests/vectors/loot.json)
- See: issues.md section 1
- Done when: `ruleset/data/loot/` with a schema; A treasure-by-level budget consistent with the price formula; seeded roll vectors.

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
