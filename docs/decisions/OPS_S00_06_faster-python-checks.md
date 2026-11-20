# OPS_S00_06_faster-python-checks Faster Python checks

Date: 2026-11-20

- **Measured first** [V 2026-11-20, development dae4244, Python 3.13, 2 CPUs]:
  - `validate.py` takes about 0.5 s; the full unit test suite about 9.5 to 11 s.
  - Four test files held most of it: `test_format_json` 2.9 s, `test_contract` 1.9 s, `test_loot` 1.5 s, and `test_builds` 1.4 s.
  - In `validate.py`, most of the time is `jsonschema`'s schema walk; parsing JSON is 0.07 s.
- **Not numpy, pandas, or Cython:** the work is nested dicts, lists, and strings checked against schemas, not numeric arrays; converting to arrays would cost more than it saves, and Cython only speeds up our own code, where little of the time is. `orjson` would only touch the 0.07 s of parsing.
- **What changed, and what it bought** [V 2026-11-20, each measured before and after, every test passing]:
  - `test_loot`: 1.46 s → 0.21 s, from caching `item_value_cp`.
  - `test_contract`: 1.87 s → 0.84 s, from one validator per command.
  - `validate.py`: 0.47 s → 0.35 s, from one validator per schema.
  - The whole suite through `run_tests.py`: 9.5 s serial → 5.6 s for the unit tests plus 0.6 s for the vectors, on 2 CPUs; GitHub's runners have 4.
  - `test_format_json` keeps its coverage: it formats each file once and leaves out npm's lockfile, which the format list never held.
- **`RefResolver` is gone:** `jsonschema` deprecated it (a warning on every run) and will remove it. `referencing` does not fetch files on its own, so `validate.py` loads a `file://` schema that is not in the store from disk, as `RefResolver` did; `schema.json` is reached that way, since the store holds only `*.schema.json`. On a copy of the data with three broken rows, the old and new `validate.py` print the same `SCHEMA` lines [V 2026-11-20].
- **`scripts/test.sh` and the devops `test` command are unchanged:** they pass unittest arguments through, which the runner does not take. CI's unit-test step uses the runner, with the golden vectors kept as their own verbose step.
- **Risk:** the caches assume the data does not change while one process runs, and the parallel runner assumes no test writes into the repo. No test does today: the tests write only into temporary directories or copies, and `test_characters.write()` (which rewrites a vectors file and a doc) runs only from `__main__` [V 2026-11-20, `git grep` of `write_text` and `json.dump` in the tests]; one that did would have to clear `item_value_cp.cache_clear()` or use a new store.
