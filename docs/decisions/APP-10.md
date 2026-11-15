# APP-10 Spell authoring in the backend, with JSON export

Date: 2026-11-15
Modified: 2026-11-15

- **Decision:** one `author` command (kind `spell`, operations `check` and `load`) checks each
  draft as `validate.py` would: the schema (`spell.schema.json`), the refs (the categories
  `validate.py` indexes, built the same way), the `spell` channel, and `parameter_category`. It
  gives the spellcraft cost (`spellcraft.Cost`, now exported) and warns when an authored mana
  differs. It exports a core file or a module file with its manifest. Spec:
  `docs/backend/authoring.md`. **[V 2026-11-15, `go test ./core`: every core spell comes back
  unchanged with no problems; a broken ref, a wrong channel, a bad parameter, and a schema error
  are each found]**
- **Decision:** the schema check uses `github.com/santhosh-tekuri/jsonschema/v6` v6.0.3
  (2026-06-28, pure Go, draft-07), a new dependency, rather than a hand-written subset that would
  drift from what `validate.py` reports. It pulls `golang.org/x/text` from v0.3.8 to v0.14.0.
  The schemas are registered under the `$id` base of `schema.json`, so their relative `$ref`s
  resolve as they do for Python. **[V 2026-11-15]**
- **Decision:** `core.Open` reads `ruleset/schemas` beside the data directory
  (`core.NewWithSchemas`); `core.New` without schemas still works, and `author` then refuses.
- **Finding, fixed:** `validate.py` did not pass the ref index to module data files, so a
  module file holding spells crashed the validator (`TypeError` in `resolve_namespaced_ref`); no
  module had spells before. `validate_modules` now takes and passes it. **[V 2026-11-15,
  `test_author_export.py`]**
- **Difference from the issue:** split. The spell editor page moved to APP-22, since the whole
  issue came to about 25 files; the browser core's schemas go with it. CLI and API parity was
  checked by running one broken draft through `riftweave author` and `POST /v1/commands/author`:
  the results are the same value. **[V 2026-11-15, by hand]**
- **Difference from the steps:** the vector commit's `test_author_export.py` left the
  `ruleset/tests` folder out of its copy, so `validate.py` saw no golden vectors; the data commit
  copies it. That fixed the test, not a rule.
- **Open:** the exported rows come back with their keys sorted (Go maps), not in the draft's
  order; the values are the same. **[A]**
