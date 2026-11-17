# Settings

Modified: 2026-11-17

The Angular app and the human CLI hold the same settings and trade them as one file (APP-25):
settings exported from the frontend import into `riftweave settings import`, and the other way
round. The file is defined once, as a JSON Schema, in
[`backend/contract/settings.schema.json`](../../backend/contract/settings.schema.json), beside the
command contract (it is a file format, not a command).

```json
{
  "bestiary": { "players_see_only_known": true },
  "locale": "de",
  "modules": ["factions", "wuxing"],
  "network": { "mode": "server", "server_url": "http://127.0.0.1:8080" },
  "riftweave": "settings",
  "theme": "dark",
  "version": 1
}
```

| Key | Holds |
| --- | --- |
| `riftweave`, `version` | `"settings"` and `1`: what the file is and its format. Any other version is refused. |
| `locale` | An i18next language code (APP-24, APP-28). |
| `theme` | `light`, `dark`, or `system` (APP-23); the CLI maps it to its own colours (APP-26). |
| `modules` | The enabled optional modules, by manifest id. Core is always loaded; a file that lists `core` is refused. |
| `network` | `mode`: `local` (offline, the default), `server` (an API at `server_url`), or `room` (`room_code`, APP-17). |
| `bestiary` | `players_see_only_known` (APP-29). |

## Rules both apps keep

- **Nothing is lost.** A key an app does not use is kept on import and written back on export,
  so a file from a later app survives a round trip through an older one.
- **One layout.** Both write the file with keys sorted, a two-space indent, and a final newline,
  so a file exported by either app is byte for byte the same. The Go rules are
  `backend/internal/settings`, the frontend's `frontend/src/app/settings-file.ts`; they check the
  same things with the same messages.
- **A bad file changes nothing.** A file that is not JSON, not a settings file, of another
  version, or that breaks a rule above is refused with a message naming the rule, and the saved
  settings stay as they were.

The shared fixture, `ruleset/tests/fixtures/settings/settings.json` (with an unknown key), is
validated against the schema (`test_contract.py`) and read and written back unchanged by a Go
test and a Vitest test. [V 2026-11-17]

## In the human CLI

The file lives in `os.UserConfigDir()/riftweave/settings.json` (`$RIFTWEAVE_CONFIG_DIR/settings.json`
when that is set). With none yet, the defaults hold: `en`, `system`, no modules, `local`.

```
riftweave settings export [file]   write the settings to the file, or to stdout
riftweave settings import <file>   check the file, then make it the CLI's settings
```

## In the frontend

`AppContext` (APP-23) is the one source. `exportSettings()` gives the file to download as
`riftweave-settings.json`: the last imported file, or the defaults, with the current theme and
language. `importSettings(text)` checks a picked file, applies its theme and language, keeps the
whole file in `localStorage`, and returns the message of the rule it breaks, if any. The buttons
belong to the settings screen of the home page (APP-27).

## Not done yet

- **Loading only the enabled modules** (`core.Open` with a selection): today both apps load every
  module. The CLI's play TUI (APP-26) and the home page (APP-27) apply `modules` when they gain a
  settings screen. [U]
- **The alert before module data is lost** on an import from a device with other modules is
  APP-30's, which names each exported file's modules and versions.
