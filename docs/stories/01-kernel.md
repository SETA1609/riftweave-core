# M1 Kernel closed

Roadmap: [M1](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M1 Kernel closed

**Goal:** every kernel function has a vector, combat and progression are data, and CI enforces it. Content written before this milestone must be rechecked after it.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 1 and 2 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold.

Done: —

#### RW-3 Reconcile docs/modules/combat.md with the kernel
- Status: todo
- Priority: high
- Needs: —
- Do: `combat.md` still describes Action, Bonus Action, and Reaction (about 30 mentions) and separate TTRPG and video-game modes. SET-8 and SET-9 removed both: one resolution function, one AP budget. Two authorities for combat is a defect in an SRD.
- See: docs/modules/combat.md, docs/core/actions.md, docs/core/resolution.md
- Done when: no bonus action, reaction slot, or per-mode rule remains in `docs/modules/combat.md`, `docs/combat/*`, or the combat examples; each rule in `combat.md` links to the `docs/core/` file that owns it, or is moved into `docs/core/`; worked examples are recomputed against current data and match a vector (RW-4).

#### RW-4 Combat pipeline as data, formulas, and vectors
- Status: todo
- Priority: high
- Needs: —
- Do: Attack roll, defense, DR, damage type, and HP loss are specified in prose. Express the pipeline with the formula AST and modifier rows, and pin it with golden vectors.
- See: ruleset/data/combat/, docs/core/formula.md, docs/core/vectors.md
- Done when: hit, miss, crit, and fumble each have a vector; DR, block, resist, and damage-type interaction each have a vector; A reference evaluator reproduces every vector; the prose is a reading of the tree.

#### RW-5 Progression as data: XP, skill points, perk cadence
- Status: todo
- Priority: high
- Needs: —
- Do: XP thresholds, skill points per level, tag-skill ratio, perk cadence, and the level 30 skill cap live in `progression.md`; `derived.md` says they stay there. A second implementation cannot read prose. Move them to data and formulas.
- See: docs/modules/progression.md, docs/core/derived.md, docs/core/loop.md
- Done when: `ruleset/data/progression/` (or equivalent) holds the XP table and per-level gains, validated by a schema; level-up 1 to 30 has vectors, including the tag-skill and cap rules; `progression.md` becomes commentary that cites the data.

#### RW-6 Rest and recovery rules as data
- Status: todo
- Priority: medium
- Needs: —
- Do: The only recorded recovery rule is that a short rest refills AP. Specify short rest, long rest, and out-of-combat recovery for HP, MP, stamina, and conditions.
- See: docs/core/contract.md, docs/core/derived.md
- Done when: recovery is rows plus formulas, not prose; every resource in the actor snapshot has a recovery rule or an explicit 'none'; vectors cover a short rest and a long rest.

#### RW-7 Modifier stacking and condition vectors
- Status: todo
- Priority: medium
- Needs: —
- Do: The contract states one stacking rule. Prove it: ordering of `add`, `mul`, `set`, `grant`, and `resist`, duplicate sources, expiry, and condition application and removal.
- See: docs/core/modifier.md, docs/core/contract.md, ruleset/data/conditions/core.json
- Done when: vectors for each operator, mixed operators, and same-source duplicates; vectors for each of the 24 conditions' apply and remove paths, or a documented subset with a reason; A contradiction between vector and docs fails CI.

#### RW-8 Resolve the traits vs features fork
- Status: todo
- Priority: medium
- Needs: —
- Do: `traits/core.json` (18) and `features/core.json` (23) both exist, and the review (data-schemas M1) flagged the fork. CLAUDE.md says `features` is the single perk table. Pick one model before more racial content lands.
- See: docs/review/findings-data-schemas.md M1, docs/review/findings-docs-hygiene.md M3
- Done when: one table owns racial traits; the other is removed or documented as a view; all references migrated; `validate.py` is green; `docs/modules/race.md` and the schemas describe the decision.

#### RW-9 One test runs every vector file
- Status: todo
- Priority: high
- Needs: —
- Do: `vectors.md` says Python runs the vectors. Confirm a single test discovers and runs every file in `ruleset/tests/vectors/`, and that every kernel function has at least one vector.
- See: ruleset/tests/vectors/, ruleset/scripts/test_validate.py
- Done when: adding a vector file needs no test edit; A checklist of kernel functions (resolution, derived, formula, modifiers, actions, progression) shows no function without a vector.

#### RW-10 CI runs unit tests and vectors
- Status: todo
- Priority: high
- Needs: —
- Do: The workflow runs only `validate.py`, and its path filter (`ruleset/**`) skips the workflow, Dockerfile, and docs. Review tooling-H1 is still open.
- See: docs/review/findings-tooling-ci.md H1, .github/workflows/validate-ruleset.yml
- Done when: workflow runs validate, unit tests, and vectors; path filter covers `ruleset/**`, `.github/**`, `Dockerfile`, `docs/core/**`, and `docs/stories/**`; A failing vector fails the PR.
