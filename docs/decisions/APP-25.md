# APP-25 One settings file that both apps read, export, and import

Date: 2026-11-17

- **Decision:** the file is `backend/contract/settings.schema.json` (named `.schema.json`, not `settings.json`, so it is not read as a command file; `test_contract.py` lists it beside `command.schema.json` and `common.json`). The keys are the issue's: `riftweave`, `version` (1), `locale`, `theme`, `modules`, `network`, `bestiary`.
- **One layout, so "exports back unchanged" is byte for byte** [A]: keys sorted, two-space indent, final newline, as Go's encoder writes a map; the frontend sorts keys the same way. The Go rules live in `backend/internal/settings` (used by `riftweave settings`), the frontend's in `settings-file.ts`; the messages match.
- **Unknown keys are kept** in both: the CLI writes the imported file whole; the frontend keeps it in `localStorage` (`riftweave.settings`) and exports it with the current theme and language on top.
- **The CLI's directory** is `os.UserConfigDir()/riftweave/`, or `RIFTWEAVE_CONFIG_DIR` when set (which the tests use).
- **Proof** [V 2026-11-17]: the shared fixture validates against the schema (`test_contract.py`); a Go test and a Vitest test read it with the same values and write it back unchanged, its unknown key included; the CLI imports it and exports it back unchanged; an unknown version, a bad theme, `core` in `modules`, and a bad network mode are refused with a message and change nothing. `go test ./...` and the frontend's 66 Vitest tests pass (the API spec needs `riftweave-server` on port 8080).
- **Not done, deliberately:** `core.Open` with a module selection. Both apps load every module today, and switching the CLI to "only the enabled ones" with an empty default would drop the modules it loads now. The settings screens (APP-26, APP-27) apply `modules` when they exist; until then the field is stored and traded only. The issue's [U] on whether `core.Open` can load a subset stays open. The export and import buttons are APP-27's screen.
- **The module-loss alert** belongs to APP-30.
