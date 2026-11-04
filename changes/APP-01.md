### APP-01 Command contract: commands and their schemas as data
- Added: `backend/contract/`: a command file with input and result schemas for each of the 12 kernel and module functions, a meta-schema, and shared definitions; `docs/backend/README.md`; `test_contract.py` checks every vector and reference result against them.
- Changed: CI also runs on changes under `backend/` and `docs/backend/`; a new kernel function needs a contract file as well as a runner and a checklist row.
