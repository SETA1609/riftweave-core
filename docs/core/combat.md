# Combat pipeline

One attack, in order, from the roll to the HP lost. Every step reads the actor snapshot and data; none calls another system ([contract.md](contract.md)). The numbers are in `ruleset/data/combat/pipeline.json` (formula trees), `ruleset/data/combat/damage_types.json` (tags), and `ruleset/data/resolution/core.json` (the d100 function). [`docs/modules/combat.md`](../modules/combat.md) keeps the TTRPG and video-game commentary and the worked examples; this file is the order of operations they share. An attack is a weapon strike or an offensive spell; both run this order, and the attack's `delivery` decides which reactions it allows. Save rolls for instant spells, and the magic layers (absorption, phase resistance), are not specified here (M3-05). [U]

## Order

```
0. setup      forbidden: a situation that forbids the range is an error  situations.json
              attack mods += every situation that applies (rows for the range,
              the delivery, and the attacker's and defender's flags)
              attack mods += the called shot's attack row               situations.json
1. attack     target = weapon_skill + attack mods, stacked by modifier.md   resolution.md
              miss          roll > target
              fumble        roll >= fumble_at, no damage
              hit           otherwise; tag crit if roll <= window
2. reaction   on a hit, the defender may take one reaction: evade or block, not both
              delivery single: evade or block      delivery area: evade or block
              delivery instant: none (a save, M3-05)
   evade      evade roll through resolution.md, target = evasion_skill + evade mods
              evaded if the evade succeeds and evade margin >= attack margin
              single: no damage, stop        area: damage halved at step 5
   block      physical types only (tag armor_dr_applies)
              block roll through resolution.md, target = block skill
              blocked if the block succeeds and block margin >= attack margin
3. damage     damage = the caller's rolled damage dice
              if a called shot: damage = floor(damage * the location's damage_mul)
4. crit       if crit: damage = crit_damage(damage)                 formula tree
5. reduce     if evaded an area attack: damage = area_evade_damage(damage)   formula tree
              if blocked: damage = damage - shield_dr
6. armor DR   if the damage type has tag armor_dr_applies:
              dr = armor_dr, or pierce_dr(armor_dr) if the type also has armor_pierce_partial
              damage = damage - dr
7. resist     damage = damage - resist[damage_type]
8. hp         lost = max(damage, 0); hp = max(hp - lost, 0)
9. riders     on a hit (not evaded), the called shot's riders are listed for the caller
```
Each subtraction stops at zero before the next step: a step never raises damage. Dice and rolls are inputs, so a vector pins them. The reference does not roll.

## Reading the steps

- **Evasion** is a defender reaction after the attack hits, contested like block: the defender's evade margin must reach the attacker's margin, ties go to the defender. It replaces the earlier passive penalty on the attacker's target, so the attack roll reads only the attacker's numbers and "dodged" is its own outcome (`evaded`), apart from "missed". This is the defense-tree reading of [`modules/combat.md`](../modules/combat.md) (layer 1, "chance to avoid entirely") and its video-game auto-dodge; the TTRPG passive-penalty variant is dropped. The evade target is the raw `evasion` skill, the same way block reads the block skill, so points bought in Evasion buy dodges; AGI counts once, through the skill's base value. Armor and encumbrance evasion penalties are modifiers on that target; they are not in the vectors yet. [U] The earlier tree `10 + evasion_skill * 0.15 + AGI * 2` is dropped: skill points barely moved it and it topped out at 45. [V 2026-11-04, owner decision]
- **Delivery, not damage type, decides the dodge.** `single` is one target: a sword cut, an arrow, a fire bolt, a touch spell. A full evade avoids it. `area` is a radius or a cone: a fireball, a breath weapon, a thrown bomb. An evade halves the damage (rounded down) instead of avoiding it. `instant` cannot be dodged or blocked: the effect lands, and a save roll on a skill or attribute the spell names decides how it behaves (M3-05). Physical and magic attacks follow the same rule. [V 2026-11-04, owner decision] A spell effect's `target` shape maps to it: `touch` and `ranged` are `single`, `area` is `area`; `instant` is a new marker for spells, added in M3-05. [A]
- **Balance between evade and block.** At equal skill an evade avoids all damage and a block only subtracts shield DR, so the raw skill alone favours evading. The balance is not in this pipeline: perks gate stronger dodges and blocks, armor penalties land on evasion, area attacks are only halved, and instant spells skip the dodge for a save. The perk rows come with their own issues. [A]
- **One reaction per attack.** The defender declares evade or block before the attack roll, as `modules/combat.md` has it for block, and rolls it only if the attack hits; the input has a single `reaction`. A failed evade does not allow a block. [A]
- **Reaction rolls use the one d100 function.** Evade and block go through [resolution.md](resolution.md) with the defender's skill and LCK, exactly like the attack roll: a 100 fumbles and fails, and a roll inside the defender's crit window is tagged `crit` (`reaction_tags`). A crit reaction has no extra effect yet; perks may read the tag later. [V 2026-11-04, owner decision]
- **Crit** multiplies damage by 2 (`critDamageMultiplier`). It is a consumer of the `crit` tag, not a second roll. Crit multiplies before any reduction. [A]
- **Block** compares margins as the TTRPG prose does: the blocker's margin must reach the attacker's. It is the other reaction (step 2) and reduces damage instead of avoiding it. The video-game input-timed block is a client mapping of the same numbers. Block applies only to types with `armor_dr_applies`: a shield does not stop fire. Evade applies to any type the delivery allows. [A]
- **Armor DR** is the sum of worn pieces ([armor.md](../modules/armor.md)); the pipeline takes the sum as an input. Types tagged `bypasses_armor` or `bypasses_physical_armor` carry no `armor_dr_applies`, so DR skips them. A piercing hit halves DR, rounded down. [A]
- **Resist** is the contract's `resist` op: a flat subtraction from a packet whose tag matches the damage type. [V 2026-11-02, read contract.md]
- **HP** floors at 0. What happens at 0 (down, dead) is not in this file. [U]

## Situations and called shots

Cover, flanking, prone, unseen, and shooting into melee are rows in `ruleset/data/combat/situations.json`; called shots are rows in the same file. Each row is plain modifiers on the attack target, so they stack under the one rule ([modifier.md](modifier.md)). Resolution reads the attack modifiers through the same stacking: two cover rows (both origin `environment`) keep the stronger penalty, half and three-quarters cover give −20, not −30. [V 2026-11-06]

- **Range.** An attack is `melee` (the default) or `ranged`. A situation row lists its modifiers per range; a range it does not list gets nothing.
- **Explicit or flag.** Cover, flanking, and shooting into melee are named by the caller (`situations`), since positions are the caller's. Prone, blinded, and invisible come from flags (`grant` modifiers of the conditions): a row with `when` applies when the attacker or the defender has that flag. [A]
- **Forbidden.** Full cover forbids ranged attacks, and a prone attacker cannot shoot. The pipeline returns an error; the caller picks another action. [A]
- **Unseen.** A blinded attacker and an invisible target share the source `situation:unseen`, so both together are still −20. Area attacks ignore it (`deliveries`). [A]
- **Called shots.** The location's `attack` is an `add` on the attack target; `damage_mul` multiplies the damage dice before the crit doubles it (a head crit is ×1.5, then ×2), rounded down. Its `riders` (staggered, slowed, a disarm on a failed STR check, a stun on a failed END check) are listed on a hit; the caller rolls the checks and applies the conditions. [A]

## Not covered

Opportunity attacks, hit quality (`solid` and `glancing` margins), condition riders on a crit, range bands for called shots (the extra −10 beyond short range), the perks that halve called-shot penalties, and the video-game timing windows. Difficult terrain is movement, not an attack term: it belongs to an optional positioning module, and standing up is an action row (2 AP). They add terms to the same steps; none changes the order. [U]
