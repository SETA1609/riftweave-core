# Dev mode

Dev mode changes a session's actor directly, so a tester can reach a state without playing to
it (APP-08). It is off unless the binary runs with `--dev`. It holds no rule: each operation
checks its refs as any command does and writes the stored actor
([state.md](../core/state.md#the-stored-actor)), and the rules that read the actor afterwards are
the real ones. An inflicted effect is a modifier with a real origin, so it stacks with a played
one by that origin's policy ([modifier.md](../core/modifier.md#origin)).

## The `dev` command

```
input    operation, session, and the operation's fields
result   session: the actor changed, the operation in the log, and dev: true
```

| `operation` | Fields | What it changes |
| --- | --- | --- |
| `resource` | `pool` (`hp`, `mp`, `sp`, `ap`), `set` or `add` | `resources`, never below 0 |
| `xp` | `set` or `add` | `xp`; the level is not recomputed |
| `level` | `set` | `level`, and `xp` to that level's total (the `progression` command) |
| `attribute` | `key` (an attribute), `set` or `add` | `attributes`, never below 1 |
| `skill` | `key` (a skill in the data), `set` or `add` | `skills`, between 0 and 100 |
| `condition` | `key` (a condition in the data), `stacks` (0 clears it) | `conditions` |
| `effect` | `effect` (a ref in the effect registry), `magnitude`, `origin` (a ref in `origins/core.json`), optional `parameter`, optional `turns`; or `effect` and `clear: true` | one modifier with source `dev:<effect key>` on the parameter, or on the effect's own key, held `turns` turns or while applied; `clear` removes it |
| `item` | `base` (an item ref), optional `quantity` (1), optional `material`; or `remove` (an inventory id) | `inventory`, a new id one above the highest |
| `turns` | `turns` | every modifier's `turns_elapsed`; a `turns` modifier that has run its turns ends |

`set` and `add` are exclusive, and one is required where the table names them. A bad ref, an
unknown operation, or a missing field is a rule error, and the session is unchanged.

## The gate and the log

- **Off by default.** Without `--dev`, `dev` returns `{"error": "dev mode is off; run with --dev"}`,
  and a session step or a replay that holds a `dev` step is refused the same way. The flag is set
  by the binary, never by the input.
- **Logged.** Each operation is one step in the session's log, `{ "command": "dev", "input": ... }`,
  with the input as given minus `session`. A replay repeats it, so a dev session replays to the
  same actor.
- **Marked.** A session that has run a dev operation carries `dev: true` from then on, so a dev run
  is never taken for a played one. A replay keeps the mark.
- **Agent CLI.** `riftweave --dev dev <operation> [file]` reads the rest of the input from the file
  or stdin; the operation may also be in the JSON.
- **HTTP API.** It does not serve `dev` (APP-06).

## Not yet

This is the first part of APP-08. Left for the second: perks, learned spells and effects, and gold
(the stored actor has no field for them yet), advancing by rests, fixed rolls and the seed, and
the human CLI's dev panel and short commands (`:give`, `:inflict`, `:level`, `:roll`). [A]
