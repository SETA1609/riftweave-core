# Files

Modified: 2026-11-21

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
| `character` | a stored actor (`common.json#/definitions/stored_actor`), with its sync fields in `character` | character creation; the [character list](#the-character-list) |
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

- **Import** on the home page (`/import`, `frontend/src/app/file-import.ts`) opens one file of any kind. The envelope is read by `frontend/src/app/files.ts`, with the CLI's rules and messages, and the data is checked by the core: a character by the `stored_actor` command, a session by a replay, a spell by the `author` command. A character starts a save; a session becomes a save; a spell is checked and waits for spell creation (APP-22); a settings file goes to the settings import.
- **A missing module** shows the CLI's alert, word for word, with **Enable and import** (only for a module this app has) and **Import anyway**. Another rules version shows a warning and still imports.
- **Export**: **Export character** on character creation, and **Export file** for each saved game, both stamped with the rules the session runs and the enabled modules at their manifests' versions. The backup download stays the session JSON itself (APP-16).

[V 2026-11-18, `files.spec.ts` and the end-to-end specs, run in CI]

## Fixtures

`ruleset/tests/fixtures/files/` holds a file of each kind, one that names the factions module,
and one whose character breaks the contract twice (level 0, which the schema refuses, and a slot that holds no item, which the `stored_actor` command refuses). `test_contract.py` validates them against
`file.json` and checks their layout; the Go tests import each kind and export it back byte for
byte, and `files.spec.ts` reads and writes each back the same way. `character-synced.riftweave.json` is a player's character with its sync fields; `characters_test.go` carries it from a player's list to a GM's party and back. [V 2026-11-20, `go test ./cmd/riftweave`, `test_contract.py`]

## The character list

The apps work with no connection, so a GM and four players can play offline and pass characters
by hand (BE_S07_03, owner 2026-11-16). Each app keeps a **character list**: the CLI in
`os.UserConfigDir()/riftweave/characters/` (or `RIFTWEAVE_CONFIG_DIR/characters/`), one
character file per character, named `<id>.riftweave.json`; the frontend in IndexedDB
(database `riftweave-characters`, store `characters`, keyed by `character.id`; FE_S07_04). The list shows each character's name, level, last change, and
who made it, and opens, exports, imports, duplicates, and deletes them.

A character file carries its **sync fields** in `character`, beside `data`
(`common.json#/definitions/character`):

```json
"character": { "changed_at": "2026-11-19T18:30:00Z", "changed_by": "player", "id": "5f0c7a52-…", "name": "Mira Vell", "revision": 3 }
```

- `id` is a UUID made when the character is created; it never changes. [A]
- `revision` is raised by one on every change; `changed_by` (`gm` or `player`) and `changed_at`
  (RFC 3339, UTC) say who made the last one and when. [A]
- Only a character file has them. A character file without them (APP-30's) imports as a new
  character: a new id, revision 1, the name given (else "Unnamed character"), changed by the
  importing side (`player` unless told `gm`). [A]
- A duplicate is a new character: a new id, revision 1, the name plus " (copy)". Exporting
  changes nothing, so the revision stays.

### The sync rule

An imported character file is compared with the saved character of the same `id`
[A, owner 2026-11-16]:

| The file's revision | Outcome | What changes |
| --- | --- | --- |
| no saved character has the id | `added` | the file is saved |
| higher than the saved one | `replaced` | the file replaces the saved one |
| lower than the saved one | `refused` | nothing; the note names both revisions |
| the same, with the same name and actor | `unchanged` | nothing |
| the same, with another name or actor (both sides changed it) | `diverged` | nothing: both sides are shown, and the person keeps one (`kept` for the saved side, `replaced` for the incoming one) |

It never merges field by field. Who changed it and when are not compared. A module the file
names that this device does not load stops the import as for any file
([above](#rules-both-apps-keep)).

### The hand-off

Before a session each player exports their character and sends the file any way they like (a
chat, a USB stick, a shared folder). The GM imports each one and adds it to the session's
**party** ([sessions.md](sessions.md#the-party)). The session plays offline on the GM's device.
Afterwards the GM exports the party, one file per player: a character the session changed goes
out one revision higher, changed by `gm`, now; one it did not change goes out as it was loaded.
The session records the export, so a GM who exports mid-session and again at the end sends
revision 4 and then 5, never 4 twice with other content.
Each player imports theirs, a higher revision, so it replaces the saved one. [V 2026-11-20,
`characters_test.go`]

### In the frontend

`Characters` (`frontend/src/app/characters.ts`) applies the same rule with the same outcomes and
notes, so a file passes from a player's browser to a GM's CLI and back: the export of a saved
character is its file byte for byte (`character-synced.riftweave.json` round-trips), named as the
CLI names it (`mira-vell-5f0c7a52.riftweave.json`). The data is checked by the core
(`stored_actor`) and a module this browser does not load stops the import, as on the Import page.
[V 2026-11-21, `characters.spec.ts`]

### In the CLI

```
riftweave characters list                                    every saved character: id, name, level, revision, changed_by, changed_at
riftweave characters show <id>                               its sync fields and its stored actor
riftweave characters export <id> [file]                      its character file, stamped with these rules and the enabled modules
riftweave characters import <file> [--name <name>] [--as gm|player] [--keep saved|incoming] [--enable-modules | --accept-loss]
riftweave characters duplicate <id> [--as gm|player]
riftweave characters delete <id>
riftweave characters party add <session file> <id>           load a saved character into the session file's party
riftweave characters party export <session file> <directory> one character file per party member
```

An import prints its report (`outcome`, `character`, `note`, `warnings`). A `refused` or
`diverged` import exits 1 with the report on stderr and changes nothing; a diverged report holds
both sides (`saved`, `incoming`), and `--keep` picks one. The human CLI's **Characters** screen
does the same ([app-layout.md](app-layout.md#characters)).

