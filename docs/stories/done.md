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
