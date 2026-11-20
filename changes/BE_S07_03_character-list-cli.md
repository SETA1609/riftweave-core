### BE_S07_03_character-list-cli A character list, and characters travel between players and the GM as files (backend)
- Added: a character file's sync fields, `character` in the file envelope (`id`, `name`, `revision`, `changed_by`, `changed_at`); a character file without them imports as a new character.
- Added: `riftweave characters list|show|export|import|duplicate|delete`, the CLI's character list in the settings directory's `characters/`, with the sync rule on import: a higher revision replaces, a lower one is refused, and a diverged one shows both sides until one is kept (`--keep`).
- Added: a session's `party` (each member's sync fields and actor as loaded, and its actor now), carried by `start`, `step`, and `replay`; `riftweave characters party add` and `party export`, one file per player, a revision higher when the session changed it.
- Added: the Characters screen in `riftweave play`: list, open, export, import with the sync rule, duplicate, delete, add to the open session's party, and export the party.
- Changed: `menu.json` entries may name `issue_apps`; Characters is live in the CLI and pending on `FE_S07_04_character-list-page` in the browser.
