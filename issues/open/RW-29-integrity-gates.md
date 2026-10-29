---
id: RW-29
title: Integrity gates
status: open
milestone: M5
priority: high
labels: [validator, schema]
linear:
created: 2026-10-29
---

Close the review's remaining integrity items: effects referenced nowhere (review M6), `oneOf` escape hatches in schemas (M3), and reference checks for every category.

## Done when

- [ ] Every cross-reference category is checked by `validate.py`.
- [ ] Unreferenced effects are used or documented as reserved.
- [ ] No schema accepts a bare string where a key is required.

See: docs/review/findings-data-schemas.md M3, M6
