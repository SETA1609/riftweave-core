# M1 Kernel closed

Roadmap: [M1](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M1 Kernel closed

**Goal:** every kernel function has a vector, combat and progression are data, and CI enforces it. Content written before this milestone must be rechecked after it.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 1 and 2 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold.

Done: M1-02, M1-03, M1-04, M1-07, M1-08 ([done.md](done.md#m1-kernel-closed))

#### M1-01 Reconcile docs/modules/combat.md with the kernel
- Status: review
- Priority: high
- Needs: M1-02
- Was: RW-3
- Do: `combat.md` still describes Action, Bonus Action, and Reaction (about 30 mentions) and separate TTRPG and video-game modes. SET-8 and SET-9 removed both: one resolution function, one AP budget. Two authorities for combat is a defect in an SRD.
- Files: ~6 (docs/modules/combat.md, docs/combat/, docs/core/)
- See: docs/modules/combat.md, docs/core/actions.md, docs/core/resolution.md
- Done when: no bonus action, reaction slot, or per-mode rule remains in `docs/modules/combat.md`, `docs/combat/*`, or the combat examples; each rule in `combat.md` links to the `docs/core/` file that owns it, or is moved into `docs/core/`; worked examples are recomputed against current data and match a vector (M1-02).

#### M1-05 Modifier stacking and condition vectors
- Status: todo
- Priority: medium
- Needs: M1-07, M1-08
- Was: RW-7
- Do: The contract states one stacking rule. Prove it: ordering of `add`, `mul`, `set`, `grant`, and `resist`, duplicate sources, expiry, and condition application and removal.
- Files: ~8 (docs/core/modifier.md, docs/core/contract.md, ruleset/schemas/schema.json, ruleset/schemas/condition.schema.json, ruleset/data/conditions/core.json, ruleset/tests/vectors/)
- See: docs/core/modifier.md, docs/core/contract.md, ruleset/data/conditions/core.json
- Done when: vectors for each operator, mixed operators, and same-source duplicates; vectors for each of the 24 conditions' apply and remove paths, or a documented subset with a reason; A contradiction between vector and docs fails CI.

#### M1-06 Resolve the traits vs features fork
- Status: review
- Priority: medium
- Needs: —
- Was: RW-8
- Do: `traits/core.json` (18) and `features/core.json` (23) both exist, and the review (data-schemas M1) flagged the fork. CLAUDE.md says `features` is the single perk table. Pick one model before more racial content lands.
- Files: ~8 (ruleset/data/traits/, ruleset/data/features/core.json, ruleset/data/races/core.json, ruleset/schemas/trait.schema.json, ruleset/schemas/feature.schema.json, ruleset/schemas/race.schema.json, ruleset/scripts/validate.py, docs/modules/race.md)
- See: docs/review/findings-data-schemas.md M1, docs/review/findings-docs-hygiene.md M3
- Done when: one table owns racial traits; the other is removed or documented as a view; all references migrated; `validate.py` is green; `docs/modules/race.md` and the schemas describe the decision.
