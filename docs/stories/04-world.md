# M4 World

Roadmap: [M4](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M4 World

**Goal:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** prices, treasure, and monster stats are consistent under checkable balance bands.

Done: M4-01, M4-02, M4-03, M4-04, M4-05, M4-06 ([done.md](done.md#m4-world))

#### M4-07 Loot composes items: default materials and size
- Status: todo
- Priority: medium
- Needs: M2-05, M4-02
- Do: A loot row's `material` lands on every weapon and armor the row yields, so the stone golem's steel hoard can drop a steel bow. Since M2-05 each equipment base lists `default_materials`: the materials it is found and sold in. Any other material is player-crafted only (a granite bow), never loot. When a row's material is not in an item's `default_materials`, the item drops in its first default material instead, and its value follows the material that lands (`value_percent`). Decide whether a row with no `material` also takes the first default, and check the treasure budget against it. Add an optional `size` on loot rows (`undersized`, `normal` by default, `oversized`, the weapon `size` from M2-05): it applies to every weapon the row yields, moves its dice one rung (`combat/dice_ladder.json`) and its attack by +10 or −10, and is how a sized weapon enters play, since sized weapons are composed, not rows. Size changes value: the item is worth its size `value_percent` (undersized 75, oversized 150) times the material `value_percent`. Merchant stock is out of scope.
- Keys: `default_materials` on equipment bases; loot row `material` and `size`
- Files: ~8 (docs/core/loot.md, ruleset/schemas/loot.schema.json, backend/contract/loot.json, ruleset/scripts/loot_reference.py, backend/internal/data/data.go, backend/internal/loot/loot.go, ruleset/tests/vectors/loot.json, docs/decisions/M4-07.md)
- See: docs/modules/weapons.md (default materials), docs/decisions/M2-05.md
- Done when: vectors cover a row material inside an item's defaults (kept), outside them (the steel hoard yields a steel sword and a wooden bow, each valued by its own material), an item that takes no material (a potion), a row with `size: oversized` yielding a longbow at 1d10, −10 attack, and 150% value (and with a material, both factors), and a sized row yielding armor (size ignored); the Python reference and Go pass them; the treasure budget check still passes.

#### M4-08 Drops live on the monster, with defaults from the creature type
- Status: todo
- Priority: medium
- Needs: M4-03, M4-07
- Do: Today `loot/core.json` has one drop row per monster, so every new monster is a second edit in another file, a derived monster cannot inherit its parent's drop, and most rows repeat the same few tables. Move drops onto the bestiary. Each creature type in `monsters/bases/` declares a default `loot` list (undead roll `grave_goods`, and so on); a monster row writes `loot` only when it differs; a derived monster inherits its base monster's, and a template may add rows (a fire template's essence). The rolls scale with level: a table row's `rolls` defaults to a count read from the level band (a formula tree next to `budget_cp`), so a level-30 boss does not list `hoard_major` three times by hand. `loot/core.json` keeps the tables and the budget, and an optional `drops` section that, for a monster it names, replaces the resolved loot: a module changes a monster's loot there without overriding the whole monster row, which is why drops were put in the loot file in M4-02. Core ships no `drops` rows unless a monster needs one the bestiary cannot express.
- Keys: `loot` on monster rows and creature-type bases; the rolls-by-level formula; `drops` in the loot file as an override only
- Files: ~12 (docs/core/loot.md, docs/modules/monsters.md, ruleset/schemas/monster.schema.json, ruleset/schemas/loot.schema.json, ruleset/data/monsters/bases/, ruleset/data/monsters/core.json, ruleset/data/loot/core.json, ruleset/scripts/validate.py, ruleset/scripts/loot_reference.py, ruleset/scripts/test_loot.py, backend/internal/data/data.go, ruleset/tests/vectors/loot.json)
- See: docs/decisions/M4-02.md (drops in the loot file), docs/modules/monsters.md#derived-monsters
- Done when: every resolved monster has loot from its type, its row, its base monster, or a template; a variant with no `loot` of its own drops what its base monster drops; the rolls follow the level band unless a row sets them; a `drops` row in the loot file (or a module's) replaces a monster's loot; `loot/core.json` holds no core drop row that the bestiary could express; every monster still fits the budget at its level; the loot vectors pass in Python and Go.
