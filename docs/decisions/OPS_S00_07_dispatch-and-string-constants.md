# OPS_S00_07_dispatch-and-string-constants Dispatch by switch or table, and named constants for meaningful strings

Date: 2026-11-21

- **Decision:** AGENTS.md has a code section: a choice on one value is a `switch` in Go and TypeScript and a dict in Python, a state machine is a small integer state read from a table, and a string with meaning to the code is a named constant (Go typed constants, Python module constants, TypeScript `as const` objects with derived types). Why lives in [docs/theory/dispatch.md](../theory/dispatch.md). [V 2026-11-21, quoted from pinned copies of the Go compiler, the Go 1.19 notes, the Python reference and FAQ, the TypeScript handbook, and V8]
- **Why:** the owner asked for it after the S19 and S07 pull requests chose branches with long `if` chains and repeated the same strings.
- **Thresholds, read from the sources:** Go emits a jump table from 8 dense cases (binary search below); V8 from 5 compact Smi literal cases. Below those a `switch` still beats a chain or a map at these sizes in Go, and costs about the same as compare-jumps in V8. [V 2026-11-21]
- **Correction:** the FE_S07_04 refactor said V8 compiles its sync `switch` to a jump table; its cases are property reads (`SYNC.Newer`) and there are 3 plus a default, so V8 compiles compare-jumps. The comment is corrected on that branch. [V 2026-11-21]
- **Open:** that V8 reads a property of an `as const` object at a fixed offset once its shape is known is [A]; it is not in the pinned sources.
