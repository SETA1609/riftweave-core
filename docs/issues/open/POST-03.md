#### POST-03 Magic crafting module: enchanting
- Status: todo
- Priority: low
- Needs: POST-01, M3-01
- Was: RW-33
- Do: Affix vs full item model, charges, attunement, rarity, runes. Owner rule, 2026-11-12: magic items give spell-like abilities with casts or uses per day; scrolls stay gold-only consumables (M5-48). Owner rule, 2026-11-12: the module has two paths. **Enchanting** puts enchantments on weapons and armor (the base item's `enchantment_slots`, material, and phase). **Wondrous items** are made from fixed recipes, like any other crafting recipe, with the `enchanting` skill as the recipe's check [V owner 2026-11-12], for example a wand of fireball. A wondrous item can only hold a stock spell, one authored in `spells/core.json`, never a spell a player crafted, so a recipe names its spell by `core:spell/<key>`. Its casts or uses per day follow the rule above.
- Files: ~6 (ruleset/modules/<module>/, docs/modules/<module>.md)
- See: docs/modules/magic-crafting.md
- Done when: decision record, then data and vectors.
