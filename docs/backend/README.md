# Reference backend

One backend serves the ruleset to every consumer: an agent CLI, a human CLI, an HTTP API, and an Angular frontend (story APP, [07-backend.md](../stories/07-backend.md)). This file is its spec. It adds no rules: the rules are in `docs/core/` and `docs/modules/`, the numbers in `ruleset/data/`, and the vectors are the authority.

## Layers

```
backend/contract/   one file per command: name, input schema, result schema      (APP-01, data)
Go core             backend/: loads ruleset/data/, implements every command, passes every vector   (APP-02)
adapters            agent CLI, human CLI (a Charm TUI), HTTP API: parse, call the core, format   (APP-03, 05, 06)
                    the agent CLI is backend/cmd/riftweave (cli.md)
                    the HTTP API is backend/cmd/riftweave-server (api.md)
frontend            Angular, talks only to the HTTP API through a generated client       (APP-07, frontend.md)
```

The contract is language-neutral, so the Go core, both CLIs, the OpenAPI document, and the frontend are all built from it, and the Python reference is checked against it today. **An adapter only parses, calls the core, and formats. A rule in an adapter is a bug.** [V 2026-11-04, owner decision]

## Go core

`backend/` is one Go module. The core is standard library only; the human CLI alone uses Charm. The only door is `backend/core`:
`core.Open("ruleset/data")`, then `Run(name, json) (json, error)`. A rule error comes back as
`{"error": "..."}` and an error that wraps `core.ErrRule`; bad input wraps `ErrInput` (inputs
decode strictly, as the contract says); an unknown command wraps `ErrUnknownCommand`. An adapter maps these to an
exit code or a status and does nothing else.

```
backend/core/                one file: the command table, the JSON binding, Run
backend/internal/data/       the loader and the tables: no rules
backend/internal/actor/      attributes (a fixed array by ability) and the actor input
backend/internal/formula/    formula trees compiled to flat programs
backend/internal/modifier/   the one stacking function
backend/internal/<area>/     one package per kernel area: resolution, derived (and actions),
                             rest, progression, condition, wuxing, economy, combat, loot
```

**Modular.** An area package imports `data`, `formula`, `modifier`, `actor`, and the areas it
calls (combat and economy call `resolution`, loot calls `economy`, rest calls `derived`); it
never imports `core`, and it knows nothing of JSON plumbing: its command is a typed
`func(*data.Ruleset, In) (Out, error)`. A new command is one package and one line in
`core.commands`. A module's rules join the same way, and its data joins through the `fs.FS`
the loader reads.

**Data-oriented.** The rules never walk raw JSON. `data.Load` reads `ruleset/data/` once into
typed tables: dense slices of rows with a key index beside them (`data.Table`). Every
cross-reference is resolved when the table is built: attribute names to an index, a loot
row's table to a row index, a damage type's tags to two booleans, a Wuxing edge to a bit
mask, a loot roll to its entry through a 100-slot lookup, an effect ref to the conditions it
applies and removes. Every formula is compiled to a flat postfix program evaluated on a fixed
stack, with no allocation. A command is then array reads and integer math. Coefficients come
from the tables, never from constants, so a data change that breaks a rule fails a vector.

**Checks.** `go test ./...` in `backend/` runs every file in `ruleset/tests/vectors/` through
`core.Run` and compares each `expect` field, and checks that the command table and
`backend/contract/` list the same commands. CI runs it with `gofmt` and `go vet` on every
change under `ruleset/`, `backend/`, or `docs/backend/`. **A rule issue that adds or changes a
function updates the Python reference, the Go package, and the contract in the same pull
request.** [V 2026-11-06, run in this change: 256 of 256 vectors pass in Go]

**Human CLI.** The interactive mode (APP-05) is a terminal UI built with Charm: Bubble Tea, Bubbles, and Lip Gloss. It lives in `backend/cmd/` as an adapter and calls `core.Run` like every other adapter, so it holds no rule and the core and `internal/` never import Charm. The agent CLI must not import it either, so a script that calls the binary thousands of times pays no TUI start cost. [A: owner decision 2026-11-06; the exact libraries and whether `cmd/` needs its own `go.mod` to keep the core dependency-free are settled in APP-05]

Modules: `core.Open` loads each module beside `ruleset/data`, and `internal/wuxing` reads the
Wuxing module's data. Other modules load their manifest only. [V 2026-11-11]

Sessions: the `session` command runs a chain of commands on one stored actor ([sessions.md](sessions.md)). Printouts: the `print` command renders a character sheet, a spell card, a monster stat block, or a condition card as printable HTML from those commands' results ([print.md](print.md)). Dev mode changes a session's actor directly for testing, only with `--dev` ([dev-mode.md](dev-mode.md)).

## Commands

A command is one kernel or module function, named like its vector function (`combat`, `resolution`, `loot`, ...). `backend/contract/<name>.json` holds:

| field | meaning |
| --- | --- |
| `name` | the command and vector function name |
| `summary` | one line for `list` and the OpenAPI document |
| `documentation` | the `docs/` file that owns the rule |
| `input_schema` | JSON Schema of the input: the vector's fields without `name`, `function`, and `expect` |
| `result_schema` | JSON Schema of a successful result |

`command.schema.json` is the schema these files must match; `common.json` holds the shared pieces (attributes, skills, a cost, the error shape) and points into `ruleset/schemas/schema.json` for `modifier`, `phase`, and `namespacedRef`, so the contract and the data use the same definitions.

**Inputs are strict** (`additionalProperties: false`). A rule change that adds an input field adds it to the contract in the same pull request, or the contract test fails. Results are strict too, so a consumer can rely on every field it sees.

## Errors

Any command may return an error instead of its result:

```json
{ "error": "hoard_minor takes 2 rolls, got 1" }
```

`error` is a sentence for a person. The vectors pin the sentence today; an adapter maps an error to a non-zero exit code (CLI) or a 4xx status (API). Machine-readable error codes are not specified yet. [U]

## Rolls

Rolls are inputs, as in the vectors: the caller rolls, or a client draws them from a seed. A backend that draws rolls itself echoes them in the result under `rolls`, so any call can be saved as a vector and replayed. The Python reference never rolls and neither does the Go core (APP-02): both take every roll as input, so no result has `rolls` yet. Drawing rolls from a seed belongs to the adapters and sessions (APP-03, APP-04). [A]

## Checks

`ruleset/scripts/test_contract.py` runs with the other tests in CI:

- every contract file matches `command.schema.json`, and its `input_schema` and `result_schema` are valid JSON Schemas;
- every vector's input validates against its command's `input_schema`;
- the reference's full result for every vector validates against `result_schema` (or the error shape);
- a vector function with no contract file fails.
