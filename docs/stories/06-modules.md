# POST Post-1.0 modules

Roadmap: [POST](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## POST Post-1.0 modules

**Goal:** optional modules, each with a manifest, written against the [kernel contract](../core/contract.md), that do not change core. Work starts after `v1.0.0` (M5-09): every issue here without a parent in this story needs M5-09. Order is by demand, not commitment.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** each module ships, validates, and has vectors for its formulas. Parked with no issue yet: psionics, alignment and morality, birthsigns.

Done: POST-05 ([done.md](done.md#post-post-10-modules))

#### POST-01 Crafting module: recipes, quality, stations
- Status: todo
- Priority: medium
- Needs: M1-07, M5-09
- Was: RW-31
- Do: Design only today. Ship one vertical slice as a module: recipes, check, time, station, quality grade.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/crafting.md, docs/archive/issues.md section 2
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
- Needs: M1-03, M5-09
- Was: RW-34
- Do: Core level-up names no alchemy (SET-14). Breakthroughs ship as an optional module that emits modifiers.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/advancement.md
- Done when: manifest, data, and vectors; core unchanged.

#### POST-06 Companions and followers module
- Status: todo
- Priority: medium
- Needs: M1-03, M5-09
- Was: RW-36
- Do: No companion model. Design doc first.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/archive/issues.md section 3
- Done when: `docs/modules/companions.md`, schema, and data.

#### POST-07 Mounts, vehicles, and ship combat module
- Status: todo
- Priority: low
- Needs: M1-02, M5-09
- Was: RW-37
- Do: Stubs only in combat edge cases.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/archive/issues.md section 8
- Done when: design doc, then data.

#### POST-08 Survival needs module
- Status: todo
- Priority: low
- Needs: M1-04, M1-05, M2-07, M5-09
- Was: RW-38
- Do: An optional module; with it off, core plays as before. The first slice is hunger and thirst; sleep, disease, radiation, and addiction beyond basic effects come later, each with its own issue. The actor gets two stored meters that rise with game time. Crossing a threshold applies a condition (proposal: hungry and starving, thirsty and parched [A]) whose penalties are ordinary modifiers. Eating lowers hunger and drinking lowers thirst, by amounts in a module table keyed by `core:consumable/<key>`, so the module leaves core's consumable rows as they are. Core has no water row, so the module adds drinks. Meals keep the M2-07 rules: not in combat, and a new dish replaces the last one's effects, while hunger still falls for each dish eaten. Decide the rates per game hour, how rest and travel advance the meters, whether the worst stage costs HP or stops at penalties, what END and `survival` change, and how a tabletop game tracks the meters (per day [A]).
- Keys: `ruleset/modules/survival/` with its manifest, a `nutrition` table, the hunger and thirst conditions, and the water rows
- Files: ~7 (ruleset/modules/survival/, docs/modules/survival.md, ruleset/tests/vectors/)
- See: docs/archive/issues.md section 8, docs/modules/module-system.md, docs/modules/conditions.md, ruleset/data/rest/core.json
- Done when: `docs/modules/survival.md`, then the manifest, conditions, nutrition, and water rows validate; vectors cover each meter rising over time, a threshold applying its condition, a dish lowering hunger, a drink lowering thirst, and rest advancing the meters; with the module off, the food vectors are unchanged.

#### POST-09 Weapon advantage module: the guard cycle
- Status: todo
- Priority: low
- Needs: M3-03, M5-09, M5-20
- Do: A small module, the first example a module author reads: a rock-paper-scissors between the three melee weapon skills, read from the skill every weapon base already names, so the module adds no table and no field on core rows. The cycle: blunt beats blades (weight batters through a parry), blades beat piercing (an edge turns a point aside and cuts a haft), piercing beats blunt (a thrust finds the gap in a slow, heavy swing). One situation row each way: the attacker whose weapon's skill beats the defender's wielded weapon's skill gets a bonus to hit of the order of the size bonus (+10, [A]), and the same penalty the other way. Unarmed, bows, crossbows, guns, throwing weapons, and natural attacks stand outside it. Decide whether swapping weapons costs AP enough that carrying the counter is a choice, not a reflex. It runs alongside Wuxing or alone; a table may load either, both, or neither. The name is our own: no other game's term for it.
- Keys: `ruleset/modules/weapon_advantage/` with its manifest and situation rows keyed by `core:skill/blades`, `core:skill/blunt`, `core:skill/piercing`
- Files: ~5 (ruleset/modules/weapon_advantage/, docs/modules/weapon-advantage.md, ruleset/tests/vectors/)
- See: docs/modules/combat.md (situations), docs/modules/weapons.md, docs/modules/module-system.md
- Done when: the rows name only core skills; vectors cover advantage, disadvantage, the same skill, and an attack outside the cycle (a bow, a fist, a claw); with the module off, combat vectors are unchanged; `module-system.md` links it as the minimal example.
