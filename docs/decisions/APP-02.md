# APP-02 Go core: the kernel in Go, every vector in CI

Date: 2026-11-06

- **Decision:** the Go core is `backend/`, one module with no dependencies. `backend/core` is
  the only door for adapters; each kernel area is a package under `backend/internal/` with a
  typed `func(*data.Ruleset, In) (Out, error)`, bound to JSON in one table. The ruleset is
  loaded once into flat tables (`internal/data`) with every cross-reference resolved, and
  formulas are compiled to flat programs (`internal/formula`). Layout and rules for a new
  command: `docs/backend/README.md`. **[V 2026-11-06]**
- **Why:** modular so the agent CLI, human CLI, API, and later modules add packages without
  touching the others; data-oriented so a command is array reads and integer math and a CLI
  can be called thousands of times in a loop.
- **Checked:** all 256 vectors in `ruleset/tests/vectors/` pass in Go, the same count Python
  runs. Changing the crit multiplier in `combat/pipeline.json` fails 4 vectors in Go (data
  change restored). The command list equals the contract's. **[V 2026-11-06, run in this
  change, Go 1.27.3]**
- **Differences from the Python reference, intended:** a bad key (unknown table, action,
  condition, merchant, damage type, delivery, or reaction kind) is a rule error in Go where
  Python raises; a delivery that is not defined is refused before the attack roll, not after
  a hit; the loader refuses data with a loot row that names an unknown table or a table whose
  `crit_window_per_two_luck`, `per_point_divisor`, or `margin_divisor` is not positive. A
  command result is checked by the vectors only on the fields a vector names, as in Python.
- **Open:** the Go core does not validate input against the contract's JSON Schemas; it
  decodes strictly into typed structs, so an unknown field or a wrong type is refused, but a
  missing required field reads as zero in places (a modifier without `source`) **[U]**. A
  schema check at the adapter belongs to APP-03. The core draws no rolls, so no result has
  `rolls` **[A]**. Module data (`ruleset/modules/`) is not loaded; the Python kernel does not
  load it either **[U]**. Modifier values are `float64` and results are floored to integers, as
  the contract's integer result schemas require **[A]**.
- **Risk:** two implementations of each rule now change together; the vectors are the only
  check that they agree, so a rule with a thin vector can drift. Rule issues must add vectors
  before changing either side (CONTRIBUTING step 3).
- **Human CLI toolkit:** the human CLI (APP-05) will be a Charm terminal UI (Bubble Tea, Bubbles,
  Lip Gloss), kept in `backend/cmd/` behind `core.Run`, never in `core` or `internal/`. The Go
  core stays standard-library only; if Charm's dependencies would enter the core's `go.mod`,
  APP-05 puts `cmd/` in its own module. **[A, owner decision 2026-11-06]**
