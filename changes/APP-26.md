### APP-26 The human CLI opens on a main menu, and Start opens the tools
- Added: `riftweave play` opens on a main menu (Start, Characters, Settings, Modules, Network, Quit); Start opens the tools (GM tools, Character creation, Spell creation, Skill resolution, Bestiary). Entries whose issue is not merged are shown dimmed with its id.
- Added: `backend/contract/menu.json`, the menu tree both apps show.
- Added: the Settings, Modules, and Network screens edit and save the settings file; in `server` mode commands go to the HTTP API.
- Added: Skill resolution, a form over the `resolution` command that rolls a blank roll.
- Changed: Create a character and Resolve an attack moved under Start (Resolve an attack under GM tools).
