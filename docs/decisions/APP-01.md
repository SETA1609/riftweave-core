# APP-01 Command contract: commands and their schemas as data

Date: 2026-11-04

- **Decision:** `backend/contract/<command>.json` holds one file per kernel or module function
  (`name`, `summary`, `doc`, `input_schema`, `result_schema`), checked by
  `command.schema.json`; `common.json` shares the error shape and points into
  `ruleset/schemas/schema.json` for `modifier`, `phase`, and `namespacedRef`. Inputs and
  results are strict. `test_contract.py` validates every vector's input and every reference
  result, and fails a vector function without a contract. Spec: `docs/backend/README.md`.
  **[V 2026-11-04]**
- **Why:** the Go core, both CLIs, the OpenAPI document, and the frontend need one
  language-neutral interface, and the Python reference is checked against it now, before Go.
- **Merge order:** the contract includes `conditions` (M1-05) and `loot` (M4-02), whose
  vectors are still on their branches; their docs are checked once their vectors exist. The
  contract test was run on the M1-05, M1-01, and M4-02 branches and passes on each, so the four
  pull requests can merge in any order. **[V 2026-11-04]**
- **Open:** errors are a sentence, not a code **[U]**. `rolls` in results is specified for a
  backend that rolls (APP-02) and appears in no result yet **[A]**. The test uses
  `jsonschema.RefResolver`, deprecated since 4.18, like `validate.py`; both move to
  `referencing` together.
