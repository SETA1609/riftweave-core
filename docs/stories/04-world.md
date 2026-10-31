# M4 World

Roadmap: [M4](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M4 World

**Goal:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** prices, treasure, and monster stats are consistent under checkable balance bands.

Done: —

#### RW-20 Economy: currency, price formula, merchants
- Status: todo
- Priority: high
- Needs: —
- Do: Equipment has `cost` and materials have `valueFactor`, but there are no price rules, denominations, or merchant model. Core play needs buy, sell, and barter.
- See: issues.md section 1
- Done when: `docs/core/economy.md` and a price formula in the AST; merchant stock and buy/sell spread as data; vectors for a purchase and a sale with a barter check.

#### RW-21 Loot and treasure tables
- Status: todo
- Priority: high
- Needs: —
- Do: No loot data exists. Monsters need drops, and levels need treasure budgets.
- See: issues.md section 1
- Done when: `ruleset/data/loot/` with a schema; A treasure-by-level budget consistent with the price formula; seeded roll vectors.

#### RW-22 Bestiary to the content floor
- Status: todo
- Priority: high
- Needs: —
- Do: 11 monsters. An SRD needs a bestiary that spans the level range and the phases.
- See: ruleset/data/monsters/core.json, docs/modules/monsters.md
- Done when: at least 40 monsters covering levels 1 to 30; every phase and damage type is used at least once as an attack and as a resistance; each monster inherits from a base plus overrides.

#### RW-23 Traps and environmental hazards
- Status: todo
- Priority: medium
- Needs: —
- Do: Mentioned as missing in the backlog. Minimal core: a hazard is a trigger plus a modifier or damage packet, with a detect and a disarm check.
- See: issues.md section 8
- Done when: schema and at least 10 hazards; detect and disarm use the one resolution function.

#### RW-24 Encounter and balance validation
- Status: todo
- Priority: medium
- Needs: —
- Do: `combat.md` has balance guidelines (damage, DR, HP by tier, hits-to-kill). Turn them into checks so data cannot drift.
- See: docs/modules/combat.md
- Done when: A script reports monsters and gear outside the guideline bands; bands are data, not prose; report runs in CI as a warning, then as a gate.
