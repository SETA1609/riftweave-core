# Sessions

A session is one run of the game as [docs/core/state.md](../core/state.md) defines it: a starting actor and a log of commands, with the current snapshot beside them. The `session` command holds no rule. It fills a step's input from the actor, calls the same core command as any other caller, and writes the result back.

```
Session
  ruleset   { modules }      the modules the run started on
  start     a stored actor
  log[]     { command, input, as? }
  actor     the current stored actor
```

| `op` | Input | Result |
| --- | --- | --- |
| `start` | `start` (a stored actor) | a new `session` |
| `step` | `session`, `step` (`command`, `input`, optional `as`) | the new `session`, and the command's `result` |
| `replay` | `session` | the `session` rebuilt from `start` and `log`, and `matches`: whether it equals the saved `actor` |

The log holds a step's input as it was given, so a replay fills it the same way. A step that fails leaves the session as it was and adds nothing to the log. A session serializes to JSON, so it can be saved and replayed. [V 2026-11-11, `go test ./core`]

## What a step reads and writes

A field the step names wins; the session fills only what is missing. [A]

| Command | Filled from the actor | Written back |
| --- | --- | --- |
| `derived` | `actor` | nothing |
| `combat` | `attacker.attributes`, or with `as: defender` `defender.hp` and `defender.attributes` | with `as: defender`, HP |
| `rest` | `actor`, `current`, `max.stamina` | HP, MP, SP, AP, and any condition the rest cleared |
| `casting` | `caster.skills`, `caster.attributes`, `caster.mana` | MP |
| `progression` | `actor` for `level_up` | `level` sets XP and level; `level_up` adds its rolls to `growth` |
| `consume` | `current`, `max` | HP, MP, SP |
| `conditions` | nothing | the conditions and flags |

Other commands run and are logged, and change nothing on the actor. The stored actor has no coins, and loot does not add to the inventory yet. [U]
