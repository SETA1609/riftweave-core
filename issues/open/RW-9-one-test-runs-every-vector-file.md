---
id: RW-9
title: One test runs every vector file
status: open
milestone: M1
priority: high
labels: [vectors, tooling]
linear:
created: 2026-10-29
---

`vectors.md` says Python runs the vectors. Confirm a single test discovers and runs every file in `ruleset/tests/vectors/`, and that every kernel function has at least one vector.

## Done when

- [ ] Adding a vector file needs no test edit.
- [ ] A checklist of kernel functions (resolution, derived, formula, modifiers, actions, progression) shows no function without a vector.

See: ruleset/tests/vectors/; ruleset/scripts/test_validate.py
