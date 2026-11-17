# APP-37 File the build import and printing issues

Date: 2026-11-17

- **Decision:** four issues, owner request 2026-11-17: APP-38 (a character from a suggested build), APP-39 (print the current character), APP-40 (spell cards, and custom spells as cards), APP-41 (custom NPCs and monsters, printed). APP-40 and APP-41 are later work, priority low.
- **A build applies through `progression`** [A]: the app replays the build's level-ups with the rules instead of copying a stored character, so a module that changes a perk or a formula still gives a legal character, or stops at the level that breaks. M6-12's pre-built characters use the same path.
- **A build stays a suggestion:** the character keeps only the build's key, to show the suggested perk at later level-ups (APP-31); any legal choice is allowed, as M5-47 requires.
- **Printing adds no numbers:** every printout still comes from `print` (APP-11), which reads other commands; the apps only call it. A custom spell, monster, or build is checked before it prints or applies, and a failing one is refused with its problems.
- **Open questions:**
  - Should a build also suggest race and background, or only the parts it holds today (APP-38 assumes only what it holds)? Owner.
  - The spell card size and cards per page (APP-40 assumes poker size, nine per page). Owner.
  - Does a custom monster need an editor in the app, or is JSON enough for now (APP-41 assumes JSON, an `author` kind `monster` later)? Owner.
- **Overlap:** M5-47 is in review; APP-38 needs its `builds` collection and schema as they merge, and changes nothing in it. APP-39, APP-40, and APP-41 all edit `print.json` and `print.go`; they run one after another.
