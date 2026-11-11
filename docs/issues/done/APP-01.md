#### APP-01 Command contract: commands and their schemas as data
- Status: done
- Merged: #74, 2026-11-04
- Priority: medium
- Needs: M1-02
- Do: Define the backend's interface without choosing a language, so Go, both CLIs, the OpenAPI document, and the frontend are all built from one source. One file per command in `backend/contract/` (name, summary, input schema, result schema, error shape), one per reference runner (`resolution`, `combat`, `derived`, `progression`, `actions`, `formula`, `modifiers`, and each runner that lands later). Rolls are inputs, or are drawn from a seed and echoed in the result as `rolls`, so any call can be saved as a vector. A Python test checks every vector's input and the reference's result against the command's schemas. Write `docs/backend/README.md`: the layers (contract, Go core, adapters, frontend), the error shape, and that no adapter holds a rule.
- Keys: `backend/contract/<command>.json`; command `name`, `input_schema`, `result_schema`; result field `rolls`
- Files: ~12 (backend/contract/, docs/backend/README.md, ruleset/scripts/test_contract.py)
- See: ruleset/scripts/kernel_reference.py, docs/core/vectors.md
- Done when: every runner has a contract file; every vector input and every reference result validate against its command's schemas; a vector function without a contract file fails the test.
