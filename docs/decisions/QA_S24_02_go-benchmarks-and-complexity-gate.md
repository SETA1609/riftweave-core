# QA_S24_02_go-benchmarks-and-complexity-gate Benchmarks and a complexity gate before the Go refactor

Date: 2026-11-24

- **The gate** (`backend/tools/complexity`, standard library only): a non-test function fails over a cyclomatic complexity of 40 (1, each `if`, `for`, `range`, `case`, and `&&` or `||`; a function literal counts toward its holder) or a nesting depth of 4 (`if`, loops, `switch`, `select`). [V 2026-11-24, the tool's tests]
  - **The allow list** (`allow.txt`) holds the 10 functions over the limits on 2026-11-24: six by complexity (`attack.Run` 133, `combat.strike` 89, `casting.Run` 48, `progression.Run` 42, `Model.Update` 42 in the terminal app, `charactersCommand` 41) and four by depth alone (`catalog.layerModule` 6, `author.refIndex`, `mcp.standalone`, `catalog.resolver.base` 5). The issue named the first five; the depth limit found the rest.
  - **It only shrinks:** an allowed function fails if it grows, and fails if it is back under the limits and still listed.
  - **Where it runs:** as a Go test (`TestTheBackendPassesTheGate`), so `go test ./...` in `validate-ruleset.yml` runs it with no new step; `go run ./tools/complexity` runs it alone.
- **The benchmarks:** `BenchmarkCommand` in `core` replays every vector of `combat`, `attack_sequence`, `casting`, `actions`, and `derived`; `BenchmarkLoad` opens the ruleset (about 10 ms on the sandbox). `tools/benchdiff` compares two `go test -bench` outputs by ns/op, keeps a benchmark's fastest run, and fails a slowdown over a threshold. The `Go benchmarks` workflow (`workflow_dispatch`) runs both on a base and the branch, three times each. [V 2026-11-24, the benchdiff tests and BenchmarkLoad; the core benchmark runs in CI]
- **Not done here:** a baseline file is not committed; each comparison runs the base fresh, so the machine is the same on both sides.
