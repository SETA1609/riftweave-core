### APP-10 Spell authoring in the backend and frontend, with JSON export
- Added: the `author` command (`docs/backend/authoring.md`): checks draft spells as `validate.py` would (schema, refs, channels, parameters), gives the spellcraft cost with a warning when the authored mana differs, loads a spell, and exports a core or a module data file.
- Added: the Go backend depends on `github.com/santhosh-tekuri/jsonschema/v6` v6.0.3 for the schema check; `core.Open` reads `ruleset/schemas`.
- Fixed: `validate.py` crashed on a module data file holding spells; module rows now resolve against the ref index.
