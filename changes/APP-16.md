### APP-16 Saves in the browser: IndexedDB with JSON export and import
- Added: the app saves every session step to IndexedDB before the next runs and resumes the latest save; a saved-games screen lists saves and marks one with no recent backup.
- Added: "Download backup" writes the session JSON as `<name>.riftweave.json`; "Open file" or a drop replays a file and flags an edited one or one from other rules.
- Added: a session's `rules_version`, recorded at start and stamped by a replay (`backend/core/session.go`).
- Changed: the app's rolls come from `crypto.getRandomValues`.
