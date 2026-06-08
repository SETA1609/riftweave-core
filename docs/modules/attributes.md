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
| **Action Points (AP)** — *turn-based / TTRPG* | `2 + floor(AGI / 3)` per turn | refreshes each turn | **AGI** |
| **Stamina (SP)** — *action combat* | `15 + END × 5 + level × 2` | **WIL** | **END** |

The third row is one resource with two faces: a **turn-based (TTRPG)** game spends
**Action Points** each turn (how *many* things you can do — an AGI/action-economy
quantity), while an **action-combat** game drains a **Stamina** pool for sprinting,
blocking, and power moves (an END/fatigue quantity). **Willpower regenerates both Mana
and Stamina.** A game picks one mode; the rest of the system is identical.

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
| Perk prerequisites | `prerequisite.abilities`, e.g. *Alert* needs `per ≥ 6` |
| Racial modifiers | `race.abilityModifiers`, e.g. dwarf `{ end: +2, str: +1, agi: −1 }` |

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
