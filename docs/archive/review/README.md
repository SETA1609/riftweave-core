# Project Review — riftweave-core

*Review date: 2026-07-02, at commit `62ec92f` (Base + Override monster system).*

This is a full-project review covering the data & schemas, the Python tooling and CI,
and the documentation & project files. Findings are split into three documents:

- [Data & schemas](findings-data-schemas.md) — referential integrity, schema design, content gaps
- [Tooling & CI](findings-tooling-ci.md) — `validate.py`, module loader, reference scripts, CI/test coverage
- [Docs & project hygiene](findings-docs-hygiene.md) — doc-vs-reality drift, licensing, onboarding

Baseline: `python ruleset/scripts/validate.py` passes cleanly (23/23 data files,
1/1 module file) and `test_module_system.py` passes 37/37. Everything below is what
that net does **not** catch.

## Overall verdict

This is a well-conceived, unusually disciplined data-first ruleset. The core ideas —
shared vocabulary in `schema.json`, a single effect registry consumed through
`appliedEffect`, the orthogonal color/phase magic axes, race lineage with conditional
schema requirements, and monster base+override — are all coherent in design *and*
correctly executed in the data. The validator goes beyond plain JSON Schema with a
real referential-integrity layer, and the licensing story (Apache-2.0 code /
CC BY 4.0 content) is a model of clarity.

The problems cluster in three places:

1. **The integrity net has holes, and real breakage hides in them.** Duplicate
   equipment ids across `weapons.json`/`armor.json` (22–25) with live ambiguous
   references from backgrounds; dangling key-based crafting references; effect
   `channels` never enforced (and violated in three places). The validator checks
   only numeric-id existence — key refs, channel legality, and duplicate ids pass
   silently.
2. **The meta-docs lag the repo by several feature branches.** CLAUDE.md, AGENTS.md,
   README.md, and CONTRIBUTING.md all describe an older repo (string ids, no module
   loader, no traits table, AC-based armor). The module docs under `docs/modules/`
   are accurate; the entry-point docs are not. A new contributor following
   CONTRIBUTING.md rule 3 would write invalid data.
3. **CI runs less than the repo can check.** The 37-test module suite only runs via
   `test.sh` (Docker, local); the workflow runs only `validate.py`, and its paths
   filter skips changes to the workflow itself, the Dockerfile, and `test.sh`.

## Top 10 recommended actions (ordered)

1. Renumber weapons 22–25 (they collide with armor 22–25) and fix the five
   background equipment references; add a duplicate-id check per collection to
   `validate.py`. *(data-schemas H1)*
2. Fix the two dangling crafting refs (`gold_necklace`, `steel_longsword` as a
   recipe input) and extend the integrity check to key-based references.
   *(data-schemas H2, tooling M9)*
3. Implement the effect-`channels` legality check that CLAUDE.md already claims
   exists, and fix the three current violations (effects 2/32/57). *(data-schemas H3,
   tooling H2)*
4. Run `test_module_system.py` in CI and widen the workflow paths filter.
   *(tooling H1, L17)*
5. Do one synchronization pass over CLAUDE.md, AGENTS.md, README.md, and
   CONTRIBUTING.md against the current repo. *(docs H1, H2, M3, M4)*
6. Decide the traits-vs-features question: fold `traits/` into `features` as
   `racial_trait`, or deprecate the `racial_trait` feature type and document the
   separate registry — before more racial content lands. *(data-schemas M1, docs M5)*
7. Wire conflict detection into validation — `merge_data()`'s tested conflict logic
   is never called from `validate.py`, so duplicate ids between core and modules
   pass silently. *(tooling M5)*
8. Unify the damage-type vocabulary (`cold`/`lightning` in the weapon enum vs
   `frost`/`shock` in the damage-type registry) behind a shared `$ref`.
   *(data-schemas M2)*
9. Migrate off the deprecated jsonschema `RefResolver` (its DeprecationWarning
   already prints on every run) and add a single-file validation mode.
   *(tooling M7, M8)*
10. Track the backlog visibly: `issues.md` is a genuinely useful, maintained backlog
    sitting in a gitignored file no contributor can see. Convert to GitHub issues or
    commit it. *(docs M7)*

## Severity summary

| Area | High | Medium | Low |
|---|---|---|---|
| Data & schemas | 3 | 7 | 5 |
| Tooling & CI | 2 | 8 | 10 |
| Docs & hygiene | 2 | 5 | 6 |

None of the high-severity items is an emergency — the repo validates and the scripts
run — but H1–H3 in data-schemas are real content bugs shipping today, and the CI gap
means loader regressions would merge green.
