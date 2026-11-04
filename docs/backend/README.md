# Reference backend

One backend serves the ruleset to every consumer: an agent CLI, a human CLI, an HTTP API, and an Angular frontend (story APP, [07-backend.md](../stories/07-backend.md)). This file is its spec. It adds no rules: the rules are in `docs/core/` and `docs/modules/`, the numbers in `ruleset/data/`, and the vectors are the authority.

## Layers

```
backend/contract/   one file per command: name, input schema, result schema      (APP-01, data)
Go core             loads ruleset/data/, implements every command, passes every vector   (APP-02)
adapters            agent CLI, human CLI, HTTP API: parse, call the core, format         (APP-03, 05, 06)
frontend            Angular, talks only to the HTTP API through a generated client       (APP-07)
```

The contract is language-neutral, so the Go core, both CLIs, the OpenAPI document, and the frontend are all built from it, and the Python reference is checked against it today. **An adapter only parses, calls the core, and formats. A rule in an adapter is a bug.** [V 2026-11-04, owner decision]

## Commands

A command is one kernel or module function, named like its vector function (`combat`, `resolution`, `loot`, ...). `backend/contract/<name>.json` holds:

| field | meaning |
| --- | --- |
| `name` | the command and vector function name |
| `summary` | one line for `list` and the OpenAPI document |
| `doc` | the `docs/` file that owns the rule |
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

Rolls are inputs, as in the vectors: the caller rolls, or a client draws them from a seed. A backend that draws rolls itself (APP-02) echoes them in the result under `rolls`, so any call can be saved as a vector and replayed. The Python reference never rolls, so no result has `rolls` yet. [A]

## Checks

`ruleset/scripts/test_contract.py` runs with the other tests in CI:

- every contract file matches `command.schema.json`, and its `input_schema` and `result_schema` are valid JSON Schemas;
- every vector's input validates against its command's `input_schema`;
- the reference's full result for every vector validates against `result_schema` (or the error shape);
- a vector function with no contract file fails.
