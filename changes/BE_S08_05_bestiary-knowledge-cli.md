### BE_S08_05_bestiary-knowledge-cli A bestiary, and the GM decides what the players know (backend)
- Added: the `bestiary` command: `reveal` sets the table's or a player's mark of a monster (`unknown`, `seen`, `known`) or rolls a Lore check that marks it known; `view` shows monsters as the GM or a player sees them, following `settings.bestiary.players_see_only_known`.
- Added: a session keeps the marks in `bestiary.knowledge`; a `bestiary` step is filled from them, a reveal writes them back, and a replay rebuilds them.
- Added: the human CLI's Bestiary: lists and filters the monsters (`type:`, `level:`, `phase:`), marks them for the table, and shows the players' view.
- Fixed: the missing `menu.compendium` locale key in en, de, and es.
