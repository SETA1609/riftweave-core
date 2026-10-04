# Resolution

Author: SETA1609

One function. Crit and fumble are tags on the result, not a second roll and not a mode. Coefficients live in `ruleset/data/resolution/core.json`.

Morrowind resolves a spell as one check, then applies the effect list. Fallout resolves a skill as one d100 against the skill. The confirmation roll in the old combat section was a second function. It is gone.

## Function

```
target  = skill + bonuses + penalties          no cap
roll    = d100
success = roll <= target and roll < fumble_at
margin  = target - roll                        only on a success
window  = crit_window + floor(LCK / crit_window_per_two_luck)
tags    = []
if success and roll <= window:  tags += crit
if roll >= fumble_at:           tags += fumble
```

Luck does not roll again. It emits `add crit_window floor(LCK / 2)`, source `attr:lck`. Same source does not stack. A perk that wants a wider window emits its own modifier. It does not edit this file.

A bought skill caps at 100. A bonus may push the check target past 100. That spare is the buffer for a penalty. Call of Cthulhu does the same: a skill over 100 does not make a roll of 100 succeed. It absorbs a malus so the check is still roll-under.

There is no TTRPG column and no video-game column. A client that has no use for the fumble tag ignores it. It does not delete the tag.

## Example

Blades 65. LCK 6. Roll 4.

```
window  = 5 + floor(6 / 2) = 8
success = 4 <= 65 and 4 < 100
margin  = 65 - 4 = 61
tags    = [crit]
```

Blades 80, bonus +40, penalty −30. Roll 100.

```
target  = 80 + 40 - 30 = 90
100 <= 90? no
100 >= 100, tag fumble, fail
```

Blades 80, bonus +40, no penalty. Roll 40.

```
target  = 120
success = 40 <= 120 and 40 < 100
margin  = 120 - 40 = 80
```

Roll 100 against a target of 120 is still a fumble and a failure. The extra 20 only matters when something subtracts.

## What this file is not

Defense, cover, called shots, and action costs stay in `docs/modules/combat.md` and `ruleset/data/combat/base_resolution.json`. Damage dice stay on the weapon. A crit tag does not itself double damage. Doubling is a consumer of the tag, not a second roll. The buy cap of 100 stays in `docs/modules/progression.md`.
