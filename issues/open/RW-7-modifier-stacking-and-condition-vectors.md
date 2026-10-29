---
id: RW-7
title: Modifier stacking and condition vectors
status: open
milestone: M1
priority: medium
labels: [vectors, modifiers]
linear:
created: 2026-10-29
---

The contract states one stacking rule. Prove it: ordering of `add`, `mul`, `set`, `grant`, and `resist`, duplicate sources, expiry, and condition application and removal.

## Done when

- [ ] Vectors for each operator, mixed operators, and same-source duplicates.
- [ ] Vectors for each of the 24 conditions' apply and remove paths, or a documented subset with a reason.
- [ ] A contradiction between vector and docs fails CI.

See: docs/core/modifier.md; docs/core/contract.md; ruleset/data/conditions/core.json
