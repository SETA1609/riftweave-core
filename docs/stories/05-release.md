# M5 Release 1.0

Roadmap: [M5](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M5 Release 1.0

**Goal:** the SRD is documented, gated, legally clean, and tagged.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 3, 5, 6, 7, and 8 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold, and `v1.0.0` is tagged.

Done: RW-1, RW-2, RW-39, RW-40, RW-41, M5-01, M5-02, M5-07, M5-08, M5-10, M5-11, M5-12, M5-13, M5-14, M5-15, M5-16, M5-17 ([done.md](done.md#m5-release-10))

#### M5-03 Move the container runner and agent image to devops-tools
- Status: review
- Priority: medium
- Needs: M5-01
- Do: devops-tools now ships `devops-tools container` (image, exec, shell, prune, clean, named commands), a shared base image, and the agent container with a plugin start script. Replace the local runner in `tools/devops/` and `container/agent.Dockerfile` with it, configured by `[tool.devops-tools.container]`; the project Dockerfile builds `FROM ${BASE_IMAGE}`. Moving `fetch-docs` is M5-22, split off to stay under 20 counted files.
- Keys: `[tool.devops-tools.container]` (image name, Dockerfile, version file, commands `validate`, `test`, and `python`)
- Files: ~19 (tools/devops/, container/, .devcontainer/devcontainer.json, scripts/, AGENTS.md, CLAUDE.md)
- See: tools/devops/README.md, https://github.com/SETA1609/devops_tools (container)
- Done when: `./scripts/validate.sh`, `test.sh`, `shell.sh`, and `agent.sh` run through `devops-tools container`; no local container runner remains; the devcontainer's image builds; tooling tests pass.

#### M5-04 Reference playthrough vector chain
- Status: todo
- Priority: high
- Needs: M1-02, M1-03, M1-04, M1-05, M3-01, M4-01, M4-02
- Was: RW-25
- Do: Extend the vertical loop into a full chain: create, buy, fight, loot, rest, level, cast, take a condition, recover. It is the executable proof that the core is a game.
- Files: ~4 (ruleset/data/examples/, ruleset/tests/vectors/playthrough.json, docs/core/loop.md)
- See: docs/core/loop.md, ruleset/data/examples/loop.json
- Done when: one chain uses only core rows; every step is a vector; the chain passes under the Python reference.

#### M5-05 Player and GM guide
- Status: todo
- Priority: medium
- Needs: M5-04
- Was: RW-26
- Do: There is no 'how to play' or 'how to run' text. An SRD needs a quick start for a table.
- Files: ~3 (docs/guide/)
- See: issues.md section 6
- Done when: quick-start of 10 pages or fewer: make a character, take a turn, run an encounter; GM section: pacing, difficulty, adjudicating unlisted actions with the one resolution function.

#### M5-06 Entry docs and review archive sync
- Status: todo
- Priority: medium
- Needs: M5-02, M5-03
- Was: RW-27
- Do: The July review found README, AGENTS, CLAUDE, and CONTRIBUTING lagging the repo. States which of `docs/core/` and `docs/modules/` wins on conflict, and retires `plan.md`, `issues.md`, and `docs/review/` as history.
- Files: ~10 (README.md, AGENTS.md, CONTRIBUTING.md, docs/archive/, plan.md, issues.md, docs/review/)
- See: docs/review/findings-docs-hygiene.md
- Done when: README tree, AGENTS, and CONTRIBUTING match the repo; authority order is written down once; historical files are moved under `docs/archive/` with a note.

#### M5-09 Licensing and attribution audit, tag v1.0.0
- Status: todo
- Priority: high
- Needs: M2-01, M2-02, M2-03, M2-04, M2-05, M3-02, M3-03, M3-04, M4-03, M4-04, M4-05, M5-01, M5-05, M5-06, M5-07, M5-08, M5-19, M5-20
- Was: RW-30
- Do: Final pass: licence headers, `source` fields, NOTICE and CONTRIBUTORS, and the changelog. Then tag.
- Files: ~6 (LICENSING.md, NOTICE, CONTRIBUTORS.md, CHANGELOG.md, ROADMAP.md)
- See: LICENSING.md, ROADMAP.md
- Done when: all checklist items in `ROADMAP.md` 'Definition of finished' are ticked; `v1.0.0` tagged and `[Unreleased]` in the changelog becomes `[1.0.0]`.

#### M5-18 Alpha playtest through the human CLI
- Status: todo
- Priority: high
- Needs: M2-01, M2-06, M3-02, M3-06, M4-03, M5-04, APP-04, APP-05, APP-08
- Do: The alpha, before the 1.0 tag: the content floors are met and the playthrough chain passes, so play the rules as a person would, with dev mode (APP-08) to reach the states a run is slow to reach. Run at least five characters through the human CLI from creation to level 30, across builds (melee, ranged, each magic color, a hybrid), each run a saved session. Log what a person had to decide, what never came up, and where a rule was missing or read two ways. File each finding as an issue or a decision; turn a run that exposes a wrong result into a vector.
- Files: ~4 (docs/decisions/M5-18.md, docs/playtests/, ruleset/tests/vectors/playthrough.json)
- See: docs/stories/07-backend.md (APP-04, APP-05), docs/core/loop.md
- Done when: five saved sessions reach level 30 and replay to the same result; every finding is an issue, a decision, or a vector; the decision record lists, for every perk, how often the runs took it and what it changed.

#### M5-19 Perk throughput veto and distillation
- Status: todo
- Priority: high
- Needs: M5-18
- Do: Use the alpha's perk tallies to judge every perk. Veto perks that inflate throughput (damage, actions, or recovery a level band cannot answer) or that no run took; distil perks that overlap into one; rename perks whose label or key does not say what they do; remove the vetoed ones. Each perk gets one verdict: keep, merge, rename, or cut. A rename or removal is a breaking key change and is marked so in the changelog; references in examples, vectors, and docs move with it.
- Files: ~8 (ruleset/data/features/core.json, docs/modules/progression.md, docs/decisions/M5-19.md, ruleset/data/examples/, ruleset/tests/vectors/, ruleset/scripts/test_features.py)
- See: docs/decisions/M5-18.md, docs/stories/done.md (M2-01)
- Done when: every perk has a verdict and a reason in the decision record; no cut or renamed key is still referenced; the perk floor (60) still holds or the decision record moves it; vectors pass.

#### M5-20 Wuxing becomes a module, the reference for cross-cutting modules
- Status: todo
- Priority: high
- Needs: M3-02, M4-03
- Do: The contract already names Wuxing a module and says a phase is a tag a module may read, but core still requires `phase` on monsters, monster templates, playable races, gems, crafted items, and engraving jewels, and ships the cycles in `data/wuxing/core.json`. Split it. Core keeps `phase` as an optional tag (the vocabulary in `schema.json`), so rows keep their phases and nothing changes with the module off. `ruleset/modules/wuxing/` takes the cycles and multipliers, the interaction rules (`magic.md` §6, `docs/modules/wuxing.md`), `wuxing_reference.py` and its vectors, and the bestiary's phase-coverage test; its effect on an attack or a cast is a modifier through the contract, never a call into combat or casting. The Go backend learns to load modules (it loads none today), and `internal/wuxing` reads the module's data. Moving `data/wuxing/` out of core is a breaking change, so it lands before the 1.0 tag. Write `docs/modules/module-system.md` up as the guide a module author follows, with Wuxing as the worked example of a module that adds a collection, reads tags on core rows, and emits modifiers.
- Keys: `ruleset/modules/wuxing/` with its manifest; `phase` optional in the six core schemas
- Files: ~16 (ruleset/schemas/, ruleset/data/wuxing/, ruleset/modules/wuxing/, ruleset/scripts/wuxing_reference.py, ruleset/scripts/test_wuxing.py, ruleset/scripts/test_bestiary.py, ruleset/scripts/crafting_reference.py, ruleset/tests/vectors/wuxing.json, backend/internal/data/data.go, backend/internal/wuxing/, docs/modules/wuxing.md, docs/modules/magic.md, docs/modules/module-system.md, docs/core/contract.md)
- See: docs/core/contract.md, docs/core/materials.md, docs/modules/module-system.md, ruleset/modules/factions/
- Done when: validate, the unit tests, the vectors, and the Go tests pass with the module and without it; without it no core rule reads a phase; with it the wuxing vectors pass under Python and Go from the module's data; the change is marked breaking in the changelog; `module-system.md` walks through Wuxing step by step.

#### M5-21 Move merged issues to done.md: M2, M3-01, M4-03, M5-17, APP-02
- Status: review
- Priority: low
- Needs: —
- Do: Eight issues whose pull requests are merged still said `review` in their story files: M2-01 and M2-06 (both in #90, M2-06 was stacked on M2-01), M2-03, M2-05, M3-01, M4-03, M5-17, and APP-02. Move them to `done.md` with `Status: done` and `Merged: #<PR>, <date>`, and list them on each story's `Done:` line.
- Files: ~7 (docs/stories/02-characters.md, 03-combat-magic.md, 04-world.md, 05-release.md, 07-backend.md, done.md, ROADMAP.md)
- See: docs/stories/README.md#order
- Done when: no issue whose PR is merged is still in a story file; `roadmap-order check` passes.

#### M5-22 Move fetch-docs to devops-tools
- Status: todo
- Priority: low
- Needs: M5-03
- Do: devops-tools ships `devops-tools fetch-docs` (url, github, list), the same downloader as `tools/fetch-docs/`, into the gitignored docs cache, pinned to a commit. Set `docs-cache-directory = "cache/docs"` in `[tool.devops-tools]`, remove `tools/fetch-docs/` and its tests, and point AGENTS, CLAUDE, and CONTRIBUTING at the new command. Split from M5-03.
- Keys: `docs-cache-directory` in `[tool.devops-tools]`
- Files: ~6 (tools/fetch-docs/, tools/devops/pyproject.toml, AGENTS.md, CLAUDE.md, CONTRIBUTING.md)
- See: https://github.com/SETA1609/devops_tools (fetch-docs)
- Done when: a cited source downloads into `cache/docs/` through `devops-tools fetch-docs`; no local fetch-docs code remains; the docs name the new command.
