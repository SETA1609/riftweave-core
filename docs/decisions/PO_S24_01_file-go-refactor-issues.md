# PO_S24_01_file-go-refactor-issues File the Go refactor issues

Date: 2026-11-24

- **Decision** (owner, 2026-11-24): refactor the Go backend for speed, size, and readability with state machines, dispatch tables, and hash sets, as story S24.
- **What the screening found** [V 2026-11-24, a `go/ast` scan of the 387 non-test functions outside the terminal app]:
  - **Little copied code:** 9 small repeated groups, mostly in `data.Load` and the binaries (`rulesetDirectory`, `retarget`, `moduleRows`).
  - **Five branched functions:** `attack.Run` (complexity 133, 383 lines, 61 `if`s, 13 action checks), `combat.strike` (89, 255), `casting.Run` (48), `progression.Run` (42), `charactersCommand` (41). 324 of the 387 functions are at complexity 10 or under.
  - **A repeated call:** `modifier.Value(…, modifier.Target(…), …)` and its error check, 29 times in 12 files.
  - **Lookups:** ruleset tables are already hash maps (`Table.indexByKey`); flags, tags, and properties are lists read with `slices.Contains` 79 times.
  - **No module registry** (asked 2026-11-24): modules are discovered in four places and a manifest parsed in five; `requires`, `version`, and `extends` are not read in Go; only Wuxing's data reaches the ruleset, through a hard-coded id check. `BE_S24_10` adds the registry.
- **Order:** benchmarks and the complexity gate first (`QA_S24_02`), so every later PR shows it is no slower and smaller; then the modifier helper (`BE_S24_03`), which the attack, combat, and check refactors use. The dispatch tables, the loader, and the shared helpers need only the gate.
- **Speed, honestly:** the commands already run in milliseconds and spend most of their time in JSON. The refactor's main gain is readability and size; "no slower", measured, is the bar for speed, and hash sets go only where a benchmark shows a hot membership test. [A]
- **Not here:** the terminal app (`play/`, 2,672 lines) goes with S23; a new language (Zig was asked about, 2026-11-24) is not part of this story.
