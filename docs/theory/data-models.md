# Data model theory

Background for [docs/core/data-model.md](../core/data-model.md): what a data model is, which
types exist, and why Riftweave's is shaped as it is. This page explains; it sets no rule.

## What a data model is

A data model is the agreed shape of a project's information: what kinds of things exist, what
fields they have, how they connect, what is stored and what is calculated, and what makes
them valid. Code and documents change around it, and everything depends on it, so it is one
of the most expensive things to change later.

## Levels of detail

Every project has all three, written down or not.

| Level | Answers | In Riftweave |
| --- | --- | --- |
| Conceptual | What things exist? | Spells, monsters, effects, actors |
| Logical | What fields and links do they have? | Schemas and `core:` references |
| Physical | How is it actually stored? | JSON files, and Go tables in memory |

## Types by structure

| Type | Shape | Strong at | Weak at | Typical use |
| --- | --- | --- | --- | --- |
| Hierarchical | A tree; each item has one parent | Fast walks down the tree | Anything with two parents | File systems, XML |
| Network | Items may have many parents | Complex links | Rigid, hard to change | Older mainframe databases |
| Relational | Tables linked by keys, each fact in one place | Consistency, joining data in any direction | Deep nested shapes | SQL databases |
| Document | Self-contained nested records | Reading one whole thing | Facts copied between records drift | MongoDB, JSON files |
| Key-value | A name maps to an opaque value | Speed, simplicity | Asking anything about the contents | Caches |
| Wide-column, columnar | Data grouped by column | Scanning huge datasets | Writing single records | Analytics |
| Graph | Items and links are both first-class | Following many links | Bulk tabular work | Social and recommendation data |
| Object | Classes and inheritance | Matching object-oriented code | Sharing data across languages | Object databases |
| Entity-component (ECS) | Entities are ids; components attach to them | Many changing things in a running game | Reading or editing by hand | Game engines |
| Dimensional | Facts surrounded by descriptive tables | Reports and totals | Everyday updates | Data warehouses |

The relational model comes from E. F. Codd's 1970 paper, "A Relational Model of Data for Large
Shared Data Banks". [U] Real systems often mix types, one per layer, as this one does.

## Why ours is shaped this way

The repo never recorded a comparison of these options, so this section reconstructs the
reasons from what the project requires. The requirements are checked; the reasoning that links
them to the choice is an assumption. [A]

| The project needs | So the model has |
| --- | --- |
| The same rules to drive computer games and tabletop play (`AGENTS.md`) | Plain data that belongs to no engine |
| A stranger to be able to build a game from the data and vectors alone (`ROADMAP.md`) | Readable files, schemas, and exact expected results |
| Systems that do not break each other (`contract.md`) | One modifier shape, and no system calling another |
| One effect to serve spells, potions, and ingredients | A shared pool, reached by reference |
| Modders to add content without forking (`docs/core/overrides.md`) | Rows that can be added or overridden by key |
| People to review every change | Text files that read well in a git diff |
| Python and Go to agree | Numbers in tables, and rolls as inputs |

`contract.md` names the influences: Mike Acton's rule that a program's purpose is to transform
data; orthogonality, from Hunt and Thomas; Morrowind's spells and potions as lists of the same
effect records; Fallout storing primary stats and recomputing the rest; and Elin's single sheet.
It names D&D 3.5's typed bonuses as the failure to avoid. [V 2026-11-10]

### Why relational, as the logical design

Game content is many small things that refer to each other. A monster names effects, a spell
names effects, and a drop names items. Keeping each fact once, and pointing at it, means a
change lands in one place. The cost of copies is real: materials once carried the same numbers
in two columns (M5-23). [V 2026-11-10]

### Why documents, as the storage

JSON needs no server, every language reads it, and a person can edit it. Nesting is kept to
what belongs to one row, such as a spell's list of effects. [A]

### Why flat tables, at runtime

Rules that do array reads and integer math stay simple and fast, and the backend resolves
references once at load instead of on every command (`docs/backend/README.md`). [V 2026-11-10]

### The alternatives

| Alternative | Why not, as the source of truth [A] |
| --- | --- |
| A SQL database | Needs a server or binary file, reviews badly in git, and is harder for a table to read |
| Documents only | Facts copied into each record drift apart |
| Object inheritance | Ties the data to one language, and deep chains are hard to reason about |
| ECS | Built for changing state in a running engine; poor for hand-edited rules. An engine may map our rows into one. |
| Graph | Most questions here are "look up this row", not "walk many links" |

## What it costs

- JSON Schema cannot check that a reference exists, so the repo needs `validate.py` on top.
- Every rule needs a consumer in each implementation, kept honest by vectors.
- There is no query language; tools have to load the files.
- Discipline is manual: nothing stops a contributor from storing a derived value except review.
