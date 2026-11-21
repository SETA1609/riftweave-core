### BE_S09_01_spell-author-wasm A spell editor page over the author command (backend)
- Added: `author` operation `effects` lists the effects a spell of a color may use: on that color or on none, with the spell channel, each as written in effects/core.json.
- Fixed: the browser core embeds `ruleset/schemas` and opens with them, so `author` checks a draft in the page instead of refusing.
