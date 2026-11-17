# Normal forms

Modified: 2026-11-17

Background for [docs/core/data-model.md](../core/data-model.md) and
[docs/core/integrity.md](../core/integrity.md): what the first three normal forms are, why a
data set follows them, and how they apply to Riftweave's JSON rows. This page explains. The rule
lives in data-model.md (owner, 2026-11-17): the source must meet 1NF, 2NF, and 3NF, and need not
go past 3NF ([Past 3NF](#past-3nf)). The integrity gates ([integrity.md](../core/integrity.md))
enforce it. Which gaps the data still has is tracked in data-model.md and the issues, not here.

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

- **Copies drift.** When a skill is renamed, a ref to it moves with the rename, but a label
  copied beside the ref keeps the old name. This happened when the `block` skill became
  `shields`.
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

**Example.** A perk prerequisite whose `skills[].id` is a single ref in some rows and a list of
refs ("any of these") in others, with its `label` following it as a parallel list, breaks 1NF
twice: one field of two kinds, and two lists paired by position.

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

**Example.** A `label` written beside a `core:skill/...` ref in a perk prerequisite, or beside a
trait ref in a race, is the target's label copied. The ref alone is enough; a reader looks the
label up.

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

**Examples.**

- A monster row that stores `hitPoints` the HP formula would give.
- An override that sets `phase` to the base's own phase, or a subrace that repeats its parent's.
- An override merge that replaces a whole object, so a new `cost.quantity` forces `cost.unit` to
  be written again: the merge rule itself makes copies.
- A perk description that says "+10" beside a modifier whose `value` is 10.

## A domain written once

1NF speaks of a field's domain, the values it may hold. A domain is also a fact, and it can be
copied too. The eight abilities are rows of `abilities/core.json` and also an enum in
`schema.json`, and the same enum is written out again in `derived.schema.json` and
`ability.schema.json`. The damage types are rows of `combat/damage_types.json` and an enum in
`schema.json`. The cure is one source, and a check that the other copy equals it.

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
with keys that overlap and with several independent lists in one row. **The rule here: 1NF, 2NF,
and 3NF must hold; the forms past 3NF are not required**, so a design may stay denormalized
beyond 3NF when that reads better. [V owner 2026-11-17] A row that breaks BCNF, 4NF, or 5NF but
meets 3NF passes the gates. Of the patterns above, only the parallel lists come near 4NF, and
they go anyway, because they break 1NF.

