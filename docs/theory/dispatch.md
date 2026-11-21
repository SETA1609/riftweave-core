# Dispatch and named strings

Modified: 2026-11-21

Background for the code rules in [AGENTS.md](../../AGENTS.md#code-dispatch-and-string-literals):
how Go, Python, and TypeScript in V8 turn a choice between branches into machine work, which
form each one makes fastest, and why a string that means something to the code gets a name.
This page explains; the rules are in AGENTS.md.

Every quote below was read from a copy that `devops-tools fetch-docs` saved under `cache/docs/`,
pinned to the commit named with it. [V 2026-11-21]

## Three ways to pick a branch

Code that does one thing for one value and another thing for another value can be written three ways:

| Form | What runs | Cost |
| --- | --- | --- |
| A chain of tests (`if`, `else if`) | each test in turn until one holds | grows with the number of cases |
| A jump table | the value, less the smallest case, indexes an array of addresses | one bounds check and one indexed jump, whatever the count |
| A hash table (a map, a dict) | the key is hashed, then looked up | one hash and a probe; more work than a jump table, less than a long chain |

A jump table needs small whole numbers that sit close together. A hash table takes any key that
can be hashed. A chain takes anything, including conditions on several values, and that is the
one place it is the right form: guards and validation, where each test asks something different.

A **state machine** is the same choice, made again at each step: the state is a small whole
number, a table read by the state says what it is or where it goes, and a `switch` on it picks
the code. The table read is one array index.

## Go

Since Go 1.19 the compiler can emit a jump table for a `switch`. The release notes
(golang/website `0eef789c9f9a`, `_content/doc/go1.19.md`):

> The compiler now uses a jump table to implement large integer and string switch statements.
> Performance improvements for the switch statement vary but can be on the order of 20% faster.
> (`GOARCH=amd64` and `GOARCH=arm64` only)

"Large" is set in the compiler (golang/go `7566694bdde3`, `src/cmd/compile/internal/walk/switch.go`):

```go
// Try to implement the clauses with a jump table. Returns true if successful.
func (s *exprSwitch) tryJumpTable(cc []exprClause, out *ir.Nodes) bool {
	const minCases = 8   // have at least minCases cases in the switch
	const minDensity = 4 // use at least 1 out of every minDensity entries
```

With fewer cases, or cases spread too thin, the same function falls back to `binarySearch`,
which is still a handful of comparisons. A `switch` whose cases only return constants can also
become a lookup table from four cases (`tryLookupTable`, `const minCases = 4 // need enough
cases to justify a table`). A `switch {}` on conditions is none of these: it compiles as the
chain of tests it reads as.

So in Go a `switch` on the value is the fastest form at every size: a jump table from eight dense
cases, a binary search below. A `map` adds a hash and is slower for a fixed set of cases known at
compile time; it is for keys that come from data. A state table is an array indexed by a
typed integer (`type rowKind uint8`), and a named string type (`type Step string`) with
constants keeps a string's meaning in the type.

## Python

CPython has no jump table for Python code. A `match` statement tries its cases in order; the
reference (python/cpython `dc222c8d6c36`, `Doc/reference/compound_stmts.rst`):

> Guard evaluation must proceed from the first to the last case block, one at a time, skipping
> case blocks whose pattern(s) don't all succeed.

and if a case fails, "the next ``case_block`` is attempted". The design FAQ (same commit,
`Doc/faq/design.rst`) names the alternative:

> For cases where you need to choose from a very large number of possibilities, you can create a
> dictionary mapping case values to functions to call.

and says what a dict is:

> CPython's dictionaries are implemented as resizable hash tables. Compared to B-trees, this
> gives better performance for lookup (the most common operation by far) under most
> circumstances, and the implementation is simpler.

So in Python the fastest dispatch is a dict from the value to a function or a result, built
once at module level; `match` and `if`/`elif` cost one test per case passed.

## TypeScript in V8

TypeScript compiles to JavaScript, and V8 decides. Its bytecode generator (v8/v8 `31a217159f03`,
`src/interpreter/bytecode-generator.cc`) builds a jump table only for small-integer literal cases:

> This adds a jump table optimization for switch statements with Smi cases. If there are 5+
> non-duplicate Smi clauses, and they are sufficiently compact, we generate a jump table. In the
> fall-through path, we put the compare-jumps for the non-Smi cases.

A "Smi" is a small integer, and the label must be a literal: the same function stops at the
first case that is not (`// Don't consider Smi cases after a non-literal, because we need to
evaluate the non-literal.`). So `case 2:` can join a table, while `case STATE.Newer:`, a property
read, cannot, and a switch on strings is always compare-and-jump. Below five cases the
compare-jumps cost about the same as a table.

So in TypeScript a `switch` is the right form for a choice, and for a large integer choice its
cases are number literals; a fixed mapping from a string to a value is a property read on an
`as const` object, which V8 turns into a fixed offset once the object's shape is known [A].

## Named strings

A string the code compares against ("check", "diverged", "template") is a hidden enum: written
out in several places, one typo is a branch that never runs, and nothing tells the reader which
values exist. A named constant fixes both. Each language has its own form:

- **Go:** a named string type with constants, `type Step string` and `const StepRace Step =
  "race"`. The value is still the string the JSON holds, and the compiler refuses a plain
  string where a `Step` is expected.
- **Python:** module constants, `STEP_RACE = "race"`, and the tuple or dict built from them.
- **TypeScript:** an `as const` object and a type taken from it. The handbook
  (microsoft/TypeScript-Website `cf57c7ec6762`, `packages/documentation/copy/en/reference/Enums.md`):

  > In modern TypeScript, you may not need an enum when an object with `as const` could suffice

  ```ts
  const ODirection = { Up: 0, Down: 1, Left: 2, Right: 3 } as const;
  type Direction = typeof ODirection[keyof typeof ODirection];
  ```

  The values stay the plain strings or numbers the JSON holds, and the type is the union of
  them. A `const enum` is inlined at compile time, and the same page lists its pitfalls,
  starting with `isolatedModules`, which "is fundamentally incompatible with ambient const enums".

What stays text: a JSON field name that a schema or the command contract defines (a Go struct
tag, a key of the result), because the schema is its single definition already; messages; and
words shown to a person.
