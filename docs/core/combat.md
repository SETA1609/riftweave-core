# Combat pipeline

Modified: 2026-11-23

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
                            defender armor none: E += mods on unarmored_evasion_band (Unarmored Grace)
              block_rate    B = 0, or with stance block: block_rate(shields_skill + mods on shields)
                            + mods on block_band (Shield Wall)                               formula tree
                            a type without armor_dr_applies: B = 0, unless Spell Ward (tier > 0)
              crit          a crit always hits: no band
              evaded        margin < E           single: no damage, stop   area: halved at step 5
              blocked       E <= margin < E + B  no damage, stop
              warded        E <= margin < E + B, a type without armor_dr_applies: a hit,
                            ward_percent = min(spell_ward_percent(END, tier), 50)      formula tree
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
              if warded: damage = damage - floor(damage * ward_percent / 100)
8. hp         lost = max(damage, 0); hp = max(hp - lost, 0)
9. riders     on a hit (not evaded, not blocked), the called shot's riders are listed for the caller,
              and a maneuver's effects and events; with a maneuver the result also gives its ap and
              stamina; a maneuver with a save_bonus adds its attempt's save (actions.md#the-maneuver-attempt)
              rider: on a plain attack, a held rider perk's maneuver that fits it (effects, events,
              and a save with no bonus) on any hit (actions.md#rider-perks)
              weapon_effects: the weapon's lingering_damage (of the attack's damage type) and, against
              a defender with the summoned flag, banish_summoned
```
Each subtraction stops at zero before the next step: a step never raises damage. Dice and rolls are inputs, so a vector pins them. The reference does not roll. `window` is the defender's crit window from [resolution.md](resolution.md) (`crit_window + LCK // crit_window_per_two_luck`).

## Reading the steps

- **One roll per attack.** Owner request, 2026-11-16 (M3-27): the maths at the table is the attacker's d100 and one DR subtraction. The defender rolls nothing. Their defense is a band of the attacker's margin: a success that lands too close to the target is evaded (or blocked) instead of a hit, so "dodged" and "blocked" stay their own outcomes, apart from "missed". This reverses two owner decisions of 2026-11-04: the defender's reaction roll with a margin comparison, and dropping the passive penalty on the attacker's target (it returns as a band, not a penalty: the target is still only the attacker's numbers). [V 2026-11-16, owner decision]
- **The evasion rate is always on.** `evasion_rate` is half the Evasion skill, rounded down (`pipeline.json`), after the defender's modifiers on `evasion`, so every point bought in Evasion widens the band by half a number. Armor and encumbrance penalties are modifiers on `evasion` and narrow it [U: not in the vectors yet]. A defender with no `evasion_skill` has a band of 0. The tree is a formula, so a module may change it. [A] The dropped `10 + evasion_skill * 0.15 + AGI * 2` stays dropped: skill points barely moved it.
- **A crit always hits.** The crit window sits at the bottom of the d100, far from the target, so a band never reaches it in practice; the rule says so outright for a low target. [A]
- **Delivery, not damage type, decides the dodge.** `single` is one target: a sword cut, an arrow, a fire bolt, a touch spell; in the evaded band it does no damage. `area` is a radius or a cone: a fireball, a breath weapon, a thrown bomb; in the evaded band it is halved (rounded down), dodging or not (owner, 2026-11-16). `instant` meets no band: the effect lands, and a save the spell names decides how it behaves ([casting.md](casting.md#delivery-and-saves)). Physical and magic attacks follow the same rule. [V 2026-11-04, owner decision] A spell effect's `target` shape maps to it: `touch` and `ranged` are `single`, `area` is `area`. [A]
- **Stances are declared actions.** `defender.stance` is `dodge` or `block`, declared on the defender's turn ([actions.md](actions.md)) and in force until their next turn; without one the defender has only the passive band and the shield's DR. **Dodge** (3 AP) doubles the evasion rate. **Block** (2 AP, reserved, refunded if unused) adds a block band of half the `shields` skill below the evaded band, and needs a carried shield (`shield_dr`). A block band stops only types with `armor_dr_applies`: a shield does not stop fire. With Spell Ward a magic hit in the band is warded: it lands, less a share of its damage (see the perks below). A blocked hit does no damage. [V 2026-11-16, owner pick: option a for dodge; the block band is the option a this issue names]
- **A carried shield is DR.** `defender.shield_dr` (with the defender's modifiers on `shield_dr`) joins the worn armor DR against types with `armor_dr_applies`, with no action and no roll. A piercing hit halves the total, shield included. [A]
- **Balance between dodge and block.** Dodging costs more AP and is narrowed by armor; blocking needs a shield and stops only physical hits (Spell Ward softens magic in the band); a carried shield always helps through its DR. [A]
- **Attribute modifier.** An attribute's modifier is `attribute − 5` (`formulas/attribute_modifier.json`): +5 at 10, 0 at 5, −4 at 1. `damage.ability` names the attribute the attack uses, read from `attacker.attributes`; with no `ability` the roll stands alone. The modifier adds before the called-shot multiplier and the crit, so both multiply it [A], and a penalty stops at zero. Which attribute a weapon uses is in [attributes.md](../modules/attributes.md#the-attribute-modifier). [V 2026-11-11, vectors in `combat.json`]
- **Crit** multiplies damage by 2 (`crit_damage` in `pipeline.json`). It is a consumer of the `crit` tag, not a second roll. Crit multiplies before any reduction. [A]
- **Armor DR** is the sum of worn pieces ([armor.md](../modules/armor.md)); the pipeline takes the sum as an input. Types tagged `bypasses_armor` or `bypasses_physical_armor` carry no `armor_dr_applies`, so DR skips them. [A]
- **Resist** is the contract's `resist` operation: a flat subtraction from a packet whose tag matches the damage type. [V 2026-11-02, read contract.md]
- **HP** floors at 0. What happens at 0 (down, dead) is not in this file. [U]

## Block and dodge perks

The defender's perks reach the pipeline as `defender.modifiers`, stacked by [modifier.md](modifier.md): `add` on `evasion`, `shields`, `block_band`, `unarmored_evasion_band`, `spell_ward`, and `shield_dr`, and `grant` flags. The rows are perks in `ruleset/data/features/core.json`. Since GD_S05_03 most are **tiered**: one perk with `ranks` and `rank_levels`, the level each tier opens at, and taking it again raises the tier ([progression.md](progression.md)). A tier emits the perk's value times the tier, or for Shield Training the tier's own value in `rank_values`. [V 2026-11-19]

| Ladder | Perk | Tiers (levels) | Emits per tier | Effect |
| --- | --- | --- | --- | --- |
| Block | Shield Training | 1, 6, 12 | `add shield_dr` 2, 3, 4 | the carried shield subtracts 2, 3, then 4 more |
| Block | Shield Wall | 1, 6, 12 | `add block_band 3` | the block band widens by 3, 6, then 9 |
| Block | Spell Ward | 6, 12, 18 | `add spell_ward 1` | while blocking, a magic hit in the block band is warded: it lands, less END × 2 × tier % of its damage, at most 50 % (`spell_ward_table` in `pipeline.json`) |
| Block | Bulwark | one (Shield Training tier 2) | `grant bulwark` | while blocking, a block in the band's lowest rolls (as wide as the defender's crit window) also staggers the attacker: `stagger_ap` (2) AP less on their next turn |
| Block | Guardian | 6, 12, 18 | `add guardian 1` | while blocking, any shield: one adjacent ally's evasion band +10 / the block band also covers attacks on that ally / it covers every adjacent ally; pending, the kernel has no allies yet [U] |
| Block | Boss Strike | 1, 6, 12, 18 | `add boss_strike 1` | the shield joins the attack as a bash ([the bash](#the-shield-bash)) |
| Dodge | Sidestep | one | `add dodge -1` | the dodge action costs 1 AP less ([actions.md](actions.md)) |
| Dodge | Armored Evasion | 6, 12, 18 | `add evasion 5` | offsets the armor penalties on the evasion skill: +5, +10, +15 |
| Dodge | Unarmored Grace | 1, 6, 12, 18 | `add unarmored_evasion_band 5` | with no body armor (`defender.armor` `none`), the evasion band widens by 5 a tier |
| Dodge | Flowing Step | one (Unarmored Grace tier 2) | `grant flowing_step` | moving out of reach draws no opportunity attack; pending: the kernel does not track a turn's first move [U] |
| Dodge | Riposte | one | `grant riposte` | while dodging, an evade of a `single` attack in the band's lowest rolls (as wide as the defender's crit window) sets `counter`: the defender may make one attack at no AP cost; the caller runs it as its own attack |

The Spell Ward table, END down and tier across (percent of the damage warded):

| END | 1 | 2 | 3 |
| --- | --- | --- | --- |
| 1 | 2 | 4 | 6 |
| 2 | 4 | 8 | 12 |
| 3 | 6 | 12 | 18 |
| 4 | 8 | 16 | 24 |
| 5 | 10 | 20 | 30 |
| 6 | 12 | 24 | 36 |
| 7 | 14 | 28 | 42 |
| 8 | 16 | 32 | 48 |
| 9 | 18 | 36 | 50 |
| 10 | 20 | 40 | 50 |

- **The ward rounds down, after resist.** A warded hit takes resist first, then loses `floor(damage × ward_percent / 100)`. A crit is never warded, as it is never blocked. [A, GD_S05_03]
- **Gone in GD_S05_03.** Acrobat (its `dodge_stamina` had no consumer) and the half-width magic block band (`block_magic`, `magic_block_rate`). Battle Mage is Mailed Casting ([casting.md](casting.md#armor-and-shields)). [A]

### The shield bash

Boss Strike makes the carried shield (a buckler counts) a second weapon in the attack sequence ([attacks by loadout](#attacks-by-loadout)): `shield_bash` in `attack_sequences/core.json`. The off hand reads as a `light_weapon`, so the AP follow the light weapon row, and the bash rolls on the Shields skill (`off_skill`; a bash without it is an error, as without `shield_dr`, so it never rolls on the main-hand weapon's skill). Its damage is `shield_dr` plus the bash roll, bludgeoning. A shield has no weapon dice, so a bash is never the Scissor Cut strike. [V 2026-11-19]

| Tier | Bash |
| --- | --- |
| 1 | a jab: the light jab row, -10 to hit, 2 AP |
| 2 | the attack includes the bash as its off-hand strike |
| 3 | a bash that hits staggers: the target has `stagger_ap` (2) AP less on its next turn |
| 4 | after a block this turn (`after_block`), a bash jab costs 1 AP (`after_block_ap`) |

- The bash needs `attacker.shield_dr`; without it the sequence is an error. A strike that is a bash reports `bash: true`, and a staggering one `stagger_ap`. [A]

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
- **Each strike is a whole `combat` input.** A strike carries its hand's weapon: `attacker.weapon_properties` and `weapon_effects` for the main hand, `off_weapon_properties` and `off_weapon_effects` for the off hand (none when left out, never the main hand's), so a bound weapon's bane or lingering damage lands per strike, and a strike lists the effects that landed when its hand has any. `situations` and `summoning` go to every strike, as in `combat` (flanking, `beside_summon` for Pincer Strike), and so do `attacker.perks`: a rider perk rides every strike it fits, each its own attempt, and the strike reports `rider`, `effects`, `events`, and `save` as `combat` does ([rider perks](actions.md#rider-perks)). [V 2026-11-19]
- **AP follows the loadout.** A two-handed attack stays 3 AP, because an AGI 3 character has only 3 AP; a two-handed weapon pays for itself by leaving no hand for a shield. A maneuver's AP is the loadout's attack AP plus its own when the `combat` input names a `loadout`. [A]
- **Numbers are [A]** until the M5-18 playtest driver reruns a dual-wield build against a one-weapon build of the same level.

| Ladder | Perk | Level | Emits | Effect |
| --- | --- | --- | --- | --- |
| Carry Through | Full Weight | 1 | `multiply melee_damage 1.25` | the old Power Attack bonus, open to Blades, Blunt, or Piercing 40 |
| Carry Through | Carry Through | 6 | `grant carry_through` | one carry per action |
| Carry Through | Reaping Arc | 12 | `grant reaping_arc` | a carry for every drop |
| Carry Through | Circle Cut | 18 | `grant can_circle_cut` | the circle cut action |
| Extra strikes | Follow-Up | 6 | `add full_attack_strikes 1` | a full attack strikes once more at -20, melee or ranged |
| Extra strikes | Press On | 18 | `add full_attack_strikes 1` | and once more at -30 |
| Two weapons | Off-Hand Drill | 4 | `add mainhand_penalty 10`, `add offhand_penalty 10` | both two-weapon penalties are 10 smaller |
| Two weapons | Crossed Guard | 8 | `add two_weapon_evasion_band 5` | the evasion band +5 while holding two weapons |
| Two weapons | Twin Rhythm | 14 | `add full_attack_off_strikes 1` | a second off-hand strike in a full attack |
| Two weapons | Scissor Cut | 20 | `grant scissor_cut` | the off-hand dice once more |

[V 2026-11-19, `attack_sequence.json` and `combat.json` vectors in Python and Go]

### Ranged loadouts

Owner request, 2026-11-18 (GD_S17_05): the full attack and the extra-strike perks work at range.
Four hand kinds join the vocabulary for ranged weapons: a main hand `bow`, `hand_gun` (a hand
crossbow or a pistol), `long_gun` (a two-handed crossbow, the arquebus, or the blunderbuss), or
`light_thrown` (a light thrown weapon), and an off hand `hand_gun` or `light_thrown`. The four are
`ranged_hands` in `attack_sequences/core.json`: a strike with one of them is a `ranged` strike
(`combat`'s `range`), so a situation reads its `ranged` row (shooting into melee). [A]

| Loadout | Attack AP | Strikes |
| --- | --- | --- |
| a bow | 3 | main 0 |
| a hand crossbow or one pistol, off hand empty, a shield, or a buckler | 2 | main 0 |
| a two-handed crossbow or a long gun | 3 | main 0 |
| two pistols or two hand crossbows | 3 | main -10, off -20 |
| one light thrown weapon in each hand | 3 | main -10, off -20 |
| a light or one-handed melee weapon and a pistol, a hand crossbow, or a light thrown weapon (the mixed hand) | 4 | main -10, off -20 |

```
foes[i].in_reach   false: the foe is in range, not in reach (default true)
aim                a melee strike goes to a foe in reach; a ranged strike to any foe
carry              a melee strike carries to the first standing foe in reach, a ranged one to the
                   first standing foe (carry_through and reaping_arc read "in range" at range)
ammunition         optional, per hand: { main: { rounds, loaded }, off: { rounds, loaded } }
                   each ranged strike of a counted hand spends one round
loading            a loading hand fires once, then reloads before its next strike
reload AP          the reload action's cost (2, Fast Reload -1), +1 when the other hand holds
                   anything (full_attack.one_handed_reload_extra)
in a full attack   a reload spends AP from the budget over the row's attack AP; a reload that
                   does not fit ends the full attack, and its strikes left are stopped
in an attack       a loading hand that is not loaded, or a hand with no rounds left, is an error
result             ammunition: the rounds left per counted hand; reloads; stopped
opportunity        a ranged main hand makes no opportunity attack (reason ranged)
```

- **The mixed hand** (owner 2026-11-18, the ranger): the melee strike goes to a foe in reach and
  the ranged strike to any target in range, so one attack can cut one foe and shoot another. A
  shot at a foe an ally is fighting takes the `shooting_into_melee` situation, as any shot does;
  it reads only on the ranged strike. GD_S17_13's perks soften the penalty and the reload. [A]
- **The extra-strike ladder is any weapon:** `follow_up` and `press_on` add main-hand strikes,
  melee or ranged; a loading weapon's added strike needs a reload first. `circle_cut` stays
  melee. [A]
- **No opportunity attack at range** (owner, 2026-11-23, from GD_S19_04): a shooter's answer to a
  moving foe is Trigger Watch (GD_S17_06); GD_S17_13's `ready_steel` is the perk exception.
- **Aimed and full:** a full attack cannot be aimed; the held-breath maneuver (GD_S17_06) is the
  aimed shot that spends a turn. [A]

## Fighting stances

Owner request, 2026-11-16 (GD_S05_02). A fighting stance is a toggle a perk unlocks, not a passive: the perk names `core:stance/<key>` in its `unlocks`, and the stance is a row of `ruleset/data/stances/core.json` with what the hands must hold and the modifiers it emits. The `stance` command takes, changes, and ends one, and drops it when the loadout stops fitting; its `modifiers` are what the caller passes to `combat` (as the defender's) and `attack_sequence` (as the attacker's).

```
one          at most one fighting stance is active [A]; taking another replaces it
cost         taking or changing one at the start of the actor's own turn: rules.turn_start_ap (0);
             later in the turn (mid_turn): rules.mid_turn_ap (1) [A]; taking the active one, or ending one, is free
lasts        until changed or ended; a loadout it no longer fits (a weapon drawn, dropped, or swapped) ends it;
             the event's by names the action that paid (draw 1, sheathe 1, drop 0, swap 1, actions.md#changing-the-loadout)
loadout      main and off from the loadout vocabulary (any when left out); main_skills and off_skills name the
             weapon's skill (the loadout's main_skill and off_skill) [A]
modifiers    origin core:origin/martial, source stance:<key>, duration while_applied; validate.py checks all three
with dodge   a fighting stance stands together with the dodge or block stance: those are declared defenses
or block     (actions.md#pool), this is how the hands are held [A]
parry        a parry grant reads the defender's weapon_skill: a parry band of floor(weapon_skill / rules.parry_band_divisor)
             (4) under the block band, physical hits only, none for an instant attack; a margin in it is blocked
             and parried; a crit always hits
```

| Stance | Perk level | Hands | Emits | Effect |
| --- | --- | --- | --- | --- |
| Fencer's Measure | 12 | a light or one-handed blade or piercing weapon, off hand empty | `grant parry`, `add attack_single_opponent 10` | a parry band; +10 on every strike while only one foe is in reach. Replaces Duelist |
| Espada y Daga | 8 | a blade or piercing weapon and a piercing light weapon (a dagger) | `grant parry`, `grant off_hand_parries`, `add attack_after_parry 10` | the dagger makes no strike in a full attack and parries; after a parry (`after_parry`) the first main-hand strike gains +10 |
| Buckler Play | 6 | a light or one-handed weapon and a buckler | `grant block_with_weapon` | while blocking, the block band reads the better of Shields and the weapon skill. The bash is `boss_strike` (GD_S05_03) |
| Phalanx | 8 | a piercing (spear) weapon and a shield | `add reactions_per_round 1`, `grant shield_cover_to_ally` | one more reaction a round, so one more opportunity attack ([opportunity attacks](#opportunity-attacks)); an adjacent ally gains the shield's cover, which is pending: no kernel step has cover yet [U] |
| Long Guard | 10 | a weapon in both hands | `grant full_weight_doubled`, `grant carry_on_crit`, `add evasion_band -10` | Full Weight counts twice (pending, as `melee_damage` is); Carry Through also fires on a crit that leaves its foe standing (in an attack, only for a next strike aimed at that foe); the evasion band -10 |
| Bare Knuckle | 6 | both hands empty | `add unarmed_damage 2`, `add full_attack_strikes 1` | +2 damage on each unarmed strike; one more strike in a full attack at -20 |
| Forward Guard | 4 | any | `add attack_roll 10`, `add evasion_band -10` | +10 on every strike; the evasion band -10 |
| Low Guard | 4 | any | `add attack_roll -10`, `add evasion_band 10` | -10 on every strike; the evasion band +10 |

- **The evasion band** term applies after the dodge stance doubles the band and after Crossed Guard, never below 0.
- **A crit carry** (Long Guard) spends a carry as a drop would: the next strike of the action, or a free strike after the last, goes to the first other foe still standing; a full attack's later strikes follow it.
- **Numbers are [A]** until the playtest driver runs a stance build.

[V 2026-11-19, `stance.json`, `combat.json`, and `attack_sequence.json` vectors in Python and Go]

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

## Opportunity attacks

Owner request, 2026-11-19 (GD_S19_04), changed 2026-11-23. An opportunity attack is one **melee** attack of the loadout, the same strikes at the same values, made outside your turn when a foe in your melee reach leaves it without disengaging. It is a **reaction**: it costs no AP, no stamina, and no guard, and spends one of the reactions you have between two of your turns (owner, 2026-11-23).

```
trigger      a foe in your melee reach leaves it, and did not disengage
reactions    reactions_used < most, else no attack (no_reaction_left)
most         reactions_per_round (1), or with reactions_from_agi max(1, AGI - 5),
             then + each reactions_per_round modifier (Phalanx +1)
costs        0 AP, 0 stamina, no guard
fires        one attack, the strikes of the row, as attack_sequence runs on your turn
not fired    nothing is spent: no reaction
```

- **A base rule, not a perk** (owner, 2026-11-19): opportunity attacks make reach, a held doorway, and Disengage matter. Disengage (1 AP) prevents one. A reach weapon (`attack_reach` 2) threatens 2 units, others 1 ([actions.md](actions.md#how-far-a-move-goes)).
- **Melee only** (owner, 2026-11-23): a ranged loadout makes no opportunity attacks; the shooter's answer to a foe moving in its range is overwatch. Ranged loadouts arrive with GD_S17_05, which refuses them an opportunity attack; GD_S17_13's `ready_steel` (draw a melee weapon for one) is the perk exception. [A]
- **Reactions** (owner, 2026-11-23): everyone has `reactions_per_round` (1, `actions/core.json`) between two of their turns. The caller counts them (`reactions_used`), as it tracks reach. Watchful Blade (level 6, AGI 7) grants `reactions_from_agi`: the most becomes the AGI modifier (`formulas/attribute_modifier.json`, AGI − 5), never below 1, so AGI 8 has 3. Phalanx adds 1 (`add reactions_per_round 1`) on top of either. A character built around reactions, high AGI with both perks, strikes every foe that slips past and keeps its AP for its own turn. Riposte is not a reaction and does not count. [A]
- **Perks extend them**, never grant them: Watchful Blade and Phalanx allow more a round; later perks may add triggers (casting, drinking, or standing up in reach) or a wider threat. [A]
- **At a table, not in real time** (owner, 2026-11-23): reactions are the turn-based rule. A real-time game fires an opportunity attack automatically when a foe leaves reach, or leaves opportunity attacks out; the engine decides, and the kernel keeps no clock for it.
- **Counter-attacks stay perks** (owner, 2026-11-19): Riposte (dodge stance only) is the only counter. A base counter would double the attacks in a fight, add a decision to every miss, and make dodging the best stance.

**Overwatch is not an opportunity attack** (owner, 2026-11-23). Overwatch is the ranged attacker's: AP held on its turn fires when an enemy moves within its weapon's range. It is paid from held AP, spends no reaction, and needs the ranged loadouts: it is the Trigger Watch maneuver of GD_S17_06_ranged-maneuvers.

The `attack_sequence` command runs it as the action `opportunity`, with `disengaged` and `reactions_used`, and returns `fired`, the `reason` it did not fire (`disengaged` or `no_reaction_left`), and `ap` and `stamina` 0. [V 2026-11-23, `attack_sequence` vectors in Python and Go]

## Ammunition

Owner request, 2026-11-18 (GD_S17_02). A weapon with the `ammunition` property fires rounds of one
**kind**, which its base names in `weapon.ammo`: `arrow` (bows), `bolt` (crossbows), `bullet`
(ball and powder: the pistol and the arquebus), `scatter` (the blunderbuss's charge), or `stone`
(the sling). The kinds are one vocabulary, `ammoKind` in `schema.json`. [A]

**Ammunition rows** are equipment items of type `ammunition` in `ruleset/data/equipment/ammunition.json`,
so a row is `core:ammunition/<key>`. Each names its `kind` and the `count` a bundle holds: 20
arrows, 20 bolts, 10 bullets, 10 scatter charges, 20 sling stones. A row may change the strike's
damage type (`damage_type`) and add effects on a hit (`effects`, `appliedEffect`); plain
ammunition changes nothing. Elemental rounds (GD_S17_04) and conjured rounds (GD_S17_11) are such
rows. [A]

```
combat input   ammunition: { rounds, loaded }      optional, for a weapon that fires or throws
               rounds    the rounds of the weapon's kind the attacker carries (a thrown weapon: how many)
               loaded    a loading weapon: whether it holds a round
refused        rounds 0                            no attack: "the weapon has no rounds left"
               loading and not loaded              no attack: "the weapon must reload first"
result         ammunition: { rounds: rounds - 1, loaded: false for a loading weapon }
```

- **Spending:** each strike with an `ammunition` weapon spends one round of its kind; a `thrown`
  weapon is its own ammunition and spends itself. Without the `ammunition` input the attack runs
  as before: a table that does not count plain rounds leaves it out. [A]
- **Loading:** a `loading` weapon (the hand crossbow, the pistol, the arquebus, the blunderbuss)
  holds one round and must reload after every shot: the `reload` action, 2 AP
  ([actions.md](actions.md#costs)), which Fast Reload lowers by 1 (`add reload -1`). A weapon
  without `loading` draws as part of its attack. The repeating crossbow has no `loading`: its
  magazine feeds each attack, and its size is not counted. [A]
- **Recovery:** after a fight, the `ammunition` command's `recover` gives back the share of each
  kind that `ruleset/data/combat/ammunition.json` names, rounded down: half the arrows and bolts,
  none of the bullets, scatter, and stones, and every thrown weapon. GD_S17_10's perks change that
  row, so it is data, not prose. [A]
- **At a table:** counting rounds is the rule; a table may stop counting plain ammunition and count
  only the special rounds, by leaving out the `ammunition` input for plain ones. [A]

## Grenades

Owner request, 2026-11-18 (GD_S17_03). A **grenade** is a consumable of type `grenade`, delivery
`throw`, in `ruleset/data/equipment/consumables.json` (`core:consumable/<key>`). It is thrown with
`throwing_weapons` as an `area` attack at a point in range (`range_normal` 6, `range_long` 18), and
every creature within its `radius` of that point takes the area roll above: delivery, not damage
type, decides the dodge. A throw costs the `attack` AP, not `item`, so it competes with a shot. [A]

```
combat input   grenade          core:consumable/<key>, a row of type grenade
               burst_distance   squares from the burst point to this defender; default 0
               damage.roll      the grenade's dice, rolled; damage.type may be left out
refused        damage.ability, a maneuver, a called shot, or ammunition with a grenade
               a delivery other than area; a grenade with no damage (the Smoke Pot)
               a roll outside the grenade's dice, or a damage type that is not the grenade's
result         outcome outside   burst_distance over the radius: no roll, no damage
               effects           the grenade's riders on a hit, empty on a miss or an evade
```

- **Flasks are the weak end** (owner, 2026-11-23): an alchemist's flask deals 1d6, below an average
  weapon's die; real grenades come later (their own issue) and must out-damage a flask. [A]
- **Fixed damage:** the grenade's own dice and damage type, with no attribute, skill, material, or
  weapon perk added. The thrower's skill decides the hit, not the damage; only the thrower perks of
  GD_S17_09 change it. [A]
- **Riders:** each grenade's `effects` are effects from the shared pool with the `grenade`
  channel, which `validate.py` checks (`CHANNEL_REQUIREMENTS`). They land on a hit that is not
  evaded, as a maneuver's do. [A]
- **The Smoke Pot** deals no damage, so it is no attack: its effect fills the area and blocks
  sight for 3 turns (`smoke_screen`, 15 seconds). [A]
- **No toxicity:** a grenade is thrown, not drunk. Merchants stock them by level; crafting them
  belongs to the alchemy module (GD_S15_02). [A]

| Grenade | Damage | Radius | On a hit |
| --- | --- | --- | --- |
| Fire Flask | 1d6 fire | 1 | `burn` |
| Frost Flask | 1d6 frost | 1 | `slow` |
| Thunder Flask | 1d6 thunder | 2 | `stagger` |
| Acid Flask | 1d6 acid | 1 | `sunder_armor` |
| Smoke Pot | none | 2 | `smoke_screen`, 3 turns |

## Not covered

Armor and encumbrance penalties on the evasion rate in the vectors, hit quality (`solid` and `glancing` margins), condition riders on a crit, range bands for called shots (the extra −10 beyond short range), the perks that halve called-shot penalties, and the video-game timing windows. Difficult terrain is movement, not an attack term: it belongs to an optional positioning module, and standing up is an action row (2 AP). They add terms to the same steps; none changes the order. [U]
