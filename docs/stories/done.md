# Done issues

Issues whose PR is merged, moved out of the story files so those list only open work. Each keeps its full text, `Status: done`, and the PR number and merge date. IDs are never reused ([IDs](README.md#ids)).

## M0 Foundations

Tracked in Linear before this repo had a tracker, recovered from git history. The merge date is the landing commit's. SET-1 to SET-5, SET-23, and SET-24 were not found in the repo and have no entry.

#### SET-6 Kernel contract: actor, modifier, event
- Status: done
- Merged: #18, 2026-10-04
- Needs: —
- Do: One shared sentence for modules. No data or schema changes. Landed in #18. See `docs/core/contract.md`.

#### SET-7 Derived stats are functions, not stored truth
- Status: done
- Merged: #20, 2026-10-04
- Needs: —
- Do: Coefficients live in `derived/core.json`. HP, MP, AP, carry, and initiative recompute from the snapshot. A module emits a modifier; it does not write a second max. Landed in #20. See `ruleset/data/derived/core.json`.

#### SET-8 One d100 resolution function
- Status: done
- Merged: #21, 2026-10-08
- Needs: —
- Do: Crit and fumble are tags on `roll <= target`. Luck widens the window; it does not re-roll. No video-game mode that deletes the fumble tag. A target over 100 absorbs penalties; roll 100 still fails. Landed in #21. See `docs/core/resolution.md`.

#### SET-9 Action budget replaces Action, Bonus, and Reaction
- Status: done
- Merged: #34, 2026-10-08
- Needs: —
- Do: One cost per action. Unspent points carry, capped at 10; a short rest refills. Block is a hold on the same pool. A spell spends 3 AP plus its own mana cost. Landed in #34. See `docs/core/actions.md`.

#### SET-10 Conditions are modifier bundles
- Status: done
- Merged: #23, 2026-10-11
- Needs: —
- Do: Paralyzed sets AP to 0; prose is not the authority. Condition points drop the percent sign and use ASCII minus. Landed in #23. See `ruleset/data/conditions/core.json`.

#### SET-11 Modules may add effect and condition rows
- Status: done
- Merged: #24, 2026-10-11
- Needs: —
- Do: Test effect id 1001. Unknown opcodes are ignored. Landed in #24. See `docs/modules/module-system.md`.

#### SET-12 Armor core is slot and DR; layering is a module
- Status: done
- Merged: #25, 2026-10-12
- Needs: —
- Do: No equipment id renumber. Landed in #25. See `docs/modules/armor-layering.md`.

#### SET-13 Unify damage types and keep equipment ids unique
- Status: done
- Merged: #19, 2026-10-04
- Needs: —
- Do: `cold`/`lightning` are aliases of `frost`/`shock`, not a second vocabulary. Equipment ids 22-25 stay with armor; weapons stay at 45-48. Landed in #19. See `ruleset/schemas/damage_type.schema.json`.

#### SET-14 Level-up does not require alchemy
- Status: done
- Merged: #26, 2026-10-12
- Needs: —
- Do: Level 1 to 2 names no ingredient, core, or phase. Landed in #26. See `docs/modules/progression.md`.

#### SET-15 Vertical loop on existing rows
- Status: done
- Merged: #27, 2026-10-12
- Needs: —
- Do: Create, equip, attack, DR, XP, perk. No phase. Example in `ruleset/data/examples/loop.json`. Landed in #27. See `docs/core/loop.md`.

#### SET-16 Magic emits a roll request and an effect list
- Status: done
- Merged: #30, 2026-10-18
- Needs: —
- Do: A phase is not required to cast. Landed in #30. See `docs/modules/magic.md`.

#### SET-17 Alchemy is a crafting profile
- Status: done
- Merged: #31, 2026-10-18
- Needs: —
- Do: No private effect language. Based on SET-11, SET-14, SET-19. Landed in #31. See `docs/modules/alchemy.md`.

#### SET-18 Wuxing is a tag reader, not a core dependency
- Status: done
- Merged: #32, 2026-10-18
- Needs: —
- Do: Cycle edges unchanged. Based on SET-16 and SET-14. Landed in #32. See `docs/modules/wuxing.md`.

#### SET-19 Material identity is separate from form
- Status: done
- Merged: #28, 2026-10-18
- Needs: —
- Do: Copper is not copied into `copper_fiber`. `appliesTo` stays. A recipe does not refuse a material; `foundAs` is how nature gives it. Landed in #28. See `ruleset/data/materials/forms.json`.

#### SET-20 Formula AST
- Status: done
- Merged: #37, 2026-10-18
- Needs: —
- Do: The AP formula is a JSON tree (`ruleset/data/formulas/ap.json`), not prose. Landed in #37. See `docs/core/formula.md`.

#### SET-21 One modifier schema
- Status: done
- Merged: #35, 2026-10-18
- Needs: —
- Do: `add`, `mul`, `set`, `grant`, `resist`. Five-phase written in ASCII, not a unicode escape. Landed in #35. See `docs/core/modifier.md`.

#### SET-22 Golden vectors
- Status: done
- Merged: #36, 2026-10-18
- Needs: —
- Do: Golden vectors for resolution and derived HP. Landed in #36. See `ruleset/tests/vectors/`.

#### SET-25 Module overrides
- Status: done
- Merged: #39, 2026-10-18
- Needs: —
- Do: A module may override a vanilla row by key. Manifest dash written as ASCII. Landed in #39. See `docs/core/overrides.md`.
