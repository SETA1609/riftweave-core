---
id: RW-3
title: Reconcile docs/modules/combat.md with the kernel
status: open
milestone: M1
priority: high
labels: [docs, combat]
linear:
created: 2026-10-29
---

`combat.md` still describes Action, Bonus Action, and Reaction (about 30 mentions) and separate TTRPG and video-game modes. SET-8 and SET-9 removed both: one resolution function, one AP budget. Two authorities for combat is a defect in an SRD.

## Done when

- [ ] No bonus action, reaction slot, or per-mode rule remains in `docs/modules/combat.md`, `docs/combat/*`, or the combat examples.
- [ ] Each rule in `combat.md` links to the `docs/core/` file that owns it, or is moved into `docs/core/`.
- [ ] Worked examples are recomputed against current data and match a vector (RW-4).

See: docs/modules/combat.md; docs/core/actions.md; docs/core/resolution.md
