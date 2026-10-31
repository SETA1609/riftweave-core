# M5 Release 1.0

Roadmap: [M5](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M5 Release 1.0

**Goal:** the SRD is documented, gated, legally clean, and tagged.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 3, 5, 6, 7, and 8 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold, and `v1.0.0` is tagged.

Done: —

#### RW-1 Dev toolchain container and host devops CLI
- Status: review
- Needs: —
- Do: Containerised Python toolchain and a host-side CLI that logs every run (`scripts/*.sh`, `tools/devops/`, `container/`, `.devcontainer/`).
- See: tools/devops/README.md

#### RW-2 In-repo changelog, issue tracker, archive, and roadmap
- Status: review
- Needs: —
- Do: `CHANGELOG.md` (one entry per issue), `issues/` with a `done/` archive, and `ROADMAP.md` defining when the core SRD is finished. Linear points at these files.
- See: issues/README.md, ROADMAP.md

#### RW-39 Adopt the workflow, PR rules, verification tags, and branch rebasing
- Status: review
- Needs: —
- Do: Every PR now edits `CHANGELOG.md` and `ROADMAP.md`, so parallel PRs conflict, and nothing in `CONTRIBUTING.md` says how big a PR may be or how a rule moves from prose to a vector. Adopt the process that already works in a sibling project, adapted to a data repo.
- See: CONTRIBUTING.md#workflow, docs/decisions.md, scripts/rebase_open_branches.py, tools/fetch-docs/
- Done when: `CONTRIBUTING.md` has the five-step workflow (spec, schema, vector, data, review), PR rules (one issue, at most 20 counted files), verification tags, and naming rules; `.github/workflows/rebase-open-branches.yml` rebases open PRs onto `main`; add-only `CHANGELOG.md` conflicts keep both sides. Tests in `tools/devops/tests/test_rebase_open_branches.py`; every GitHub Action is pinned to a commit SHA; `tools/fetch-docs/` fetches cited sources pinned to a commit, with offline tests; `docs/decisions.md` logs decisions, open questions, and risks; `python ruleset/scripts/validate.py` is green; `CHANGELOG.md` entry added.

#### RW-25 Reference playthrough vector chain
- Status: todo
- Priority: high
- Needs: —
- Do: Extend the vertical loop into a full chain: create, buy, fight, loot, rest, level, cast, take a condition, recover. It is the executable proof that the core is a game.
- See: docs/core/loop.md, ruleset/data/examples/loop.json
- Done when: one chain uses only core rows; every step is a vector; the chain passes under the Python reference.

#### RW-26 Player and GM guide
- Status: todo
- Priority: medium
- Needs: —
- Do: There is no 'how to play' or 'how to run' text. An SRD needs a quick start for a table.
- See: issues.md section 6
- Done when: quick-start of 10 pages or fewer: make a character, take a turn, run an encounter; GM section: pacing, difficulty, adjudicating unlisted actions with the one resolution function.

#### RW-27 Entry docs and review archive sync
- Status: todo
- Priority: medium
- Needs: —
- Do: The July review found README, AGENTS, CLAUDE, and CONTRIBUTING lagging the repo. States which of `docs/core/` and `docs/modules/` wins on conflict, and retires `plan.md`, `issues.md`, and `docs/review/` as history.
- See: docs/review/findings-docs-hygiene.md
- Done when: README tree, AGENTS, and CONTRIBUTING match the repo; authority order is written down once; historical files are moved under `docs/archive/` with a note.

#### RW-28 Key stability and versioning policy
- Status: todo
- Priority: high
- Needs: —
- Do: Cross-references are by key, so renaming a key silently breaks consumers. Define what is breaking before tagging a version.
- See: docs/core/keys.md
- Done when: `docs/core/versioning.md`: semver, key rename and removal are breaking, deprecation window; A tool diffs keys between tags and lists removals; CHANGELOG marks breaking entries.

#### RW-29 Integrity gates
- Status: todo
- Priority: high
- Needs: —
- Do: Close the review's remaining integrity items: effects referenced nowhere (review M6), `oneOf` escape hatches in schemas (M3), and reference checks for every category.
- See: docs/review/findings-data-schemas.md M3, M6
- Done when: every cross-reference category is checked by `validate.py`; unreferenced effects are used or documented as reserved; no schema accepts a bare string where a key is required.

#### RW-30 Licensing and attribution audit, tag v1.0.0
- Status: todo
- Priority: high
- Needs: —
- Do: Final pass: licence headers, `source` fields, NOTICE and CONTRIBUTORS, and the changelog. Then tag.
- See: LICENSING.md, ROADMAP.md
- Done when: all checklist items in `ROADMAP.md` 'Definition of finished' are ticked; `v1.0.0` tagged and `[Unreleased]` in the changelog becomes `[1.0.0]`.

#### RW-40 Track issues in story files like project-new-world
- Status: review
- Priority: medium
- Needs: RW-39
- Do: Replace one file per issue (`issues/open/`, `issues/done/`) with one story file per milestone in `docs/stories/` and a `done.md` archive, so the story's goal and order sit next to its issues. Issues keep their `RW-` and `SET-` ids. ROADMAP.md links to story files instead of repeating issue tables.
- Files: ~70 (most are the 57 deleted `issues/` files, a mechanical move; over the 20-file rule by owner request)
- See: docs/stories/README.md
- Done when: every issue from `issues/` is in exactly one story file or `done.md`; no link points into `issues/`; `validate.py` is green; `CHANGELOG.md` entry added.
