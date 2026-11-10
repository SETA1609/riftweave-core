# Vectors

Author: SETA1609

One file is the contract between implementations. An actor, a modifier list, a seeded roll, an expected result. Python runs it, and the Go core in `backend/` runs the same file. A second implementation that disagrees fails the vector, not a review comment.

## Running them

One test runs every file:

```
python -m unittest ruleset/scripts/test_vectors.py
```

`./scripts/test.sh` runs it with the other `ruleset/scripts/test_*.py` tests. The Go core runs every file too: `cd backend && go test ./...`.

The test globs `ruleset/tests/vectors/*.json`. Adding a vector file, or a vector to a file, needs no test edit. A new kernel function needs a runner in `ruleset/scripts/kernel_reference.py`, a row in the checklist below, and a command file in `backend/contract/` and a Go command ([backend](../backend/README.md)); the tests fail until all four exist. A new input field on an existing function goes into its command's `input_schema` in the same pull request.

## File format

A file holds one vector object or a JSON array of them. Each vector has:

```
name       what it checks, unique in the file
function   the kernel function; defaults to the file name without .json
expect     the result fields to check; fields not named are not checked
```

The other fields are the input the function reads. They are listed per function below. The test compares each `expect` field with the reference result for equality. A file that is not valid JSON, a vector without `name` or `expect`, or a `function` with no runner fails the test. [V 2026-11-02, run in this change]

## Inputs per function

| Function | Input | Result |
| --- | --- | --- |
| `resolution` | `actor.skills` (one skill, or `skill` names it), `actor.attributes.lck`, `modifiers`, `roll` | `target`, `success`, `margin` (null on a failure), `tags` |
| `derived` | `actor.attributes`, `actor.level`, `modifiers` | one field per derived key (`hp`, `mp`, `ap`, `carry`, `initiative`) |
| `formula` | `formula` (a key in `ruleset/data/formulas/`) or `expr` (an inline tree), `actor.attributes` | `value` |
| `modifiers` | `value` of `target`, `modifiers`, optional `elapsed_turns`, `removed_sources`, `packet` (`amount`, `tag`) | `value`, `flags`, `packet` (the amount after resist) |
| `conditions` | optional `base` (target values), `modifiers` (the actor's own), `conditions` (active key and stacks), `events` (`effect` refs or `remove` refs, in order) | `conditions` (key and stacks), `flags`, `values` (one per `base` target) |
| `actions` | `actor.attributes`, `actions` (a list of action keys) | `pool`, `cost` |
| `economy` | `op`: `convert` (`cost`) or `trade` (`side` buy or sell, `merchant`, `item_cost`, optional `value_percent`, optional `barter` with `skill_value` and `roll`) | `cp`; or `base_cp`, `shift`, `percent`, `price_cp` |
| `rest` | `occasion` (a key in `ruleset/data/rest/core.json`), `actor.attributes`, `actor.level`, `modifiers`, `current` (hp, mp, stamina, ap), `max.stamina`, `conditions` (key and duration) | the four resources after the rest, and `conditions` (the keys left) |
| `loot` | `table` (a key) or `monster` (a ref), `rolls` (one d100 per chance roll and per table roll, in row order), optional looter `actor.attributes.lck` and `modifiers` (on `loot_roll`) | `table` (a table input) or `tables` (the tables a monster rolled), `items` (item, quantity, and `material` if a row set one), `coins_cp`, `value_cp` (coins plus items at list price), `budget_cp` (at the table's or the monster's level), `looter_shift`, or `error` |
| `combat` | `attacker` (`skill`, `roll`, `attributes.lck`), `defender` (`evasion_skill`, `attributes.lck` for reaction crits, `armor_dr`, `hp`, `resist` by damage type, optional `reaction`: `kind` `evade` with `roll`, or `kind` `block` with `skill`, `roll`, `shield_dr`), `damage` (`roll`, `type`), `delivery` (`single`, the default, `area`, or `instant`), `modifiers` on the attack, optional `range` (`melee` or `ranged`), `situations` (keys), `called_shot` (a key), `attacker.flags` and `defender.flags` | `outcome` (`hit`, `miss`, `fumble`, `evaded`), `target`, `margin`, `crit`, `evaded`, `blocked`, `reaction_tags` (`crit`, `fumble` on the reaction roll), `damage` after reductions, `hp`, `riders` (on a called-shot hit), or `error` |
| `stored_actor` | `actor` (a stored actor, [state.md](state.md#the-stored-actor)) | `actor` unchanged, or `error` |
| `monster_sheet` | `monster` (a composed row: `attributes`, `level`, optional `skills`, `hitPoints`), optional `read_skills` | the derived stats and `skills` (or `error`) |
| `progression` | `op`: `xp` (`level`), `level` (`xp`), `level_up` (`actor.attributes`, `luck_roll`, `hp_roll`, `mp_roll`, `sp_roll`), `spend` (`skill`, `tagged`, `points`) | `xp_total`, `xp_step`; `level`; `skill_points`, `perks`, `gain` (or `error`); `ranks`, `spent`, `left` |
| `hazard` | `hazard` (a key), `op`: `detect`, `disarm`, or `trigger`; for a check: `actor.skills` (the check's skill), `actor.attributes.lck`, `modifiers`, `roll`; for a trigger: optional `actor` and `roll` (the avoid check) and `damage_roll` | `success`, `target`, `margin`, `tags` (and `triggered` for a disarm); or `avoided`, `packet` (`amount`, `type`, or null), `effects`; or `error` |
| `casting` | `spell` (a key), `caster` (`skills` with the spell's `<color>_magic`, `attributes.lck`, `mana`, optional `flags`, `armor` (`none`, `light`, `medium`, `heavy`), `reagents` (refs), `concentration` (the held spell key or null)), `modifiers` on the skill or on `spell_cost`, `roll`, optional `standing` (spell keys on the target, for dispel) | `outcome` (`cast`, `failed`, `prevented`, `no_mana`, `no_reagents`), `mana`, `consumed`, `check` (`target`, `success`, `margin`, `tags`, or null), `effects`, `concentration`, `ended`, `dispelled`; or `error` |

Coefficients come from `ruleset/data/`, not from the vector, so a data change that breaks a rule fails a vector.

Notes on the reference, for a second implementation to check against:

- `resolution` adds the `add` modifiers to the target after the stacking of [modifier.md](modifier.md): same source keeps the strongest, one `highest` origin keeps its strongest bonus and penalty. Resolution sees bonuses and penalties as one signed list. [V 2026-11-06]
- `modifiers` and `conditions` follow the order in [modifier.md](modifier.md): expire, dedupe by origin, source, target, and op (strongest wins), keep the strongest bonus and penalty of each `highest` origin, then add, mul (floored), set; `grant` makes flags and `resist` reduces a packet. [V 2026-11-06]
- `derived` and `resolution` read a missing attribute as 0, and `derived` a missing level as 1. [A] Each vector names every value its `expect` fields depend on.
- `formula` evaluates `lit`, `attr`, and the ops `add`, `sub`, `mul`, `div`. `div` is integer floor division, the `floor` rounding named in `ap.json`. [A] Other ops raise an error.
- `actions` pool is `base + attr * per_point / per_point_divisor` from `ruleset/data/actions/core.json`; cost is the sum of the named actions' costs, so `attack` plus `aimed` is 5. The turn cap and carry are not checked yet. [A]

Module functions are not kernel functions. The `wuxing` function's vectors are in `ruleset/modules/wuxing/tests/vectors.json`, and `test_wuxing.py` and the Go tests run them from the module's data.

## Checklist

Every kernel function and the vector file that covers it. The test reads this table: a `covered` row with no vector fails, a `pending` row that gained a vector fails until the row says `covered`, and a vector whose function is not in the table fails.

| Function | Vector file | Status |
| --- | --- | --- |
| `resolution` | `resolution.json` | covered |
| `derived` | `derived.json` | covered |
| `monster_sheet` | `monster_sheet.json` | covered |
| `formula` | `formula.json` | covered |
| `modifiers` | `modifiers.json` | covered |
| `conditions` | `conditions.json` | covered |
| `actions` | `actions.json` | covered |
| `economy` | `economy.json` | covered |
| `rest` | `rest.json` | covered |
| `combat` | `combat.json` | covered |
| `loot` | `loot.json` | covered |
| `progression` | `progression.json` | covered |
| `hazard` | `hazard.json` | covered |
| `casting` | `casting.json` | covered |
| `stored_actor` | `stored_actor.json` | covered |

The modifier vectors cover every op, mixed ops, same-source duplicates, origins, and expiry. The condition vectors apply and remove each of the 24 conditions once, plus stacks and multi-condition effects.
