# Combat pipeline

One attack, in order, from the roll to the HP lost. Every step reads the actor snapshot and data; none calls another system ([contract.md](contract.md)). The numbers are in `ruleset/data/combat/pipeline.json` (formula trees), `ruleset/data/combat/damage_types.json` (tags), and `ruleset/data/resolution/core.json` (the d100 function). [`docs/modules/combat.md`](../modules/combat.md) keeps the TTRPG and video-game commentary and the worked examples; this file is the order of operations they share. The physical tree only. The magic tree (absorption, resistance, phase resistance) is not specified here. [U]

## Order

```
1. evasion    penalty = evasion(evasion_skill, AGI)           formula tree
2. attack     target = weapon_skill + attack mods - penalty   resolution.md
              miss          roll > target
              fumble        roll >= fumble_at, no damage
              hit           otherwise; tag crit if roll <= window
3. damage     damage = the caller's rolled damage dice
4. crit       if crit: damage = crit_damage(damage)           formula tree
5. block      defender reaction, physical types only
              blocked if block roll <= block skill and block margin >= attack margin
              if blocked: damage = damage - shield_dr
6. armor DR   if the damage type has tag armor_dr_applies:
              dr = armor_dr, or pierce_dr(armor_dr) if the type also has armor_pierce_partial
              damage = damage - dr
7. resist     damage = damage - resist[damage_type]
8. hp         lost = max(damage, 0); hp = max(hp - lost, 0)
```

Each subtraction stops at zero before the next step: a step never raises damage. Dice and rolls are inputs, so a vector pins them. The reference does not roll.

## Reading the steps

- **Evasion** is a passive penalty on the attacker's target, as `modules/combat.md` says for the TTRPG column. The coefficients are the ones in `base_resolution.json` (`baseEvasion` 10, skill factor 0.15, AGI factor 2), written as integers: `10 + floor(evasion_skill * 15 / 100) + AGI * 2`. [A]
- **Crit** multiplies damage by 2 (`critDamageMultiplier`). It is a consumer of the `crit` tag, not a second roll. Crit multiplies before any reduction. [A]
- **Block** compares margins as the TTRPG prose does: the blocker's margin must reach the attacker's. The video-game input-timed block is a client mapping of the same numbers. Block applies only to types with `armor_dr_applies`. [A]
- **Armor DR** is the sum of worn pieces ([armor.md](../modules/armor.md)); the pipeline takes the sum as an input. Types tagged `bypasses_armor` or `bypasses_physical_armor` carry no `armor_dr_applies`, so DR skips them. A piercing hit halves DR, rounded down. [A]
- **Resist** is the contract's `resist` op: a flat subtraction from a packet whose tag matches the damage type. [V 2026-11-02, read contract.md]
- **HP** floors at 0. What happens at 0 (down, dead) is not in this file. [U]

## Not covered

Called shots, cover and flanking, opportunity attacks, hit quality (`solid` and `glancing` margins), condition riders on a hit, and the video-game timing windows. They add terms to the same steps; none changes the order.
