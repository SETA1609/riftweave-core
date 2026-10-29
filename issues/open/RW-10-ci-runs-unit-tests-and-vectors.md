---
id: RW-10
title: CI runs unit tests and vectors
status: open
milestone: M1
priority: high
labels: [ci, tooling]
linear:
created: 2026-10-29
---

The workflow runs only `validate.py`, and its path filter (`ruleset/**`) skips the workflow, Dockerfile, and docs. Review tooling-H1 is still open.

## Done when

- [ ] Workflow runs validate, unit tests, and vectors.
- [ ] Path filter covers `ruleset/**`, `.github/**`, `Dockerfile`, `docs/core/**`, and `issues/**`.
- [ ] A failing vector fails the PR.

See: docs/review/findings-tooling-ci.md H1; .github/workflows/validate-ruleset.yml
