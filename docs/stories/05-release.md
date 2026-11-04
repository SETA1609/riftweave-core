# M5 Release 1.0

Roadmap: [M5](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M5 Release 1.0

**Goal:** the SRD is documented, gated, legally clean, and tagged.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 3, 5, 6, 7, and 8 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold, and `v1.0.0` is tagged.

Done: RW-1, RW-2, RW-39, RW-40, RW-41, M5-01, M5-02, M5-07, M5-08, M5-10, M5-12, M5-13, M5-15 ([done.md](done.md#m5-release-10))

#### M5-03 Move the container runner, agent image, and fetch-docs to devops-tools
- Status: todo
- Priority: medium
- Needs: M5-01
- Do: devops-tools now ships `devops-tools container` (image, exec, shell, prune, clean, named commands), a shared base image, the agent container with a plugin start script, and `fetch-docs`. Replace the local runner in `tools/devops/`, `container/agent.Dockerfile`, and `tools/fetch-docs/` with it, configured by `[tool.devops-tools.container]`; the project Dockerfile builds `FROM ${BASE_IMAGE}`. Split into two issues if the counted files pass 20.
- Keys: `[tool.devops-tools.container]` (image name, Dockerfile, version file, commands `validate` and `test`)
- Files: ~20 (tools/devops/, tools/fetch-docs/, container/, .devcontainer/devcontainer.json, scripts/, AGENTS.md)
- See: tools/devops/README.md, https://github.com/SETA1609/devops_tools (container, fetch-docs)
- Done when: `./scripts/validate.sh`, `test.sh`, `shell.sh`, and `agent.sh` run through `devops-tools container`; no local container or fetch-docs code remains; the devcontainer builds; tooling tests pass.

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
- Needs: M2-01, M2-02, M2-03, M2-04, M2-05, M3-02, M3-03, M3-04, M4-03, M4-04, M4-05, M5-01, M5-05, M5-06, M5-07, M5-08
- Was: RW-30
- Do: Final pass: licence headers, `source` fields, NOTICE and CONTRIBUTORS, and the changelog. Then tag.
- Files: ~6 (LICENSING.md, NOTICE, CONTRIBUTORS.md, CHANGELOG.md, ROADMAP.md)
- See: LICENSING.md, ROADMAP.md
- Done when: all checklist items in `ROADMAP.md` 'Definition of finished' are ticked; `v1.0.0` tagged and `[Unreleased]` in the changelog becomes `[1.0.0]`.

#### M5-11 Gate post-1.0 modules on the 1.0 release
- Status: review
- Priority: high
- Needs: —
- Do: The roadmap puts every post-1.0 module after `v1.0.0`, but those issues only needed kernel issues, so `roadmap-order show` listed three of them as ready next to core work, and one was merged before 1.0. Add `M5-09` to the `Needs` of every post-1.0 module issue that has no parent in its story, so the computed order matches the roadmap.
- Files: ~6 (docs/stories/06-modules.md, docs/stories/05-release.md, docs/stories/done.md, ROADMAP.md, docs/decisions/M5-11.md)
- See: ROADMAP.md#post-10-modules, docs/stories/README.md#order
- Done when: `roadmap-order show` lists no post-1.0 module issue as ready and blocks each one without a parent on M5-09; `roadmap-order check` passes.

#### M5-16 Move merged issues to done.md
- Status: review
- Priority: low
- Needs: —
- Do: Ten issues whose pull requests are merged (M1-02, M1-03, M1-04, M1-07, M1-08, M3-04, M4-01, M5-07, M5-08, M5-13) and M5-15 still said `review` in their story files. Move them to `done.md` with `Status: done` and `Merged: #<PR>, <date>`, and list them on each story's `Done:` line.
- Files: ~6 (docs/stories/01-kernel.md, 03-combat-magic.md, 04-world.md, 05-release.md, done.md, ROADMAP.md)
- See: docs/stories/README.md#workflow
- Done when: no issue whose PR is merged is still in a story file; `roadmap-order check` passes.

#### M5-14 Work the backend alongside the rules
- Status: review
- Priority: medium
- Needs: M5-12
- Do: The APP story waited on core issues: the Go core on M1-01, M1-04, and M1-05, sessions on M5-04. The owner wants the backend built alongside the rules instead: a rule issue that adds or changes a function also changes the backend. Drop the core issues from APP's `Needs` (APP keeps its own order), and say in CONTRIBUTING that a rule's data step updates the contract and, once it exists, the Go core.
- Files: ~5 (docs/stories/07-backend.md, docs/stories/05-release.md, ROADMAP.md, CONTRIBUTING.md, docs/decisions.md)
- See: docs/stories/07-backend.md, docs/backend/README.md
- Done when: no APP issue needs an issue outside the APP story except APP-01's `M1-02`; `roadmap-order show` lists APP-02 as blocked only by APP-01; `roadmap-order check` passes.
