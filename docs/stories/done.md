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

## M1 Kernel closed

#### M1-01 Reconcile docs/modules/combat.md with the kernel
- Status: done
- Merged: #70, 2026-11-04
- Priority: high
- Needs: M1-02
- Was: RW-3
- Do: `combat.md` still describes Action, Bonus Action, and Reaction (about 30 mentions) and separate TTRPG and video-game modes. SET-8 and SET-9 removed both: one resolution function, one AP budget. Two authorities for combat is a defect in an SRD.
- Files: ~6 (docs/modules/combat.md, docs/combat/, docs/core/)
- See: docs/modules/combat.md, docs/core/actions.md, docs/core/resolution.md
- Done when: no bonus action, reaction slot, or per-mode rule remains in `docs/modules/combat.md`, `docs/combat/*`, or the combat examples; each rule in `combat.md` links to the `docs/core/` file that owns it, or is moved into `docs/core/`; worked examples are recomputed against current data and match a vector (M1-02).

#### M1-02 Combat pipeline as data, formulas, and vectors
- Status: done
- Merged: #60, 2026-11-04
- Priority: high
- Needs: M1-07
- Was: RW-4
- Do: Attack roll, defense, DR, damage type, and HP loss are specified in prose. Express the pipeline with the formula AST and modifier rows, and pin it with golden vectors.
- Files: ~10 (docs/core/, ruleset/data/combat/, ruleset/schemas/combat.schema.json, ruleset/tests/vectors/combat.json, ruleset/scripts/combat_reference.py)
- See: ruleset/data/combat/, docs/core/formula.md, docs/core/vectors.md
- Done when: hit, miss, crit, and fumble each have a vector; DR, block, resist, and damage-type interaction each have a vector; A reference evaluator reproduces every vector; the prose is a reading of the tree.

#### M1-03 Progression as data: XP, skill points, perk cadence
- Status: done
- Merged: #57, 2026-11-02
- Priority: high
- Needs: M1-07
- Was: RW-5
- Do: XP thresholds, skill points per level, tag-skill ratio, perk cadence, and the level 30 skill cap live in `progression.md`; `derived.md` says they stay there. A second implementation cannot read prose. Move them to data and formulas.
- Files: ~8 (ruleset/data/progression/, ruleset/schemas/progression.schema.json, ruleset/tests/vectors/progression.json, docs/modules/progression.md, docs/core/derived.md)
- See: docs/modules/progression.md, docs/core/derived.md, docs/core/loop.md
- Done when: `ruleset/data/progression/` (or equivalent) holds the XP table and per-level gains, validated by a schema; level-up 1 to 30 has vectors, including the tag-skill and cap rules; `progression.md` becomes commentary that cites the data.

#### M1-04 Rest and recovery rules as data
- Status: done
- Merged: #54, 2026-11-04
- Priority: medium
- Needs: M1-07
- Was: RW-6
- Do: The only recorded recovery rule is that a short rest refills AP. Specify short rest, long rest, and out-of-combat recovery for HP, MP, stamina, and conditions.
- Files: ~6 (docs/core/, ruleset/data/rest/, ruleset/schemas/rest.schema.json, ruleset/tests/vectors/rest.json)
- See: docs/core/contract.md, docs/core/derived.md
- Done when: recovery is rows plus formulas, not prose; every resource in the actor snapshot has a recovery rule or an explicit 'none'; vectors cover a short rest and a long rest.

#### M1-06 Resolve the traits vs features fork
- Status: done
- Merged: #52, 2026-11-02
- Priority: medium
- Needs: —
- Was: RW-8
- Do: `traits/core.json` (18) and `features/core.json` (23) both exist, and the review (data-schemas M1) flagged the fork. CLAUDE.md says `features` is the single perk table. Pick one model before more racial content lands.
- Files: ~8 (ruleset/data/traits/, ruleset/data/features/core.json, ruleset/data/races/core.json, ruleset/schemas/trait.schema.json, ruleset/schemas/feature.schema.json, ruleset/schemas/race.schema.json, ruleset/scripts/validate.py, docs/modules/race.md)
- See: docs/review/findings-data-schemas.md M1, docs/review/findings-docs-hygiene.md M3
- Done when: one table owns racial traits; the other is removed or documented as a view; all references migrated; `validate.py` is green; `docs/modules/race.md` and the schemas describe the decision.

#### M1-07 One test runs every vector file
- Status: done
- Merged: #48, 2026-11-02
- Priority: high
- Needs: —
- Was: RW-9
- Do: `vectors.md` says Python runs the vectors. Confirm a single test discovers and runs every file in `ruleset/tests/vectors/`, and that every kernel function has at least one vector.
- Files: ~3 (ruleset/scripts/, ruleset/tests/vectors/, docs/core/vectors.md)
- See: ruleset/tests/vectors/, ruleset/scripts/test_validate.py
- Done when: adding a vector file needs no test edit; A checklist of kernel functions (resolution, derived, formula, modifiers, actions, progression) shows no function without a vector.

#### M1-08 CI runs unit tests and vectors
- Status: done
- Merged: #61, 2026-11-03
- Priority: high
- Needs: M1-07
- Was: RW-10
- Do: The workflow runs only `validate.py`, and its path filter (`ruleset/**`) skips the workflow, Dockerfile, and docs. Review tooling-H1 is still open.
- Files: ~2 (.github/workflows/validate-ruleset.yml)
- See: docs/review/findings-tooling-ci.md H1, .github/workflows/validate-ruleset.yml
- Done when: workflow runs validate, unit tests, and vectors; path filter covers `ruleset/**`, `.github/**`, `Dockerfile`, `docs/core/**`, and `docs/stories/**`; A failing vector fails the PR.

## M2 Characters

#### M2-02 Skill audit: no orphan skills
- Status: done
- Merged: #49, 2026-11-02
- Priority: medium
- Needs: —
- Was: RW-12
- Do: 44 skills. Each needs a governing attribute and at least one consumer: a weapon, spell color, check, or perk.
- Files: ~4 (ruleset/data/skills/core.json, docs/modules/attributes.md, ruleset/scripts/)
- See: ruleset/data/skills/core.json, docs/modules/attributes.md
- Done when: script lists skills with zero references; each is wired up or removed; `attributes.md` seeds match the data.

## M3 Combat and magic playable

#### M3-04 Wuxing interaction vectors
- Status: done
- Merged: #59, 2026-11-04
- Priority: medium
- Needs: M1-07
- Was: RW-19
- Do: Wuxing is a tag reader (SET-18). Pin the four cycles and their multipliers with vectors, including the reverse-cycle math the review called fragile.
- Files: ~3 (ruleset/data/wuxing/core.json, ruleset/tests/vectors/wuxing.json, docs/modules/wuxing.md)
- See: docs/modules/wuxing.md, ruleset/data/wuxing/core.json
- Done when: vectors for generating, overcoming, weakening, and insulting; reverse-cycle logic exists once.

## M4 World

#### M4-01 Economy: currency, price formula, merchants
- Status: done
- Merged: #58, 2026-11-04
- Priority: high
- Needs: M1-07
- Was: RW-20
- Do: Equipment has `cost` and materials have `valueFactor`, but there are no price rules, denominations, or merchant model. Core play needs buy, sell, and barter.
- Files: ~6 (docs/core/economy.md, ruleset/data/economy/, ruleset/schemas/economy.schema.json, ruleset/data/formulas/, ruleset/tests/vectors/economy.json)
- See: issues.md section 1
- Done when: `docs/core/economy.md` and a price formula in the AST; merchant stock and buy/sell spread as data; vectors for a purchase and a sale with a barter check.

## M5 Release 1.0

#### RW-1 Dev toolchain container and host devops CLI
- Status: done
- Merged: #43, 2026-10-29
- Needs: —
- Do: Containerised Python toolchain and a host-side CLI that logs every run (`scripts/*.sh`, `tools/devops/`, `container/`, `.devcontainer/`).
- See: tools/devops/README.md

#### RW-2 In-repo changelog, issue tracker, archive, and roadmap
- Status: done
- Merged: #45, 2026-11-01
- Needs: —
- Do: `CHANGELOG.md` (one entry per issue), `issues/` with a `done/` archive, and `ROADMAP.md` defining when the core SRD is finished. Linear points at these files.
- See: issues/README.md, ROADMAP.md

#### RW-39 Adopt the workflow, PR rules, verification tags, and branch rebasing
- Status: done
- Merged: #45, 2026-11-01
- Needs: —
- Do: Every PR now edits `CHANGELOG.md` and `ROADMAP.md`, so parallel PRs conflict, and nothing in `CONTRIBUTING.md` says how big a PR may be or how a rule moves from prose to a vector. Adopt the process that already works in a sibling project, adapted to a data repo.
- See: CONTRIBUTING.md#workflow, docs/decisions.md, scripts/rebase_open_branches.py, tools/fetch-docs/
- Done when: `CONTRIBUTING.md` has the five-step workflow (spec, schema, vector, data, review), PR rules (one issue, at most 20 counted files), verification tags, and naming rules; `.github/workflows/rebase-open-branches.yml` rebases open PRs onto `main`; add-only `CHANGELOG.md` conflicts keep both sides. Tests in `tools/devops/tests/test_rebase_open_branches.py`; every GitHub Action is pinned to a commit SHA; `tools/fetch-docs/` fetches cited sources pinned to a commit, with offline tests; `docs/decisions.md` logs decisions, open questions, and risks; `python ruleset/scripts/validate.py` is green; `CHANGELOG.md` entry added.

#### RW-40 Track issues in story files like project-new-world
- Status: done
- Merged: #45, 2026-11-01
- Priority: medium
- Needs: RW-39
- Do: Replace one file per issue (`issues/open/`, `issues/done/`) with one story file per milestone in `docs/stories/` and a `done.md` archive, so the story's goal and order sit next to its issues. Issues keep their `RW-` and `SET-` ids. ROADMAP.md links to story files instead of repeating issue tables.
- Files: ~70 (most are the 57 deleted `issues/` files, a mechanical move; over the 20-file rule by owner request)
- See: docs/stories/README.md
- Done when: every issue from `issues/` is in exactly one story file or `done.md`; no link points into `issues/`; `validate.py` is green; `CHANGELOG.md` entry added.

#### RW-41 Use the shared devops-tools instead of local copies
- Status: done
- Merged: #46, 2026-11-02
- Priority: medium
- Needs: RW-1, RW-39
- Do: The run log (`tools/devops/src/devops/run_log.py`) and the open-PR rebase (`scripts/rebase_open_branches.py`) are copies of what [SETA1609/devops_tools](https://github.com/SETA1609/devops_tools) now ships and tests for several repos. Depend on it, pinned to a commit, and delete the copies. The container runner stays local until devops-tools has one.
- Files: ~17 (tools/devops/, scripts/, .github/workflows/rebase-open-branches.yml, AGENTS.md)
- See: tools/devops/README.md#shared-devops-tools, https://github.com/SETA1609/devops_tools
- Done when: `devops-tools` is a dependency of `tools/devops` pinned to a full SHA in `pyproject.toml` and `uv.lock`; no local run log or rebase script remains; the rebase workflow calls the reusable `rebase-open-branches.yml` from devops-tools at the same SHA; `cd tools/devops && uv run pytest`, `ruff check`, and `ruff format --check` pass; `CHANGELOG.md` entry added.

#### M5-01 Story-scoped issue IDs and a computed order
- Status: done
- Merged: #46, 2026-11-02
- Priority: high
- Needs: RW-41
- Do: Every open issue said `Needs: —`, so `devops-tools roadmap-order` showed all 36 as ready, and nothing told two agents which issues collide. Use story-scoped IDs as in new-world (`M1-01`; the old id stays as `Was:`), write real `Needs` edges and `Files:` on every open issue, add the owner's `Focus:` line, generate the story table in `ROADMAP.md`, and fail CI when it is stale. Story files merge like the changelog in the rebase workflow when both sides only add lines.
- Keys: `[tool.devops-tools]` in `tools/devops/pyproject.toml` (`roadmap-file`, `keep-both-added-lines`)
- Files: ~14 (docs/stories/, ROADMAP.md, docs/decisions.md, CONTRIBUTING.md, AGENTS.md, tools/devops/pyproject.toml, .github/workflows/)
- See: docs/stories/README.md#order, https://github.com/SETA1609/devops_tools
- Done when: every open issue has a story-scoped id, a `Was:` line, `Needs`, and `Files`; `devops-tools roadmap-order show` lists a critical path and more than one wave; `roadmap-order check` passes and runs in CI on story changes; no link points at an old `rw-NN` anchor; `CHANGELOG.md` entry added.

#### M5-02 Changelog fragments instead of CHANGELOG edits
- Status: done
- Merged: #50, 2026-11-02
- Priority: medium
- Needs: M5-01
- Do: Every PR edits `CHANGELOG.md`, so parallel PRs conflict on it (the rebase workflow only merges pure additions). Each PR adds `changes/<issue-id>.md` instead; the owner releases the fragments into `CHANGELOG.md` with `devops-tools changelog-fragments release` in a PR of its own, as in new-world.
- Keys: `changes/<issue-id>.md`, `changes-directory` setting
- Files: ~8 (changes/README.md, CHANGELOG.md, CONTRIBUTING.md, docs/stories/README.md, AGENTS.md, tools/devops/pyproject.toml, .github/workflows/changelog-fragments.yml)
- See: https://github.com/SETA1609/devops_tools (changelog-fragments, reusable workflow)
- Done when: `changes/README.md` states the rules; CI runs `changelog-fragments check --branch` on every PR through the reusable workflow at the pinned SHA; CONTRIBUTING and the stories README say to add a fragment, not a CHANGELOG line; this PR's own entry is a fragment.

#### M5-07 Key stability and versioning policy
- Status: done
- Merged: #56, 2026-11-04
- Priority: high
- Needs: —
- Was: RW-28
- Do: Cross-references are by key, so renaming a key silently breaks consumers. Define what is breaking before tagging a version.
- Files: ~4 (docs/core/versioning.md, ruleset/scripts/, CHANGELOG.md)
- See: docs/core/keys.md
- Done when: `docs/core/versioning.md`: semver, key rename and removal are breaking, deprecation window; A tool diffs keys between tags and lists removals; CHANGELOG marks breaking entries.

#### M5-08 Integrity gates
- Status: done
- Merged: #62, 2026-11-03
- Priority: high
- Needs: M1-06
- Was: RW-29
- Do: Close the review's remaining integrity items: effects referenced nowhere (review M6), `oneOf` escape hatches in schemas (M3), and reference checks for every category.
- Files: ~6 (ruleset/scripts/validate.py, ruleset/scripts/test_validate.py, ruleset/schemas/, ruleset/data/effects/core.json)
- See: docs/review/findings-data-schemas.md M3, M6
- Done when: every cross-reference category is checked by `validate.py`; unreferenced effects are used or documented as reserved; no schema accepts a bare string where a key is required.

#### M5-10 Keep both sides of decision-log additions in the rebase
- Status: done
- Merged: #53, 2026-11-02
- Priority: high
- Needs: M5-02
- Do: The rebase workflow has failed on every push to `main` since 2026-11-01: two open PRs that each add a dated section at the top of `docs/decisions.md` conflict, and the tool leaves the branch for a person. new-world has no shared append-only log, so it never hit this. Add `docs/decisions.md` to `keep-both-added-lines`, so the rebase keeps both sections like the changelog and story files. Pin devops-tools to v0.2.0, as new-world does.
- Keys: `keep-both-added-lines` in `[tool.devops-tools]`
- Files: ~8 (tools/devops/pyproject.toml, tools/devops/uv.lock, .github/workflows/, docs/decisions.md, AGENTS.md, CONTRIBUTING.md)
- See: https://github.com/SETA1609/devops_tools (rebase-open-branches, `keep-both-added-lines`)
- Done when: `keep-both-added-lines` lists `docs/decisions.md`; a local `rebase-open-branches --no-push` of two branches that both add a decision section succeeds; devops-tools, the rebase workflow, and the fragments workflow are pinned to the same v0.2.0 SHA; tooling tests pass.

#### M5-12 Put the reference backend in scope
- Status: done
- Merged: #65, 2026-11-04
- Priority: medium
- Needs: —
- Do: The roadmap put every UI out of scope, so there was no way to call the ruleset except through the vector test. The owner wants one fast backend, in Go, that an agent CLI, a human CLI, an API server, and an Angular frontend all use. Add the APP story with its issues, and change the scope lines in ROADMAP, AGENTS, and CLAUDE: a reference backend is in, not a 1.0 gate; game engines and VTT code stay out.
- Files: ~8 (docs/stories/07-backend.md, docs/stories/README.md, docs/stories/05-release.md, ROADMAP.md, AGENTS.md, CLAUDE.md, docs/decisions.md)
- See: ROADMAP.md#app-reference-backend
- Done when: `roadmap-order show` lists the APP issues with their edges; ROADMAP, AGENTS, and CLAUDE say the same about scope; `roadmap-order check` passes.

#### M5-13 Use devops-tools v0.3: regenerate the roadmap table on rebase
- Status: done
- Merged: #68, 2026-11-04
- Priority: high
- Needs: M5-10
- Do: Every pull request regenerates the story table in `ROADMAP.md`, so two open pull requests that change the same story's row conflict, and the rebase workflow leaves the second for a person; `docs/core/vectors.md` conflicts the same way when two pull requests add checklist rows. devops-tools v0.3.0 adds `regenerate-roadmap-on-rebase`: the rebase takes the main branch's table and regenerates it after each branch. Pin devops-tools and both reusable workflows to v0.3.0, turn the setting on, and add `docs/core/vectors.md` to `keep-both-added-lines`.
- Keys: `regenerate-roadmap-on-rebase` in `[tool.devops-tools]`
- Files: ~8 (tools/devops/pyproject.toml, tools/devops/uv.lock, .github/workflows/, AGENTS.md, CONTRIBUTING.md, docs/decisions.md)
- See: https://github.com/SETA1609/devops_tools (v0.3.0, issue 10)
- Done when: devops-tools, the rebase workflow, and the fragments workflow pin the v0.3.0 commit; a local `rebase-open-branches --no-push` of two branches that change the same story's table row succeeds where v0.2.0 left it for a person; tooling tests pass.

#### M5-15 One file per decision
- Status: done
- Merged: #75, 2026-11-04
- Priority: high
- Needs: M5-10
- Do: Every pull request inserts its section at the top of `docs/decisions.md`, so any two open pull requests conflict there. `keep-both-added-lines` only helps inside the rebase workflow; GitHub's merge button ignores it. Record each issue's decisions in its own file, `docs/decisions/<issue-id>.md`, freeze the old log, and point CONTRIBUTING, AGENTS, and the stories README at the new place.
- Keys: `docs/decisions/<issue-id>.md`
- Files: ~7 (docs/decisions/, docs/decisions.md, CONTRIBUTING.md, AGENTS.md, docs/stories/README.md, docs/stories/05-release.md)
- See: docs/decisions/README.md, changes/README.md
- Done when: `docs/decisions/README.md` states the format; no rule tells a pull request to edit `docs/decisions.md`; two branches that each add a decision file merge into `main` with plain git in either order.

## POST Post-1.0 modules

#### POST-05 Factions and reputation module
- Status: done
- Merged: #51, 2026-11-02
- Priority: medium
- Needs: —
- Was: RW-35
- Do: No faction schema or data yet. Design doc first.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: issues.md section 3
- Done when: `docs/modules/factions.md`, schema, and data.

## APP Reference backend

#### APP-01 Command contract: commands and their schemas as data
- Status: done
- Merged: #74, 2026-11-04
- Priority: medium
- Needs: M1-02
- Do: Define the backend's interface without choosing a language, so Go, both CLIs, the OpenAPI document, and the frontend are all built from one source. One file per command in `backend/contract/` (name, summary, input schema, result schema, error shape), one per reference runner (`resolution`, `combat`, `derived`, `progression`, `actions`, `formula`, `modifiers`, and each runner that lands later). Rolls are inputs, or are drawn from a seed and echoed in the result as `rolls`, so any call can be saved as a vector. A Python test checks every vector's input and the reference's result against the command's schemas. Write `docs/backend/README.md`: the layers (contract, Go core, adapters, frontend), the error shape, and that no adapter holds a rule.
- Keys: `backend/contract/<command>.json`; command `name`, `input_schema`, `result_schema`; result field `rolls`
- Files: ~12 (backend/contract/, docs/backend/README.md, ruleset/scripts/test_contract.py)
- See: ruleset/scripts/kernel_reference.py, docs/core/vectors.md
- Done when: every runner has a contract file; every vector input and every reference result validate against its command's schemas; a vector function without a contract file fails the test.
