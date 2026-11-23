### GD_S20_02_custom-monster-rules A custom monster computes like a character, without a player's limits
- Added: docs/core/actor-sheet.md#a-custom-monster says what a GM's own monster skips (point-buy, the attribute maximum, perk prerequisites and ranks, the skill pool) and what it keeps; `monster_sheet` vectors cover a monster past each player limit.
- Changed: `monster_sheet` refuses a level over `max_level` and a written skill over `skill_cap`, in Python and Go.
- Fixed: the Python reference refuses an unknown perk with Go's sentence instead of raising.
- Added: `custom_monster_problems` in `monster_reference.py` reports every problem of a written monster row at once.
