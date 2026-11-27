# BE_S23_03_tui-flow-audit Move every rule step that lives only in the terminal app into the core

Date: 2026-11-24

- **Decision:** the terminal app (`backend/cmd/riftweave/play/`, 2,672 lines) holds one rule step, a new stored actor built by hand, and it moves to the core as `stored_actor`'s `new`. Everything else it does is input, formatting, or a call to a core command, and goes with the app in `BE_S23_05`.
- **The audit**, screen by screen [V 2026-11-24, read in this change]:

| File | Screen | Calls | Its own steps | Verdict |
| --- | --- | --- | --- | --- |
| `play.go` | menu, forms, dev panel | `combat`, `resolution`, `derived`, `session`, `dev` | `startActor`: a new actor (attributes 5, level 1, xp 0, empty lists, pools full from `derived`) | **rule: moved** to `stored_actor` `new` |
| `play.go` | dev panel | `dev`, `session` | parses `:give`, `:inflict`, `:level`, `:roll`, and queues d100 rolls (1 to 100) | input |
| `resolution.go` | skill check | `resolution` | rolls a blank d100 so the input always names its roll; a target modifier becomes a `situation:gm` modifier | input: rolls are inputs, and the client rolls them |
| `creation.go` | creation | `creation` (`rules`, `check`, `finish`) | moves a point up or down, then the core checks the draft | input |
| `characters.go` | Characters | `cmd/riftweave/characters` | list, export, delete, import through the shared package | kept: the agent commands use it |
| `import.go` | Import | `cmd/riftweave/files` | keeps an imported character's attributes and level as the forms' defaults | UI state |
| `bestiary.go` | bestiary | `catalog`, `bestiary`, `session` | parses filters; hides level, phase, and type when the core's `shows` is not `full` | input and formatting: `bestiary` decides `shows` |
| `compendium.go` | compendium | `catalog` | the table and a row's JSON | formatting |
| `print.go` | print | `print` | writes the page to a file | formatting |
| `settings.go`, `network.go` | settings | none | local or remote runner, the settings file | adapter |
| `menu.go` | menus | none | the menu rows | formatting |

- **The same step in the web app:** `api.ts`'s `newCharacter` built the same actor by hand; it calls `stored_actor` `new` now, so a new pool or field is one change in the core. [V 2026-11-24, the api spec]
- **The rule:** `stored_actor` takes `actor` (checked and returned) or `new` (`attributes`, optional `level` and `tags`), never both; the contract's `oneOf` refuses both, and the command does too. [V 2026-11-24, 2 new vectors in Python and Go]
- **Not done here:** the TUI's flow tests move in `QA_S23_04`; the app itself goes in `BE_S23_05`.
