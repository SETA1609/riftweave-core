# Game state

The ruleset and a running game are two different kinds of data, and they have two models.

| | Ruleset | Game state |
| --- | --- | --- |
| What | Rows: spells, monsters, formulas | One run: an actor, what happened to it |
| Changes | Between releases, by review | On every command |
| Model | Tables ([data-model.md](data-model.md)) | A snapshot and a log |
| Lives in | `ruleset/data/` | A session, saved as JSON |

## A session

```
Session
  ruleset   the ruleset version and the modules the run started on
  start     the actor snapshot when the run began
  log[]     every command, in order: { command, input }
  actor     the current snapshot
```

## Rules

1. **The log is the truth.** `actor` is the result of running `log` from `start`, kept so a
   read does not replay. If they disagree, the log wins. [A]
2. **Every command is repeatable.** The rules never roll: a roll is an input
   ([resolution.md](resolution.md)), so a logged input holds it. The same start, log, and
   ruleset give the same actor.
3. **Nothing changes an actor outside the log.** A testing change, such as dev mode's
   inflicting a condition, is a logged entry like any other. [A]
4. **A session belongs to its ruleset version.** Replaying on another version is not promised
   to give the same result ([versioning.md](versioning.md)). [A]
5. **A log may be shortened.** A checkpoint makes the current `actor` the new `start` and
   drops the entries before it. Replay then starts at the checkpoint. [A]

What this gives: a save file, a replay, a bug report that reproduces, and a way to turn a real
run into vectors, all from one mechanism.

## The stored actor

The stored actor follows the [contract](contract.md): primary facts only, with derived stats
computed. A row in `ruleset/data/` is a definition; the actor holds instances that point at
rows. One schema says what an actor stores: `stored_actor` in `backend/contract/common.json`.

| Stored | Field |
| --- | --- |
| Attributes, level, and experience | `attributes`, `level`, `xp` |
| Skills and tag skills | `skills` (value by key), `tags` (skill keys) |
| Growth from level-up rolls | `growth` (`hp`, `mp`, `sp`) |
| Current pools | `resources` (`hp`, `mp`, `sp`, `ap`) |
| Active conditions | `conditions` (`key`, `stacks`) |
| Active modifiers | `modifiers` (`modifier`, with its source and origin, and `turns_elapsed`) |
| Flags | `flags` (keys) |
| Item instances | `inventory` (`id`, `base`, optional `material` and `size`, `mods`, `quantity`) |
| What is equipped | `equipped` (an inventory `id` by slot) |

Computed, never stored: maximum pools and every derived stat, from the snapshot and the data
([derived.md](derived.md)). An `id` is unique in the inventory, and `equipped` names only ids
that exist. Pools use the derived keys, so stamina is `sp`. There is no actor `id`: a session
holds one actor. [A] The `stored_actor` vectors round-trip it through Python and Go.
[V 2026-11-11]

The session's JSON schema and commands are in the backend's contract.
