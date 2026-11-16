# APP-16 Saves in the browser: IndexedDB with JSON export and import

Date: 2026-11-16
Modified: 2026-11-16

- **Decision:** `Saves` writes each save to IndexedDB (database `riftweave`, store `saves`) when
  a session starts and after every step, and the next step runs only after the write; the app
  resumes the latest save on start. A backup file, `<name>.riftweave.json`, is the session JSON
  itself, so the agent CLI and the API replay it unchanged. Opening a file replays it and says
  whether it matches, was edited, or came from other rules. **[V 2026-11-16, `saves.spec.ts`,
  `e2e/app.spec.ts` over HTTP and WebAssembly, `go test ./core`]**
- **The rules version lives on the session.** The issue asked for "the session JSON plus the
  rules version"; nothing named a rules version yet, and the core refuses unknown fields, so the
  session gains an optional `rules_version`: `start` records the core's `RulesVersion` (`0.1.0`
  until the 1.0 release moves it, M5-09) and `replay` stamps the current one. A file from other
  rules therefore shows by comparing the two. **[A]**
- **No IndexedDB library.** The store is the plain IndexedDB API behind a small `SaveStore`
  interface; the unit specs run on an in-memory store, and the Playwright specs on the browser's
  IndexedDB. No package was added.
- **Differences from the issue:**
  - There is one saved-games screen (a new tab) rather than a list on each screen; resuming a save
    makes it the game the sheet shows.
  - A file that replays to its actor is kept as the replayed game; "Keep the saved character" is
    offered only when the replay differs, as the issue's choice exists for that case.
  - The agent CLI replays a backup as `{"operation": "replay", "session": <file>}`; a Go test checks
    that exact input rather than the CLI binary.
- **Not checked:** which browsers clear site data and when (the issue's Safari note) is still
  **[U]**; `navigator.storage.persist()` is asked once, and a refusal shows a note.
- **Risk:** a session now carries `rules_version`; an older client that sends a session without it
  still works (the field is optional), and a client that strips unknown fields loses only the
  version.
