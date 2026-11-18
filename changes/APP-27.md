### APP-27 The Angular app opens on a home page, and Start opens the tools
- Changed: the tab bar gives way to the Angular router; the app opens on the home page (Start, Characters, Settings, Modules, Network), and Start (`/play`) opens the tools. The character sheet, the attack resolver, the playthrough, and the saved games each have a route under `/play`.
- Added: `rw-menu-grid`, one component for every menu, built from `backend/contract/menu.json`; pending entries are disabled with their issue id. Menu entries take `apps` when only one app shows them (Quit in the CLI; Saved games and Playthrough in the browser).
- Added: the Settings, Modules, and Network pages edit the settings file; the network setting switches the API transport between the core in the page and a server.
- Added: Skill resolution, a form over the `resolution` command.
