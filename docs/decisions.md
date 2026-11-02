# Decisions, risks, and open questions

A dated log of process and design decisions that are not obvious from the data. Newest
first. A decision that changes a rule also updates the doc for that rule; this file only
records that it was decided, when, and why. Open questions move to **Decided** once an
issue settles them.

## Decided

### 2026-11-02: The rebase keeps both sides of decision-log additions (M5-10)

- `docs/decisions.md` is in `keep-both-added-lines`: when two open PRs each add a section at
  the top, the rebase keeps both, the branch's first, like the changelog and story files.
  Before this, the rebase workflow failed on every push to `main` since 2026-11-01 and left
  each such branch for a person. Reproduced locally with `rebase-open-branches --no-push`;
  `docs/decisions.md` was the only conflicting file. **[V 2026-11-02]**
- new-world never hit this: it has no shared append-only log; decisions go into the topic
  doc they change. Kept the single log here because it is short and newest-first. **[A]**
- devops-tools, the rebase workflow, and the fragments workflow move to `v0.2.0`
  (`15c1bf1`), the tag new-world uses; the code is the same as `fde087b` apart from the
  version bump.
- Risk: two PRs that edit the same existing lines of the log still need a person.

### 2026-11-02: Changelog fragments (M5-02)

- Each PR adds `changes/<issue-id>.md` instead of a `CHANGELOG.md` line, as in new-world. Every
  PR used to add a line at the same places in the changelog, so parallel PRs conflicted there;
  the rebase workflow only resolves pure additions. Rules: [changes/README.md](../changes/README.md).
- CI runs `devops-tools changelog-fragments check --branch` on every PR through the reusable
  workflow at the devops-tools pin (`fde087b`, unchanged; it already has the command and the
  workflow). A branch whose issue has no fragment fails.
- The owner releases the fragments into `CHANGELOG.md` with `changelog-fragments release` in
  a PR of its own. The release writes `### <issue-id> <title> (#N)` entries under a new
  `## <title>` section above `[Unreleased]`; the categorized `[Unreleased]` list holds the
  entries written before M5-02 and is folded into the first release by hand. **[A]**
- Risk: a release PR deletes every fragment, its own included, so it adds its own fragment
  after `release` or its branch check fails.

### 2026-11-02: Racial traits are features (M1-06)

- `features` is the only table for racial traits. The 18 rows of `data/traits/core.json`
  move to `data/features/core.json` as `type: racial_trait`, keeping their keys;
  `traits/core.json` and `trait.schema.json` are removed. Races keep their `traits` field,
  which now holds `core:feature/<key>` refs.
- Why: `AGENTS.md` names `features` the single perk table, `feature.schema.json` already had
  the `racial_trait` type, and the traits table repeated the row shape without `effects`.
  One table means one ref category and one effect shape for everything a character has.
- Rejected: keeping `traits` as its own table and dropping `racial_trait`. It would make
  the docs match the data, but keep two row shapes and two ref categories for one concept.
- Breaking: `core:trait/<key>` refs no longer resolve; use `core:feature/<key>`. The race
  field name `traits` is unchanged, so no data key is renamed.
- `validate.py` checks that a race's `traits` point at `racial_trait` features and that
  every core `racial_trait` is used by a race.

### 2026-11-02: Every skill has a consumer; checks are rows (M2-02)

- A skill needs a governing attribute and at least one consumer: a weapon, an armor
  category, a spell or effect color, a check, or a perk ([checks.md](core/checks.md)).
  Background bonuses do not count. `ruleset/scripts/skill_audit.py` enforces it.
- Checks get their own collection, `ruleset/data/checks/core.json`. There was no row a
  non-combat skill could be rolled through, so 18 of the 44 skills had no consumer.
- No skill is removed. Every orphan was a skill the design docs already use (TTRPG
  social, knowledge, and survival skills, the crafting trades, the nine color schools), so
  each gets a check row, and `orange_magic` gets its first effect. Removing them would
  have been a breaking key change for no gain.
- The audit is a separate script, not part of `validate.py`, so it can land without
  touching the validator while M1-06 changes it. Folding it into `validate.py` is M5-08.

### 2026-11-01: Story-scoped issue IDs and a computed order (M5-01)

- Open issues are renumbered per story (`M1-01`, `POST-03`), as in new-world. Owner decision;
  it reverses the RW-40 line "Issue ids stay `RW-NN`". Done `RW-` and `SET-` ids stay in
  `done.md`, and each renumbered issue keeps its old id in `Was:`.
- Order is computed by `devops-tools roadmap-order` from `Status`, `Needs`, and the `Focus:`
  line, so parallel waves and the critical path are never maintained by hand. CI fails when
  the story table in `ROADMAP.md` is stale.
- The rebase workflow keeps both sides of a story-file conflict where both only added lines.
  Status edits are changes, not additions, and still go to a person.
- Changelog fragments (M5-02) and the container runner move (M5-03) are their own issues.

### 2026-11-01: Shared devops-tools (RW-41)

- Generic devops code (run logs, log pruning, open-PR rebasing, changelog fragments) lives in
  [SETA1609/devops_tools](https://github.com/SETA1609/devops_tools) and is shared by several
  repos. This repo depends on it through `tools/devops/pyproject.toml`, pinned to a full commit
  SHA, and keeps only Riftweave-specific code (the containers, cache cleanup).
- The pin moves only in its own PR, after its tests pass here. The rebase workflow calls the
  shared reusable workflow at the same SHA as `tools/devops/pyproject.toml`; the two move together.
- Not adopted yet: changelog fragments (`changes/<id>.md`) and `keep-both-added-lines` for
  `docs/stories/*.md`. Either changes the RW-39 and RW-40 process, so each needs its own issue
  (adopted in M5-01 and M5-02).

### 2026-10-31: Story files (RW-40)

- Issues live in one story file per milestone in `docs/stories/`, with a `done.md`
  archive, as in project-new-world. It replaces one file per issue (RW-2). Owner decision:
  the story's goal and the issue order belong next to the issues. Accepted cost: two open
  PRs in the same milestone can conflict on the story file; the rebase workflow leaves those
  for a person.
- Issue ids stay `RW-NN` and `SET-NN`; they are not renumbered per story.

### 2026-10-31: Process (RW-39)

- One issue, one PR, at most 20 counted changed files, five steps (spec, schema, vector,
  data, review), one commit per step ([CONTRIBUTING.md](../CONTRIBUTING.md#workflow)).
- Design docs tag claims **[V]**, **[A]**, or **[U]**
  ([CONTRIBUTING.md](../CONTRIBUTING.md#verification-tags)).
- Open pull requests are rebased onto `main` after every push to `main`, keeping both
  sides of an add-only `CHANGELOG.md` conflict.
- GitHub Actions are pinned to a commit SHA with the tag in a comment.

### 2026-10-29: Tracking (RW-2)

- The repo owns issues, the roadmap, and the changelog. Linear mirrors them.
- The changelog has one entry per closed issue, not per commit.

## Open questions

- **Opposed checks.** A check row may name an `opposed_by` skill, but how two results
  compare (margin against margin, ties) is not specified
  ([checks.md](core/checks.md)). Decide with the casting and combat contest rules in M3.

- **Traits versus features.** Whether creation traits stay a `features` type or become
  their own table ([M1-06](stories/01-kernel.md#m1-06-resolve-the-traits-vs-features-fork)).
- **Casting scope.** What the core casting procedure includes and what is a module
  ([M3-01](stories/03-combat-magic.md#m3-01-casting-procedure-decide-scope-then-specify)).
- **Content floors.** The targets in [ROADMAP.md](../ROADMAP.md#definition-of-finished) are
  proposals **[A]**, to be changed only by a decision recorded here.
- **Schema property casing.** Older schemas use `camelCase`, newer ones `snake_case`.
  Unifying them is breaking; decide before 1.0, because after the key-stability policy
  ([M5-07](stories/05-release.md#m5-07-key-stability-and-versioning-policy)) it costs a major
  version.

## Risks

- **Silent cross-reference breaks.** Refs are strings. `validate.py` checks most namespaced
  refs, but not every category yet
  ([M5-08](stories/05-release.md#m5-08-integrity-gates)).
- **Rebased branches are not revalidated by CI.** Pushes made with the workflow's
  `GITHUB_TOKEN` do not trigger other workflows, so a rebased PR's new head has no
  validation run. Run validation (or push a commit) before merging a rebased PR.
- **Docs outrunning data.** `docs/modules/` holds long design prose with few vectors. Until
  it is tagged, treat untagged claims there as **[A]**.
