---
id: RW-39
title: Adopt the workflow, PR rules, verification tags, and branch rebasing
status: done
milestone: M5
labels: [process, tooling, ci]
linear:
created: 2026-10-31
closed: 2026-10-31
---

Every PR now edits `CHANGELOG.md` and `ROADMAP.md`, so parallel PRs conflict, and nothing in `CONTRIBUTING.md` says how big a PR may be or how a rule moves from prose to a vector. Adopt the process that already works in a sibling project, adapted to a data repo.

## Done when

- [x] `CONTRIBUTING.md` has the five-step workflow (spec, schema, vector, data, review), PR rules (one issue, at most 20 counted files), verification tags, and naming rules.
- [x] `.github/workflows/rebase-open-branches.yml` rebases open PRs onto `main`; add-only `CHANGELOG.md` conflicts keep both sides. Tests in `tools/devops/tests/test_rebase_open_branches.py`.
- [x] Every GitHub Action is pinned to a commit SHA.
- [x] `tools/fetch-docs/` fetches cited sources pinned to a commit, with offline tests.
- [x] `docs/decisions.md` logs decisions, open questions, and risks.
- [x] `python ruleset/scripts/validate.py` is green.
- [x] `CHANGELOG.md` entry added.

See: CONTRIBUTING.md#workflow; docs/decisions.md; scripts/rebase_open_branches.py; tools/fetch-docs/
