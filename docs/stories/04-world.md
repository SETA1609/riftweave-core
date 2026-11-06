# M4 World

Roadmap: [M4](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M4 World

**Goal:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** prices, treasure, and monster stats are consistent under checkable balance bands.

Done: M4-01, M4-02, M4-04, M4-05, M4-06 ([done.md](done.md#m4-world))

#### M4-03 Bestiary to the content floor
- Status: doing
- Priority: high
- Needs: M1-02
- Was: RW-22
- Do: 11 monsters. An SRD needs a bestiary that spans the level range and the phases.
- Files: ~4 (ruleset/data/monsters/, ruleset/schemas/monster.schema.json)
- See: ruleset/data/monsters/core.json, docs/modules/monsters.md
- Done when: at least 40 monsters covering levels 1 to 30; every phase and damage type is used at least once as an attack and as a resistance; each monster inherits from a base plus overrides.
