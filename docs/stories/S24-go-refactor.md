# S24 Go refactor: state machines, dispatch tables, and hash sets

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S24 Go refactor: state machines, dispatch tables, and hash sets

**Goal:** The Go backend is faster to change and to read, and no slower to run. The branched functions become state machines, pipelines, and lists of named rules; a handler picked by a name is a table lookup; a membership test that is hot is a hash set; a value with its modifiers is one line; modules come from one registry that orders them by `requires` and loads every module's rows. Behaviour does not change: every vector passes before and after, the benchmarks show each change is no slower, and a complexity gate keeps the functions small from then on.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S24_01_file-go-refactor-issues` | PO | File the Go refactor issues | — | — |
| `QA_S24_02_go-benchmarks-and-complexity-gate` | QA | Benchmarks and a complexity gate before the Go refactor | — | — |
| `BE_S24_03_modifier-applier` | BE | One helper applies a target's modifiers, with one error | — | QA_S24_02_go-benchmarks-and-complexity-gate |
| `BE_S24_04_attack-sequence-state-machine` | BE | The attack sequence as a state machine with a per-action table | — | BE_S24_03_modifier-applier |
| `BE_S24_05_combat-strike-pipeline` | BE | One combat strike as a pipeline of stages | — | BE_S24_03_modifier-applier |
| `BE_S24_06_command-dispatch-tables` | BE | Dispatch tables for commands, operations, routes, and dev steps | — | QA_S24_02_go-benchmarks-and-complexity-gate |
| `BE_S24_07_table-driven-data-loader` | BE | A table-driven ruleset loader | — | QA_S24_02_go-benchmarks-and-complexity-gate |
| `BE_S24_08_rule-check-lists` | BE | The casting and creation checks as lists of named rules | — | BE_S24_03_modifier-applier |
| `BE_S24_09_shared-adapter-helpers` | BE | One copy of the helpers the binaries share | — | — |
| `BE_S24_10_module-registry` | BE | One module registry that discovers, orders, and loads every module | — | QA_S24_02_go-benchmarks-and-complexity-gate |

**Done when:** every issue above is done; the complexity gate's allow list is empty; every vector passes in Python and Go; the benchmarks of the six hot paths are no slower than on 2026-11-24.
