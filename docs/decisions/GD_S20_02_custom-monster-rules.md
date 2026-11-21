# GD_S20_02_custom-monster-rules A custom monster computes like a character, without a player's limits

Date: 2026-11-21

- **Decision:** a custom monster is a monster row a GM writes, and it computes through `monster_sheet` exactly as a core row does; the rule is [actor-sheet.md#a-custom-monster](../core/actor-sheet.md#a-custom-monster).
- **What was already true:** `monster_sheet` never applied point-buy, an attribute maximum, perk prerequisites, a rank limit, or `perks_per_level`; the new vectors pin that down. [V 2026-11-21, `monster_sheet` vectors in Python and Go]
- **What changed in the rule:** `monster_sheet` now refuses a level over `max_level` (30) and a written skill over `skill_cap` (100), in Python and Go with the same sentence. No core monster writes a skill or a level over either, so no data moves. [V 2026-11-21, `validate.py`, `test_bestiary.py`]
- **A fix:** the Python reference raised a `KeyError` on an unknown perk; it now returns the error Go already gave (`perk "x" is not defined`).
- **Refuse, not clamp:** a written skill over the cap is an error, so the sheet never shows a number the GM did not write.
- **No new attribute bound:** `monster.schema.json` already sets at least 1 and no maximum, so no `custom_monster` bound was added to the data.
- **Where the check lives:** `custom_monster_problems` is in `monster_reference.py`, not `kernel_reference.py` as the issue said, because it reuses validate.py's monster composition and reference index, which that module already imports.
- **Open:** should the balance report refuse a custom monster far outside its role's band, or only note it? The spec says note it [A]; the owner decides. The Go check for the editor is BE_S20_03.
