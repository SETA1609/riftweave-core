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
