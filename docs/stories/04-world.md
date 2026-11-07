# M4 World

Roadmap: [M4](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M4 World

**Goal:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** prices, treasure, and monster stats are consistent under checkable balance bands.

Done: M4-01, M4-02, M4-04, M4-05, M4-06 ([done.md](done.md#m4-world))

#### M4-03 Bestiary to the content floor
- Status: review
- Priority: high
- Needs: M1-02
- Was: RW-22
- Do: 11 monsters. An SRD needs a bestiary that spans the level range and the phases.
- Files: ~4 (ruleset/data/monsters/, ruleset/schemas/monster.schema.json)
- See: ruleset/data/monsters/core.json, docs/modules/monsters.md
- Done when: at least 40 monsters covering levels 1 to 30; every phase and damage type is used at least once as an attack and as a resistance; each monster inherits from a base plus overrides; a family member derives from its base monster, and an element variant is a base monster plus a template; no name comes from another game.

#### M4-07 Loot composes items: default materials and size
- Status: todo
- Priority: medium
- Needs: M2-05, M4-02
- Do: A loot row's `material` lands on every weapon and armor the row yields, so the stone golem's steel hoard can drop a steel bow. Since M2-05 each equipment base lists `default_materials`: the materials it is found and sold in. Any other material is player-crafted only (a granite bow), never loot. When a row's material is not in an item's `default_materials`, the item drops in its first default material instead, and its value follows the material that lands (`value_percent`). Decide whether a row with no `material` also takes the first default, and check the treasure budget against it. Add an optional `size` on loot rows (`undersized`, `normal` by default, `oversized`, the weapon `size` from M2-05): it applies to every weapon the row yields, moves its dice one rung (`combat/dice_ladder.json`) and its attack by +10 or −10, and is how a sized weapon enters play, since sized weapons are composed, not rows. Size changes value: the item is worth its size `value_percent` (undersized 75, oversized 150) times the material `value_percent`. Merchant stock is out of scope.
- Keys: `default_materials` on equipment bases; loot row `material` and `size`
- Files: ~8 (docs/core/loot.md, ruleset/schemas/loot.schema.json, backend/contract/loot.json, ruleset/scripts/loot_reference.py, backend/internal/data/data.go, backend/internal/loot/loot.go, ruleset/tests/vectors/loot.json, docs/decisions/M4-07.md)
- See: docs/modules/weapons.md (default materials), docs/decisions/M2-05.md
- Done when: vectors cover a row material inside an item's defaults (kept), outside them (the steel hoard yields a steel sword and a wooden bow, each valued by its own material), an item that takes no material (a potion), a row with `size: oversized` yielding a longbow at 1d10, −10 attack, and 150% value (and with a material, both factors), and a sized row yielding armor (size ignored); the Python reference and Go pass them; the treasure budget check still passes.
