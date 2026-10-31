# Decisions, risks, and open questions

A dated log of process and design decisions that are not obvious from the data. Newest
first. A decision that changes a rule also updates the doc for that rule; this file only
records that it was decided, when, and why. Open questions move to **Decided** once an
issue settles them.

## Decided

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
  their own table ([RW-8](../issues/open/RW-8-resolve-the-traits-vs-features-fork.md)).
- **Casting scope.** What the core casting procedure includes and what is a module
  ([RW-16](../issues/open/RW-16-casting-procedure-decide-scope-then-specify.md)).
- **Content floors.** The targets in [ROADMAP.md](../ROADMAP.md#definition-of-finished) are
  proposals **[A]**, to be changed only by a decision recorded here.
- **Schema property casing.** Older schemas use `camelCase`, newer ones `snake_case`.
  Unifying them is breaking; decide before 1.0, because after the key-stability policy
  ([RW-28](../issues/open/RW-28-key-stability-and-versioning-policy.md)) it costs a major
  version.

## Risks

- **Silent cross-reference breaks.** Refs are strings. `validate.py` checks most namespaced
  refs, but not every category yet
  ([RW-29](../issues/open/RW-29-integrity-gates.md)).
- **Rebased branches are not revalidated by CI.** Pushes made with the workflow's
  `GITHUB_TOKEN` do not trigger other workflows, so a rebased PR's new head has no
  validation run. Run validation (or push a commit) before merging a rebased PR.
- **Docs outrunning data.** `docs/modules/` holds long design prose with few vectors. Until
  it is tagged, treat untagged claims there as **[A]**.
