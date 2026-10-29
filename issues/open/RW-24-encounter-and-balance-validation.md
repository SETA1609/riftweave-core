---
id: RW-24
title: Encounter and balance validation
status: open
milestone: M4
priority: medium
labels: [balance, tooling]
linear:
created: 2026-10-29
---

`combat.md` has balance guidelines (damage, DR, HP by tier, hits-to-kill). Turn them into checks so data cannot drift.

## Done when

- [ ] A script reports monsters and gear outside the guideline bands.
- [ ] Bands are data, not prose.
- [ ] Report runs in CI as a warning, then as a gate.

See: docs/modules/combat.md
