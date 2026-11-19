# Combat

Modified: 2026-11-19

Commentary on the combat rules: what they are for, how they fit together, and worked examples. It holds no rule of its own. Every rule here is owned by a file in `docs/core/`, has its numbers in `ruleset/data/`, and is pinned by a vector in `ruleset/tests/vectors/`. **If this file and `docs/core/` disagree, `docs/core/` wins** and this file is the bug.

There is one set of rules. A turn-based table, a party-based cRPG, and a real-time action game run the same functions on the same numbers; a real-time client only presents them differently (§11). There is no bonus action, no reaction slot, and no per-mode column.

## Who owns what

| Rule | Owner | Data | Vectors |
| --- | --- | --- | --- |
| The d100 check, crit and fumble tags | [resolution.md](../core/resolution.md) | `resolution/core.json` | `resolution.json` |
| The turn: one AP pool, action costs | [actions.md](../core/actions.md) | `actions/core.json` | `actions.json` |
| One attack, from the roll to the HP lost | [combat.md](../core/combat.md) | `combat/pipeline.json`, `combat/damage_types.json` | `combat.json` |
| Modifier stacking, conditions | [modifier.md](../core/modifier.md), [contract.md](../core/contract.md) | `conditions/core.json` | `modifiers.json` |
| HP, MP, AP, carry, initiative | [derived.md](../core/derived.md) | `derived/core.json` | `derived.json` |
| Rest and recovery | [rest.md](../core/rest.md) | `rest/core.json` | `rest.json` |
| Armor slots, layers, DR | [armor.md](./armor.md) | `equipment/armor.json` | — |

Rules still in prose here are marked **[A]** and named with the issue that turns them into rows. Until then a table and an engine use them as written.

---

## 1. The turn

A turn spends action points. The pool is `2 + floor(AGI / 3)`, unspent points carry up to 10, and a short rest refills it ([actions.md](../core/actions.md)). Every action has one cost:

| Action | AP |
| --- | --- |
| attack | by loadout, 2 to 5 (3 with one one-handed weapon) |
| jab | by loadout, 2 or 3: the off-hand weapon alone |
| full attack | every AP left, at least the attack + 1 |
| cast | 3 (plus the spell's mana) |
| move | 1 per move |
| disengage | 1 |
| dodge | 3 |
| item | 2 |
| block | 2, reserved, refunded if no attack comes |
| aimed | +2 on the attack |

There is no bonus action and no reaction slot. An attack strikes with everything held: with two weapons it is a main-hand and an off-hand strike, each its own roll, at the loadout's AP and penalties ([core/combat.md](../core/combat.md#attacks-by-loadout)); the jab is the off hand alone. A second move is two moves. Readying a block or a dodge spends its cost from the pool before the enemy's turn (§3).

## 2. Opportunity attacks

Leaving an enemy's reach without disengaging provokes one attack from that enemy, resolved as an ordinary attack ([combat.md](../core/combat.md)). Reach weapons (`attackReach` 2) threaten 2 units, others 1.

- The attack costs the attacker its normal 3 AP from its carried pool. An attacker with fewer than 3 AP carried does not get it. **[A]** This replaces the old "reaction" for it; it is logged as an open question in [`docs/decisions/M1-01.md`](../decisions/M1-01.md).
- Disengage (1 AP) prevents it. Perks may add other triggers (casting in reach, standing up, drinking a potion). **[A]**

## 3. Defense

The order of one attack is fixed in [docs/core/combat.md](../core/combat.md): one attack roll read against the defense bands, damage, crit, area halving, armor and shield DR, resist, HP. This section explains the pieces.

### Defense: one roll, two bands, two stances

The defender rolls nothing. The attacker's d100 decides it all: a miss, a hit, or a success that lands inside the defender's band (owner, 2026-11-16, M3-27). The bands are the attacker's margin read against the defender's numbers:

- **Evasion, always on.** The evasion rate is half the `evasion` skill. A success whose margin is under it is **evaded**: a single-target attack does nothing, an area attack is halved. Instant spells meet no band; their target makes the save the spell names.
- **A carried shield, always on.** Its DR adds to the armor DR against physical types: buckler DR 1, standard shield DR 2, tower shield DR 4 (and grants cover to an adjacent ally, **[A]**, M3-03). No action, no roll.
- **Dodge, a stance** (3 AP, until your next turn): the evasion rate doubles. An area attack in the band is still halved, not avoided.
- **Block, a stance** (2 AP, reserved, refunded if nobody swings; needs a shield): a block band of half the `shields` skill sits under the evaded band, and a success in it is **blocked**: no damage. It works against physical types only, unless Spell Ward gives a band of half the width against magic.
- A crit always hits: no band reaches it.
- Delivery decides what is possible, not the damage type: a fire bolt can be dodged; a fireball dodged is halved.

Balance between the two is perks and armor, not the pipeline: gated block and dodge perks (M2-06), armor evasion penalties on the evasion skill, a shield's DR that helps without an action, and instant spells that skip the dodge.

Not yet rows, **[A]**:

- **Parry.** Some weapons (rapiers, swords) and the *Duelist* perk let a defender block with the weapon skill instead of the `shields` skill. A parried hit is blocked.
- **Two weapons.** An off-hand weapon can parry but not block. It strikes as part of the attack (GD_S05_01) or alone as a jab; parrying with it is GD_S05_02's stances.
- **Unarmed.** The *Deflect Arrows* perk blocks missiles with `unarmored`. No unarmed melee block without a perk. A natural weapon (a bite, a claw, a race's claws) is an unarmed attack: it rolls with `unarmed`, and the unarmed perks change its damage ([monsters.md](./monsters.md#natural-weapons)).
- A defender who is stunned, incapacitated, paralyzed, petrified, or unconscious keeps no stance and has no evasion band. **[A]** The conditions grant those flags ([modifier.md](../core/modifier.md)); the pipeline does not read them yet.

### Armor DR and resist

Armor DR is the sum of all worn pieces ([armor.md](./armor.md)), plus a carried shield's DR, and applies only to damage types tagged `armor_dr_applies`; a piercing hit halves the total. Resist is the contract's `resist` operation: a flat subtraction from damage of the matching type. Both stop at zero.

### Magic

A spell attack runs the same order as a weapon: the cast check is the attack roll, delivery decides the bands it meets, then damage, crit, resist. Physical armor DR does not apply to fire, frost, shock, or other types without `armor_dr_applies`.

Further magic layers are not in the kernel yet **[U]**: spell absorption (a chance to absorb a spell into mana, `core:effect/spell_absorption`), WIL-based magic resistance, and phase resistance from the [wuxing](./wuxing.md) cycles. They will add terms to the same steps (M3-01, M3-05); none changes the order.

---

## 4. Critical hits and fumbles

One function ([resolution.md](../core/resolution.md)): a roll at or under the crit window (5 + floor(LCK / 2)) is tagged `crit`; a roll of 100 is tagged `fumble` and fails. There is no confirmation roll and no effects table. The combat pipeline doubles the damage of a crit before any reduction ([combat.md](../core/combat.md)). A fumble does no damage.

Condition riders on a crit (bleed on a blade, stagger on a blunt weapon) are not rows yet **[A]**, M3-03.

## 5. Cover, flanking, terrain, prone

The rows are in `ruleset/data/combat/situations.json`; the order and the stacking are in [core/combat.md](../core/combat.md#situations-and-called-shots). This table is the reading of those rows. Difficult terrain is movement and belongs to an optional positioning module; standing up is the `stand_up` action (2 AP).

| Situation | Effect |
| --- | --- |
| Half cover | −10 on attacks against the target |
| Three-quarters cover | −20 on attacks against the target |
| Full cover | cannot be targeted by ranged attacks |
| Flanking (two allies on opposite sides) | +10 on melee attacks |
| Beside your summon (you and a summon of yours attack the same target) | nothing, unless a perk adds to `beside_summon_attack` (Pincer Strike: +10, melee or ranged) |
| Difficult terrain | each move costs 2 AP; no disengage |
| Attacker prone | −20 on melee attacks; no ranged attacks without a perk |
| Target prone | +10 on ranged attacks against it |
| Standing up | 2 AP |
| Attacker blind or target invisible | −20 on attacks; area attacks are unaffected |
| Shooting into melee with an ally engaged | −10; *Sharpshooter* removes it |

## 6. Called shots

Declared before the roll. The rows are in `ruleset/data/combat/situations.json` (`called_shots`); the penalty is an `add` on the attack target and the multiplier applies to the damage dice before the crit and the reductions ([core/combat.md](../core/combat.md#situations-and-called-shots)).

| Location | Attack | Damage | On a hit |
| --- | --- | --- | --- |
| Torso | 0 | ×1 | — |
| Head | −20 | ×1.5 | the target is staggered |
| Arms | −15 | ×0.75 | disarm: the target drops its weapon on a failed STR check |
| Legs | −15 | ×0.75 | slowed for 2 rounds |
| Groin, eyes | −30 | ×2 | staggered; stunned 1 round on a failed END check |

Beyond short range a ranged called shot takes another −10; at long range it is not possible. A crit on a called shot doubles the location's damage too (×1.5 then ×2). Perks halve the penalties: *Marksman* (ranged), *Street Samurai* (all), *Surgeon* (no damage loss on arms and legs).

### Maneuvers

A maneuver changes how an attack is made rather than where it lands, so it stacks with a called
shot: a Heavy Swing to the head is −30 to hit and ×1.5 then ×1.5 on the damage. The rows, their
costs, and who may use them are in [core/actions.md](../core/actions.md#maneuvers).

## 7. Initiative

Initiative is a derived stat: `PER` plus modifiers ([derived.md](../core/derived.md)). Highest acts first. Ties go to the higher AGI, then to a d100 roll-off, lowest first. **[A]** The *Alert* perk adds 10. There is no d20.

## 8. Encumbrance and armor action costs

Encumbrance comes from worn weight against STR ([armor.md](./armor.md) §9). Those thresholds do not use the derived `carry` stat (`25 + STR × 10`, [derived.md](../core/derived.md)); which one is the authority is open **[U]**. Armor category comes from the heaviest piece worn, shields included. Both add AP to demanding actions, and they add together. **[A]** They become `add` modifiers on action costs.

| | Light | Medium | Heavy |
| --- | --- | --- | --- |
| Dodge | 3 | 4 | 5 |
| Stand up | 2 | 3 | 4 |
| Jump | 2 | 3 | 4 |
| Aimed | +2 | +3 | +4 |

| Worn weight | Moves | Evasion penalties from armor | Other |
| --- | --- | --- | --- |
| Up to STR × 10 | normal | ×1 | — |
| Over STR × 10 | each move +1 AP | ×2 | no full jump |
| Over STR × 15 | each move +1 AP | ×3 | no dodge, −20 athletics |
| Over STR × 20 | none | evade always fails | only drop items and talk |

A knight in full plate (heavy, AGI 6, pool 4) cannot afford a 5 AP dodge: he blocks (2 AP) and trusts his DR. A skirmisher in leather (light, AGI 9, pool 5) can dodge once and still move twice. *Heavy Armor Expert*, *Light Foot*, and *Acrobat* reduce the surcharges (M2-06).

---

## 9. Worked examples

Each step below is a vector in `ruleset/tests/vectors/combat.json` (the names start with `example`), so the numbers are checked on every PR. Dice and d100 rolls are given; the pipeline does the rest.

### Example 1: knight against bandit

- **Sir Aldric:** Blades 72, Shields 55, Evasion 30, LCK 4, END 7, AGI 6, PER 4. Derived at level 1: HP 75, AP 4, initiative 4. Iron plate cuirass, helm, gauntlets, and greaves: armor DR 22. Longsword (1d10 slashing), standard shield (DR 2).
- **Bandit:** Blades 45, Evasion 25 (evasion rate 12). Leather jerkin and cap: armor DR 2. Shortsword (1d6). HP 30.

1. **Initiative.** Aldric 4 against the bandit's PER 3: Aldric acts first.
2. **Aldric attacks** (3 AP). The bandit had declared a dodge: evasion rate 12 doubles to 24. Blades 72, roll 47: margin 25, not under 24, a hit. Damage 8 − DR 2 = **6**. Bandit HP 30 → **24**. *(example 1a)*
3. **The bandit attacks.** Aldric had declared a block (2 AP): evasion rate 15, block band 27, so a margin under 15 is evaded, 15 to 41 is blocked, 42 and up hits. Blades 45, roll 25: margin 20, blocked, no damage. Aldric HP stays **75**. Had the bandit rolled 2 (a crit, margin 43), the hit would land: 5 × 2 = 10 − armor 22 and shield 2 = **0**. *(example 1b)*
4. **Next round, Aldric attacks again.** Roll 4: inside the crit window (5 + floor(4 / 2) = 7), tagged `crit`, and a crit always hits, dodge or not. Damage 8 × 2 = 16 − DR 2 = **14**. Bandit HP 24 → **10**. *(example 1c)*

### Example 2: mage against fire

- **Lyra:** Red Magic 65. Knows *Fire Bolt* (`damage_fire` 8, `ranged`, so delivery single) and *Fire Burst* (`damage_fire` 28, `area`).
- **Cinder,** a fire salamander: Evasion 30 (rate 15), scales DR 3, fire resist 3, HP 40.
- **Two goblins:** a nimble one with Evasion 40 (rate 20) and one with Evasion 20 (rate 10), DR 1, HP 15.

1. **Fire Bolt at Cinder** (3 AP, 5 mana). Cast check 65, roll 38: margin 27, over Cinder's rate 15, a hit. Fire is not tagged `armor_dr_applies`, so the scales do nothing. Damage 8 − fire resist 3 = **5**. Cinder HP 40 → **35**. *(example 2a)*
2. **Fire Burst on both goblins** (3 AP, 25 mana). One cast check for the burst: roll 50, margin 15. Each goblin reads it against its own band:
   - Goblin A: rate 20, margin 15 is under it, evaded. An area attack is halved, not avoided: 28 → **14**. HP 15 → **1**. *(example 2b)*
   - Goblin B: rate 10, margin 15 is not under it, a hit. Damage **28**. HP 15 → **0**. *(example 2c)*

---

## 10. Balance guidelines

Reference values for building encounters and gear. They are not rules; tune them per campaign. The bands are data in `ruleset/data/balance/core.json`, and a report checks monsters and gear against them ([core/balance.md](../core/balance.md)).

### Hit points

`hp = 15 + END × 8 + 4 per level after the first` ([derived.md](../core/derived.md)), from `ruleset/data/derived/core.json`. **[V 2026-11-04]** computed with the reference.

| Role | END | Level 1 | Level 10 | Level 20 | Level 30 |
| --- | --- | --- | --- | --- | --- |
| Mage | 4 | 51 | 87 | 127 | 167 |
| Skirmisher | 5 | 59 | 95 | 135 | 175 |
| Fighter | 7 | 75 | 111 | 151 | 191 |
| Tank | 9 | 91 | 127 | 167 | 207 |

HP grows slowly with level: END and gear matter more than level. The older table here (551 HP at level 30) assumed `END × 8 + 4` per level; the data never said that. The table has no luck growth; each level-up can add up to `⌊LCK / 2⌋` more ([progression.md](../core/progression.md#resource-growth)).

### Damage and DR by tier

| Tier | Levels | Weapon dice | Expected damage before DR | Light DR | Medium DR | Heavy DR |
| --- | --- | --- | --- | --- | --- | --- |
| Early | 1–5 | 1d6–1d8 | 4–7 | 1–3 | 2–4 | 3–6 |
| Mid | 6–14 | 1d8–1d10 | 7–12 | 2–4 | 3–6 | 5–10 |
| High | 15–20 | 1d10–2d8 | 10–16 | 3–5 | 5–8 | 8–14 |
| Endgame | 21–30 | 2d8–2d10 | 14–22 | 4–6 | 6–10 | 10–18 |

Total worn DR is roughly 3–8 light, 8–18 medium, 14–28 heavy. A fighter at level 10 (111 HP) hit for 9 a time goes down in 13 hits unarmored, 23 in light armor (DR 4), and 111 in medium (DR 8); heavy armor (DR 14) takes no damage at all. Heavy armor is meant to stop low-tier physical damage; the counters are magic, called shots, armor-piercing types, and conditions.

### Skill benchmarks and heuristics

| Level | Weapon skill | Evasion | Shields |
| --- | --- | --- | --- |
| 1 | 20–30 | 15–25 | 15–25 |
| 5 | 35–50 | 25–40 | 25–40 |
| 10 | 50–65 | 35–55 | 35–55 |
| 20 | 70–85 | 55–75 | 55–75 |
| 30 | 85–100 | 70–90 | 70–90 |

- A fair one-on-one fight ends in 4–6 hits on each side, after misses, evades, blocks, and DR. A boss takes 8–12 party hits, a minion 1–2.
- Expected DR should stay below expected damage per hit, or the fight stalls.
- With LCK 0 the crit window is 5 (5 % of rolls); at LCK 10 it is 10.
- Two equal fighters at 100 attack and 100 evasion (rate 50): rolls 1 to 50 hit, 51 to 99 are evaded, 100 fumbles, so about half land, half are dodged, 1 % fumble.

---

## 11. Real-time clients

A real-time game runs the same functions. These notes are presentation, not rules: a client that disagrees with a vector is wrong.

- **Time instead of turns.** An action costs its AP times `stamina_per_ap` in stamina, and a duration is its seconds; [actions.md](../core/actions.md#one-rule-set-two-clocks) has the conversion. A client may show the cost as a cooldown in the same proportion (an attack costs three times a move). Initiative is not needed when everyone acts at once; aggro decides who an enemy targets.
- **Stances.** A dodge or block stance becomes holding the dodge or block input when the hit lands: the same bands on the attacker's one roll. Input timing may decide whether the stance is held, not whether the band catches the roll. **[A]** A narrower parry window for the parry perks (M2-06).
- **Dodge feedback.** A successful evade plays a dodge animation; an area dodge plays a stumble and shows the halved number.
- **Damage popups.** Show the pipeline: `16 − 2 (DR) = 14`. Colours: white physical, orange crit, blue magic, green healing, red damage over time, grey fully absorbed.
- **Animation guidance.** Attack recovery 0.3–0.5 s (heavy weapons 0.6–0.8 s), hit stun 0.1–0.2 s, knockdown 0.5–0.8 s plus 0.5 s to stand, dodge animation 0.3–0.5 s.
- **Threat.** Damage dealt ×1.0, healing ×0.5, taunt (`core:effect/taunt`) fixed and high; threat halves after 5 s out of reach. A tank should generate 2–3 times a damage dealer's threat.

---

## 12. Deferred

Mounted combat, underwater combat, vehicles, and mass battles are not specified. Mounts and vehicles belong to a post-1.0 module. Notes for when they are picked up:

- **Mounted:** the rider uses the mount's movement and acts on the rider's initiative; attacks may hit rider or mount; a charge of 4+ units adds damage.
- **Underwater:** moves halved; slashing and blunt deal half; breath `END × 15` seconds; fire-phase effects fail, water-phase ones are amplified.
- **Mass battles:** units of 10–50 with pooled HP and morale, commander checks, player characters as hero units.

## 13. Combat perks and traits

Perks are rows in `ruleset/data/features/core.json`; each emits modifiers ([contract.md](../core/contract.md)) and so works the same at a table and in an engine. Existing combat perks include Toughness (+15 HP per rank), Alert (+10 initiative), Full Weight, Deadeye, Shield Wall, Fast Shot (creation), and Field Medic. The gated block and dodge ladders are M2-06. The melee ladders (Carry Through, extra strikes, two weapons) are in [core/combat.md](../core/combat.md#attacks-by-loadout).

When adding one:

1. One rank improves one thing by about 10–25 %; rarely more than 50 %.
2. Gate it on the matching attribute and a skill of 30–50, and on prerequisite perks for a ladder.
3. Express it as modifiers on existing targets (an `add` on a skill, a `grant` of a flag, a `resist`), not as a new rule.
4. Two perks on the same number add (different sources), they do not multiply.

---

## 14. Settled questions

- **Slots and layers.** The slot and layer model in [armor.md](./armor.md) is the data. A client may hide it (auto-equip a skin layer, a three-zone screen); the DR is the same.
- **Shields.** `slot: "offhand"`, `layer: null`. A carried shield's DR adds to the armor DR against physical types (M3-27); a block stance adds a block band on top. Its `evasionPenalty` applies whenever it is worn. A buckler leaves room for a light off-hand item.
- **Conditions and effects.** Every condition's `appliedBy` and `removedBy` names a real effect in `ruleset/data/effects/core.json`; the integrity check fails on one that does not resolve ([integrity.md](../core/integrity.md)).
