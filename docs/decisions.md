# Decisions, risks, and open questions

A dated log of process and design decisions that are not obvious from the data. Newest
first. A decision that changes a rule also updates the doc for that rule; this file only
records that it was decided, when, and why. Open questions move to **Decided** once an
issue settles them.

## Decided

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

- **Traits versus features.** Whether creation traits stay a `features` type or become
  their own table ([RW-8](stories/01-kernel.md#rw-8-resolve-the-traits-vs-features-fork)).
- **Casting scope.** What the core casting procedure includes and what is a module
  ([RW-16](stories/03-combat-magic.md#rw-16-casting-procedure-decide-scope-then-specify)).
- **Content floors.** The targets in [ROADMAP.md](../ROADMAP.md#definition-of-finished) are
  proposals **[A]**, to be changed only by a decision recorded here.
- **Schema property casing.** Older schemas use `camelCase`, newer ones `snake_case`.
  Unifying them is breaking; decide before 1.0, because after the key-stability policy
  ([RW-28](stories/05-release.md#rw-28-key-stability-and-versioning-policy)) it costs a major
  version.

## Risks

- **Silent cross-reference breaks.** Refs are strings. `validate.py` checks most namespaced
  refs, but not every category yet
  ([RW-29](stories/05-release.md#rw-29-integrity-gates)).
- **Rebased branches are not revalidated by CI.** Pushes made with the workflow's
  `GITHUB_TOKEN` do not trigger other workflows, so a rebased PR's new head has no
  validation run. Run validation (or push a commit) before merging a rebased PR.
- **Docs outrunning data.** `docs/modules/` holds long design prose with few vectors. Until
  it is tagged, treat untagged claims there as **[A]**.
