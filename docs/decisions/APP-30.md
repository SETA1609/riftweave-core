# APP-30 A file made in one app opens in the other

Date: 2026-11-18

- **Decision:** one envelope, `{riftweave, rules_version, modules, data}` (`backend/contract/file.json`), for `character`, `session`, and `spell`; `data` is what the contract already returns. Spec: `docs/backend/files.md`.
- **Split by area (CONTRIBUTING rule 2):** this is the backend half (`APP-30-backend-…`): the contract, `riftweave export` and `import`, the TUI's Import, the fixtures, and the docs. The frontend half (Export on each screen, Import on the home page, the Vitest and Playwright specs) follows on the same fixtures and adds its bullet to the fragment. The issue stays in `review` until both merge.
- **The settings file keeps its format:** it already names its kind (APP-25), so an import that finds `"riftweave": "settings"` hands it to `settings import` instead of wrapping it. [A]
- **The check:** a character and a session validate against `common.json` with the JSON Schema library the core's author command already uses, compiled with `ruleset/schemas` beside it (`common.json` points into `schema.json`); a spell goes through the `author` command's check. [V 2026-11-18, `test_contract.py`; the Go tests run in CI]
- **Missing modules in the CLI:** the import stops (exit 2, the alert and the rows on stderr) unless `--enable-modules` (turns them on in the settings file, only modules this ruleset has) or `--accept-loss`. The TUI's Import takes ` --accept-loss` after the path. [A]
- **The new package** `backend/cmd/riftweave/files` holds the envelope rules, so the agent CLI and the TUI share them; it stays out of `backend/core` and `internal/`.
- **Not done here:** a character file's sync fields (APP-33, not started); a spell exported by the frontend (its spell creation screen is APP-22). [U]
- **Not run here:** the Go tests of `backend/cmd/riftweave`, which need module downloads; the sandbox has no network. CI runs them. [U]
