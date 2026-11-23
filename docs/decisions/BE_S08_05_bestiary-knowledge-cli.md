# BE_S08_05_bestiary-knowledge-cli A bestiary, and the GM decides what the players know (backend)

Date: 2026-11-23

- **Decision:** a `bestiary` command holds the knowledge rule, `reveal` and `view` ([bestiary.md](../backend/bestiary.md)); listing and filtering the monsters stay the `catalog`'s (BE_S08_10), so the bestiary adds no second monster list.
- **Marks:** a mark is a monster at `unknown`, `seen`, or `known`, for the table or one player; a player knows the higher of the table's and their own. A reveal replaces the holder's mark in place, so Python and Go replay the same list. [V 2026-11-23, `bestiary` vectors in Python and Go]
- **Campaign state in the session:** the marks live in the session (`bestiary.knowledge`), not the settings, through a new `sessionStateAdapters` table beside the actor's `sessionAdapters`: the session fills a `bestiary` step and keeps a reveal's marks, and a replay rebuilds them. `matches` still compares the actor only. [V 2026-11-23, `TestARevealStepKeepsTheMarksAndReplays`]
- **Lore check:** `resolution` with the `lore` skill, on the player's actor and roll; a success marks the monster `known` for that player, a failure changes nothing. [A] The issue left whether the check reads a monster difficulty open; it does not, so the target is the player's Lore alone.
- **A view takes the monsters to show** (`monsters`), the refs a catalog list returns, so the filter is the catalog's and the view never lists the whole ruleset by itself.
- **The CLI screen marks for the table** (`s`, `k`, `u`) and shows the players' view (`p`); per-player marks and Lore checks go through the command or a session step until the browser page and rooms add them. [A]
- **Menu:** Bestiary is no longer pending in the CLI; it stays pending in the browser on `FE_S08_06_bestiary-page`.
- **Also fixed:** `menu.compendium` was missing from the locale files, so the web menu showed the raw key and the frontend's menu specs failed on `development` since BE_S08_10 and FE_S11_01 merged apart. [V 2026-11-23, `npm test`, 123 of 123]
- **Open:** whether a monster carries a Lore difficulty that lowers the target; showing a `seen` monster's picture once monsters have one.
