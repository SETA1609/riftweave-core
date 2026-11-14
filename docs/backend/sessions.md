# Sessions

A session is one run of the game as [docs/core/state.md](../core/state.md) defines it: a starting actor and a log of commands, with the current snapshot beside them. The `session` command holds no rule. It fills a step's input from the actor, calls the same core command as any other caller, and writes the result back.

```
Session
  ruleset   { modules }      the modules the run started on
  start     a stored actor
  log[]     { command, input, as? }
  actor     the current stored actor
  dev?      true once a dev operation ran ([dev-mode.md](dev-mode.md))
```

| `operation` | Input | Result |
| --- | --- | --- |
| `start` | `start` (a stored actor) | a new `session` |
| `step` | `session`, `step` (`command`, `input`, optional `as`) | the new `session`, and the command's `result` |
| `replay` | `session` | the `session` rebuilt from `start` and `log`, and `matches`: whether it equals the saved `actor` |

The log holds a step's input as it was given, so a replay fills it the same way. A step that fails leaves the session as it was and adds nothing to the log. A session serializes to JSON, so it can be saved and replayed. [V 2026-11-11, `go test ./core`]

## What a step reads and writes

A field the step names wins; the session fills only what is missing. [A]

| Command | Filled from the actor | Written back |
| --- | --- | --- |
| `derived` | `actor`, `modifiers` (the actor's, below) | nothing |
| `combat` | `attacker.attributes`, `attacker.skill` (the equipped weapon's skill), and `modifiers` (the actor's on that skill); or with `as: defender` `defender.hp`, `defender.attributes`, `defender.armor_dr` (the worn armor), and `defender.modifiers` (the actor's) | with `as: defender`, HP |
| `rest` | `actor`, `current`, `max.stamina` | HP, MP, SP, AP, and any condition the rest cleared |
| `casting` | `caster.skills`, `caster.attributes`, `caster.mana`, `modifiers` (the actor's) | MP |
| `progression` | `actor` for `level_up`; for `spend` with `key`, `skill` (the skill's value, as `read_skill` reads it) and `tagged` (whether `key` is a tag); for `take_perk`, `actor` and `perks` | `level` sets XP and level; `level_up` adds its rolls to `growth`; `spend` with `key` adds the bought `ranks` to that skill; `take_perk` sets the perk to its new `rank` when `taken` |
| `spellcraft` | for `learn`, `caster.known` (the effects the actor knows) | `learn` adds the spell to `spells` and sets `known_effects` to `known` |
| `consume` | `current`, `max` | HP, MP, SP |
| `conditions` | `conditions` | the conditions and flags |
| `dev` | nothing | what the operation names ([dev-mode.md](dev-mode.md)); only with `--dev` |

### What the actor brings to a step (APP-13)

The session fills a step with what the stored actor holds, so a perk or a piece of armor changes a
command's result without the caller restating it. A field the step names still wins.

- **The actor's modifiers** are its perks and its held modifiers, in that order. A perk at rank
  `r` emits each of its `modifiers` as [progression.md](../modules/progression.md#perks-and-traits)
  says: an `add` at `value × r`, a `multiply` of `m` at `1 + (m − 1) × r`, a `grant` or a `set`
  once. A held modifier is the stored `modifier` as it is. The commands stack them by origin
  ([modifier.md](../core/modifier.md#origin)), so the session adds no rule.
- **Which reach which command:** `derived`, `casting`, and the defender's side of `combat` read
  each modifier's target themselves, so they get all of them. The attacker's side of `combat`
  treats every modifier as one on the attack, so it gets only those whose target is the attack's
  skill.
- **The attack's skill** is the equipped weapon's `skill` (the first equipped item that is a
  weapon, by slot name, since the stored actor does not name its slots [A]), read as
  `read_skill` reads it: the written value, else its start. With no weapon it is `unarmed`.
- **Armor DR** is the sum over every equipped armor piece of `drBase` plus the piece's material
  `armor.defense` ([armor.md](../modules/armor.md)); a piece with no material takes the first of
  its `default_materials`.
- **Not filled:** the stored actor has no race or background, so their modifiers are not added;
  a caller passes them. [A]

A level-up therefore plays with no dev step: `progression` `level_up` for the pool, `spend` with each skill's `key`, `take_perk`, and `spellcraft` `learn` for each new spell, all logged and replayed as any other step. A perk whose requirements fail is logged with `taken: false` and changes nothing. The session does not count the pool or the perks a level grants; the caller does. [A] The playtest driver plays all eight M5-18 builds to level 30 this way, with the binary run without `--dev`, and each session replays to the same actor. [V 2026-11-14, `go test ./core`, `playtest_driver.py`]

Each row is one entry in `sessionAdapters` in `backend/core/session.go`, so adding a command to sessions is one entry. Other commands run and are logged, and change nothing on the actor. The stored actor has no coins, and loot does not add to the inventory yet. [U]
