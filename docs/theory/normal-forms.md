# Normal forms

Modified: 2026-11-17

Background for [docs/core/data-model.md](../core/data-model.md) and
[docs/core/integrity.md](../core/integrity.md): what the first three normal forms are, why a
data set follows them, and how they apply to Riftweave's JSON rows. This page explains; it sets
no rule. The rules are the integrity gates, which the issues listed at the end add one by one.

Sources, downloaded with `devops-tools fetch-docs` into the gitignored `cache/docs/` on
2026-11-17 and quoted from there:

- William Kent, "A Simple Guide to Five Normal Forms in Relational Database Theory",
  *Communications of the ACM* 26(2), February 1983 (`https://www.bkent.net/Doc/simple5.htm`).
- Wikipedia, "Database normalization", "First normal form", and "Third normal form"
  (`https://en.wikipedia.org/wiki/...`). These articles quote E. F. Codd's papers; the papers
  themselves were not read for this page. [U]

## What normalization is

> Database normalization is the process of structuring a relational database in accordance
> with a series of normal forms to reduce data redundancy and improve data integrity. It was
> first proposed by British computer scientist Edgar F. Codd as part of his relational model.
>
> — Wikipedia, "Database normalization" [V 2026-11-17]

Kent says the normal forms are design guidelines, useful without a database:

> The normal forms defined in relational database theory represent guidelines for record
> design. [...] The design guidelines are meaningful even if one is not using a relational
> database system.
>
> — Kent 1983 [V 2026-11-17]

That fits Riftweave. The logical design is relational, but the storage is JSON files and there
is no database ([data-models.md](data-models.md)). The normal forms still say where each fact
belongs.

## Words used below

| Word | Meaning | In Riftweave |
| --- | --- | --- |
| Table (relation) | A set of rows of one kind | A collection, such as `features` or `skills` |
| Key | The fields that pick out exactly one row | A row's `key`; for a nested list, the parent's key plus the item |
| Non-key field | Any other field | `label`, `phase`, `hitPoints`, `description` |
| Fact about X | A value that is fixed once X is known (a functional dependency, X → value) | A skill's label is fixed by the skill's key |
| Domain | The set of values a field may hold | An enum, a `$ref` type, or the rows of a lookup table |

## Why use them

Every normal form removes one way for the same fact to be written twice. Kent lists what goes
wrong when a fact is copied into rows that are about something else (his example is a
warehouse address copied into every part record):

> The warehouse address is repeated in every record that refers to a part stored in that
> warehouse. If the address of the warehouse changes, every record referring to a part stored
> in that warehouse must be updated. Because of the redundancy, the data might become
> inconsistent, with different records showing different addresses for the same warehouse. If
> at some point in time there are no parts stored in the warehouse, there may be no record in
> which to keep the warehouse's address.
>
> — Kent 1983 [V 2026-11-17]

These are the update anomalies: a change has to be made in many places, the copies drift apart,
and some facts have nowhere to live. Codd stated the aims of going past the first normal form:

> To free the collection of relations from undesirable insertion, update and deletion
> dependencies. To reduce the need for restructuring the collection of relations, as new types
> of data are introduced, and thus increase the life span of application programs.
>
> — Codd, "Further Normalisation of the Data Base Relational Model" (1971), as quoted in
> Wikipedia, "Database normalization" [V 2026-11-17]

For Riftweave the reasons are concrete:

- **Copies drift, and this has already happened.** When M5-56 renamed the `block` skill to
  `shields`, the ref in four perk prerequisites moved with it, but the label copied beside the
  ref still says "Block". [V 2026-11-17, a script over `ruleset/data/features/core.json` on
  `main` at `c99c7a8`] Earlier, materials carried the same numbers in two columns (M5-23).
- **A rename should be one edit.** Content is edited by hand and reviewed as a git diff. A
  fact written once changes in one line of the diff.
- **Modules override by key** ([overrides.md](../core/overrides.md)). If a module changes a
  skill's label, a copy of that label in a perk shows the old one.
- **Two implementations must agree.** Python and Go both load the data. A value that is
  computed from other fields, but also stored, gives two answers when they disagree.

## First normal form (1NF): one value of one kind per field

> Under first normal form, all occurrences of a record type must contain the same number of
> fields. First normal form excludes variable repeating fields and groups.
>
> — Kent 1983 [V 2026-11-17]

> A relation is said to be in first normal form if each field is atomic, containing a single
> value rather than a set of values or a nested table.
>
> — Wikipedia, "First normal form" [V 2026-11-17]

**How it reads for JSON rows.** A JSON row can hold a list, which a strict relational table
cannot. Riftweave treats a nested list that belongs to one row as a **child table**: a spell's
`effects` is a table whose key is the spell plus the position in the list, stored inside the
spell for readability ([data-models.md](data-models.md#why-documents-as-the-storage)). That is
allowed. What breaks 1NF here [A]:

- **A field whose kind changes from row to row.** A field holds one domain: always a ref, or
  always a list of refs, never one or the other.
- **Parallel lists**: two lists whose items pair up by position (`id: [a, b]` beside
  `label: [A, B]`). The pairing is a fact that no field holds.
- **Several facts packed into one string**, such as `"str 6, blades 40"` in place of fields.

**Today.** Every field path in `ruleset/data/` holds one JSON type except one: a perk
prerequisite's `skills[].id` is a single ref in 59 rows and a list of refs (meaning "any of
these") in 10, and its `label` follows it as a parallel list. [V 2026-11-17, a script over every
data file] Issue M5-70.

## Second normal form (2NF): no fact about part of the key

> Second normal form is violated when a non-key field is a fact about a subset of a key. It is
> only relevant when the key is composite, i.e., consists of several fields.
>
> — Kent 1983 [V 2026-11-17]

**How it reads for JSON rows.** A top-level row has a one-field key (`key`), so it cannot
break 2NF. A child table can: its key is the parent plus the item. A perk prerequisite row is
keyed by the perk and the skill it names, and its `min` is a fact about both (this perk needs
this much of this skill). A `label` there is a fact about the skill alone, so it belongs in the
`skills` table, not in the perk.

**Today.** 77 references carry a copy of the target's label: 59 in perk prerequisites and 18 in
race `traits`. Four of them are already wrong (the "Block" labels above). [V 2026-11-17] The
2026-11-01 review flagged the race copies as DOD-23 and left them open
(`docs/archive/review/findings-data-oriented-design.md`). Issue M5-71.

## Third normal form (3NF): no fact about another non-key field

> Third normal form is violated when a non-key field is a fact about another non-key field
> [...] The EMPLOYEE field is the key. If each department is located in one place, then the
> LOCATION field is a fact about the DEPARTMENT -- in addition to being a fact about the
> EMPLOYEE.
>
> — Kent 1983 [V 2026-11-17]

Kent's summary of 2NF and 3NF together, in the form Wikipedia quotes:

> William Kent summarised 3NF with the dictum that "a non-key field must provide a fact about
> the key, the whole key, and nothing but the key".
>
> — Wikipedia, "Third normal form" [V 2026-11-17]

**How it reads for JSON rows.** Three patterns break it here:

- **A value that a formula computes from other fields of the row.** A monster's HP is a fact
  about its level, attributes, and creature type, not about the monster's key directly.
  Storing it is the employee's location again. Riftweave's rule already says so: store the
  primary facts and compute the rest (`contract.md`).
- **An override that repeats its base.** An override row names a `base` (a non-key field) and
  lists only what differs. A value equal to the base's is a fact about the base, copied.
- **Prose that restates a number.** "+10 to deception" in a description is a copy of the
  modifier's `value`. When the value changes, the text does not.

**What already follows it.** A spell stores its `color`, not its skill: the skill is always
`<color>_magic`. An armor piece stores its weight class, and the casting penalty for each class
lives once in `casting/core.json` (`armor_penalty`), not on each piece. [V 2026-11-17]

**Today.**

- 55 of 59 monster rows store `hitPoints`. [V 2026-11-17] Issue M5-74.
- 74 override fields equal the base field they replace (`phase` 26, `size` 24, `tags` 9,
  `level` 7, `hitPoints` 6, `enchantment_slots` 2), and 4 subraces repeat their parent's
  `phase`. Another 41 values are repeated because an override replaces a whole object: a new
  `cost.quantity` forces `cost.unit` to be written again. [V 2026-11-17] Issue M5-73.
- 79 of 124 feature descriptions restate a number that is also a modifier value of the same
  row. [V 2026-11-17, a script matching numbers in the text to the row's modifiers; it can
  count a coincidence] Issue M5-75.

## A domain written once

1NF speaks of a field's domain, the values it may hold. A domain is also a fact, and it can be
copied too. The eight abilities are rows of `abilities/core.json` and also an enum in
`schema.json`, and the same enum is written out again in `derived.schema.json` and
`ability.schema.json`. The 13 damage types are rows of `combat/damage_types.json` and an enum in
`schema.json`. They match today. [V 2026-11-17] Issue M5-72: one source, and a check that the
other is the same.

## When copies are fine

Normalization is a rule for the **source**, the files people edit. A copy that is generated
from the source and never edited by hand is a cache, not a second fact:

- **The runtime tables.** The backend merges bases, templates, and overrides into flat rows at
  load ([data-models.md](data-models.md#why-flat-tables-at-runtime)). That is denormalized on
  purpose, for speed, and rebuilt from the source every time.
- **Generated files**: `docs/issues/order.md`, `docs/core/data-model-diagram.md`, and the
  playtest reports.
- **Vectors.** A vector copies inputs and expected results on purpose. It pins what the rules
  give, so a change in the data that alters a result has to change the vector too
  ([vectors.md](../core/vectors.md)).

## Past 3NF

Kent goes on to Boyce-Codd normal form (BCNF), the fourth (4NF), and the fifth (5NF). They deal
with keys that overlap and with several independent lists in one row. The goal here is 3NF
**at least**. Of the patterns above, only the parallel lists come near 4NF, and M5-70 removes
them. [A]

## The issues

| Form | Issue | What it fixes | Gate it adds |
| --- | --- | --- | --- |
| 1NF | M5-70 | Perk prerequisite `skills[].id` is a ref or a list; the label list runs beside it | A field path holds one JSON type in every row |
| 2NF | M5-71 | 77 labels copied beside references, 4 already stale | No field of the target row beside a reference |
| Domain | M5-72 | Ability and damage-type enums written beside their tables | Each enum equals its table |
| 3NF | M5-73 | 74 override values equal to the base; 4 subrace copies; whole-object overrides | An override that equals its base fails |
| 3NF | M5-74 | 55 monsters store `hitPoints` | HP is computed, never authored |
| 3NF | M5-75 | 79 descriptions restate modifier numbers | A number in a description is one of the row's values, or the text is filled from them |
