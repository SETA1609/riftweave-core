# Versioning

The ruleset is versioned with [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`. A consumer (an engine, a module, a tabletop table) depends on **keys**: `core:weapon/shortsword` in a save file or a module row names an entry by its key, and a rename breaks it silently, because JSON Schema does not check cross-references ([keys.md](keys.md)). This file says what counts as breaking and how a key leaves.

## What is public

| Surface | Example |
| --- | --- |
| A **key** in a data collection | `shortsword`, `fire_bolt`, `persuasion`; the string `id` of a phase or cycle |
| A **namespaced reference** built from it | `core:effect/damage_fire` |
| A **schema property or enum member** | `ability` in `schema.json`, a property of `race.schema.json` |
| A **vector's expected result** | `ruleset/tests/vectors/*.json`; a second implementation passes them |

Numeric `id`s are internal bookkeeping and are not public ([keys.md](keys.md)). Labels, descriptions, and prose are not public.

## What is breaking

A change is **breaking** (a MAJOR bump) if data or a save that was valid and meant something at the previous release no longer is, or means something else:

- A key is removed or renamed. A rename is a removal plus an addition: the old reference no longer resolves.
- A schema property or enum member is removed or renamed, or a property becomes required, or a type or range narrows.
- A vector's expected result changes for the same input, other than a correction recorded in [decisions.md](../decisions.md) as a bug fix. [A]

Moving an entry to another file, or splitting a file, is not breaking: the key is unchanged.

## What is not

- **MINOR:** a new key, collection, module, optional schema property, or enum member; a new vector.
- **PATCH:** a fix to a label, description, or doc; a data correction that no vector and no key depends on.

Before `1.0.0` any minor release may break, but every breaking change is still marked **(breaking)** in its changelog fragment so a consumer can read what moved. `[Unreleased]` becomes `1.0.0` when the roadmap's definition of finished holds ([ROADMAP.md](../../ROADMAP.md#definition-of-finished)).

## Deprecation

A key is not removed in the release that stops using it:

1. A MINOR release keeps the key and says in its changelog fragment that it is **Deprecated**, naming the replacement.
2. The next MAJOR release may remove it, with a **Removed** bullet marked **(breaking)**.

A rename is the same two steps: add the new key in the MINOR, deprecate the old one, remove it in the MAJOR. The window is at least one MINOR release. Deprecation is recorded in the changelog only; there is no `deprecated` field in the schemas yet. [A]

## Tool

`ruleset/scripts/key_diff.py` lists the keys that disappeared or appeared between two versions:

```
python ruleset/scripts/key_diff.py v0.9.0 v1.0.0       # two tags (any git ref)
python ruleset/scripts/key_diff.py v1.0.0 WORKTREE     # a tag against the files on disk
```

It reads every `ruleset/data/**/*.json` at each ref. A key is the pair (collection, key), where the collection is the file's top-level array name, or the folder name for a file that is a single entry. It exits 1 if any key was removed, so a release job can fail on it, and 0 otherwise. Added keys are listed and never fail. Module data under `ruleset/modules/` is outside the tool: a module owns its namespace and versions itself. [A]

## Release check

Before tagging, run the tool from the previous tag to the release commit. Every removed key needs a **Removed** bullet marked **(breaking)** in the changelog; every deprecated key, a **Deprecated** bullet in an earlier release. The owner does this when moving fragments into `CHANGELOG.md` ([changes/README.md](../../changes/README.md#releasing)).
