#### APP-30 A file made in one app opens in the other
- Status: todo
- Priority: medium
- Needs: APP-16, APP-26
- Modified: 2026-11-16
- Do: Owner request, 2026-11-16: anything a person makes in the frontend can be exported as JSON and imported into the human CLI, and the other way round. One envelope for every file: `{"riftweave": "<kind>", "rules_version", "data"}` with kind `character` (a stored actor, `stored_actor.json`), `spell` (APP-10's author export), `session` (APP-16's `.riftweave.json`), and `settings` (APP-25) [A]; `data` is exactly what the command contract already returns, so no format is new. A character file also carries its sync fields (APP-33). The frontend gets "Export" on each screen that makes one and one "Import" on the home page that reads the kind; the CLI gets `riftweave export <kind> ...` / `riftweave import <file>` and an Import entry in the main menu. A file whose `rules_version` differs warns and still opens when the contract validates it [A]; a file that does not validate is refused with the schema's messages. A module item in a file needs that module enabled, or the import says which one is missing.
- Keys: the file envelope (`backend/contract/file.json`); `riftweave export|import`
- Files: ~12 (backend/contract/file.json, backend/cmd/riftweave/files.go, files_test.go, backend/cmd/riftweave/play/import.go, frontend/src/app/files.ts, frontend/src/app/home/, app.spec.ts, e2e/files.spec.ts, ruleset/tests/fixtures/files/, docs/backend/files.md, changes/APP-30.md)
- See: docs/issues/open/APP-16.md, docs/issues/open/APP-25.md, backend/contract/stored_actor.json, backend/contract/author.json
- Done when: a fixture of each kind imports in both apps; a character and a spell exported by the frontend import into the CLI and export back byte-identical after formatting; a missing module and an invalid file are reported; the Go, Vitest, and Playwright tests pass.
