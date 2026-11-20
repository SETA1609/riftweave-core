# App layout

Modified: 2026-11-18

The human CLI (`riftweave play`) and the Angular app (APP-27) show the same menus, so a player
who learns one finds their way in the other. The menu tree is one file,
[`backend/contract/menu.json`](../../backend/contract/menu.json): the CLI embeds it, and the
Angular app imports it. Everything works offline: a GM and four players can run a whole session
with no connection, and the network only adds what a shared table needs online (APP-35).

## The menu tree

Each entry has a `key`, the `parent` menu it sits in (`main` for the main menu), a `label`, the
`issue` that adds it while that issue is not merged (with `issue_apps` when the issue is one
app's half, so the other app already shows the entry), and `apps` when only one app shows it
(`cli` or `web`): Quit is the CLI's, and Saved games and Playthrough are the browser's, which
keeps its saves in IndexedDB (APP-27). The app adds **Back** at the end of every menu but the
main one; it is not in the file.

```
Main menu
├── Start
│   ├── GM tools
│   │   ├── Resolve an attack
│   │   ├── Encounter tracker          APP-34
│   │   └── Playthrough                web only
│   ├── Character creation
│   ├── Spell creation                 APP-22
│   ├── Skill resolution
│   ├── Bestiary                       APP-29
│   └── Saved games                    web only
├── Characters                         FE_S07_04 (web only; the CLI has it)
├── Settings
├── Modules
├── Network
├── Import                           APP-30
└── Quit                             CLI only
```

In the Angular app the main menu is the home page at `/`, Start is `/play`, a tool is
`/play/<key>`, and the other main entries are `/<key>` ([frontend.md](frontend.md#screens)).

- **A pending entry** (one with an `issue`) is shown dimmed with that issue's id, and choosing
  it says which issue brings it, so the menus are the same in both apps from the start. [A]
- **Esc** goes back one menu, to the entry it came from. From a form or a screen it goes back
  to the menu that opened it.
- **Keys** are the menu keys of the file. The i18n keys of APP-28 hang off the same keys. [A]

## Start: the tools

- **GM tools** holds the attack screen (`combat`) and, with APP-34, the offline encounter
  tracker.
- **Character creation** is eight attributes and a level, run through `derived`.
- **Skill resolution** is one d100 check through `resolution`: the skill and its value, LCK, a
  target modifier (an `add` on the skill, source `situation:gm`), and the roll. A blank roll is
  rolled by the app, and the input it sends names that roll, so the check can be replayed with
  the agent CLI and gives the same result.

## Characters

The character list ([files.md](files.md#the-character-list), BE_S07_03). In the human CLI the
screen lists each saved character with its level, its last change (local time), who made it,
and its revision, then **Import a character file**, **Export the party of the open session**, and
**Back**. On a character, enter opens it (it becomes the character a dev session starts from, as
Import does), `e` exports it to a path, `d` duplicates it, `x` deletes it after a typed `yes`, and
`p` adds it to the party of the open session (one imported from a file or started in dev mode).
Import applies the sync rule; on a divergence it shows both sides' last change, and enter on
Import again asks which to keep (`saved` or `incoming`; anything else keeps both as they were).
[V 2026-11-20, `play_test.go`] The frontend page is `FE_S07_04_character-list-page`.

## Settings, Modules, and Network

These three screens edit the settings file of APP-25 ([settings.md](settings.md)). Every change
is checked by the settings rules and saved at once; a change the rules refuse (a language code
that is not one, a module named twice) is shown, and nothing is saved.

- **Settings:** the language, the theme (system, light, dark; the CLI sets its colours for a
  light or dark background), whether players see only the monsters they know (APP-29), and
  export and import of the whole file.
- **Modules:** every module in `ruleset/modules/`, by manifest id, with a check box, saved in
  `settings.modules`. Core heads the list, always checked and locked, because core is always
  loaded. Today the choice is saved for both apps; loading only the checked modules is not part
  of APP-26. [A]
- **Network:** `local` (the default: the core in the program, offline) or `server` (the HTTP API
  at `server_url`, [api.md](api.md)). In server mode every command goes to
  `POST <server_url>/v1/commands/<name>` with the same input, so each screen works the same way
  with either. Server mode needs a URL. `room` is listed and disabled until APP-17. In the
  Angular app `local` is the build's own transport (the core in the page, or the API a dev
  build names), and the network setting switches the `Transport` of `api.ts` on the next call.
  A server on another origin than the page needs CORS, which the API does not send yet
  ([api.md](api.md#not-covered)). [U]
