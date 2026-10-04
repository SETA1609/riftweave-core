# Resolution

Author: SETA1609

One function. Crit and fumble are tags on the result, not a second roll and not a mode. Coefficients live in `ruleset/data/resolution/core.json`.

Morrowind resolves a spell as one check, then applies the effect list. Fallout resolves a skill as one d100 against the skill. The confirmation roll in the old combat section was a second function. It is gone.

## Function

```
roll    = d100
success = roll <= target
margin  = target - roll
window  = crit_window + floor(LCK / crit_window_per_two_luck)
tags    = []
if roll <= window:     tags += crit
if roll >= fumble_at:  tags += fumble
```

Luck does not roll again. It emits `add crit_window floor(LCK / 2)`, source `attr:lck`. Same source does not stack. A perk that wants a wider window emits its own modifier. It does not edit this file.

There is no TTRPG column and no video-game column. A client that has no use for the fumble tag ignores it. It does not delete the tag.

## Example

Blades 65. LCK 6. Roll 4.

```
window  = 5 + floor(6 / 2) = 8
success = 4 <= 65
margin  = 65 - 4 = 61
tags    = [crit]
```

Roll 100 against the same target is a miss and a fumble. Roll 9 is a hit and not a crit. The old confirmation target of 68 is not computed.

## What this file is not

Defense, cover, called shots, and action costs stay in `docs/modules/combat.md` and `ruleset/data/combat/base_resolution.json`. Damage dice stay on the weapon. A crit tag does not itself double damage. Doubling is a consumer of the tag, not a second roll.
