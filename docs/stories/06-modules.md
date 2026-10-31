# POST Post-1.0 modules

Roadmap: [POST](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## POST Post-1.0 modules

**Goal:** optional modules, each with a manifest, written against the [kernel contract](../core/contract.md), that do not change core. Order is by demand, not commitment.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** each module ships, validates, and has vectors for its formulas. Parked with no issue yet: psionics, alignment and morality, birthsigns.

Done: —

#### RW-31 Crafting module: recipes, quality, stations
- Status: todo
- Priority: medium
- Needs: —
- Do: Design only today. Ship one vertical slice as a module: recipes, check, time, station, quality grade.
- See: docs/modules/crafting.md, issues.md section 2
- Done when: module manifest and at least 30 recipes; quality formula in the AST with vectors.

#### RW-32 Alchemy module: quality formula and discovery
- Status: todo
- Priority: medium
- Needs: —
- Do: Quality and magnitude formula is TBD. Define discovery, stations, and multi-effect brews.
- See: docs/modules/alchemy.md
- Done when: formula and vectors; discovery rules as data.

#### RW-33 Magic crafting module: enchanting
- Status: todo
- Priority: low
- Needs: —
- Do: Affix vs full item model, charges, attunement, rarity, runes.
- See: docs/modules/magic-crafting.md
- Done when: decision record, then data and vectors.

#### RW-34 Breakthrough advancement module
- Status: todo
- Priority: medium
- Needs: —
- Do: Core level-up names no alchemy (SET-14). Breakthroughs ship as an optional module that emits modifiers.
- See: docs/modules/advancement.md
- Done when: manifest, data, and vectors; core unchanged.

#### RW-35 Factions and reputation module
- Status: todo
- Priority: medium
- Needs: —
- Do: No faction schema or data yet. Design doc first.
- See: issues.md section 3
- Done when: `docs/modules/factions.md`, schema, and data.

#### RW-36 Companions and followers module
- Status: todo
- Priority: medium
- Needs: —
- Do: No companion model. Design doc first.
- See: issues.md section 3
- Done when: `docs/modules/companions.md`, schema, and data.

#### RW-37 Mounts, vehicles, and ship combat module
- Status: todo
- Priority: low
- Needs: —
- Do: Stubs only in combat edge cases.
- See: issues.md section 8
- Done when: design doc, then data.

#### RW-38 Survival needs module
- Status: todo
- Priority: low
- Needs: —
- Do: Hunger, thirst, sleep, disease, radiation, addiction beyond basic effects.
- See: issues.md section 8
- Done when: design doc, then conditions and recovery rows.
