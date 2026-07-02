# PR Tasks — Project Review (2026-07-02)

## PR-1: Equipment reference migration (H1)

**Status:** In progress on `review/project-review`

**Problem:** Numeric equipment `id` values 22–25 collided between `weapons.json` and `armor.json`, making background `starting_equipment` refs ambiguous.

**Solution:** Migrate equipment cross-references to namespaced string keys instead of renumbering alone.

### Adopted identifier format

```text
core:<category>/<key>
```

Examples:

- `core:weapons/shortsword`
- `core:armors/padded_tunic`
- `core:consumables/potion_minor_healing`

Categories map to equipment source files: `weapons.json` → `weapons`, `armor.json` → `armors`, `consumables.json` → `consumables`.

### Changes

- `backgrounds/core.json` — all `starting_equipment.item` values use `core:category/key`
- `validate.py` — equipment key index, namespaced ref resolution, duplicate key detection (numeric id check retained)
- `background.schema.json`, `docs/modules/backgrounds.md`, `docs/modules/character-creation.md` — document the convention

Numeric `id` fields remain in equipment data during transition; legacy numeric and bare-key refs still validate.