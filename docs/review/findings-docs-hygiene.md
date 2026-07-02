# Findings — Documentation & Project Hygiene

Scope: README.md, AGENTS.md, CLAUDE.md, docs/ (modules + combat), CONTRIBUTING.md,
LICENSING.md / LICENSE / LICENSE-CONTENT / NOTICE, SECURITY.md, plan.md, issues.md.

Context that reframes several items: `.gitignore` deliberately excludes `CLAUDE.md`,
`plan.md`, and `issues.md`, and AGENTS.md declares itself the canonical shareable
ruleset with CLAUDE.md as a personal local override. So CLAUDE.md drift is
local-file staleness, not repo rot — but it is severe, and AGENTS.md (which *is*
tracked) has drift of its own.

The systemic problem in one sentence: the repo's reality (data, schemas, scripts)
evolved fast through the recent feature branches, and the meta files describing it
(CLAUDE.md, AGENTS.md, README.md, CONTRIBUTING.md) were each updated at different
times, so they now tell different stories. One synchronization pass resolves most of
this document.

## High severity

### H1. CLAUDE.md is stale on almost every structural claim

- "Modules (planned, not implemented) … there is no module loader, dependency model,
  or multi-root validator" — **false**: `module_loader.py`,
  `test_module_system.py`, `module.manifest.schema.json`, and
  `ruleset/modules/example-module/` all exist, and `validate.py` itself discovers
  and validates module data ("Modules: 1 passed, 0 failed").
- "there is no separate 'trait' type" / "`features` is the single perk table" —
  **false**: `traits/core.json` (18 entries) + `trait.schema.json` exist as a
  racial-traits registry referenced by races; `features/core.json` has zero
  `racial_trait` entries.
- "Entry ids follow `^[a-z_][a-z0-9_]*$`" — **false**: every collection now uses
  integer `id` + snake_case `key` + `label`.
- "Weapons/armor are still on the original D&D-derived shape (AC/`acDexBonus`) …
  known follow-up" — **false**: armor is fully on the DR + slot + layer model;
  `progression.md` explicitly says "No legacy AC fields remain."
- "The repo's referential-integrity check (run after `validate.py`)" — it's now
  *inside* `validate.py`.
- Missing from its file-pair list: `backgrounds`, `conditions`, `combat`,
  `crafting`, `gems`, `tiers`, `traits`.

**Recommendation:** regenerate CLAUDE.md from the current repo, or reduce it to
"read AGENTS.md" plus truly personal notes, since AGENTS.md is declared canonical.

### H2. AGENTS.md contradicts itself and lags the repo (and it's the tracked, canonical file)

- Line 64 says integer ids with "(A future referential integrity check will
  validate…)"; line 92 says cross-refs "are **by string ID only**" — a direct
  internal contradiction, and both stale: the integrity check exists today.
- "Modules (Planned / Aspirational) … Today the ruleset is monolithic" — false;
  `docs/modules/README.md` itself says the foundation "is **implemented**".
- "`features` is the **single** perk table. Do not create separate
  trait/class/race tables" — contradicted by the shipped `traits` table.
- The collections list omits `backgrounds`, `combat`, `conditions`, `crafting`,
  `traits`.

**Recommendation:** one consistency pass; drift here propagates to every agent and
contributor who is told this file is canonical.

## Medium severity

### M1. README.md project-structure tree is badly out of date

Lists only 4 of 19 docs in `docs/modules/`, omits `docs/combat/` entirely, omits
eleven data directories and the monster `bases/`, omits `ruleset/modules/` and the
three non-validate scripts. Closing line still calls modules "planned".

**Recommendation:** trim the tree to top-level directories (so it can't rot this
fast) and update the modules sentence to "implemented foundation".

### M2. CONTRIBUTING.md rule 3 teaches the old ID convention

Lines 52–54: "IDs are snake_case matching `^[a-z_][a-z0-9_]*$` … cross-references
are by string id and are **not** schema-validated." A new contributor following this
rule would write invalid data.

**Recommendation:** rewrite rule 3 to the id/key/label triple and mention the
integrity check.

### M3. `racial_trait` feature type vs `traits` registry is an unresolved design fork

`feature.schema.json` still defines and documents `racial_trait` ("pinned to a
race"), but no data uses it; racial traits actually live in `traits/core.json`.
Docs disagree with each other (feature schema description vs `race.md`).
Same issue as data-schemas M1 — the docs side of it.

**Recommendation:** deprecate `racial_trait` in the feature schema or document when
each mechanism applies.

### M4. plan.md is self-declared historical (457 lines, gitignored)

Header: "This plan document is historical — the combat system has been implemented."
Everything in it is done.

**Recommendation:** delete it; anything still open is captured in issues.md.

### M5. issues.md is a genuinely useful, maintained backlog — in a gitignored file

284 lines, well-structured (state / why-it-matters / next-steps per area), recently
maintained. But it's invisible to contributors. Minor staleness: § 7 "No actual
module loading system" contradicts the implemented loader; § 2 "crafting.md is a
design document only" is contradicted by that file's own "Implemented Slice" table
and by `data/crafting/*`.

**Recommendation:** convert to GitHub issues or track it in-repo (e.g.
`docs/BACKLOG.md`); fix the two stale sections either way.

## Low severity

- **L1.** `docs/modules/README.md` line 7 says core focuses on "abilities,
  **classes**, races…" — the system is emphatically classless everywhere else.
  One-word fix.
- **L2.** SECURITY.md scope says "the only executable code … is `validate.py`" —
  now also the loader, two reference scripts, and the test suite. The placeholder
  `<security-contact@example.com>` was never replaced.
- **L3.** LICENSE-CONTENT's suggested attribution still has the placeholder
  `https://github.com/<owner>/riftweave-core`.
- **L4.** `ruleset/docs/` is an empty directory — delete or populate.
- **L5.** README never mentions the module system, `test.sh`, or the two reference
  scripts (`docs/combat/examples.md` advertises them; the front door doesn't).
- **L6.** `ruleset/scripts/__pycache__/` sits on disk (untracked, harmless noise).

## Doc-vs-data spot checks — mostly clean

- **combat:** `docs/modules/combat.md` + `docs/combat/{integration,examples}.md`
  accurately describe the damage-type and base-resolution data, schema names, effect
  ids 65–77, and the reference script. Current.
- **monsters:** `monsters.md` matches the base+override data including `$extend`
  semantics. Current (matches the latest commit).
- **materials/gems/tiers:** `materials.md` matches the 10 materials, phases,
  modifiers, and tier ladder; tiers have no dedicated doc but `materials.md` § 4
  covers them adequately.
- **magic:** `magic.md` matches the color/phase two-axis model, wuxing data, and
  naming rules. Current.
- **crafting:** matches the data, though the doc's "Open Design Questions" sections
  now sit oddly *above* its "Implemented Slice" — minor restructure would help.
- No orphan gaps found: backgrounds, conditions, gems, advancement, armor, weapons,
  race all have dedicated docs matching data.

## Licensing — good

The dual-license split (Apache-2.0 for scripts/schemas/CI, CC BY 4.0 for data/docs/
prose) is sensible and consistently stated across LICENSING.md, README, NOTICE,
CONTRIBUTING, and AGENTS.md, with a fallback rule for ambiguous files, a correct
note that mechanics aren't copyrightable, a trademark carve-out, and
inbound=outbound contribution terms. Only defects: the two placeholders (L2, L3).
One of the best-executed parts of the repo.

## Onboarding — mostly yes

A new contributor gets working setup, the validate loop, schema-first rules, and the
CI story from README + CONTRIBUTING alone. The two trip hazards: CONTRIBUTING's
stale ID rule (M2) and README's stale directory map (M1). Neither doc mentions
modules or the reference scripts.

## Strengths

- **The module docs are unusually thorough and current:** combat.md (1316 lines with
  dual TTRPG/video-game resolution for every mechanic, worked examples, balance
  tables), conditions, armor, race lineage, magic — all verified against actual data
  and found accurate.
- **Cross-linking discipline:** docs link to exact data/schema paths, and per-doc
  status headers make maturity explicit.
- **The validator story is honest and complete:** one command, CI parity, Docker
  parity, module validation integrated.
- **LICENSING.md** is a model of clarity for a code+content split.
- **issues.md shows real backlog hygiene** — items get marked done with pointers to
  the delivering docs.
