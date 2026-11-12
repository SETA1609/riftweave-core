# Agent CLI

`riftweave` is one static Go binary for agents and scripts (APP-03). It only parses, calls `core.Run`, and formats ([README.md](README.md)): no rule lives here. No prompts, no colour, and one small start, so an agent can call it thousands of times in a loop. The human CLI (APP-05) is a separate mode and the agent path imports none of it.

Build it from `backend/`: `go build ./cmd/riftweave`.

## Commands

```
riftweave list                      every command, with its summary
riftweave describe <command>        the command's contract file (backend/contract/<command>.json)
riftweave <command> [file]          run a command on JSON from the file, or from stdin
riftweave help                      usage, on stdout
riftweave --dev dev <operation> [file]   change a session's actor for testing ([dev-mode.md](dev-mode.md))
```

- **Input** is one JSON object, the command's `input_schema`: a vector without `name`, `function`, and `expect`. `-` as the file means stdin.
- **Output** is the result, one line of JSON on stdout, then a newline. Nothing else is written to stdout.
- `--dev` turns on dev mode for this call; without it the `dev` command is refused. `--dev` and `--data` come before the command, in any order.
- `list` prints `[{"name": ..., "summary": ...}]`, sorted by name; `describe` prints the contract file as it is. Both read the contract built into the binary, so they need no ruleset.

## The ruleset

A command reads `ruleset/data/` once per call. The directory is, in order: the `--data <dir>` flag (before the command), the `RIFTWEAVE_DATA` environment variable, or the first `ruleset/data` found from the current directory upward. [A]

## Errors and exit codes

An error is one line of JSON on stderr, `{"error": "..."}`, the sentence the core gives; stdout stays empty.

| Exit | Meaning |
| --- | --- |
| 0 | the result is on stdout |
| 1 | a rule error: the command ran and refused (the wrong number of rolls, an unknown key). The `error` is the one a vector pins. |
| 2 | bad input or usage: not JSON, an unknown field, more than one value, an unknown command, a missing argument |
| 3 | the ruleset could not be read or did not load |

[V 2026-11-08, `go test ./cmd/riftweave`]

## Rolls

Rolls are inputs, as everywhere in the core. The CLI does not draw rolls, and there is no seed: a logged input holds every roll. In dev mode the human CLI can queue its next roll prompts ([dev-mode.md](dev-mode.md#rolls-and-the-seed)).

## Checks

`go test ./cmd/riftweave` runs every vector in `ruleset/tests/vectors/` through the CLI, as a process would see it (arguments, stdin, stdout, stderr, exit code), and compares the JSON with `core.Run`; it checks `list` against the contract, `describe`, and that bad input exits 2 with a JSON error. [V 2026-11-08]

## Human mode

`riftweave play` is the interactive mode of the same binary (APP-05): a terminal UI built with Bubble Tea, Bubbles, and Lip Gloss. It holds no rule. It prompts for inputs, builds the same JSON the agent CLI reads, calls `core.Run`, and shows the result as a table with a short log of the rolls.

- **Create a character:** eight attributes and a level, run through `derived`.
- **Resolve an attack:** the rolls and the two sides, run through `combat`. The LCK defaults to the character's.

Rolls stay inputs: you type each one, so a result matches the agent CLI for the same inputs and rolls. With `--dev`, a dev panel and short commands change a session's actor ([dev-mode.md](dev-mode.md#the-human-cli)). It needs a terminal. Charm is imported only by `backend/cmd/riftweave/play`, and `backend/core` and `backend/internal/` stay free of it.
