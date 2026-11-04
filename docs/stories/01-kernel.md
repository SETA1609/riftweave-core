# M1 Kernel closed

Roadmap: [M1](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M1 Kernel closed

**Goal:** every kernel function has a vector, combat and progression are data, and CI enforces it. Content written before this milestone must be rechecked after it.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 1 and 2 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold.

Done: M1-01, M1-02, M1-03, M1-04, M1-06, M1-07, M1-08 ([done.md](done.md#m1-kernel-closed))

#### M1-05 Modifier stacking and condition vectors
- Status: review
- Priority: medium
- Needs: M1-07, M1-08
- Was: RW-7
- Do: The contract states one stacking rule. Prove it: ordering of `add`, `mul`, `set`, `grant`, and `resist`, duplicate sources, expiry, and condition application and removal.
- Files: ~8 (docs/core/modifier.md, docs/core/contract.md, ruleset/schemas/schema.json, ruleset/schemas/condition.schema.json, ruleset/data/conditions/core.json, ruleset/tests/vectors/)
- See: docs/core/modifier.md, docs/core/contract.md, ruleset/data/conditions/core.json
- Done when: vectors for each operator, mixed operators, and same-source duplicates; vectors for each of the 24 conditions' apply and remove paths, or a documented subset with a reason; A contradiction between vector and docs fails CI.
