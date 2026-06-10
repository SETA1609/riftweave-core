# Attributes

**Status:** Core (implemented as data + schema)

The eight attributes are the foundation a character is built on: every skill is
seeded by one, every resource and derived statistic is computed from them, and many
perks gate on them. This document explains the scale, what each attribute does, and
the formulas that read it. For where attributes sit in the build order see
[`character-creation.md`](./character-creation.md); for the skills they seed see
[`progression.md`](./progression.md).

Data: [`abilities/core.json`](../../ruleset/data/abilities/core.json) (schema
`ability.schema.json`). The set of valid attribute ids is the enum
`schema.json#/definitions/ability` — **changing the attributes means editing that one
enum**, which every schema that accepts an ability inherits from.

---

## The set

Riftweave's eight attributes are seven physical and mental stats plus **Willpower**,
which fuels resource regeneration and resistance to hostile magic:

| Id | Attribute | One-line role |
| --- | --- | --- |
| `str` | Strength | Melee force and how much you can carry. |
| `per` | Perception | Aim, awareness, and acting first. |
| `end` | Endurance | How much punishment you can take, and your Stamina. |
| `int` | Intelligence | How fast you learn, how deep your mana runs, and the **power of all magic**. |
| `wil` | Willpower | Regenerates your resources and resists hostile magic. |
| `agi` | Agility | Speed, finesse, stealth, and Action Points. |
| `cha` | Charisma | Sway over other people. |
| `lck` | Luck | A thumb on the scale of everything. |

## The scale (1–10)

Attributes run **1–10**: **1** is debilitating, **4** is the unmodified human
baseline, **5–6** is trained/above-average, **8** is exceptional, **10** is the mortal
peak. At creation you start every attribute at **4**, distribute **10** points
(cap 10 each), then apply racial `abilityModifiers` and any creation-perk bonuses — those
layers may push a stat above 10 or below 4 (see
[`character-creation.md`](./character-creation.md) §2).

Attributes change rarely after creation (perks, rare items, magic) — unlike skills,
which grow every level. They are the slow-moving frame; skills are the dial.

---

## The three resources

Every character runs on **three pools**. The first two are universal; the **third
changes shape with the combat mode**:

| Resource | Pool (max) | Regenerates via | Governing attribute |
| --- | --- | --- | --- |
| **Health Points (HP)** | `15 + END × 8 + level × 4` | rest / healing | **END** |
| **Mana Points (MP)** | `INT × 8 + level × 2` | **WIL** | **INT** (pool) |
| **Action Points (AP)** — *turn-based / TTRPG* | Max 10. Pool carries from turn to turn but cannot exceed 10. (Typical starting value 2 + floor(AGI / 3) or as determined at the start of an encounter.) | carries over (capped at 10) | **AGI** |
| **Stamina (SP)** — *action combat* | `15 + END × 5 + level × 2` | **WIL** | **END** |

The third row is one resource with two faces. In **turn-based (TTRPG)** mode the
character has an **Action Point** pool (max 10) that carries over from turn to turn
but cannot exceed 10; this governs how many discrete actions can be taken. In
**action-combat** mode a **Stamina** pool is used for movement, blocking, and power
moves (an END/fatigue quantity). **Willpower regenerates both Mana and Stamina.** A
game picks one mode; the rest of the system is identical.

> **Magic strength is INT, not WIL.** INT sets the size of the mana pool *and* the
> power of every spell. Willpower's job in magic is purely to **regenerate** mana (and
> to resist incoming magic) — a high-WIL caster sustains casting longer, but a high-INT
> caster hits harder and has more to spend.

---

## What each attribute does

Each entry lists what it **governs**, the **skills it seeds** (a skill's starting
rating is `5 + associatedAbility × 2`), and the **resources/derived stats** it feeds.

### Strength (STR)
- **Governs:** melee damage, carrying capacity, Strength-gated equipment.
- **Seeds skills:** `blades`, `blunt`, `piercing`, `athletics`.
- **Feeds:** Carry weight `25 + STR × 10`.

### Perception (PER)
- **Governs:** ranged accuracy, detection, initiative order.
- **Seeds skills:** `marksmanship`, `insight`, `survival`.
- **Feeds:** Initiative `= PER`. (The *Alert* perk adds +10.)

### Endurance (END)
- **Governs:** hit points, fatigue, resistance to poison and disease.
- **Seeds skills:** `block`.
- **Feeds:** **Hit points** `15 + END × 8 + level × 4`; **Stamina** pool
  `15 + END × 5 + level × 2` (action-combat mode). END is the single biggest lever on
  survivability.

### Intelligence (INT)
- **Governs:** the **skill-point gain rate**, the **mana pool**, the **power of all
  nine magic schools**, and knowledge/tech skills. This is the magic attribute and the
  learning attribute at once — the most load-bearing stat in the system.
- **Seeds skills:** **all nine** `<color>_magic` schools (`red` `orange` `yellow`
  `green` `blue` `indigo` `violet` `white` `black`), plus `lore`, `investigation`,
  `medicine`, `repair`, `alchemy`.
- **Feeds:** Skill points per level `5 + INT × 2 + random(0…LCK)`; **Mana pool**
  `INT × 8 + level × 2`. INT compounds — it raises *every* future level's skill points
  on top of powering magic.

### Willpower (WIL)
- **Governs:** **mana and stamina regeneration**, and **magic resistance**. WIL does
  *not* set magic power (that is INT) — it sets how fast you recover the resources you
  spend, and how well you shrug off hostile magic.
- **Seeds skills:** `animal_handling`.
- **Feeds:** Mana regeneration rate; Stamina regeneration rate; resistance to incoming
  spells.

### Agility (AGI)
- **Governs:** action economy, stealth, and dexterous skills.
- **Seeds skills:** `unarmed`, `stealth`, `lockpick`, `sleight_of_hand`.
- **Feeds:** **Action Points** `2 + floor(AGI / 3)` per turn (turn-based mode);
  movement/attack speed in action combat.

### Charisma (CHA)
- **Governs:** persuasion, barter prices, leadership.
- **Seeds skills:** `persuasion`, `deception`, `intimidation`.
- **Feeds:** social outcomes and prices (no single combat formula; it is the social
  pillar's attribute).

### Luck (LCK)
- **Governs:** fortune across the board. Deliberately **diffuse** — it always helps
  but never dominates, and **seeds no skill**.
- **Feeds:**
  - **+floor(LCK / 2)%** to **every** check's target number (applied at roll time, not
    stored in skill ratings).
  - **+1% critical-hit chance** per point (`crit = 1% + LCK%`).
  - A **random bonus to skill points** each level: the `random(0…LCK)` term.
  - Better loot quality and more favourable random outcomes.

---

## Magic & attributes

Magic leans on exactly two attributes, doing two different jobs:

| Attribute | Role in magic |
| --- | --- |
| **INT** | Seeds **all nine** color schools, sets the **mana pool**, and drives **spell power**. The caster's reservoir and punch. |
| **WIL** | **Regenerates** mana and **resists** hostile magic. The caster's stamina, not their punch. |

A dedicated caster invests primarily in **INT** (to cast hard and often), with **WIL**
to sustain longer fights and weather enemy spells. See [`magic.md`](./magic.md).

---

## How attributes feed the rest of the system

| Reads an attribute | Formula / rule |
| --- | --- |
| Skill starting rating | `5 + associatedAbility × 2` |
| Skill points per level | `5 + INT × 2 + random(0…LCK)` |
| Hit points (END) | `15 + END × 8 + level × 4` |
| Mana pool (INT) | `INT × 8 + level × 2` |
| Mana / Stamina regen (WIL) | scales with WIL |
| Stamina pool (END, action combat) | `15 + END × 5 + level × 2` |
| Action Points (AGI, turn-based) | `2 + floor(AGI / 3)` per turn |
| Carry weight (STR) | `25 + STR × 10` |
| Critical chance (LCK) | `1% + LCK%` |
| Initiative (PER) | `PER` |
| Every check (LCK) | `+floor(LCK / 2)%` to the target |
| Movement / Attack Speed (AGI) | Base from race + AGI modifiers; % or units in combat (see Secondary Statistics) |
| Magic Resistance (WIL) | WIL-based % or resist bonus + effects |
| Physical / Elemental Resist (END + effects) | Armor DR + phase resists via Wuxing |
| Perk prerequisites | `prerequisite.abilities`, e.g. *Alert* needs `per ≥ 6` |
| Racial modifiers | `race.abilityModifiers`, e.g. dwarf `{ end: +2, str: +1, agi: −1 }` |

## Secondary & Combat Statistics (Derived)

These are computed from attributes, race, equipment, perks, and effects. They are the "surface" numbers an engine or character sheet uses in play. Many can be further modified by the shared effect system (see `effects/core.json` and `magic.md`).

### Movement Speed
Racial base speed (in feet per round for exploration) is the foundation. Combat movement depends on mode.

- **Base**: Race `speed` (e.g. human 30, dwarf ~25).
- **Exploration**: Base speed. Modifiers from AGI (+1 ft per 2 AGI above 4?), encumbrance (see Carry weight), heavy armor (penalties via equipment properties), effects/perks (e.g. Woodland Stride ignores plant difficult terrain).
- **Turn-based combat movement**: Units per turn = floor(base speed / 5) + floor(AGI / 3) – encumbrance penalties. Typical human: 6 units.
- **Action-combat movement**: Speed rating (derived from AGI and base speed) that affects positioning, dodge chance, and how quickly you close distance or escape. High speed can grant "free" movement or reduce enemy attack opportunities.
- **Modifiers**: AGI primary; STR/END for carrying load; racial traits; effects (e.g. feather for carry, haste-like effects).

Races declare the base. Equipment and effects apply deltas or multipliers.

### Attack Speed
Governs how frequently or quickly attacks resolve.

- **Base**: Primarily from AGI (finesse/speed) and weapon properties (light weapons faster than heavy).
- **Turn-based**: May reduce "recovery" time between attacks or grant extra attacks when high (e.g. via AP economy or specific perks). Fast Shot perk gives +20% ranged speed.
- **Action combat**: Directly affects attack recovery, combo windows, or number of attacks in a flurry. Higher speed = more actions before opponent reacts.
- **Modifiers**: AGI, perks (Fast Shot), effects (haste/slow), weapon category/length, encumbrance.
- Reference: AGI "movement/attack speed in action combat." Can be expressed as a percentage multiplier on base recovery.

### Magic Resistance
Reduces incoming hostile magic (damage, duration, success chance).

- **Base**: WIL × 3–5 (or similar) + racial traits (e.g. Fey Ancestry).
- **Formula reference**: `Magic Resistance % = (WIL × 4) + bonuses from equipment/effects/perks – penalties`.
- **Application**: Flat % reduction on magic damage/effects, or bonus to resist rolls (d100 under effective WIL + MR). Can be general or per color/phase.
- **Ties to system**: WIL already "resists hostile magic." Effects like `resist` (with parameter "magic", "undead", etc.) provide the implementation. Wuxing phases can interact (e.g. certain phases resist better against opposing colors).
- High WIL casters are harder to shut down.

### Damage Resistance (Physical & Elemental)
Reduces incoming damage after a hit is confirmed.

- **Physical DR**: Primarily from armor (base rating + material modifiers + quality/enchants). Flat subtraction or % reduction. Block skill can add temporary DR.
  - Example: Armor provides "damage-reduction rating". Heavy armor high flat DR but speed/stealth penalties.
- **Elemental / Phased DR**: Per Wuxing phase (wood/fire/earth/metal/water). Governed by race phase, equipment materials (phase-tagged), effects, and perks.
  - Leverages the wuxing matrix for interactions (e.g. generating cycle amplifies resistance or weakness).
  - Effects provide `resist` (parameter e.g. "fire", "poison", "physical", "undead"). Innate or temporary.
- **General formula**: Incoming damage reduced by DR (flat or %). Can stack physical + specific elemental.
- **Modifiers**: END (general toughness), armor, effects (resist_*), racial traits (Dwarven Resilience), material phase vs. incoming effect phase.
- Note: Separate from "to-hit" (skill roll) and Block (active defense).

### Other Recommended Secondary Statistics

These fit the existing design (attributes + effects + Wuxing + dual combat modes + classless perks) and close gaps:

- **Evasion / Dodge**: AGI-based chance or bonus to avoid being hit entirely (before armor DR). Complements Block. "Dodge vs block vs armor" is noted as a follow-up in progression.
- **Block Value / Power**: How much damage a successful Block negates (END or block skill + shield material). Already referenced in perks (e.g. +15% damage blocked).
- **Critical Damage Multiplier**: LCK for chance; add multiplier (e.g. 1.5× + LCK/10 or weapon-based). Currently only chance is defined.
- **Regeneration Rates** (formal): 
  - Mana/Stamina: "scales with WIL" – e.g. WIL × 2 per turn/round or per 6 seconds.
  - HP: Slow natural (END-based) or effect-driven only.
- **Encumbrance Penalty**: When carry > weight, apply speed reduction (–1 unit/move or % speed), attack penalties, skill penalties (athletics, stealth). Ties directly to Carry weight.
- **Elemental Resistances** (explicit per phase): Even without full DR, % resist vs wood/fire/etc. damage/effects. Strong Wuxing synergy (your race phase + equipment phase + incoming effect phase via cycles).
- **Status / Effect Resistance**: Bonus to resist specific categories (paralyze, poison, fear, charm) beyond general magic resist. Useful for effects system.
- **Mana / Stamina Efficiency**: % reduction in resource cost for spells/actions (INT or WIL + perks/effects). Helps casters and action-combat characters.
- **Initiative Bonus**: Already PER base + perks (Alert +10). Can be expanded with effects or racial traits for "acting first" in turn-based.

These can all be implemented primarily through the existing **effect system** (innate, applied, or temporary) + attribute formulas + equipment properties. No new top-level collections needed initially — extend the effect vocabulary and document the derived formulas here.

Racial `speed` and traits already provide starting points. Perks and the shared `effects` table handle most modifiers. Wuxing phases give elemental stats flavorful interactions without new mechanics.

> **Note — attributes vs. element.** A character's **`phase`** (its Wuxing element)
> comes from its **race**, not from an attribute. Attributes drive your numbers;
> element drives how effects interact with you (see [`race.md`](./race.md) and
> [`magic.md`](./magic.md)).

---

## Data & schema map

| Concern | Schema | Data |
| --- | --- | --- |
| Attribute definitions | `ability.schema.json` | `data/abilities/core.json` |
| Valid attribute id enum | `schema.json#/definitions/ability` | — |
| Skill ↔ attribute (`associatedAbility`) | `skill.schema.json` | `data/skills/core.json` |
| Attribute-gated perks (`prerequisite.abilities`) | `feature.schema.json` | `data/features/core.json` |
| Racial modifiers (`abilityModifiers`) | `race.schema.json` | `data/races/core.json` |

## Open items

- **Combat mode picks the third resource.** Turn-based → Action Points (AGI);
  action-combat → Stamina (END). The AP-per-turn and Stamina formulas above are
  reference values; the mode itself is a per-game choice (`progression.md` open items).
- **INT load.** With all magic, the skill-point rate, the mana pool, and five
  knowledge skills riding on INT, it is by far the strongest attribute for a caster.
  A game wanting flatter builds may redistribute some knowledge skills or cap INT's
  reach; the ruleset leaves it concentrated by design.
- **Above-10 attributes.** Racial and creation-perk stacking can exceed 10; whether a game caps
  the effective value is left to the game.
