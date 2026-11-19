# Combat pipeline

Modified: 2026-11-19

One attack, in order, from the roll to the HP lost. Every step reads the actor snapshot and data; none calls another system ([contract.md](contract.md)). The numbers are in `ruleset/data/combat/pipeline.json` (formula trees), `ruleset/data/combat/damage_types.json` (tags), and `ruleset/data/resolution/core.json` (the d100 function). [`docs/modules/combat.md`](../modules/combat.md) keeps the TTRPG and video-game commentary and the worked examples; this file is the order of operations they share. An attack is a weapon strike or an offensive spell; both run this order, and the attack's `delivery` decides which defense bands it meets. A cast's damage runs steps 2 to 8 with the cast check as its attack roll ([casting.md](casting.md#damage-to-a-target), M3-18). Save rolls for instant spells are in [casting.md](casting.md#delivery-and-saves); the magic layers (absorption, phase resistance) are not specified here. [U]

## Order

```
0. setup      forbidden: a situation that forbids the range is an error  situations.json
              attack mods += every situation that applies (rows for the range,
              the delivery, and the attacker's and defender's flags)
              attack mods += the called shot's attack row               situations.json
              maneuver: an error when its ranges or weapon properties do not fit;
              attack mods += its attack; ap = attack cost + its ap      maneuvers/core.json
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
              if a maneuver: damage = floor(damage * the maneuver's damage_multiplier)
              bane: damage += magnitude of each weapon bane whose parameter is the defender's creature_type
4. crit       if crit: damage = crit_damage(damage)                 formula tree
5. area       if evaded an area attack: damage = area_evade_damage(damage)   formula tree
6. DR         if the damage type has tag armor_dr_applies:
              dr = armor_dr + shield_dr (the carried shield's, plus defender mods on shield_dr)
              dr = pierce_dr(dr) if the type also has armor_pierce_partial
              dr = dr - floor(dr * pierce / 100), pierce = the weapon's armor_pierce magnitudes, at most 100
              damage = damage - dr
7. resist     damage = damage - resist[damage_type]
8. hp         lost = max(damage, 0); hp = max(hp - lost, 0)
9. riders     on a hit (not evaded, not blocked), the called shot's riders are listed for the caller,
              and a maneuver's effects; with a maneuver the result also gives its ap and stamina
              weapon_effects: the weapon's lingering_damage (of the attack's damage type) and, against
              a defender with the summoned flag, banish_summoned
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

## Attacks by loadout

Owner request, 2026-11-16 and 2026-11-17 (GD_S05_01). An attack action strikes with everything held: once with the main hand and once with an off-hand weapon, if there is one. Each strike is its own roll through the order above (one `combat` run per strike), and the `attack_sequence` command runs the action. The loadout is what the two hands hold, in the `mainHand` and `offHand` vocabulary of `ruleset/schemas/schema.json`: a main hand `empty`, `light`, `one_handed`, or `two_handed` (a two-handed weapon, or a versatile one in both hands), and an off hand `empty`, `shield`, `buckler`, `light_weapon`, `weapon`, or `two_handed` (the main hand's grip). [A]

How strikes stack is a table, `ruleset/data/attack_sequences/core.json`, not prose:

| Loadout | Attack AP | Strikes | Jab AP (off hand only) |
| --- | --- | --- | --- |
| both hands empty | 2 | main 0 | - |
| one light weapon, off hand empty, a shield, or a buckler | 2 | main 0 | - |
| one one-handed weapon, off hand empty, a shield, or a buckler | 3 | main 0 | - |
| one weapon in both hands | 3 | main 0 | - |
| two light weapons | 3 | main -10, off -20 | 2, at -10 |
| a one-handed main and a light off hand (sword and dagger) | 4 | main -10, off -20 | 2, at -10 |
| two one-handed weapons (a light main counts as one-handed here) | 5 | main -10, off -30 | 3, at -20 |

```
attack       the row's AP; its strikes in order; each strike may go to its own foe in reach (targets)
jab          the off-hand weapon alone, at the row's jab AP and value; a row without an off-hand weapon has none
full_attack  spends every AP left, at least the row's attack AP + full_attack.min_extra_ap (1), against one foe:
             the row's strikes, then full_attack_strikes main-hand strikes at extra_main_attacks (-20, -30),
             then, with two weapons and full_attack_off_strikes, one off-hand strike at the off value + extra_off_attack (-10)
circle_cut   with can_circle_cut and a two_handed main hand: every AP left (at least attack + 1), one main-hand
             strike at circle_cut.attack (-10) at every foe in reach; nothing carries
penalties    mainhand_penalty and offhand_penalty add to a negative value of the table (attack and jab), never past 0;
             the strikes perks add keep their own values
drop         a strike whose foe is down is lost, unless a carry is left: it goes to the first foe in reach still
             standing (carried), and a full attack's later strikes follow it
carries      carry_through: one per action; reaping_arc: one for every drop; none without them
free strike  when the action's last strike drops a foe and a carry is left, one more strike with the same hand and
             value goes to the next foe still standing (free)
scissor      scissor_cut, full attack only: the first off-hand strike that hits a foe a main-hand strike of the same
             full attack already hit rolls its damage dice twice (the damage roll doubled), once per full attack
crossed      a defender holding a weapon in each hand adds its modifiers on two_weapon_evasion_band to the
             evasion band, after the dodge stance doubles it
```

- **Rolls are inputs.** `rolls` holds one attack roll and one damage roll per strike made, in order; a lost strike takes none. Too few rolls is an error. [A]
- **Each strike is a whole `combat` input.** A strike carries its hand's weapon: `attacker.weapon_properties` and `weapon_effects` for the main hand, `off_weapon_properties` and `off_weapon_effects` for the off hand (none when left out, never the main hand's), so a bound weapon's bane or lingering damage lands per strike, and a strike lists the effects that landed when its hand has any. `situations` and `summoning` go to every strike, as in `combat` (flanking, `beside_summon` for Pincer Strike). [V 2026-11-19]
- **AP follows the loadout.** A two-handed attack stays 3 AP, because an AGI 3 character has only 3 AP; a two-handed weapon pays for itself by leaving no hand for a shield. A maneuver's AP is the loadout's attack AP plus its own when the `combat` input names a `loadout`. [A]
- **Numbers are [A]** until the M5-18 playtest driver reruns a dual-wield build against a one-weapon build of the same level.

| Ladder | Perk | Level | Emits | Effect |
| --- | --- | --- | --- | --- |
| Carry Through | Full Weight | 1 | `multiply melee_damage 1.25` | the old Power Attack bonus, open to Blades, Blunt, or Piercing 40 |
| Carry Through | Carry Through | 6 | `grant carry_through` | one carry per action |
| Carry Through | Reaping Arc | 12 | `grant reaping_arc` | a carry for every drop |
| Carry Through | Circle Cut | 18 | `grant can_circle_cut` | the circle cut action |
| Extra strikes | Follow-Up | 6 | `add full_attack_strikes 1` | a full attack strikes once more at -20 |
| Extra strikes | Press On | 18 | `add full_attack_strikes 1` | and once more at -30 |
| Two weapons | Off-Hand Drill | 4 | `add mainhand_penalty 10`, `add offhand_penalty 10` | both two-weapon penalties are 10 smaller |
| Two weapons | Crossed Guard | 8 | `add two_weapon_evasion_band 5` | the evasion band +5 while holding two weapons |
| Two weapons | Twin Rhythm | 14 | `add full_attack_off_strikes 1` | a second off-hand strike in a full attack |
| Two weapons | Scissor Cut | 20 | `grant scissor_cut` | the off-hand dice once more |

[V 2026-11-19, `attack_sequence.json` and `combat.json` vectors in Python and Go]

## Weapon effects

`attacker.weapon_effects` are the on-hit effects of the weapon, such as a bound weapon's perk
enchantments ([casting.md](casting.md#bound-weapons)). Each is an applied effect:

| Effect | Step | What it does |
| --- | --- | --- |
| `armor_pierce` | 6 | ignores `magnitude` percent of the armor and shield DR, rounded down in the defender's favour; several add, at most 100 |
| `bane` | 3 | `+magnitude` damage when its `parameter` is the defender's `creature_type` (`undead`, `beast`, ...); a bane with no `parameter` adds nothing |
| `lingering_damage` | 9 | listed in `weapon_effects` on a hit, with the attack's damage type as its parameter (the attack's type always wins, even over a parameter the input carries; a bound weapon's own effect already carries its type): damage over time for the caller to run |
| `banish_summoned` | 9 | listed on a hit when the defender has the `summoned` flag; the caller ends that summon. Against a creature that was not summoned it does nothing |

On a miss, an evade, or a block `weapon_effects` is empty; without weapon effects the field is
left out. Other effects in the list are passed through on a hit, for the caller. [A]

## Situations and called shots

Cover, flanking, prone, unseen, and shooting into melee are rows in `ruleset/data/combat/situations.json`; called shots are rows in the same file. Each row is plain modifiers on the attack target, so they stack under the one rule ([modifier.md](modifier.md)). Resolution reads the attack modifiers through the same stacking: two cover rows (both origin `environment`) keep the stronger penalty, half and three-quarters cover give −20, not −30. [V 2026-11-06]

- **Range.** An attack is `melee` (the default) or `ranged`. A situation row lists its modifiers per range; a range it does not list gets nothing.
- **Explicit or flag.** Cover, flanking, and shooting into melee are named by the caller (`situations`), since positions are the caller's. Prone, blinded, and invisible come from flags (`grant` modifiers of the conditions): a row with `when` applies when the attacker or the defender has that flag. [A]
- **Beside a summon.** `beside_summon` is named by the caller when a summon of the attacker attacks the same target. Its row has no modifier of its own but `adds: beside_summon_attack`: while a row with `adds` applies, the attacker's modifiers on that target join the attack at any range; outside it they add nothing ([magic.md](../modules/magic.md#the-hybrid-summoner)). [A]
- **Forbidden.** Full cover forbids ranged attacks, and a prone attacker cannot shoot. The pipeline returns an error; the caller picks another action. [A]
- **Unseen.** A blinded attacker and an invisible target share the source `situation:unseen`, so both together are still −20. Area attacks ignore it (`deliveries`). [A]
- **Called shots.** The location's `attack` is an `add` on the attack target; `damage_multiplier` multiplies the damage dice before the crit doubles it (a head crit is ×1.5, then ×2), rounded down. Its `riders` (staggered, slowed, a disarm on a failed STR check, a stun on a failed END check) are listed on a hit; the caller rolls the checks and applies the conditions. [A]

## Not covered

Opportunity attacks, armor and encumbrance penalties on the evasion rate in the vectors, hit quality (`solid` and `glancing` margins), condition riders on a crit, range bands for called shots (the extra −10 beyond short range), the perks that halve called-shot penalties, and the video-game timing windows. Difficult terrain is movement, not an attack term: it belongs to an optional positioning module, and standing up is an action row (2 AP). They add terms to the same steps; none changes the order. [U]
