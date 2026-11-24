# Sessions

A session is one run of the game as [docs/core/state.md](../core/state.md) defines it: a starting actor and a log of commands, with the current snapshot beside them. The `session` command holds no rule. It fills a step's input from the actor, calls the same core command as any other caller, and writes the result back.

```
Session
  ruleset   { modules }      the modules the run started on
  start     a stored actor
  log[]     { command, input, as? }
  actor     the current stored actor
  dev?      true once a dev operation ran ([dev-mode.md](dev-mode.md))
  party?[]  { character, start, actor }   the characters the GM loaded (below)
  bestiary? { knowledge[] }               what the players know of each monster (below)
  encounter? { round, order[], turn, acted[], surprise[] }   the fight the GM runs (below)
```

| `operation` | Input | Result |
| --- | --- | --- |
| `start` | `start` (a stored actor) | a new `session` |
| `step` | `session`, `step` (`command`, `input`, optional `as`) | the new `session`, and the command's `result` |
| `replay` | `session` | the `session` rebuilt from `start` and `log`, and `matches`: whether it equals the saved `actor` |

The log holds a step's input as it was given, so a replay fills it the same way. A step that fails leaves the session as it was and adds nothing to the log. A session serializes to JSON, so it can be saved and replayed. [V 2026-11-11, `go test ./core`]

## The party

A GM's session may hold a **party**: the characters the GM loaded beside the session's own
actor and the GM's monsters (BE_S07_03, [files.md](files.md#the-hand-off)). Each member is the
character file's sync fields as loaded (`character`, with the `revision` it was loaded at, so a
replay knows which version it started from), its stored actor as loaded (`start`), and its
stored actor now (`actor`). `riftweave characters party add` loads one from the character list
and `party export` writes them back out, one file per player, a revision higher when `actor`
differs from `start`. The export also records itself in the session: each member's `character`
becomes the exported sync fields and `start` the exported actor, so a second export counts on
from the first instead of sending the same revision with other content.

The session carries the party through `start`, `step`, and `replay` unchanged and holds no rule
for it: no step reads or writes a member yet. The encounter tracker
(`BE_S08_07_encounter-tracker-cli`) will change members through steps. [A] [V 2026-11-20, a
`step` keeps the party, `characters_test.go`]

## Campaign state

Some steps change the campaign beside the actor. The session keeps that state and fills it into
the step, so the log holds the step as it was given and a replay rebuilds the state from it.
Each such command is one entry in `sessionStateAdapters` in `backend/core/session.go`.

| Command | Filled from the session | Written back |
| --- | --- | --- |
| `bestiary` | `knowledge`, from `bestiary.knowledge` | a `reveal`'s `knowledge` into `bestiary.knowledge`; a `view` writes nothing |
| `encounter` | `encounter`, from `encounter`, for `turn_end` | the result's `encounter` into `encounter` |

A replay's `matches` still compares the actor only; the rebuilt state replaces the saved one.
[A] [V 2026-11-23, `TestARevealStepKeepsTheMarksAndReplays`, BE_S08_05, [bestiary.md](bestiary.md)]

## Encounters

Owner request, 2026-11-16 (BE_S08_07): the GM runs a fight on one device with no connection, in
the web app or through the MCP tools. The `encounter` command holds the initiative list and whose
turn it is; as a session step it keeps the encounter in the session, so the fight replays and
exports with it. Damage, AP, and conditions stay the existing steps (`combat`, `actions`,
`conditions`) on each actor. No grid offline: positions are what the table says. [A]

```
encounter input    operation   start | turn_end
  start            combatants[]  { id, initiative, agi, lck, rolls?[], surprise? }
                                 initiative is the derived value, no roll (derived.md); agi and lck
                                 after modifiers; rolls the d100 roll-off, made when the encounter
                                 starts and kept in the step; surprise an ambusher
  turn_end         encounter     the state; a session step fills it from the session
result             encounter     { round, order[], turn, acted[], surprise[] }
                   current       { id, start? }   whose turn it is; start: surprise in the surprise
                                 round, combat on an actor's first turn of the fight, else absent
                   next          the id after the current one
refused            an id twice; a tie the rolls do not break
```

- **Order:** highest initiative first; a tie breaks by [derived.md](../core/derived.md#initiative-ties).
  Each actor's `rolls` are its roll-off, compared first to first, then second to second for a
  tie on the roll, lowest first. [A]
- **The surprise round:** with any ambusher, the fight opens on round 0, in which only the
  ambushers act, in initiative order, each on `start: surprise`. Round 1 is everyone's. [A]
- **The first turn:** an actor's first turn of round 1 or later that it has not acted in reads
  `start: combat`: the caller passes it to `actions` for the partial pool
  ([actions.md](../core/actions.md#pool)). An ambusher that acted in the surprise round has a
  normal turn in round 1. [A]
- **A turn ends** with `turn_end`: the current actor joins `acted`, the turn moves on, and after the
  last actor the round goes up by one and the order starts again.

## The rules version

`rules_version` names the rules a session ran on. `start` records the core's `RulesVersion`
(`backend/core/session.go`, `0.1.0` before the 1.0 release [A]), a step keeps it, and `replay`
stamps the rebuilt session with the current one, whatever the saved session said, so a caller
that compares the two knows a save came from other rules. A save from other rules is replayed
under the current ones; `matches` says whether the result is the same character. [V 2026-11-16,
`go test ./core`]

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
- **Armor DR** is the sum over every equipped armor piece of `dr_base` plus the piece's material
  `armor.defense` ([armor.md](../modules/armor.md)); a piece with no material takes the first of
  its `default_materials`.
- **Not filled:** the stored actor has no race or background, so their modifiers are not added;
  a caller passes them. [A]

[V 2026-11-14, the playthrough vectors through `go test ./core`: Toughness at rank 2 adds 30 HP,
the attack reads the shortsword's Blades (17), and the iron breastplate gives DR 7]

A level-up therefore plays with no dev step: `progression` `level_up` for the pool, `spend` with each skill's `key`, `take_perk`, and `spellcraft` `learn` for each new spell, all logged and replayed as any other step. A perk whose requirements fail is logged with `taken: false` and changes nothing. The session does not count the pool or the perks a level grants; the caller does. [A] The playtest driver plays all eight M5-18 builds to level 30 this way, with the binary run without `--dev`, and each session replays to the same actor. [V 2026-11-14, `go test ./core`, `playtest_driver.py`]

Each row is one entry in `sessionAdapters` in `backend/core/session.go`, so adding a command to sessions is one entry. Other commands run and are logged, and change nothing on the actor. The stored actor has no coins, and loot does not add to the inventory yet. [U]
