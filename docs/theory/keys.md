# Keys, hashing, and interning

Modified: 2026-11-15

Background for [docs/core/keys.md](../core/keys.md): why the data names every row with a
string, what hashing and interning are, and how a program should turn those strings into
something fast. This page explains; it sets no rule. The rule stays in `keys.md`: a reference
is a namespaced string such as `core:weapon/shortsword`.

## The question

Data-oriented design (DOD) advises against strings as ids in a running program, for good
reasons:

- **Size.** A Go string is 16 bytes on a 64-bit machine (a pointer and a length), and its
  characters live elsewhere; `core:weapon/shortsword` adds 22 more. An `int32` index is 4
  bytes. **[V 2026-11-15, `unsafe.Sizeof` on go1.26.1 linux/amd64]**
- **Speed.** Two integers compare in one instruction. Two strings compare their lengths, then
  their bytes. A map lookup by string hashes every byte of the key first. **[A]**
- **Cache.** An array of indices packs many to a cache line. Strings are pointers to memory
  scattered around the heap, so following them misses the cache. **[A]**

All three are about **memory in a running program**. None is about the file a person edits.
Riftweave keeps the two apart: strings in the data, numbers in memory.

## Why the data uses strings

| Need | String key | Numeric id |
| --- | --- | --- |
| A reviewer reads a diff | `"base": "core:weapon/shortsword"` says what it is | `"base": 412` needs a lookup |
| A row is added, removed, or reordered | Every other key is unchanged | Numbers shift, or gaps and a counter must be managed by hand |
| Two modules add rows | `module:a/weapon/x` and `module:b/weapon/x` cannot collide | Two modules pick the same number |
| A module replaces a core row | It names the key it overrides (`docs/core/overrides.md`) | It must know the core's number, which a core update can move |
| A save from an older version loads | The key still names the row | The number may now name another row |
| The validator checks a reference | It knows the category from `weapon/` and reports the wrong kind | A bare number has no category |
| An engine in another language reads it | Any JSON reader works | Same, but the meaning lives in a separate table |

The repo tried numeric ids and took them out. M5-29 and M5-30 removed every numeric row `id`,
following the data-oriented review (`docs/archive/review/findings-data-oriented-design.md`).
The one numeric field left is a tier's `id`. It is a rank (1 common, 3 rare), not a row
handle. **[V 2026-11-15, `docs/decisions/M5-30.md`]**

The data is small: about 1,000 rows with a `key` and about 1,000 `core:` references in
`ruleset/data/`. **[V 2026-11-15, grep count]** Storing strings costs nothing measurable
there. The cost DOD warns about comes from looking a string up in a loop that runs every frame.

## Interning

**What it is.** Interning maps each distinct string to one small number once, and uses the
number from then on. At load, a program reads every row of a category into an array and
builds a table from key to array index. Every reference in the data is then resolved through
that table, once, and stored as the index.

```text
load:   "core:weapon/shortsword"  --table-->  17
run:    weapons[17].damage            (array access, no string, no hash)
```

**What it is for.** Dense ids: 0, 1, 2, ... with no gaps, one range per category. They index
arrays directly, so a column of data (every weapon's damage, every actor's HP) can sit in one
contiguous array, which is what DOD's structure of arrays needs.

| Pros | Cons |
| --- | --- |
| No collisions: each key gets its own index | Needs the table, built at load |
| Dense, so it indexes arrays and bitsets | An index is only valid for one load of one ruleset with one set of modules |
| 4 bytes or fewer per reference | Turning an index back into a name needs the table (keep `rows[i].Key`) |
| A bad reference fails once, at load, with the key in the message | Adding a row at run time means growing the table |

## Hashing

**What it is.** A hash function turns a string into a fixed-size number (32 or 64 bits)
computed from its bytes, with no table. `fnv1a("core:weapon/shortsword")` gives the same
number in every program, on every run. **[A]**

**What it is for.** A number you can compute anywhere without loading anything:

- **Constants in code.** Some engines hash a string literal when the code compiles
  (`"fire_bolt"_id` in C++), so the code can switch on a key without a lookup. **[A]**
- **Map buckets.** A hash map hashes the key to pick a bucket. Go's `map[string]T` does this
  on every lookup. **[A]**
- **Network messages** between two programs that cannot share a table.

| Pros | Cons |
| --- | --- |
| No table, and the same number everywhere | Two keys can collide. Rare at 64 bits, likely enough at 32 bits to need a check |
| Can be computed when code compiles | Not dense: a hash cannot index an array, so you still need a map |
| Stable across loads, unlike an interned index | Turning a hash back into a name needs a table anyway |
| | A collision changes meaning silently unless the loader refuses it |

A hash is safe only if the loader hashes every key, finds two keys with the same number, and
fails. A 32-bit hash of 1,000 keys collides with a probability of about 1 in 8,600; a 64-bit
hash of the same keys, about 1 in 37 trillion. **[V 2026-11-15, birthday bound n²/2ᵇ⁺¹]**

## Interning or hashing?

| Situation | Use |
| --- | --- |
| A ruleset is loaded and then used in a loop | Interning |
| Data laid out as arrays per category (DOD) | Interning, because it needs dense indices |
| A key named in code (`fire_bolt` in a test or a special case) | Intern it at start-up into a variable, or hash it with a collision check |
| Two processes exchange a reference with no shared table | The string key, or a hash both sides check |
| A save file, the API, the command contract, a module | The string key, always |

For this ruleset, interning is the default. Hashing is a tool for an engine that needs
compile-time constants. It adds the collision check to the loader and gives up dense indices.

## How code should handle keys

1. **Strings at every edge.** Data files, saves, the command contract, the HTTP API, the MCP
   server, and modules carry `core:<category>/<key>`. That is the stable name.
2. **Resolve once, at load.** Build one table per category (key to index), resolve every
   reference in the data through it, and fail on a key that does not resolve or names the
   wrong category. After load, no code looks a string up in a hot path.
3. **Indices in memory, per category.** A weapon index and a spell index are different types
   where the language allows it (`type WeaponIndex int32` in Go), so one cannot be passed for
   the other.
4. **Never write an index out.** A save stores keys. Loading a save resolves them again, so
   rows added by a newer version or another module set cannot shift what an old save means.
5. **Keep the name next to the index.** Errors, logs, and debug output print `rows[i].Key`,
   not `17`.
6. **Hash only with a collision check.** If an engine hashes keys, its loader hashes all of
   them and refuses two that collide.
7. **Measure before you optimise.** Swap a string for an index in a hot path only after a
   profile shows the lookup costs something.

## What the Go backend does

The Go reference backend already follows this pattern for its tables. `data.Table` holds
`rows []Row` and `indexByKey map[string]int32`, built once per collection when the ruleset
loads. `Get(key)` looks a key up, and `Index` and `At` let a caller keep the index. The
condition tables store `AppliedBy` and `RemovedBy` as `[]int32` row indices.
**[V 2026-11-15, `backend/internal/data/data.go`]**

Commands still look keys up by string at the start of each request, which is fine. One
command runs over a few hundred rows, and the backend's job is to be correct and readable,
not to run a frame loop. **[A]** An engine that runs combat every frame for many actors
would add the next step itself: resolve the actor's equipment, perks, and conditions to
indices when it loads the save, then run on arrays of indices.

## Summary

- Files keep strings: people, diffs, modules, versions, and saves need names.
- Programs intern: one table per category at load, dense indices after that, keys back out at
  every edge.
- Hashes are for constants in code and for programs with no shared table. Always check for
  collisions.
- The DOD advice is right, and it applies to memory, not to the data format.
