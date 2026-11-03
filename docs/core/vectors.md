# Vectors

Author: SETA1609

One file is the contract between Python and Zig. An actor, a modifier list, a seeded roll, an expected result. Python runs it now. Zig runs the same file later. A second implementation that disagrees fails the vector, not a review comment.

## Running them

One test runs every file:

```
python -m unittest ruleset/scripts/test_vectors.py
```

`./scripts/test.sh` runs it with the other `ruleset/scripts/test_*.py` tests.

The test globs `ruleset/tests/vectors/*.json`. Adding a vector file, or a vector to a file, needs no test edit. A new kernel function needs a runner in `ruleset/scripts/kernel_reference.py` and a row in the checklist below; the test fails until both exist.

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
| `modifiers` | `value`, `modifiers` | `value` |
| `actions` | `actor.attributes`, `actions` (a list of action keys) | `pool`, `cost` |
| `combat` | `attacker` (`skill`, `roll`, `attributes.lck`), `defender` (`evasion_skill`, `attributes.agi`, `armor_dr`, `hp`, `resist` by damage type, optional `reaction`: `kind` `evade` with `roll`, or `kind` `block` with `skill`, `roll`, `shield_dr`), `damage` (`roll`, `type`), `modifiers` on the attack | `outcome` (`hit`, `miss`, `fumble`, `evaded`), `target`, `margin`, `crit`, `evaded`, `blocked`, `damage` after reductions, `hp` |
| `progression` | `op`: `xp` (`level`), `level` (`xp`), `level_up` (`actor.attributes`, `luck_roll`), `spend` (`skill`, `tagged`, `points`) | `xp_total`, `xp_step`; `level`; `skill_points`, `perks` (or `error`); `ranks`, `spent`, `left` |

Coefficients come from `ruleset/data/`, not from the vector, so a data change that breaks a rule fails a vector.

Notes on the reference, for a second implementation to check against:

- `resolution` adds every `add` modifier's value to the target. Resolution sees bonuses and penalties as one signed list. [V 2026-11-02, matches `resolution.md`]
- `modifiers` applies the contract stacking rule for `add` and `mul` only: same source and op, the latest wins; different sources add; `mul` after `add`. Two `mul` from different sources multiply. [A] `set`, `grant`, and `resist` raise an error in the reference until M1-05 fixes their order with vectors.
- `derived` and `resolution` read a missing attribute as 0, and `derived` a missing level as 1. [A] Each vector names every value its `expect` fields depend on.
- `formula` evaluates `lit`, `attr`, and the ops `add`, `sub`, `mul`, `div`. `div` is integer floor division, the `floor` rounding named in `ap.json`. [A] Other ops raise an error.
- `actions` pool is `base + attr * per_point / per_point_divisor` from `ruleset/data/actions/core.json`; cost is the sum of the named actions' costs, so `attack` plus `aimed` is 5. The turn cap and carry are not checked yet. [A]

## Checklist

Every kernel function and the vector file that covers it. The test reads this table: a `covered` row with no vector fails, a `pending` row that gained a vector fails until the row says `covered`, and a vector whose function is not in the table fails.

| Function | Vector file | Status |
| --- | --- | --- |
| `resolution` | `resolution.json` | covered |
| `derived` | `derived.json` | covered |
| `formula` | `formula.json` | covered |
| `modifiers` | `modifiers.json` | covered |
| `actions` | `actions.json` | covered |
| `combat` | `combat.json` | covered |
| `progression` | `progression.json` | covered |

The modifier vectors here cover `add`, `mul`, and same-source duplicates; M1-05 adds `set`, `grant`, `resist`, expiry, and conditions.
