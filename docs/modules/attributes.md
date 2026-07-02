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
- **Seeds skills:** `blades`, `blunt`, `heavy_armor`, `athletics`.
- **Feeds:** Carry weight `25 + STR × 10`.

### Perception (PER)
- **Governs:** ranged accuracy, detection, initiative order.
- **Seeds skills:** the archery skills (`bows`, `crossbows`, `guns`, `throwing_weapons`), `insight`, `survival`.
- **Feeds:** Initiative `= PER`. (The *Alert* perk adds +10.)

### Endurance (END)
- **Governs:** hit points, fatigue, resistance to poison and disease.
- **Seeds skills:** `block`, `medium_armor`.
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
- **Seeds skills:** `piercing`, `unarmed`, `light_armor`, `unarmored`, `evasion`, `stealth`, `lockpick`, `sleight_of_hand`.
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
Racial base speed in abstract units is the foundation. Combat movement depends on mode.

- **Base**: Race `speed` in units (e.g. human 6, dwarf ~5).
- **Exploration**: Base speed in units. Modifiers from AGI (+1 unit per 2 AGI above 4?), encumbrance (see Carry weight), heavy armor (penalties via equipment properties), effects/perks (e.g. Woodland Stride ignores plant difficult terrain).
- **Turn-based combat movement**: Units per turn = base speed (units) + floor(AGI / 3) – encumbrance penalties. Typical human: 6 units.
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

### Resistance System (General, Disease & Poison)

A character's resistance to harmful effects, damage types, and conditions is derived primarily from attributes (especially END and WIL), modified by race, equipment, perks, and the shared effect system.

**Base Formulas (reference):**
- **General Physical / Poison / Disease Resistance**: `END × 5 + bonuses` (as % reduction or bonus to d100 resistance roll). END governs toughness against physical harm, toxins, and illness.
- **Magic Resistance**: `WIL × 4 + bonuses` (as above). WIL governs mental and arcane resilience.
- **Per-type / Phased Resistance**: Handled via `resist` effects with `parameter` (e.g. "poison", "fire", "undead", "physical", or Wuxing phase). These stack with or override base.
- **Status Resistance** (paralyze, fear, disease, poison): `END × 3 + specific bonuses` (e.g. from Dwarven Resilience or creation perks). Reduces chance or severity of application.

**At Character Creation** (see character-creation.md):
- Start with base from final END/WIL after racial + creation perk modifiers.
- Add flat % or roll bonuses from race traits, backgrounds, and creation perks (modeled as innate `resist` effects with appropriate `parameter` and `value` as %).
- Example: A dwarf with END 8 starts with ~40 base disease/poison resistance + 25% from racial trait.

**Resolution**:
- When exposed to a disease or poison (via effect application, injury, contact), make a d100 roll under (base resistance + modifiers).
- Success: effect is negated, reduced in magnitude, or delayed.
- Failure: full effect applied (often as damage/drain/control effects over time or with duration).
- Wuxing phase of the disease/poison vs character's phase can amplify/weaken via cycles.
- Ongoing diseases may require periodic resistance rolls.

Resistances are a core part of character sheet and creation — they make END and certain racial/creation choices meaningful for survival and exploration play.

### Disease System

Diseases are persistent harmful conditions, often contracted through contact, injury, or specific effects (e.g. monster abilities, blighted areas, poor hygiene). They are modeled using the shared `effects` system but form a distinct "system" for resolution and progression.

**Core Mechanics**:
- **Contraction**: Triggered by failing a resistance roll against a disease effect (see Resistance System). Common vectors: poison channel, contact with diseased creatures, contaminated food/water, or environmental effects.
- **Incubation**: Optional delay (hours to days, represented by effect `duration` or separate timer).
- **Symptoms**: Applied as one or more effects (typically `damage_health`, `drain_stamina`, `damage_attribute`, `control` like fatigue or paralysis). Severity scales with the disease's magnitude or stage.
- **Progression**: Periodic checks (e.g. daily or per rest) or fixed duration. Failing resistance may worsen symptoms (increase magnitude) or spread (contagion tag).
- **Resistance**: As above — END-based + specific disease resistance bonuses. High resistance can prevent contraction or reduce severity.
- **Cure**: Via `cure` effect with `parameter` "disease" or specific (e.g. "blight"). Also herbal remedies, rest in clean conditions, or magic. Some diseases may require specific cures or have permanent effects if untreated.
- **Wuxing Interaction**: A disease's phase (often wood for blight/growth-related) interacts with the victim's phase.

**Sample Diseases** (defined in `effects/core.json` as reusable effects with `category: "control"`, `tags: ["disease"]`, and `channels` including "innate", "poison"):
- Blight: Drains vitality (damage_health over time), wood phase. Common in blighted regions or from undead/plants.
- Fever: Drains stamina and causes weakness (drain_stamina + damage_attribute "end").
- Plague: Highly contagious, high magnitude damage + attribute drain.

**In Character Creation**:
- Racial traits and creation perks can grant starting resistance bonuses (e.g. Dwarven Resilience: +25% poison/disease resistance modeled as innate `resist` effect with parameter "poison" / "disease").
- Backgrounds or perks may provide immunity or vulnerability.
- Calculate base disease resistance from END + bonuses as part of derived stats (step 8 in character-creation).

Diseases add risk to exploration and survival play, making END and resistance choices meaningful beyond combat.

### Poison System

Poisons are fast-acting harmful substances, typically delivered via weapons (coating), traps, consumables, or monster attacks. Like diseases, they leverage the effects system but have distinct delivery and timing.

**Core Mechanics**:
- **Application**: Via "poison" channel (e.g. `coating` on weapons, `poison` delivery in consumables, or direct effect). On hit or consumption, the target makes a resistance roll.
- **Resistance**: END-based (poison resistance) + bonuses. Success may negate, halve magnitude, or shorten duration.
- **Effects**: Usually immediate or short-term `damage_health`, `drain_stamina`, `damage_attribute`, `control` (paralyze, weakness). Magnitude determines strength; duration for lingering.
- **Duration & Stages**: Instant damage or over-time (DoT). Some have stages (e.g. initial damage, then secondary effect if untreated).
- **Cure**: `cure` effect with parameter "poison" (antidotes, alchemy remedies). Time or rest may also mitigate weak poisons.
- **Wuxing**: Poison phase (often wood or earth) vs victim phase.
- **Delivery Channels**: Explicit in effects (poison, coating) and equipment (poisons as consumables with delivery "apply_weapon").

**Sample Poisons** (in `effects/core.json` with `tags: ["poison"]`, channels including "poison", "coating"):
- Weakness Poison: Drains stamina and strength (drain_stamina + damage_attribute "str").
- Paralytic Poison: Applies `paralyze` effect (already in effects, usable via poison channel).
- Health Drain Poison: Direct `damage_health` over short duration.

**In Character Creation**:
- Same as disease: base poison resistance from END + racial/creation bonuses (e.g. creation perks like Undead Phobia or Silver Sensitivity grant % resistance via `resist` effects).
- Some creation perks or backgrounds grant poison kits or vulnerabilities.
- Poisons tie into Alchemy skill for brewing (see alchemy.md) and combat (coatings).

Poisons make weapons and consumables dangerous, rewarding resistance investment and cure preparation. They interact with the generalized `resist` and `cure` effects for consistency.

**Poison, diseases, and resistances are handled entirely as effects from the shared registry (`data/effects/core.json`).**

- **Resistances**: The generalized `resist` effect (core:effect/resist) with `parameter` for the type ("poison", "disease", "physical", "magic", Wuxing phase, "undead", etc.) and `value` as the % reduction or bonus to resistance rolls. These are granted as innate effects (channel "innate") at creation via race traits (e.g. Dwarven Resilience), backgrounds, or creation perks (see `features/core.json` and `traits/core.json`, which reference the resist effect by string key with the appropriate parameter).
- **Diseases**: Specific effects tagged ["disease"] (e.g. blight core:effect/blight, fever core:effect/fever, plague core:effect/plague in effects/core.json), typically category "control" or "damage", long duration, symptoms delivered as other effects (damage_health, drain_*, control). Applied on exposure (failed resistance or via poison channel/spell). Cured via the generalized `cure` effect (core:effect/cure) with `parameter` "disease" or specific (e.g. "blight").
- **Poisons**: Handled via effects tagged ["poison"] or the generic `poison` effect (core:effect/poison). These use "poison" or "coating" channels. The generic poison (and specific poison entries) can list `sub_effects` (array of other effect refs) which are applied as the symptoms/payload when the poison takes hold (e.g. sub_effects: [5] for damage_health, [6] for damage_stamina, [11] for paralyze). This allows recursive application of other effects (damage, drain, control, etc.) as symptoms. However, a poison effect MUST NOT list any other effect tagged "poison" in its sub_effects (enforced in data and documented to prevent infinite recursion or self-poisoning). Specific poisons (e.g. poison_paralytic core:effect/poison_paralytic with sub_effects [11], etc.) are concrete instances using the generic mechanism. Applied via weapon coatings, consumables, or monster attacks. Cured via `cure` with `parameter` "poison". Resistance via `resist` effect with parameter "poison".

The full systems (resistance checks on exposure using d100 under base resistance from attributes + resist effect bonuses; symptoms as applied effects; progression and contagion via duration/tags; cures; Wuxing phase interactions) are all implemented through the effects system for consistency with spells, alchemy, monsters, and perks. No separate top-level collections for "diseases" or "poisons" — they are effects.

**In Character Creation** (see `character-creation.md`):
- Base resistance calculated from END (for poison/disease/physical) and WIL (for magic/status), plus bonuses granted as the resist effects (core:effect/resist with parameter).
- Creation perks and racial traits (in traits/core.json) grant these resist effects explicitly (e.g. Undead Phobia and Dwarven Resilience grant resist for "poison" and "disease").
- The systems are part of derived stats and make END and certain creation choices meaningful for survival.

See the sample effects in `effects/core.json` and the resist/cure generalized effects (ids 19 and 20) for the building blocks. The "other 3 points in #5" (full derived stats, encumbrance, light/vision) are deferred.

(The "Other Recommended Secondary Statistics" and formula sections above provide the attribute baselines that feed into these effect-based systems.)

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

**Note**: The Resistance, Disease, and Poison systems (detailed above) have been implemented as full mechanics tied to character creation. They close several of the prior gaps in #5. The remaining points in #5 (complete derived stats formulas, encumbrance rules, and light/vision systems) are deferred per instructions and will be addressed after these.
