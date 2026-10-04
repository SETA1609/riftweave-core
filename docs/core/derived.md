# Derived stats

Author: SETA1609

Derived stats are functions of the actor snapshot. They are not fields a module may write. Fallout stores SPECIAL and recomputes HP, AP, and carry. This file is that recompute. The coefficients live in `ruleset/data/derived/core.json`. A module that wants more HP emits a modifier. It does not edit this file.

## Function

```
attr_term = (attr * per_point) / per_point_divisor
value = base + attr_term + level * (attr * per_level_attr + per_level_flat)
final = apply(modifiers, value)
```

Division is integer division. `per_point_divisor` defaults to 1. `attr` is the current attribute after its own modifiers. Derived modifiers apply after the function. Stacking is the kernel contract: same source does not stack, different sources add, `mul` after `add`.

| key | attr | formula |
| --- | --- | --- |
| hp | END | `15 + END * 8 + level * 4` |
| mp | INT | `INT * 8 + level * 2` |
| ap | AGI | `2 + floor(AGI / 3)` |
| carry | STR | `25 + STR * 10` pounds |
| initiative | PER | `PER` |

These closed forms are the ones already printed under Derived statistics in `docs/modules/progression.md`. That table is now a copy. The JSON is the authority. The level-up bullets (`END × 8 + 4` and so on) are the grant procedure, not a second max. Stamina stays in that prose. This file does not own a second action economy.

## Example

Level 1. No modifiers. END 5, INT 4, AGI 6, STR 7, PER 5.

```
hp         = 15 + 5 * 8 + 1 * 4 = 59
mp         = 4 * 8 + 1 * 2       = 34
ap         = 2 + floor(6 / 3)    = 4
carry      = 25 + 7 * 10         = 95 lb
initiative = 5
```

A perk `source: perk:tough` with `add hp 8` makes HP 67. It does not change `per_point`. Two copies of the same perk do not stack.

## What this file is not

Level-up order, XP, and perk cadence stay in `docs/modules/progression.md`. The turn cap of 10 on the action-point pool stays there too. This function is the starting pool. The cap is a consumer rule.
