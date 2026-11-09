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

The stored actor follows the [contract](contract.md): primary facts only, with derived stats
computed. The session's JSON schema and commands are in the backend's contract.
