---
id: RW-4
title: Combat pipeline as data, formulas, and vectors
status: open
milestone: M1
priority: high
labels: [combat, vectors]
linear:
created: 2026-10-29
---

Attack roll, defense, DR, damage type, and HP loss are specified in prose. Express the pipeline with the formula AST and modifier rows, and pin it with golden vectors.

## Done when

- [ ] Hit, miss, crit, and fumble each have a vector.
- [ ] DR, block, resist, and damage-type interaction each have a vector.
- [ ] A reference evaluator reproduces every vector; the prose is a reading of the tree.

See: ruleset/data/combat/; docs/core/formula.md; docs/core/vectors.md
