# Combat pipeline

One attack, in order, from the roll to the HP lost. Every step reads the actor snapshot and data; none calls another system ([contract.md](contract.md)). The numbers are in `ruleset/data/combat/pipeline.json` (formula trees), `ruleset/data/combat/damage_types.json` (tags), and `ruleset/data/resolution/core.json` (the d100 function). [`docs/modules/combat.md`](../modules/combat.md) keeps the TTRPG and video-game commentary and the worked examples; this file is the order of operations they share. An attack is a weapon strike or an offensive spell; both run this order, and the attack's `delivery` decides which reactions it allows. Save rolls for instant spells, and the magic layers (absorption, phase resistance), are not specified here (M3-05). [U]

## Order

```
1. attack     target = weapon_skill + attack mods                   resolution.md
              miss          roll > target
              fumble        roll >= fumble_at, no damage
              hit           otherwise; tag crit if roll <= window
2. reaction   on a hit, the defender may take one reaction: evade or block, not both
              delivery single: evade or block      delivery area: evade or block
              delivery instant: none (a save, M3-05)
   evade      evade_target = evasion_skill + evade mods             resolution.md
              evaded if evade roll <= evade_target and evade margin >= attack margin
              single: no damage, stop        area: damage halved at step 5
   block      physical types only (tag armor_dr_applies)
              blocked if block roll <= block skill and block margin >= attack margin
3. damage     damage = the caller's rolled damage dice
4. crit       if crit: damage = crit_damage(damage)                 formula tree
5. reduce     if evaded an area attack: damage = area_evade_damage(damage)   formula tree
              if blocked: damage = damage - shield_dr
6. armor DR   if the damage type has tag armor_dr_applies:
              dr = armor_dr, or pierce_dr(armor_dr) if the type also has armor_pierce_partial
              damage = damage - dr
7. resist     damage = damage - resist[damage_type]
8. hp         lost = max(damage, 0); hp = max(hp - lost, 0)
```
Each subtraction stops at zero before the next step: a step never raises damage. Dice and rolls are inputs, so a vector pins them. The reference does not roll.

## Reading the steps

- **Evasion** is a defender reaction after the attack hits, contested like block: the defender's evade margin must reach the attacker's margin, ties go to the defender. It replaces the earlier passive penalty on the attacker's target, so the attack roll reads only the attacker's numbers and "dodged" is its own outcome (`evaded`), apart from "missed". This is the defense-tree reading of [`modules/combat.md`](../modules/combat.md) (layer 1, "chance to avoid entirely") and its video-game auto-dodge; the TTRPG passive-penalty variant is dropped. The evade target is the raw `evasion` skill, the same way block reads the block skill, so points bought in Evasion buy dodges; AGI counts once, through the skill's base value. Armor and encumbrance evasion penalties are modifiers on that target; they are not in the vectors yet. [U] The earlier tree `10 + evasion_skill * 0.15 + AGI * 2` is dropped: skill points barely moved it and it topped out at 45. [V 2026-11-04, owner decision]
- **Delivery, not damage type, decides the dodge.** `single` is one target: a sword cut, an arrow, a fire bolt, a touch spell. A full evade avoids it. `area` is a radius or a cone: a fireball, a breath weapon, a thrown bomb. An evade halves the damage (rounded down) instead of avoiding it. `instant` cannot be dodged or blocked: the effect lands, and a save roll on a skill or attribute the spell names decides how it behaves (M3-05). Physical and magic attacks follow the same rule. [V 2026-11-04, owner decision] A spell effect's `target` shape maps to it: `touch` and `ranged` are `single`, `area` is `area`; `instant` is a new marker for spells, added in M3-05. [A]
- **Balance between evade and block.** At equal skill an evade avoids all damage and a block only subtracts shield DR, so the raw skill alone favours evading. The balance is not in this pipeline: perks gate stronger dodges and blocks, armor penalties land on evasion, area attacks are only halved, and instant spells skip the dodge for a save. The perk rows come with their own issues. [A]
- **One reaction per attack.** The defender declares evade or block before the attack roll, as `modules/combat.md` has it for block, and rolls it only if the attack hits; the input has a single `reaction`. A failed evade does not allow a block. [A]
- **Crit** multiplies damage by 2 (`critDamageMultiplier`). It is a consumer of the `crit` tag, not a second roll. Crit multiplies before any reduction. [A]
- **Block** compares margins as the TTRPG prose does: the blocker's margin must reach the attacker's. It is the other reaction (step 2) and reduces damage instead of avoiding it. The video-game input-timed block is a client mapping of the same numbers. Block applies only to types with `armor_dr_applies`: a shield does not stop fire. Evade applies to any type the delivery allows. [A]
- **Armor DR** is the sum of worn pieces ([armor.md](../modules/armor.md)); the pipeline takes the sum as an input. Types tagged `bypasses_armor` or `bypasses_physical_armor` carry no `armor_dr_applies`, so DR skips them. A piercing hit halves DR, rounded down. [A]
- **Resist** is the contract's `resist` op: a flat subtraction from a packet whose tag matches the damage type. [V 2026-11-02, read contract.md]
- **HP** floors at 0. What happens at 0 (down, dead) is not in this file. [U]

## Not covered

Called shots, cover and flanking, opportunity attacks, hit quality (`solid` and `glancing` margins), condition riders on a hit, and the video-game timing windows. They add terms to the same steps; none changes the order.
