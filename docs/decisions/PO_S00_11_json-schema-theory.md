# PO_S00_11_json-schema-theory A theory page on JSON Schema: how it works, why, and for what

Date: 2026-11-27

- **Decision** (owner, 2026-11-27): a theory page, `docs/theory/json-schema.md`, on how JSON Schema works, why the ruleset uses it, and for what, with its sources.
- **Sources:** fetched with `devops-tools fetch-docs github` into `cache/docs/` and read there [V 2026-11-27]:
  - the draft-07 specification, `json-schema-org/json-schema-spec` at `1afc34b65ead` (tag `draft-handrews-json-schema-01`): `jsonschema-core.xml`, `jsonschema-validation.xml`, `schema.json`. The default branch no longer has these files, so the fetch names the tag's commit.
  - `python-jsonschema/jsonschema` at tag `v4.26.0` and `python-jsonschema/referencing` at tag `v0.37.0`, the versions installed in this session.
  - `json-schema-org/website` at `1a415f9d36c3`.
- **Checked against the installed library** [V 2026-11-27, `jsonschema` 4.26.0]: a keyword beside `$ref` is ignored; `additionalProperties: false` rejects a property declared only in an `allOf` branch; `format` is not asserted without a `FormatChecker`; `default` leaves the instance unchanged.
- **Checked in the schemas** [V 2026-11-27]: every schema declares draft-07; every keyword beside a `$ref` is a `description`; no schema uses `format`; only `schema.json` declares an `$id`.
- **The data file's `$schema`** is a repo convention: a relative path to the file's schema, read by `validate.py`. In the specification `$schema` belongs in a schema and names its meta-schema. The page says so, and nothing changes.
- **Open:** whether to move from draft-07 to 2020-12, which json-schema.org recommends. It would change what `$ref` beside other keywords means and rename `definitions` to `$defs`, across every schema and the command contract [U].
