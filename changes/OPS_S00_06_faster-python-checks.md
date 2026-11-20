### OPS_S00_06_faster-python-checks Faster Python checks
- Changed: `validate.py` reuses one schema validator per schema and resolves `$ref`s through a `referencing` registry instead of the deprecated `RefResolver`, with the same error lines.
- Changed: the unit tests run one process per file in parallel (`ruleset/scripts/run_tests.py`, used by CI); `test_contract` reuses one validator per command, the loot reference caches item values, and the format test formats each file once and leaves out npm's lockfile.
