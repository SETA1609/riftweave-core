# M5 Release 1.0

Roadmap: [M5](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M5 Release 1.0

**Goal:** the SRD is documented, gated, legally clean, and tagged.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 3, 5, 6, 7, and 8 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold, and `v1.0.0` is tagged.

Done: RW-1, RW-2, RW-39, RW-40, RW-41, M5-01, M5-02, M5-10 ([done.md](done.md#m5-release-10))

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

#### M5-07 Key stability and versioning policy
- Status: review
- Priority: high
- Needs: —
- Was: RW-28
- Do: Cross-references are by key, so renaming a key silently breaks consumers. Define what is breaking before tagging a version.
- Files: ~4 (docs/core/versioning.md, ruleset/scripts/, CHANGELOG.md)
- See: docs/core/keys.md
- Done when: `docs/core/versioning.md`: semver, key rename and removal are breaking, deprecation window; A tool diffs keys between tags and lists removals; CHANGELOG marks breaking entries.

#### M5-08 Integrity gates
- Status: review
- Priority: high
- Needs: M1-06
- Was: RW-29
- Do: Close the review's remaining integrity items: effects referenced nowhere (review M6), `oneOf` escape hatches in schemas (M3), and reference checks for every category.
- Files: ~6 (ruleset/scripts/validate.py, ruleset/scripts/test_validate.py, ruleset/schemas/, ruleset/data/effects/core.json)
- See: docs/review/findings-data-schemas.md M3, M6
- Done when: every cross-reference category is checked by `validate.py`; unreferenced effects are used or documented as reserved; no schema accepts a bare string where a key is required.

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
- Do: The roadmap puts every POST module after `v1.0.0`, but the root POST issues only needed kernel issues, so `roadmap-order show` listed POST-01, POST-04, and POST-06 as ready next to core work, and POST-05 was merged before 1.0. Add `M5-09` to the `Needs` of every POST issue that has no POST parent, so the computed order matches the roadmap.
- Files: ~6 (docs/stories/06-modules.md, docs/stories/05-release.md, docs/stories/done.md, ROADMAP.md, docs/decisions.md)
- See: ROADMAP.md#post-10-modules, docs/stories/README.md#order
- Done when: `roadmap-order show` lists no POST issue as ready and blocks every root POST issue on M5-09; `roadmap-order check` passes.
#### M5-12 Put the reference backend in scope
- Status: review
- Priority: medium
- Needs: —
- Do: The roadmap put every UI out of scope, so there was no way to call the ruleset except through the vector test. The owner wants one fast backend, in Go, that an agent CLI, a human CLI, an API server, and an Angular frontend all use. Add the APP story with its issues, and change the scope lines in ROADMAP, AGENTS, and CLAUDE: a reference backend is in, not a 1.0 gate; game engines and VTT code stay out.
- Files: ~8 (docs/stories/07-backend.md, docs/stories/README.md, docs/stories/05-release.md, ROADMAP.md, AGENTS.md, CLAUDE.md, docs/decisions.md)
- See: ROADMAP.md#app-reference-backend
- Done when: `roadmap-order show` lists the APP issues with their edges; ROADMAP, AGENTS, and CLAUDE say the same about scope; `roadmap-order check` passes.

#### M5-13 Use devops-tools v0.3: regenerate the roadmap table on rebase
- Status: review
- Priority: high
- Needs: M5-10
- Do: Every pull request regenerates the story table in `ROADMAP.md`, so two open pull requests that change the same story's row conflict, and the rebase workflow leaves the second for a person; `docs/core/vectors.md` conflicts the same way when two pull requests add checklist rows. devops-tools v0.3.0 adds `regenerate-roadmap-on-rebase`: the rebase takes the main branch's table and regenerates it after each branch. Pin devops-tools and both reusable workflows to v0.3.0, turn the setting on, and add `docs/core/vectors.md` to `keep-both-added-lines`.
- Keys: `regenerate-roadmap-on-rebase` in `[tool.devops-tools]`
- Files: ~8 (tools/devops/pyproject.toml, tools/devops/uv.lock, .github/workflows/, AGENTS.md, CONTRIBUTING.md, docs/decisions.md)
- See: https://github.com/SETA1609/devops_tools (v0.3.0, issue 10)
- Done when: devops-tools, the rebase workflow, and the fragments workflow pin the v0.3.0 commit; a local `rebase-open-branches --no-push` of two branches that change the same story's table row succeeds where v0.2.0 left it for a person; tooling tests pass.
