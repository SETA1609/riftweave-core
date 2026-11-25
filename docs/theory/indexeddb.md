# IndexedDB, and data-oriented design in the browser

Modified: 2026-11-25

Background for how the web app keeps data on the device: saved characters, saved games, and,
with story S25, the profile and creation drafts. It covers what IndexedDB is, what it promises,
and where it breaks; what data-oriented design means for a store like it; and how this repo should
use it, store by store, with what each store is for and why. This page explains; the rules for
the code are in the issues of story S27 and in [frontend.md](../backend/frontend.md).

The specification and MDN were read from copies under `cache/docs/`, cloned at the commits named
in [Sources](#sources). The web.dev and Chrome pages were read on 2026-11-25. Every claim about
this repo's code was read in its files that day. [V 2026-11-25]

## What IndexedDB is

IndexedDB is a database the browser keeps for each origin. The W3C specification, *Indexed
Database API 3.0*, defines it in a few terms, and each one shapes how data should be laid out.

| Term | What the specification says | What it means here |
| --- | --- | --- |
| Database | Has a name, a version, and a set of object stores. Stores and indexes change only in an upgrade transaction, when the page opens a higher version. | The layout is versioned. Every change to the stores is code that runs once per device, on data the page did not write itself. |
| Object store | Its records are each a key and a value, kept sorted by key. No two records share a key. | A table with one primary key. A read by key, or by a key range, is an ordered lookup, not a scan. |
| Key | A number, a date, a string, binary data, or an array of keys. Keys compare by type first, then by value, and arrays item by item. | A compound key is an array: `[save_id, step]` keeps one save's steps together and in order. |
| Index | A second sorted list of records, kept up to date by the browser on every write. It maps a key read from the value by a key path to the record's primary key. It may be `unique`, or `multiEntry` (one entry for each item of an array). | A sorted column. A filter or a sort on it reads only the matching part. Every index costs something on every write. |
| Transaction | Has a scope (the stores it may touch), a mode (`readonly`, `readwrite`, or `versionchange`), and is atomic: all of it lands, or none of it. Read-only transactions can run together; read-write ones with overlapping scope wait for each other. | One user action is one transaction over every store it changes, so the stores never disagree. |
| Lifetime | A transaction is active only while its requests run. It commits on its own once none are left and control has gone back to the event loop. The specification expects transactions to be short. | Await nothing else (a network call, a timer) inside a transaction: it will have committed, or will fail with `TransactionInactiveError`. Prepare everything first, then write. |
| Durability | A hint on a read-write transaction: `strict` (flush to disk before `complete`), `relaxed` (`complete` once the data is with the operating system), or `default`. | `relaxed` for frequent small writes, `strict` where losing a write costs more than the wait (a migration). |
| Value | Stored as a structured clone of the object. | Writing is a copy of everything in the object, made on the page's thread. |

Two retrieval calls matter for performance. `getAll(query, count)` returns every value in a key
range, up to `count`, in one request. `getAllRecords()` is new in this edition: it returns keys
and values together, and the specification says it is in Chrome and Edge 141 only. A cursor
(`openCursor`) walks records one at a time and can stop early, at one request per record.

## What it promises, and what it does not

**Space.** The data is *best-effort* by default: it stays while the origin is under its quota and
the device has room. MDN lists the quotas by browser:

- **Chrome and Edge:** up to 60% of the disk for one origin.
- **Firefox:** in best-effort mode, the smaller of 10% of the disk and 10 GiB for the whole site.
  With persistent storage, 50% of the disk.
- **Safari (macOS 14, iOS 17):** about 60% of the disk for a browser app, about 15% in an app that
  embeds a web view.

A write over the quota fails with `QuotaExceededError`. `navigator.storage.estimate()` gives
the usage and the quota, as estimates.

**Eviction.** When a device runs low on space, the browser deletes whole origins, least recently
used first, and skips origins whose storage is *persistent*. An origin is deleted all at once,
never in part. Safari also deletes script-written data proactively. With tracking prevention on,
an origin that has had no user interaction (a click, a tap) in the last seven days of browser use
loses it. `navigator.storage.persist()` asks for persistent storage:

- **Firefox** asks the user.
- **Chrome and Safari** decide without a prompt, from how the person uses the site.

Chrome's article on persistent storage (web.dev) says automatic deletion is rare and that users
clearing data themselves is far more common. Its advice is to ask when critical data is first
saved, from a user gesture, not on page load.

**Failure.** MDN's guide gives several ways a write can be lost:
- When the browser quits, or the disk goes away, every open transaction aborts.
- A transaction may never complete, even on a normal shutdown.
- No write should depend on an unload event.

Its advice is to leave the database consistent at the end of every transaction. Data can also be
changed or deleted by the user, by an extension, or by the developer tools. Web.dev's best
practices add three more points:
- Always handle a failed write, since private browsing and full disks refuse them.
- Keep the state in memory as well, so the page still works when storage does not.
- Write upgrade code for every older version a device may still hold, a buggy one included, and
  test those upgrade paths with unit tests.

**Other tabs.** A page that opens a higher version waits until every other connection to the
database closes. An open tab that never closes its connection blocks the upgrade, and the new
page gets a `blocked` event. MDN's pattern has two parts:
- Every connection closes itself on `versionchange` and asks the person to reload.
- The opening page listens for `blocked`.

**Speed.** The structured clone runs on the page's thread, and a bigger object blocks it longer.
Web.dev advises against storing a large nested object as one record and rewriting it after every
change. Instead, split the state into records and write only the ones that changed, and read and
write no more than the data you need. Chrome made `relaxed` the default durability for read-write
transactions in Chrome 121, as Firefox and Safari already did. The Chrome team measured writes 3
to 30 times faster in real examples, and reads speed up too, since reads wait behind writes. The
same article keeps `strict` for a migration, where the old data must not be dropped before the
new data is safely written.

## Data-oriented design

Data-oriented design is laying out data for the way the program reads and writes it, rather
than for the way the problem is described. It is usually traced to Noel Llopis's 2009 article and
was written up at length by Richard Fabian in *Data-Oriented Design* (2018). The
best-known example is the **struct of arrays**: each field in its own array, instead of an
**array of structs**, where each row is one object holding all its fields.

Which layout is right depends on the reader:

- **A struct of arrays** pays off when a loop reads one or two fields of many rows: a filter on
  the level of every monster, a sort by name. The loop touches only the bytes it needs, in order,
  and the processor's cache and prefetcher do the rest.
- **An array of structs** pays off when the code uses most fields of a few rows: one actor in one
  attack, one character on one sheet. Splitting those fields apart only spreads one row over many
  places.

The data-oriented question is never "structs or arrays?" but "which data does this step read,
and how often?". Fabian's book has a section on why database technology matters (2.8). Its point is that
normalized tables, each keyed and holding facts about its key alone, are already a data-oriented
layout ([normal-forms.md](normal-forms.md)).

### What that means for the Go core and the ruleset

The core runs one command on one actor at a time. The whole ruleset is about 1,350 rows, and
`data.Load` reads all of it in about 10 ms (`BenchmarkLoad`, QA_S24_02). Each table is a
`Table[Row]`: a slice of row structs and a map from key to index (`backend/internal/data`).
That is already contiguous, and a command reads most fields of the few rows it touches, so an
array of structs is the right layout there. What costs time per call is decoding and encoding JSON,
the WebAssembly boundary, and `map[string]any` rows in the catalog, not the struct layout.

The one place the core scans many rows for one field is a list: the catalog and the bestiary,
filtered and sorted. There a column index pays: per filterable field, a typed slice built once at
load, plus the row order per sort key, with whole rows built only for the page shown. S24's rule
holds: measure first (`go-benchmarks.yml`), and keep a change only if it is no slower.

### What that means for IndexedDB

IndexedDB holds objects, not columns, but its parts map onto a data-oriented layout:

| Data-oriented idea | In IndexedDB |
| --- | --- |
| Split a record by who reads which part | Separate stores: a small summary record for a list, the full record only when one is opened |
| A column sorted for a filter or a sort | An index on that field, or a compound index for two (`[race, level]`) |
| Append, don't rewrite | One record per event, keyed `[owner, sequence]`, read back by a key range |
| Batch the work | One transaction per user action, over every store it touches |
| Read only what the step needs | `getAll(range, count)` or a cursor on an index, not `getAll()` on the whole store |

## How Riftweave should use it

### Where we are [V 2026-11-25]

- **Two databases, each at version 1:**
  - `riftweave` holds the store `saves`, keyed by `id` (`frontend/src/app/saves.ts`).
  - `riftweave-characters` holds the store `characters`, keyed by `character.id`
    (`frontend/src/app/characters.ts`).
- **No indexes.** Each upgrade handler only creates its store, and reads are `getAll()` on the
  whole store.
- **Autosave rewrites the whole game.** It `put`s the whole save record (the session's start, its
  full log, and its actor) after every step. A game of n steps writes about n²/2 steps in total,
  and every write clones the whole log on the page's thread.
- **No connection closes on `versionchange`**, and no open handles `blocked`, so the first upgrade
  to version 2 can hang while an old tab is open.
- **Persistent storage is never asked for**, so saved characters are best-effort, and Safari's
  seven-day rule applies.

### The stores, and what, for what, and why

A store's version comes from an upgrade that is a `switch` on the old version, one case per step
(0 to 1, 1 to 2, and so on), each with its own test (story S27).

| Store | What | For what | Why |
| --- | --- | --- | --- |
| `characters` | The full stored actor with its sync fields, keyed by id | The sheet, export, play, a session's start | The record of truth. It changes only when the character changes, so one write per change is cheap. |
| `character_summaries` | Id, name, lowercased name, sex, race, background, level, revision, `changed_at`. Indexes on `name_lower`, `race`, `level`, `changed_at`, and `[race, level]` | The roster and its filters and sorts (FE_S25_06) | The list reads small records through an index instead of cloning every actor. It is written in the same transaction as `characters`, so the two never disagree. |
| `profile` | One record: the current character's id and when it was chosen (BE_S25_04) | Every play page's start (FE_S25_07) | A pointer, not a copy, so editing the character changes no second record. It is per device, as the id means nothing on another one. |
| `drafts` | One record per creation draft, keyed by draft id | The step-by-step flow (FE_S25_05) surviving a reload | One small write per step. A draft is not a character until finished, so it stays out of the roster. |
| `saves` | Per game: id, the session's start, the step count, the current actor, the character id, `changed_at`. Index on `changed_at` | The saved-games list, and loading a game | The list reads only this. The log is not in it, so a step does not rewrite it. |
| `steps` | One record per session step, keyed `[save_id, index]` | Autosave, replay, export | Append only: a step is one small write. Replay reads the key range `[save_id, 0]` to `[save_id, n]` in order. Export joins the steps back into today's session file, so the file format does not change. |

### The rules, with their reasons

1. **One user action, one transaction**, over every store it changes. A finished draft writes
   `characters` and `character_summaries`, deletes the draft, and sets `profile` in one
   transaction. It is atomic (the spec), and leaves the database consistent if the tab dies (MDN).
2. **No `await` on anything but the transaction's own requests inside it.** Read and prepare
   first, then open the transaction and write. It commits on its own (the spec).
3. **`relaxed` for autosave steps and drafts, `strict` for an upgrade that moves data.** Speed and
   battery for frequent writes, safety where a half-done migration would lose records (Chrome).
4. **Small records, split by reader.** Lists read summaries; a step writes a step. The structured
   clone blocks the page in proportion to the object (web.dev).
5. **Indexes only for filters and sorts a page really has.** Each index is updated on every write
   (the spec).
6. **Every read is a key range or an index query**, with a `count` when a page shows a page.
7. **Close on `versionchange`, handle `blocked`**, and say "reload to update" instead of hanging
   (MDN).
8. **Upgrades are tested from every old version** against a fake IndexedDB in Vitest (web.dev).
9. **Handle every failed write** (`QuotaExceededError`, private browsing) and keep the state in
   memory, so the page still works when the write did not (web.dev, MDN).
10. **Ask for persistent storage when the first character is saved**, from that click, once
    (web.dev). Keep export one click away, since a refused or cleared store loses everything at
    once (MDN).

## Sources

- W3C, *Indexed Database API 3.0*, Editor's Draft: `w3c/IndexedDB` at `f9f83f196f85`, `index.bs`.
  It covers the sections on the database, object store, keys, index, transactions (lifecycle,
  durability, upgrade transactions), `getAll`, `getAllRecords`, and `commit()`.
- MDN Web Docs, `mdn/content` at `eeebbb874aeb`:
  - `files/en-us/web/api/indexeddb_api/using_indexeddb/index.md`, the sections "Version changes
    while a web app is open in another tab" and "Warning about browser shutdown".
  - `files/en-us/web/api/storage_api/storage_quotas_and_eviction_criteria/index.md`.
  - `files/en-us/web/api/idbobjectstore/getall/index.md`.
- Philip Walton, "Best practices for persisting application state with IndexedDB", web.dev:
  https://web.dev/articles/indexeddb-best-practices-app-state
- "Persistent storage", web.dev: https://web.dev/articles/persistent-storage
- Thomas Steiner, "A change to the default durability mode in IndexedDB", Chrome for Developers,
  updated 2023-11-03: https://developer.chrome.com/blog/indexeddb-durability-mode-now-defaults-to-relaxed
- Richard Fabian, *Data-Oriented Design*, 2018: https://dataorienteddesign.com/dodbook.pdf (section 2.8, "Why does database technology matter?")
- This repo: `frontend/src/app/saves.ts`, `frontend/src/app/characters.ts`,
  `backend/internal/data/data.go`, and `BenchmarkLoad` in `backend/internal/data/data_test.go`.
