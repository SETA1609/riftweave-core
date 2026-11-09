# Findings — Tooling & CI

Scope: `ruleset/scripts/` (validate.py, module_loader.py, combat_reference.py,
crafting_reference.py, test_module_system.py), `test.sh`, `Dockerfile`,
`.github/workflows/validate-ruleset.yml`.

Baseline: everything runs — `validate.py` exits 0 (23/23 files),
`test_module_system.py` passes 37/37, both reference scripts execute cleanly.

## High severity

### H1. CI never runs the test suite

`.github/workflows/validate-ruleset.yml` runs only `validate.py`.
`test_module_system.py` (37 tests) is executed only by `test.sh`, which requires
Docker and is not invoked in CI. A regression in `module_loader.py` merge/conflict
logic would merge green.

**Recommendation:** add a `python -m unittest ruleset/scripts/test_module_system.py`
step to the workflow.

### H2. The documented `channels` check does not exist

CLAUDE.md claims "the repo's referential-integrity check … confirms effect ids
resolve and respect `channels`." No script mentions channels;
`check_references()` in `validate.py` only checks id existence, never
delivery-vector legality. A spell can reference a potion-only effect and validation
passes — and three real violations exist in the data today (see data-schemas H3).

**Recommendation:** implement the channels check in `check_references` (and fix the
data), or correct the docs.

## Medium severity

### M1. Latent shared-cache mutation in `_resolve_monster_bases`

`validate.py:363` does `merged = dict(base_data)` (shallow copy). If a base template
lacked `growth.attributes`, `merged["attributes"]` would alias the cached base dict,
and the `merged["attributes"].update(v)` at line 387 would mutate `_BASE_CACHE`,
contaminating every later monster on that base. Currently masked because all 6 bases
define `growth.attributes`.

**Recommendation:** `copy.deepcopy(base_data)` or copy the attributes dict before
updating.

### M2. Summary counts can disagree with reality

Two paths increment `failed` without incrementing `total`: `validate.py:551-555`
(schema-not-in-store) and `validate.py:493-504` (manifest failure counts each data
file as failed but never adds to `total`). The final line can read e.g.
"22 file(s): 22 passed, 3 failed".

**Recommendation:** count every failed file in `total`.

### M3. Conflict detection is dead code in validation

`validate_modules()` (`validate.py:463-531`) initializes `conflicts = []` and
returns it untouched; `merge_data()`'s tested conflict detection is never called
from `validate.py`. Duplicate ids between core and a module — or within a single
core collection (see data-schemas H1!) — pass validation silently.

**Recommendation:** build merged data via `merge_data` during validation and fail on
conflicts; add an intra-collection duplicate-id check.

### M4. KeyError bug in `crafting_reference.py:159`

`mag_mult = item["effect_magnitude_mult"]` inside the enchantment loop, while line
142 treats the same field as optional (`item.get("effect_magnitude_mult", 1.0)`).
Any crafted item with enchantments but without that field crashes. Masked because
all 6 current items define it.

**Recommendation:** reuse the already-computed `eff_mag_mult`.

### M5. Deprecated `RefResolver` + ineffective warning suppression

`validate.py:14` imports `RefResolver` (deprecated, slated for removal from
jsonschema), and the `warnings.catch_warnings()` at lines 585-587 wraps only
`main()` — the DeprecationWarning fires at *import* time and prints on every run
(verified).

**Recommendation:** migrate to the `referencing` library; the suppression block then
goes away.

### M6. No single-file validation mode

`validate.py` takes no CLI args; iterating on one data file means revalidating
everything. (CLAUDE.md documents the gap.)

**Recommendation:** accept optional file-path arguments filtering
`collect_data_files()`.

### M7. String-key references are entirely unchecked

`check_references` only handles integer ids. `crafted_items.json` references
`material_key`, `gem_key`, `base_key`, and weapons reference skill keys — none
validated; both reference scripts would crash at runtime on a typo instead. Two
dangling key refs exist today (data-schemas H2).

**Recommendation:** extend the integrity check to key-based cross-references.

### M8. Reverse-cycle Wuxing math is duplicated and fragile

`get_interaction`'s `1.0 / mult` (`combat_reference.py:68`,
`crafting_reference.py:67`) divides by zero if a cycle multiplier is ever 0, and the
code's own TODO admits the semantics are wrong for the `insulting` cycle
(reverse-insulting currently yields ×0.8).

**Recommendation:** extract to one shared module and handle insulting explicitly.

## Low severity

- **L1. Duplication:** `find_cycle`/`get_interaction` are copy-pasted between the two
  reference scripts; a ~30-line Wuxing-message block appears 4 times
  (`combat_reference.py:625-651, 688-714`; `crafting_reference.py:162-193, 207-234`);
  `build_id_index` (`validate.py:60-93`) duplicates its core/module loops verbatim.
- **L2. Unused imports:** `sys` in both reference scripts; `os` in
  `module_loader.py` and `test_module_system.py`.
- **L3. Hardcoded magic numbers in `combat_reference.py`:** effect id 20 (`cure`) at
  line 357; stack cap `6` at line 381 (should read the condition's `maxStacks`);
  demo effect ids at lines 791-851 will silently misbehave if data ids shift.
- **L4.** `apply_condition`'s `parameter` argument (`combat_reference.py:250`) is
  accepted and never used.
- **L5. Tests leak temp dirs:** every `tempfile.mkdtemp()` in
  `test_module_system.py` lacks cleanup — use `self.addCleanup(shutil.rmtree, …)` or
  `TemporaryDirectory`.
- **L6.** `.ruff_cache/` isn't in the project `.gitignore` (it's invisible only
  because ruff writes a `.gitignore` inside its own cache dir). Nothing spurious is
  actually tracked — `git ls-files` is clean.
- **L7. CI paths-filter blind spots:** the workflow triggers only on `ruleset/**`,
  so changes to the workflow itself, `Dockerfile`, or `test.sh` never run CI.
- **L8.** `validate.py:369`: `overrides.get("level") or base_data.get("level", 1)` —
  a legitimate `level: 0` override is ignored (falsy). Use an `is None` check.
- **L9.** `_resolve_monster_bases` runs before the `$schema` check (line 419 vs
  423) — a schema-less file with a bad base fails instead of being
  skipped-with-warning, inconsistent with the documented skip behavior.
- **L10.** CLAUDE.md documents string snake_case ids; the tooling's int-based
  `build_id_index` reflects the real (numeric) convention — the doc is wrong, not
  the code.

## CI / test.sh / Dockerfile alignment

- `test.sh` ⊃ CI: test.sh runs validation *and* unit tests (in Docker); CI runs only
  validation natively. Environments also differ (Docker `python:3.12-slim` vs `3.x`
  on ubuntu-latest) — minor; the test gap (H1) is the real issue.
- The Dockerfile works as documented: `docker run --rm riftweave-validate` matches
  the ENTRYPOINT; `COPY ruleset/` includes `modules/` so module validation and (via
  test.sh's `--entrypoint` override) the tests work inside the image.

## Strengths

- `validate.py`'s module-aware pipeline is genuinely good: shared schema store,
  per-file `$schema` resolution, base+override monster merging with `$extend` and
  growth scaling, and a real (if partial) referential-integrity layer.
- `test_module_system.py` is a thorough, well-organized suite — negative cases,
  semver/snake_case manifest enforcement, non-mutation assertions on `merge_data`,
  integration tests against the real example module.
- `module_loader.py` has clean separation (discover / load / merge), precise error
  types, and `merge_data` correctly returns a new dict with first-source-wins
  conflict reporting.
- `combat_reference.py` builds condition combat modifiers from data
  (`_build_combat_modifiers`) rather than hardcoding them, with honest TODOs marking
  the seams.
- Ops hygiene: non-root Docker user, layer-cached deps, `set -euo pipefail` in
  test.sh with correct aggregate exit codes, workflow requests only
  `contents: read`.
