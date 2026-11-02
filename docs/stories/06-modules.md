# POST Post-1.0 modules

Roadmap: [POST](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## POST Post-1.0 modules

**Goal:** optional modules, each with a manifest, written against the [kernel contract](../core/contract.md), that do not change core. Order is by demand, not commitment.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** each module ships, validates, and has vectors for its formulas. Parked with no issue yet: psionics, alignment and morality, birthsigns.

Done: —

#### POST-01 Crafting module: recipes, quality, stations
- Status: review
- Priority: medium
- Needs: M1-07
- Was: RW-31
- Do: Design only today. Ship one vertical slice as a module: recipes, check, time, station, quality grade.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/crafting.md, issues.md section 2
- Done when: module manifest and at least 30 recipes; quality formula in the AST with vectors.

#### POST-02 Alchemy module: quality formula and discovery
- Status: todo
- Priority: medium
- Needs: POST-01
- Was: RW-32
- Do: Quality and magnitude formula is TBD. Define discovery, stations, and multi-effect brews.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/alchemy.md
- Done when: formula and vectors; discovery rules as data.

#### POST-03 Magic crafting module: enchanting
- Status: todo
- Priority: low
- Needs: POST-01, M3-01
- Was: RW-33
- Do: Affix vs full item model, charges, attunement, rarity, runes.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/magic-crafting.md
- Done when: decision record, then data and vectors.

#### POST-04 Breakthrough advancement module
- Status: todo
- Priority: medium
- Needs: M1-03
- Was: RW-34
- Do: Core level-up names no alchemy (SET-14). Breakthroughs ship as an optional module that emits modifiers.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/advancement.md
- Done when: manifest, data, and vectors; core unchanged.

#### POST-05 Factions and reputation module
- Status: review
- Priority: medium
- Needs: —
- Was: RW-35
- Do: No faction schema or data yet. Design doc first.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: issues.md section 3
- Done when: `docs/modules/factions.md`, schema, and data.

#### POST-06 Companions and followers module
- Status: todo
- Priority: medium
- Needs: M1-03
- Was: RW-36
- Do: No companion model. Design doc first.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: issues.md section 3
- Done when: `docs/modules/companions.md`, schema, and data.

#### POST-07 Mounts, vehicles, and ship combat module
- Status: todo
- Priority: low
- Needs: M1-02
- Was: RW-37
- Do: Stubs only in combat edge cases.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: issues.md section 8
- Done when: design doc, then data.

#### POST-08 Survival needs module
- Status: todo
- Priority: low
- Needs: M1-04, M1-05
- Was: RW-38
- Do: Hunger, thirst, sleep, disease, radiation, addiction beyond basic effects.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: issues.md section 8
- Done when: design doc, then conditions and recovery rows.
