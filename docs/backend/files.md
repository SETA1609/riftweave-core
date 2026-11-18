# Files

Modified: 2026-11-18

Anything a person makes in one app can be exported as a file and imported into the other
(APP-30). One envelope holds every kind, defined once in
[`backend/contract/file.json`](../../backend/contract/file.json):

```json
{
  "data": { "...": "exactly what the command contract already returns" },
  "modules": [{ "id": "factions", "version": "0.1.0" }],
  "riftweave": "character",
  "rules_version": "0.1.0"
}
```

| `riftweave` | `data` is | Made by |
| --- | --- | --- |
| `character` | a stored actor (`common.json#/definitions/stored_actor`) | character creation; the character list (APP-33) |
| `session` | a session (`common.json#/definitions/session`), the save of APP-16 | a saved game |
| `spell` | a spell row as `spells/core.json` holds it, which the `author` command accepts (APP-10) | spell creation (APP-22) |

The **settings** file keeps its own format, which already names its kind (`"riftweave":
"settings"`, [settings.md](settings.md)); an import that finds it hands it to the settings
import. No format is new: `data` is what a command already returns, so a file can be fed to the
next command as it is. [A]

## Rules both apps keep

- **One layout.** Keys sorted, a two-space indent, and a final newline, as the settings file, so
  a file exported by either app is byte for byte the same.
- **Checked before it applies.** The envelope must name a kind, `rules_version`, `modules`, and
  `data`; the data must validate against the contract (a character and a session against
  `common.json`, a spell through the `author` command's check). A file that fails is refused
  with the schema's messages, and nothing changes.
- **Another rules version warns.** A file made with other rules still opens when its data
  validates, with a warning that names both versions. [A]
- **A missing module stops the import.** `modules` lists the modules enabled where the file was
  made. A module this device does not load would lose the rows that need it, so the import stops
  with an alert naming the module and every `module:<id>/…` ref in the data, and goes on only
  when the person enables the module or confirms the loss (owner, 2026-11-16). A module loaded
  here at another version warns.

## In the CLI

```
riftweave export <character|session|spell> [file]          the data from the file or stdin, in the envelope, on stdout
riftweave import <file>                                       check it, then print {"kind", "data", "warnings"}
riftweave import <file> --enable-modules                      enable the missing modules in the settings file, then import
riftweave import <file> --accept-loss                         import without the missing modules
```

`export` stamps the file with these rules and the settings file's enabled modules, at the
versions their manifests name. The main menu of `riftweave play` has **Import**: a path, and the
same check; a character becomes the character a dev session starts from, and a session the dev
panel's session. Adding ` --accept-loss` to the path goes on past a missing module. [V
2026-11-18, `files_test.go` and `play_test.go`, run in CI]

## In the frontend

"Export" on each screen that makes a file and one "Import" on the home page that reads the kind
are the frontend half of APP-30 (`APP-30-frontend-…`), on the same envelope and fixtures. A
spell is exported from spell creation, which is APP-22. [U]

## Fixtures

`ruleset/tests/fixtures/files/` holds a file of each kind, one that names the factions module,
and one whose character breaks the contract. `test_contract.py` validates them against
`file.json` and checks their layout; the Go tests import each kind and export it back byte for
byte. A character file's sync fields come with APP-33. [U]
