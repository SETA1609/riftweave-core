# Vectors

Modified: 2026-11-19

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
| `derived` | `actor.attributes`, `actor.level`, `modifiers` | one field per derived key (`hp`, `mp`, `ap`, `carry`, `initiative`, `toxicity_limit`) |
| `formula` | `formula` (a key in `ruleset/data/formulas/`) or `expression` (an inline tree), `actor.attributes`, optional `variables` (named integer inputs) | `value` |
| `modifiers` | `value` of `target`, `modifiers`, optional `elapsed_turns`, `removed_sources`, `summoning` (whether a summon is out), `packet` (`amount`, `tag`) | `value`, `flags`, `packet` (the amount after resist) |
| `conditions` | optional `base` (target values), `modifiers` (the actor's own), `conditions` (active key and stacks), `events` (`effect` refs or `remove` refs, in order) | `conditions` (key and stacks), `flags`, `values` (one per `base` target) |
| `actions` | `actor.attributes`, `actions` (a list of action keys) | `pool`, `cost`, `stamina` (`cost` × `stamina_per_ap`) |
| `economy` | `operation`: `convert` (`cost`) or `trade` (`side` buy or sell, `merchant`, `item_cost`, optional `value_percent`, optional `barter` with `skill_value` and `roll`) | `cp`; or `base_cp`, `shift`, `percent`, `price_cp` |
| `rest` | `occasion` (a key in `ruleset/data/rest/core.json`), `actor.attributes`, `actor.level`, `modifiers`, `current` (hp, mp, stamina, ap, optional toxicity), `max.stamina`, `conditions` (key and duration), optional `running` (`ticks`, `modifiers`) and `meal` | the resources after the rest, and `conditions` (the keys left), `running` after the food row and the meal, `refused` |
| `loot` | `table` (a key) or `monster` (a ref), `rolls` (one d100 per chance roll and per table roll, in row order), optional looter `actor.attributes.lck` and `modifiers` (on `loot_roll`) | `table` (a table input) or `tables` (the tables a monster rolled), `items` (item, quantity, and `material` if a row set one), `coins_cp`, `value_cp` (coins plus items at list price), `budget_cp` (at the table's or the monster's level), `looter_shift`, or `error` |
| `combat` | `attacker` (`skill`, `roll`, `attributes.lck`), `defender` (`evasion_skill`, `attributes.lck` for the riposte and bulwark windows, `armor_dr`, optional `shield_dr` (a carried shield), `shields_skill`, and `weapon_skill` (a parry band, Buckler Play), optional `armor` (`none`, the default, `light`, `medium`, `heavy`; Unarmored Grace), `hp`, `resist` by damage type, optional `stance`: `dodge` or `block`, `modifiers`), `damage` (`roll`, `type`, optional `ability`: the attacker's attribute whose modifier adds), `delivery` (`single`, the default, `area`, or `instant`), `modifiers` on the attack, optional `range` (`melee` or `ranged`), `situations` (keys), `called_shot` (a key), `attacker.flags` and `defender.flags`, optional `attacker.weapon_effects` (applied effects) and `defender.creature_type`, optional `maneuver` (a key) and `attacker.perks` (refs; a rider perk, GD_S05_04), optional `loadout` (a maneuver's AP follows it) and `defender.loadout` (Crossed Guard) | `outcome` (`hit`, `miss`, `fumble`, `evaded`, `blocked`), `target`, `margin`, `crit`, `evaded`, `blocked`, `evasion_rate`, `block_rate`, and `parry_rate` (the bands), `parried`, `warded` and `ward_percent` (Spell Ward), `counter`, `stagger_ap`, `damage` after reductions, `hp`, `riders` (on a called-shot hit), `ap`, `stamina`, `effects` and `events` (with a maneuver or a rider), `save` (`attribute`, `dc`, `target`: a maneuver attempt on a hit), `rider` (the maneuver a rider perk put on a plain attack), `weapon_effects` (with weapon effects), or `error` |
| `attack_sequence` | `action` (`attack`, `jab`, `full_attack`, `circle_cut`), `loadout` (`main`, `off`), optional `ap` (AP left), `attacker` (`skill`, optional `off_skill`, `attributes`, `flags`, `modifiers`: the perks and the fighting stance), `foes` (combat defenders, in carry order), `damage` and optional `off_damage` (`type`, optional `ability`), `rolls` (`attack` and `damage` per strike made), optional `targets` (a foe per strike of an attack or a jab), `modifiers` on every strike, optional `after_parry` and `after_block`, optional `attacker.shield_dr` (a shield bash, Boss Strike) | `action`, `sequence` (the row), `ap`, `stamina`, `strikes` (`hand`, `foe`, `attack`, the combat fields `outcome`, `target`, `margin`, `crit`, `damage`, `hp`, and `carried`, `free`, `scissor`, and on a shield bash `bash` and `stagger_ap`), `hp` per foe, `lost`, or `error` |
| `stance` | `perks` (feature refs), `loadout` (`main`, `off`, optional `main_skill`, `off_skill`), optional `stance` (active before), `events` in order: `take` (a stance key, optional `mid_turn`), `end`, or `loadout` | `stance` (active after), `ap`, `modifiers` (the active stance's), `steps` (`stance`, `ap`, `ended`), or `error` |
| `consume` | `current` and `max` (`hp`, `mp`, `stamina`, optional `toxicity`; `current.ap` at a table), optional `at_table`, `events` in order: `item` (a consumable key, optional `in_combat`), `advance` (seconds), or `turns` | `hp`, `mp`, `stamina`, `ticks` (`source`, `origin`, `pool`, `magnitude`, `seconds_left`), `modifiers` (`source`, `origin`, `target`, `operation`, `value`, `seconds_left`), `refused` (item keys), `toxicity`, `conditions` (`core:condition/overdosed` while toxicity is above `max.toxicity`), and with `at_table` `ap` and `pending_ap` |
| `stored_actor` | `actor` (a stored actor, [state.md](state.md#the-stored-actor)) | `actor` unchanged, or `error` |
| `monster_sheet` | `monster` (a composed row: `attributes`, `level`, optional `skills`, `hitPoints`, `perks`, `maneuvers`), optional `read_skills` | the derived stats, `skills`, `maneuvers` (listed and unlocked by its perks), and `riders` (with rider perks) (or `error`, as on two perks that exclude each other) |
| `progression` | `operation`: `xp` (`level`), `level` (`xp`), `level_up` (`actor.attributes`, `luck_roll`, `hp_roll`, `mp_roll`, `sp_roll`), `spend` (`skill`, `tagged`, `points`, optional `group`), `read_skill` (`actor`, `read`, optional `groups`), `take_perk` (`actor`, `perks`, `perk`), `point_buy` (`actor.attributes`, as assigned before race) | `xp_total`, `xp_step`; `level`; `skill_points`, `perks`, `gain` (or `error`); `ranks`, `spent`, `left` (or `error`); `value`, `read_from` (`sheet`, `group`, `start`) (or `error`); `taken`, `perk`, `rank`, `unmet`, and `replaced` (a perk of the same group, GD_S05_04) (or `error`); `spent`, `left` (or `error`) |
| `hazard` | `hazard` (a key), `operation`: `detect`, `disarm`, or `trigger`; for a check: `actor.skills` (the check's skill), `actor.attributes.lck`, `modifiers`, `roll`; for a trigger: optional `actor` and `roll` (the avoid check) and `damage_roll` | `success`, `target`, `margin`, `tags` (and `triggered` for a disarm); or `avoided`, `packet` (`amount`, `type`, or null), `effects`; or `error` |
| `casting` | `spell` (a key), `caster` (`skills` with the spell's `<color>_magic`, `attributes.lck`, `mana`, optional `flags`, `armor` (`none`, `light`, `medium`, `heavy`), optional `shield` (`none`, `buckler`, `standard`, `tower`), `reagents` (refs), `concentration` (the held spell key or null), optional `perks` (feature refs), optional `summons` and `summon_templates` (refs)), optional `coating` and `enchantment` (effect refs, for a bound weapon), `modifiers` on the skill or on `spell_cost`, `roll`, optional `standing` (spell keys on the target, for dispel), optional `target` (`attributes`, `skills`, `roll`, for an instant spell's save) | `outcome` (`cast`, `failed`, `prevented`, `no_mana`, `no_reagents`), `mana`, `consumed`, `check` (`target`, `success`, `margin`, `tags`, or null), `effects`, `concentration`, `ended`, `dispelled`, `save`, `bound_weapon` (for a bound weapon spell); or `error` |
| `spellcraft` | `operation`: `craft` (`caster.known`, `caster.skills`, `spell` with `color` and `effects`), `learn` (`caster.known`, `spell` a key), or `research` (`caster.known`, `caster.skills`, `caster.attributes.lck`, `effect` a ref, optional `color`, `modifiers`, `roll`) | `valid`, `rejected` (`effect`, `reason`), `mana`, `cast_threshold`, `craft_threshold`, `skill_ok`; or `taught`, `known`; or `outcome` (`learned`, `failed`, `refused`), `reason`, `check`, `known`, `days`, `cost` |
| `command_summons` | `commander` (`name`, optional `perk` with a `command` object, `modifiers`, `concentration`), `summons` (names), optional `order` (`targets`) | `ap` (the order's cost), `modifiers` per summon name, or `error` |

Coefficients come from `ruleset/data/`, not from the vector, so a data change that breaks a rule fails a vector.

Notes on the reference, for a second implementation to check against:

- `resolution` adds the `add` modifiers to the target after the stacking of [modifier.md](modifier.md): same source keeps the strongest, one `highest` origin keeps its strongest bonus and penalty. Resolution sees bonuses and penalties as one signed list. [V 2026-11-06]
- `modifiers` and `conditions` follow the order in [modifier.md](modifier.md): expire, dedupe by origin, source, target, and operation (strongest wins), keep the strongest bonus and penalty of each `highest` origin, then add, multiply (floored), set; `grant` makes flags and `resist` reduces a packet. [V 2026-11-06]
- `derived` and `resolution` read a missing attribute as 0, and `derived` a missing level as 1. [A] Each vector names every value its `expect` fields depend on.
- `formula` evaluates `literal`, `attribute`, `variable`, and the operations `add`, `subtract`, `multiply`, `divide`. `divide` is integer floor division, the `floor` rounding named in `ap.json`. [A] Other operations raise an error.
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
| `attack_sequence` | `attack_sequence.json` | covered |
| `stance` | `stance.json` | covered |
| `loot` | `loot.json` | covered |
| `progression` | `progression.json` | covered |
| `hazard` | `hazard.json` | covered |
| `casting` | `casting.json` | covered |
| `spellcraft` | `spellcraft.json` | covered |
| `consume` | `consume.json` | covered |
| `stored_actor` | `stored_actor.json` | covered |
| `share` | `share.json` | covered |
| `command_summons` | `command_summons.json` | covered |

The modifier vectors cover every operation, mixed operations, same-source duplicates, origins, and expiry. The condition vectors apply and remove each of the 24 conditions once, plus stacks and multi-condition effects.
