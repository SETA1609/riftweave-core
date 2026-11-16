# Combat pipeline

Modified: 2026-11-16

One attack, in order, from the roll to the HP lost. Every step reads the actor snapshot and data; none calls another system ([contract.md](contract.md)). The numbers are in `ruleset/data/combat/pipeline.json` (formula trees), `ruleset/data/combat/damage_types.json` (tags), and `ruleset/data/resolution/core.json` (the d100 function). [`docs/modules/combat.md`](../modules/combat.md) keeps the TTRPG and video-game commentary and the worked examples; this file is the order of operations they share. An attack is a weapon strike or an offensive spell; both run this order, and the attack's `delivery` decides which defense bands it meets. A cast's damage runs steps 2 to 8 with the cast check as its attack roll ([casting.md](casting.md#damage-to-a-target), M3-18). Save rolls for instant spells are in [casting.md](casting.md#delivery-and-saves); the magic layers (absorption, phase resistance) are not specified here. [U]

## Order

```
0. setup      forbidden: a situation that forbids the range is an error  situations.json
              attack mods += every situation that applies (rows for the range,
              the delivery, and the attacker's and defender's flags)
              attack mods += the called shot's attack row               situations.json
1. attack     one d100 roll: target = weapon_skill + attack mods, stacked by modifier.md   resolution.md
              miss          roll > target
              fumble        roll >= fumble_at, no damage
              success       otherwise, margin = target - roll; tag crit if roll <= window
2. bands      the defense bands, read on the same roll (none for delivery instant)
              evasion_rate  E = evasion_rate(evasion_skill + defender mods on evasion)     formula tree
                            stance dodge: E = dodge_evasion_rate(E)                         formula tree
              block_rate    B = 0, or with stance block: block_rate(shields_skill + mods on shields)
                            a type without armor_dr_applies: B = 0, or with block_magic
                            B = magic_block_rate(B)                                          formula tree
              crit          a crit always hits: no band
              evaded        margin < E           single: no damage, stop   area: halved at step 5
              blocked       E <= margin < E + B  no damage, stop
              hit           margin >= E + B
              riposte       stance dodge, single, evaded, margin >= E - window: counter
              bulwark       stance block, blocked, margin >= E + B - window: stagger
3. damage     damage = the caller's rolled damage dice
              if the attack names an ability: damage = max(damage + attribute_modifier, 0)
              if a called shot: damage = floor(damage * the location's damage_multiplier)
4. crit       if crit: damage = crit_damage(damage)                 formula tree
5. area       if evaded an area attack: damage = area_evade_damage(damage)   formula tree
6. DR         if the damage type has tag armor_dr_applies:
              dr = armor_dr + shield_dr (the carried shield's, plus defender mods on shield_dr)
              dr = pierce_dr(dr) if the type also has armor_pierce_partial
              damage = damage - dr
7. resist     damage = damage - resist[damage_type]
8. hp         lost = max(damage, 0); hp = max(hp - lost, 0)
9. riders     on a hit (not evaded, not blocked), the called shot's riders are listed for the caller
```
Each subtraction stops at zero before the next step: a step never raises damage. Dice and rolls are inputs, so a vector pins them. The reference does not roll. `window` is the defender's crit window from [resolution.md](resolution.md) (`crit_window + LCK // crit_window_per_two_luck`).

## Reading the steps

- **One roll per attack.** Owner request, 2026-11-16 (M3-27): the maths at the table is the attacker's d100 and one DR subtraction. The defender rolls nothing. Their defense is a band of the attacker's margin: a success that lands too close to the target is evaded (or blocked) instead of a hit, so "dodged" and "blocked" stay their own outcomes, apart from "missed". This reverses two owner decisions of 2026-11-04: the defender's reaction roll with a margin comparison, and dropping the passive penalty on the attacker's target (it returns as a band, not a penalty: the target is still only the attacker's numbers). [V 2026-11-16, owner decision]
- **The evasion rate is always on.** `evasion_rate` is half the Evasion skill, rounded down (`pipeline.json`), after the defender's modifiers on `evasion`, so every point bought in Evasion widens the band by half a number. Armor and encumbrance penalties are modifiers on `evasion` and narrow it [U: not in the vectors yet]. A defender with no `evasion_skill` has a band of 0. The tree is a formula, so a module may change it. [A] The dropped `10 + evasion_skill * 0.15 + AGI * 2` stays dropped: skill points barely moved it.
- **A crit always hits.** The crit window sits at the bottom of the d100, far from the target, so a band never reaches it in practice; the rule says so outright for a low target. [A]
- **Delivery, not damage type, decides the dodge.** `single` is one target: a sword cut, an arrow, a fire bolt, a touch spell; in the evaded band it does no damage. `area` is a radius or a cone: a fireball, a breath weapon, a thrown bomb; in the evaded band it is halved (rounded down), dodging or not (owner, 2026-11-16). `instant` meets no band: the effect lands, and a save the spell names decides how it behaves ([casting.md](casting.md#delivery-and-saves)). Physical and magic attacks follow the same rule. [V 2026-11-04, owner decision] A spell effect's `target` shape maps to it: `touch` and `ranged` are `single`, `area` is `area`. [A]
- **Stances are declared actions.** `defender.stance` is `dodge` or `block`, declared on the defender's turn ([actions.md](actions.md)) and in force until their next turn; without one the defender has only the passive band and the shield's DR. **Dodge** (3 AP) doubles the evasion rate. **Block** (2 AP, reserved, refunded if unused) adds a block band of half the `shields` skill below the evaded band, and needs a carried shield (`shield_dr`). A block band applies only to types with `armor_dr_applies`: a shield does not stop fire unless the defender has Spell Ward, which gives a block band of half the width against magic. A blocked hit does no damage. [V 2026-11-16, owner pick: option a for dodge; the block band is the option a this issue names]
- **A carried shield is DR.** `defender.shield_dr` (with the defender's modifiers on `shield_dr`) joins the worn armor DR against types with `armor_dr_applies`, with no action and no roll. A piercing hit halves the total, shield included. [A]
- **Balance between dodge and block.** Dodging costs more AP and is narrowed by armor; blocking needs a shield and stops only physical hits without a perk; a carried shield always helps through its DR. [A]
- **Attribute modifier.** An attribute's modifier is `attribute − 5` (`formulas/attribute_modifier.json`): +5 at 10, 0 at 5, −4 at 1. `damage.ability` names the attribute the attack uses, read from `attacker.attributes`; with no `ability` the roll stands alone. The modifier adds before the called-shot multiplier and the crit, so both multiply it [A], and a penalty stops at zero. Which attribute a weapon uses is in [attributes.md](../modules/attributes.md#the-attribute-modifier). [V 2026-11-11, vectors in `combat.json`]
- **Crit** multiplies damage by 2 (`crit_damage` in `pipeline.json`). It is a consumer of the `crit` tag, not a second roll. Crit multiplies before any reduction. [A]
- **Armor DR** is the sum of worn pieces ([armor.md](../modules/armor.md)); the pipeline takes the sum as an input. Types tagged `bypasses_armor` or `bypasses_physical_armor` carry no `armor_dr_applies`, so DR skips them. [A]
- **Resist** is the contract's `resist` operation: a flat subtraction from a packet whose tag matches the damage type. [V 2026-11-02, read contract.md]
- **HP** floors at 0. What happens at 0 (down, dead) is not in this file. [U]

## Block and dodge perks

The defender's perks reach the pipeline as `defender.modifiers`, stacked by [modifier.md](modifier.md): `add` on `evasion` and `shields` (the skills the bands read) and on `shield_dr`, and three `grant` flags. The rows are perks in `ruleset/data/features/core.json`, each gated by a skill threshold and the perk below it. [V 2026-11-06]

| Ladder | Perk | Emits | Effect |
| --- | --- | --- | --- |
| Block | Shield Training | `add shield_dr 2` | the carried shield subtracts 2 more |
| Block | Spell Ward | `grant block_magic` | while blocking, a block band of half the width stops a single or area attack of a type without `armor_dr_applies` (fire, a force dart) |
| Block | Bulwark | `grant bulwark` | while blocking, a block in the band's lowest rolls (as wide as the defender's crit window) also staggers the attacker: `stagger_ap` (2) AP less on their next turn |
| Dodge | Sidestep | `add dodge -1` | the dodge action costs 1 AP less ([actions.md](actions.md)) |
| Dodge | Armored Evasion | `add evasion 10` | offsets the armor penalties on the evasion skill: the band widens by 5 |
| Dodge | Riposte | `grant riposte` | while dodging, an evade of a `single` attack in the band's lowest rolls (as wide as the defender's crit window) sets `counter`: the defender may make one attack at no AP cost; the caller runs it as its own attack |

- **Instant stays out.** An `instant` attack meets no band, so no perk blocks or evades it; its defense is the save ([casting.md](casting.md#delivery-and-saves)). [V 2026-11-06, owner rule in M2-06]
- **Perks hook on the stances.** With no defender roll there is no crit reaction, so Riposte and Bulwark read a window of the band instead, and only in their stance (owner pick, 2026-11-16). The band's lowest rolls are its highest margins, next to the hit band: the attacker only just failed to hit clean. [A]
- **The caller applies the stagger.** `stagger_ap` is the AP the attacker loses on their next turn; the pipeline reports it and the caller spends it, as it runs a riposte's `counter`. [A]

## Situations and called shots

Cover, flanking, prone, unseen, and shooting into melee are rows in `ruleset/data/combat/situations.json`; called shots are rows in the same file. Each row is plain modifiers on the attack target, so they stack under the one rule ([modifier.md](modifier.md)). Resolution reads the attack modifiers through the same stacking: two cover rows (both origin `environment`) keep the stronger penalty, half and three-quarters cover give −20, not −30. [V 2026-11-06]

- **Range.** An attack is `melee` (the default) or `ranged`. A situation row lists its modifiers per range; a range it does not list gets nothing.
- **Explicit or flag.** Cover, flanking, and shooting into melee are named by the caller (`situations`), since positions are the caller's. Prone, blinded, and invisible come from flags (`grant` modifiers of the conditions): a row with `when` applies when the attacker or the defender has that flag. [A]
- **Forbidden.** Full cover forbids ranged attacks, and a prone attacker cannot shoot. The pipeline returns an error; the caller picks another action. [A]
- **Unseen.** A blinded attacker and an invisible target share the source `situation:unseen`, so both together are still −20. Area attacks ignore it (`deliveries`). [A]
- **Called shots.** The location's `attack` is an `add` on the attack target; `damage_multiplier` multiplies the damage dice before the crit doubles it (a head crit is ×1.5, then ×2), rounded down. Its `riders` (staggered, slowed, a disarm on a failed STR check, a stun on a failed END check) are listed on a hit; the caller rolls the checks and applies the conditions. [A]

## Not covered

Opportunity attacks, armor and encumbrance penalties on the evasion rate in the vectors, hit quality (`solid` and `glancing` margins), condition riders on a crit, range bands for called shots (the extra −10 beyond short range), the perks that halve called-shot penalties, and the video-game timing windows. Difficult terrain is movement, not an attack term: it belongs to an optional positioning module, and standing up is an action row (2 AP). They add terms to the same steps; none changes the order. [U]
