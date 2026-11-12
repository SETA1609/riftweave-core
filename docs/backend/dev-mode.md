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
| `rest` | `key` (an occasion in `rest/core.json`) | what that rest gives, by the `rest` command and the session's rest adapter |
| `perk` | `key` (a perk in `features/core.json`), optional `remove: 1` | `perks`, one rank more up to the perk's `ranks`, or the perk gone; prerequisites are not checked |
| `learn` | `spell` (a spell ref) or `effect` (an effect ref), optional `clear: true` | `spells` or `known_effects`, the key added once or removed |
| `gold` | `set` or `add`, in copper | `coins`, never below 0 |

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
- **HTTP API.** It does not serve `dev`: `POST /v1/commands/dev` is a 404, and the command is not in
  the list or the OpenAPI document ([api.md](api.md)).

## Rolls and the seed

The rules never roll: every roll is an input ([state.md](../core/state.md#rules)), so there is no
seed for dev mode to fix, and the agent CLI already takes each roll it needs. What a tester fixes
is the human CLI's prompts: `:roll` queues d100 values, and the next roll prompts open with them,
first in first out. The queue lives in the terminal UI, not in the session, because it changes
nothing until a command runs with the roll, and then the roll is in that command's logged input.
[V 2026-11-12, `go test ./cmd/riftweave/play`]

## The human CLI

`riftweave --dev play` adds a dev panel: the session's actor (level, XP, pools, conditions, held
effects, and items) under the menu, and a command line opened with `:`. The panel's session starts
from the last character created, or from every attribute at 5 at level 1. Each short command is
one `dev` operation, run through the same core, so the agent CLI gives the same session for the
same operations. [V 2026-11-12, `go test ./cmd/riftweave/play`]

| Command | Operation |
| --- | --- |
| `:give <item ref> [quantity]` | `item` |
| `:inflict <condition key> [stacks]` | `condition` |
| `:inflict <effect ref> <magnitude> <origin ref> [turns]` | `effect` |
| `:level <level>` | `level` |
| `:roll <d100> [<d100> ...]` | queues the rolls; no operation |
