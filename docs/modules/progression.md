# Core System: Attributes, Skills, Perks, Magic

Riftweave is a **classless, d100 roll-under cRPG ruleset**. A character is the sum
of eight attributes, a spread of point-buy skills (some tagged), and a stack of
perks. The numeric formulas below are the **ruleset reference**; a consuming game or
tabletop group may override them via its own data.

## Resolution — d100 roll-under

- To attempt something, roll **1d100** and succeed if the roll is **≤ your target
  number**. The target is your skill rating plus situational modifiers plus Luck.
- **Margin** = `target − roll` gives degrees of success (e.g. for crafting quality,
  spell potency, persuasion strength).
- **Critical success** on `01–05`, **fumble (critical failure)** on `96–00`.
- **Polyhedral dice are used for damage**, not resolution. A weapon hits via a d100
  weapon-skill roll, then rolls its damage dice (`1d8`, `2d6`, …).

## Attributes (1–10)

Eight attributes, defined in `ruleset/data/abilities/core.json`:

| Id | Attribute | Governs |
| --- | --- | --- |
| `str` | Strength | melee damage, carry weight |
| `per` | Perception | ranged accuracy, detection, initiative |
| `end` | Endurance | hit points, Stamina, fatigue, poison/disease resistance |
| `int` | Intelligence | skill-point rate, mana pool, **all magic power**, knowledge skills |
| `wil` | Willpower | mana & stamina regeneration, magic resistance |
| `agi` | Agility | action economy & Action Points, stealth, dexterous skills |
| `cha` | Charisma | persuasion, barter, leadership |
| `lck` | Luck | crits, loot, a flat bonus to every check, bonus skill points |

The eight attributes are seven physical and mental stats plus **Willpower**, which
fuels resource regeneration and magic resistance. Suggested creation: each attribute
starts at **4**, distribute **10** points (max 10 each), then apply racial modifiers
from `races/core.json`.

## Skills (0–100, point-buy)

Defined in `ruleset/data/skills/core.json`, each `associatedAbility`-seeded and
tagged with a `category` (combat, magic, stealth, social, knowledge, survival,
utility).

- **Starting value** = `5 + associatedAbility × 2` (Luck is applied at check time,
  not stored in the rating).
- **XP is almost entirely event-driven**. Direct experience from killing creatures is
  extremely limited by design (see Experience section below). Successful resolution of
  meaningful skill challenges grants experience, but *using* a skill does **not** train it
  — skills advance through earned XP spent at level-up, never through repeated use.
  Crafting is explicitly excluded from experience gain (see Experience section).
- On **level-up** you get a pool of skill points to spend:
  - **skill points/level = `5 + INT × 2 + random(0…LCK)`** — Luck adds a random
    bonus, so a luckier character gains more points per level on average, but
    unpredictably.
- **Tag skills** — choose **3** at creation. Normal skills are bought at a 2:1 ratio
  (2 points for +1 rating). Tagged skills are bought at a 1:1 ratio (1 point for +1
  rating). Same pool, double efficiency on your focus skills — an ongoing discount,
  **not** a flat bonus at creation. Mark a skill non-taggable with `"taggable": false`.
- Skills cap at **100** in v1 (above-100 cost escalation is a per-game extension).

**Level 30 cap and max-INT/LCK characters (with and without regular breakthroughs)**

With a hard level cap of 30, even a character who starts with maximum INT (10) and LCK (10) will only ever receive 30 skill-point pools (including the level-1 pool). Each pool is 25–35 points, for a career total of roughly **750–1,050 skill points** (average ~900). This is the normal source only.

There are currently 36 skills. Bringing every skill from its starting value (base = 5 + associatedAbility × 2, typically 13–25) all the way to the 100 cap requires 75–87 ranks per skill.

- Total ranks needed (best-case, all governing abilities 10): 36 × 75 = 2,700 ranks.
- Minimum normal point cost (3 tagged at 1:1, rest 2:1): **~5,175 points**.

**Without regular breakthroughs:** A max-INT/LCK character has only ~1/5 of the points needed. They can deeply specialize (easily maxing their 3 tagged + 8–12 others) but cannot max *all* 36.

**With grand/legendary items on every breakthrough (max 3 per level, high success chance with max LCK + high grade):**

Each successful high-grade item on a skill target gives 4–5 ranks from the current 0–5 range? No — using the updated 0–10 range (higher grades give ranges like 6–8 or 8–10, average ~7–8.5 per successful high-grade item on a skill target):

Conservative: ~2.5–3 successful skill bonuses per level × ~7.5–8.5 average = **19–25.5 free ranks per level**.

Over 30 levels: **~570–765 free ranks** directly added to skills (player can spread or focus; the target-type weighting makes skill targets the easiest to succeed on).

**Total effective ranks (normal points + breakthroughs):**
- Normal points buy ~450–900 ranks (depending on tagging efficiency).
- + ~570–765 free ranks from breakthroughs.
- **Grand total: roughly 1,020–1,665 ranks.**

This is still short of the ~2,700 needed for all 36 (roughly 38–62% of the way).

**Practical outcome at level 30 with perfect high-grade breakthroughs every level:**  
You can max significantly more skills than with normal points alone — roughly **22–30+ skills fully capped**, depending on how you allocate the free bonuses (focus them on untagged skills to save normal points) and how many normal points you dump into the 3 tagged skills at 1:1 cost. The remaining skills will be at good but not capped levels.

You will **not** be able to maximize *all* current skills (let alone when more are added later) under the current formulas and level-30 cap.

The level-30 cap + current formulas + breakthrough power level (even with grand/legendary every time) still enforces meaningful specialization rather than omni-competence. This is intentional.

To allow a max-INT/LCK character to cap everything by level 30 (or with a higher cap while adding more skills), you would need substantially more free ranks per breakthrough (e.g. higher per-item bonuses, more than 3 items, or occasional extra point pools from breakthroughs) or adjustments to the normal per-level pool / 2:1 ratio / total number of skills.

(See also the XP thresholds table above and the Base Gains on a Normal Level section for the underlying numbers. The proposed success formula and exact 0–10 grade ranges are in `advancement.md`.)

## Experience

Riftweave deliberately discourages "murder hobo" gameplay and endless combat grinding for power.

- **Direct combat experience is extremely limited.** Only the very first monster a character ever kills grants experience.
- After that first kill, slaying creatures no longer awards direct XP, regardless of enemy type or number of kills.
- Example: Killing a bandit for the first time may grant a small amount of experience. Every bandit killed after that (outside of a quest objective) grants **zero** direct experience.
- **All meaningful experience comes from event-driven sources**, including:
  - Completion of story quests and major narrative milestones
  - Radiant / minor quests (especially repeatable "subjugation" or clearing quests taken from notice boards or patrons)
  - Exploration milestones and significant discoveries
  - Major world events and achievements
  - Successful resolution of meaningful **skill challenges** (perception, social negotiation, investigation, puzzle-solving, trap disarming, clever non-combat problem solving, etc.)
- Killing enemies remains highly valuable, but for different reasons:
  - Looting equipment, materials, and treasure
  - Harvesting **monster cores** (the primary source of raw essence for Wuxia-style breakthroughs and alchemical refining — see `advancement.md`)
  - Fulfilling explicit quest objectives (e.g. accepting a "Bandit Subjugation" radiant quest and turning it in upon completion)
  - Narrative consequences, reputation, and roleplaying

**Skill challenges** (as listed above) grant experience when they represent meaningful obstacles or opportunities. However, **crafting and item creation do not award experience** through skill rolls or repeated use. Crafting uses its own dedicated resolution mechanics focused on quality grades, material properties, station tiers, time invested, and character skill investment rather than the general experience system.

This system ensures that players progress by engaging with the world, taking on quests, and participating in stories rather than repeatedly clearing the same bandit camps or monster spawns.

Consuming engines should implement a one-time "first kill" experience award (either globally for the character's life or per distinct creature archetype) and route all other experience through quest/event completion handlers.

## Experience Progression and Thresholds

The total experience required to reach level **N** is given by the formula:

**Total XP to reach level N = 450 × N × (N − 1)**

This uses 450 as the baseline constant for a normal difficulty curve. The constant can be tuned for different campaign tones:

- 500 = Hard / slow progression
- **450 = Normal** (recommended default)
- 400 = Easy / faster progression
- 350 = Very easy / quick power fantasy

### Sample XP Thresholds (Normal / 450 constant)

| Level | Total XP to Reach | XP Needed from Previous Level |
|-------|-------------------|-------------------------------|
| 1     | 0                 | —                             |
| 2     | 900               | 900                           |
| 3     | 2,700             | 1,800                         |
| 4     | 5,400             | 2,700                         |
| 5     | 9,000             | 3,600                         |
| 6     | 13,500            | 4,500                         |
| 7     | 18,900            | 5,400                         |
| 8     | 25,200            | 6,300                         |
| 9     | 32,400            | 7,200                         |
| 10    | 40,500            | 8,100                         |
| 11    | 49,500            | 9,000                         |
| 12    | 59,400            | 9,900                         |
| 13    | 70,200            | 10,800                        |
| 14    | 81,900            | 11,700                        |
| 15    | 94,500            | 12,600                        |
| 16    | 108,000           | 13,500                        |
| 17    | 122,400           | 14,400                        |
| 18    | 137,700           | 15,300                        |
| 19    | 153,900           | 16,200                        |
| 20    | 171,000           | 17,100                        |
| 21    | 189,000           | 18,000                        |
| 22    | 207,900           | 18,900                        |
| 23    | 227,700           | 19,800                        |
| 24    | 248,400           | 20,700                        |
| 25    | 270,000           | 21,600                        |
| 26    | 292,500           | 22,500                        |
| 27    | 315,900           | 23,400                        |
| 28    | 340,200           | 24,300                        |
| 29    | 365,400           | 25,200                        |
| 30    | 391,500           | 26,100                        |

XP to reach level N from level N-1 is always **900 × (N-1)** under the normal constant.

(See the detailed note below the skills cap at 100 for calculations on whether a max-INT/LCK character can master skills by level 30, including with regular grand/legendary breakthroughs.)

### Recommended XP Awards by Event Type

These values are guidance for quest designers and GMs. With the strict "only the very first monster kill grants direct combat XP" rule, the vast majority of experience should come from event-driven sources (including meaningful skill challenges) rather than grinding or repetitive crafting. Crafting does not award experience.

On every normal level the character also receives the base gains package described in the "Base Gains on a Normal Level" section below (1 perk, skill points, resource growth, and TTRPG Action Point rules).

| Event Type                              | Recommended XP | Notes |
|-----------------------------------------|----------------|-------|
| Minor radiant / small task              | 300 – 800      | Filler |
| Normal side quest                       | 1,000 – 2,500  | Most common |
| Major story quest / big event           | 3,500 – 8,000+ | Main driver |
| Significant discovery (major new area/ruin) | 800 – 2,500 | Rewards exploration |
| First/meaningful defeat of dangerous enemy or boss | 600 – 3,000 | Only important fights |
| Hard non-combat resolution              | 700 – 2,000    | Rewards clever play |
| Meaningful skill challenge (non-crafting) | 400 – 1,500  | Perception, social, investigation, puzzles, traps, etc. |
| Major world event resolution            | 2,000 – 6,000  | Dynamic content |

**Note on tuning**: The 450 constant (and the award ranges above) can be adjusted globally for campaign difficulty. Higher constants make the same events feel more significant; lower constants make progression faster.

## Luck

Luck is deliberately diffuse so it always helps but never dominates:

- **+`floor(LCK / 2)`%** to every check's target number.
- **+1% critical-hit chance** per point of Luck.
- A **random bonus to skill points** each level (see formula above).
- Better loot quality and more favorable random outcomes.

## Base Gains on a Normal Level

On every normal level (i.e. when gaining a level without using a breakthrough), a
character receives the following base package:

- **1 perk** chosen from `features/core.json` (the primary character-defining
  advancement).
- A pool of **skill points**: `5 + INT × 2 + random(0…LCK)`. Normal skills are
  bought at a 2:1 ratio (2 points for +1 rating). Tagged skills are bought at a 1:1
  ratio (1 point for +1 rating). Same pool, double efficiency on focus skills.
- **Resource growth** (added to maximums):
  - Hit Points: `END × 8 + 4`
  - Mana (MP): `INT × 8 + 2`
  - Stamina (SP, action-combat mode): `END × 5 + 2`
- In **TTRPG / turn-based** mode: Action Point pool rules (maximum 10; the pool
  carries over from turn to turn but cannot exceed 10).

These base gains always occur. A breakthrough (see `advancement.md`) can provide
additional permanent bonuses on top when the player chooses to offer cores or
elixirs.

(See the note just below the skills cap at 100 for the long-term implication of a level-30 cap on a max-INT/LCK character's ability to master many skills.)

## Full Level-Up Procedure

When a character’s cumulative experience reaches the threshold for the next level
(`450 × N × (N − 1)` for level N under normal difficulty), they gain a level.
This process can be performed immediately upon reaching the threshold or as a
narrative downtime activity tied to the level-up.

### Step-by-Step Procedure

1. **Confirm Level Gain**
   - Add 1 to the character’s level.
   - Record the new total XP threshold for the following level.

2. **Apply Base Level Gains** (always received)
   - Gain **1 perk** (chosen from `features/core.json` entries with `type: "perk"` that meet prerequisites).
   - Receive a **skill point pool**: `5 + INT × 2 + random(0…LCK)`.
   - Apply **resource growth** (added to current maximums):
     - Hit Points: `+ (END × 8 + 4)`
     - Mana (MP): `+ (INT × 8 + 2)`
     - Stamina (SP, if using action-combat mode): `+ (END × 5 + 2)`
   - In **TTRPG / turn-based** mode: Update the Action Point pool (maximum 10; the pool carries over from turn to turn but cannot exceed 10).

3. **Allocate Skill Points**
   - Spend the skill point pool to increase skill ratings.
   - Normal skills are bought at a **2:1** ratio (2 points for +1 rating).
   - Tagged skills (up to 3 chosen at creation) are bought at a **1:1** ratio (1 point for +1 rating).
   - Skills cannot exceed 100 in the base ruleset.

4. **Optional: Perform a Breakthrough**
   - If the character has spirit cores or refined elixirs and chooses to use them (this can be done immediately or as a short downtime activity tied to the level-up), they may attempt a breakthrough.
   - Select up to a maximum of **3** eligible items (cores and/or elixirs) to offer. These items are consumed.
   - For each offered item:
     - Its `phase` determines the governed options (attributes, resource pools such as HP/Mana/Stamina, or skill areas).
     - The player selects **one target** from those options for that item (restrictions apply: attributes only if grade > common, and attributes may only rise by +1 total from all breakthroughs this level).
     - Breakthrough success for this item is luck-based: base chance from the item's `qualityGrade` (a petty_ember_core has a much lower success % than a refined legendary essence), modified by phase synergy (with foundation or target) and the character's Luck (LCK). Higher Luck and better grade/phase dramatically improve odds, but failure is always possible for any item.
     - On success:
       - **Attributes**: +1 (only if eligible; max +1 total this level from breakthroughs).
       - **Skills**: bonus in the 0–10 range using the reference tables in `advancement.md` (grade sets floor and ceiling; e.g. petty 0–2, greater 6–8, grand/legendary 8–10). Higher grades also improve success chance on the luck roll.
       - **Resources** (HP, Mana, or Stamina): the normal level-up growth for that resource is rolled first. A second "growth of a level" roll is performed for the breakthrough (same growth formula). The higher of the two rolls is used for the final growth at this level-up.
     - Wuxing cycle interactions (from `data/wuxing/core.json`) may amplify or reduce results across multiple items.
   - Apply any resulting **extra permanent bonuses** (to the selected targets) on top of the base package. See `docs/modules/advancement.md` for full resolution details.

5. **Update Derived Statistics**
   - Recalculate all derived stats based on the new level, any breakthrough bonuses, and current attributes (HP, Mana, Stamina, carry weight, etc.).
   - In TTRPG mode, note the current Action Point pool (capped at 10).

6. **Record and Narrate**
   - Update the character sheet.
   - Narrate the level-up (and any breakthrough) in a way that fits the story. Breakthroughs especially should feel significant.

### Player Checklist (at Level-Up)

- [ ] Have I reached the required total XP for the next level?
- [ ] Choose 1 perk (check prerequisites in `features/core.json`).
- [ ] Calculate and spend skill point pool (remember 2:1 for normal skills, 1:1 for tagged skills).
- [ ] Apply base resource growth (HP, Mana, Stamina).
- [ ] Decide whether to perform a breakthrough (do I have good cores/elixirs? Max 3 items).
- [ ] If breakthrough: Select up to 3 items → for each, choose one governed target (attribute only if grade > common and +1 max total this level; resource or skill otherwise) → roll for luck-based success using the formula in `advancement.md` (base from grade + phase + LCK, with target-type weighting: skills easiest, then resources, then attributes hardest) → on success:
  - Attributes: +1 (if eligible).
  - Skills: 0–10 range bonus using the reference tables in `advancement.md` (grade sets floor and ceiling; e.g. petty 0–2, greater 6–8, grand/legendary 8–10). Higher grades also improve success chance.
  - Resources: second level-growth roll; take the higher of the normal growth and the breakthrough growth.
  (Wuxing may modify across items) → apply extras.
- [ ] Update all derived stats and current pools (especially AP in TTRPG mode).
- [ ] Note any new capabilities or story implications.

### GM / Engine Notes

- Award XP only through approved event-driven sources after the very first monster kill (quests, radiants, skill challenges, story milestones, significant discoveries, etc.).
- Core drops follow the rules in `advancement.md` (eligible monsters only, base chance + level + Luck modifiers, two-roll process for drop then grade).
- Breakthrough resolution (exact bonuses) follows the design in `advancement.md`. Concrete numeric tables for bonus application are still in development.
- The 450 constant in the XP formula can be tuned for campaign difficulty (500 = hard/slower, 450 = normal, 400 = easy, 350 = very easy).
- Level-ups (especially breakthroughs) are excellent opportunities for narrative spotlight and character development.

### Open Notes on This Procedure

- Reference tables for success % per grade and per-target bonus ranges (0–10 for skills, second growth for resources, +1 for eligible attributes) are defined in `advancement.md`.
- Any requirements for performing a breakthrough (e.g., safe location, time investment, special catalysts) can be added per campaign or module.

## Perks (and traits)

Perks are the **character-defining** progression in this classless system, defined
in a single `features` table (`ruleset/data/features/core.json`). They are gated by
`prerequisite` — any of `level`, `abilities` (min attribute values), `skills` (min
skill values), and `perks` (prerequisite perk ids) — and may have multiple `ranks`.
Their `effects` compose into the character's active modifier set. The `type` field
says how each is acquired:

- **`perk`** — chosen on level-up (one per normal level). This is the primary
  character-defining advancement on every level.
- **`creation`** — chosen at character creation (the former "traits"), usually with a
  tradeoff (e.g. *Gifted*: +1 to all attributes but −10% skill points; *Genius*: +1
  tag skill but −15 starting skill points). Not offered on level-up.
- **`racial_trait`** — granted automatically by ancestry (pinned to a race).
- **`universal`** — always available.

All four live in the same perk table; "trait" is no longer a separate concept.
Racial features may also be authored inline on a race in `races/core.json`.

## Magic — 9 color schools

Magic uses **nine color schools**, each a normal point-buy skill in
`skills/core.json` named `<color>_magic`:

`red` · `orange` · `yellow` · `green` · `blue` · `indigo` · `violet` · `white` · `black`

(Engine vocabulary is color-named only; do **not** introduce school names like
"destruction" in ids or keys.)

- A **spell** (`spells/core.json`) is a composition of one or more atomic
  **effects** drawn from the shared **effect registry** (see below), each with
  magnitude, duration, and target shape, tagged with one `color`.
- The matching `<color>_magic` skill gates the spell's cost, power, and success
  chance. **Casting does not grant skill XP** — magic skills rise by point-buy like
  any other.
- **All nine schools are INT-seeded**: Intelligence sets the mana pool and the power
  of every spell, while **Willpower** regenerates mana and resists hostile magic
  (see `attributes.md`).
- **Spellmaking**: players combine effects into custom spells; cost derives from the
  sum of effects (`custom: true`).

## Effect registry — the shared pool

Effects are **atomic, reusable** building blocks defined once in
`ruleset/data/effects/core.json` (schema `effect.schema.json`) and drawn from by
every system that applies an effect — **magic, potions, poisons, coatings, oils,
food, ingredients, and enchantments**. This is the single source of truth — one
shared effect table for the whole game.

- Each effect declares a `category` (damage, restore, fortify, resist, cure,
  conjure, control, …), a `polarity` (beneficial / harmful / neutral — alchemy
  sorts beneficial → potions, harmful → poisons), an optional default `color`
  school, a `magnitudeUnit`, and a `channels` list naming which delivery vectors
  may use it (`spell`, `potion`, `poison`, `coating`, `ingredient`, …).
- Consumers reference an effect by id through the shared
  `schema.json#/definitions/appliedEffect` shape (`effect` + `magnitude` +
  `duration`, plus `target`/`radius` for spells):
  - **Spells** (`spells/core.json`) — `effects[]`.
  - **Consumables** (`equipment/consumables.json`) — `consumable.effects[]`, with a
    `delivery` of drink / eat / apply_weapon / throw.
  - **Ingredients** (`ingredients/core.json`) — `effects[]` (ids only; magnitude is
    derived at brew time from Alchemy skill). Brewing combines ingredients that
    share an effect.
- These id references are **not** schema-validated across files, so keep effect ids
  stable; a referential-integrity check (every referenced id exists and respects
  its `channels`) is run alongside `validate.py`.

## Elemental interaction — the five-phase cycle

Layered **on top of** the color schools (it does not replace them) is a five-phase
interaction system modeled on the Wuxing five-element cycle, defined in
`ruleset/data/wuxing/core.json` (schema `wuxing.schema.json`). The two axes are
**orthogonal**:

- **Color** picks the governing `<color>_magic` skill — what you train, the
  cost/power/success axis.
- **Phase** picks the *elemental relationship* — how an effect interacts with other
  effects in play. A Red-school `damage_fire` (phase **Fire**) and `damage_frost`
  (phase **Water**) share a skill but interact with the world differently.

`phase` is an **optional** field on effects (`effect.schema.json`), one of `wood`,
`fire`, `earth`, `metal`, `water`. Non-elemental effects (invisibility, telekinesis,
summons, attribute buffs) omit it and sit out the cycles entirely.

The five phases relate through four cycles (each a set of directed `from → to`
edges, where *from* acts upon *to*):

| Cycle | Interaction | Effect | Reading |
| --- | --- | --- | --- |
| **Generating** | amplify (×1.5 recipient) | mother feeds child | Wood→Fire→Earth→Metal→Water→Wood |
| **Overcoming** | suppress (×0.5 recipient) | controller restrains/dispels | Wood→Earth→Water→Fire→Metal→Wood |
| **Weakening** | drain (×0.75 recipient) | child saps mother (reverse generating) | Fire→Wood, Earth→Fire, Metal→Earth, Water→Metal, Wood→Water |
| **Insulting** | backlash (×1.25 recipient, conditional) | over-strong controlled rebels (reverse overcoming) | Earth→Wood, Water→Earth, Fire→Water, Metal→Fire, Wood→Metal |

The multipliers are **reference values** the consuming engine tunes; the cycle table
is the single source of truth for the matrix. This is a **runtime interaction
system** — riftweave-core supplies the vocabulary, the per-effect tags, and the
matrix; the engine implements resolution (it has no five-phase concept yet, so this
data leads that design). Alchemy/crafting can reuse the same cycles (generating →
potency, overcoming → cures/antidotes).

## Derived statistics (reference formulas)

A character runs on **three resources**. The third changes shape with the combat
mode: a turn-based (TTRPG) game uses **Action Points** (a shared pool), while an
action-combat game drains a **Stamina** pool. See `attributes.md` § The three
resources.

| Stat | Formula / Rule |
| --- | --- |
| Hit points | `15 + END × 8 + level × 4` (gains `END × 8 + 4` per level after 1st) |
| Mana (MP) | `INT × 8 + level × 2` (gains `INT × 8 + 2` per level after 1st; regen scales with WIL) |
| Stamina (SP, action combat) | `15 + END × 5 + level × 2` (gains `END × 5 + 2` per level after 1st; regen scales with WIL) |
| Action Points (TTRPG / turn-based) | Max 10 per turn. The pool carries over from turn to turn but cannot exceed 10. Starting pool at the beginning of combat or after a full rest is typically 2 + floor(AGI / 3) or as set by the table. |
| Carry weight | `25 + STR × 10` |
| Critical chance | `1% + LCK%` (+ perks) |
| Initiative | `PER` (+ perks such as *Alert*) |

### Additional Secondary & Combat Statistics

See `attributes.md` § Secondary & Combat Statistics for full details and rationale. The most relevant new or expanded ones for character sheets and engines are:

- **Movement Speed** (racial base + AGI/encumbrance; exploration vs combat units or speed rating).
- **Attack Speed** (AGI + weapon; % faster or extra actions/recovery; e.g. Fast Shot).
- **Magic Resistance** (WIL-based; % or resist bonus via the generalized `resist` effect with parameter "magic").
- **Physical & Elemental Damage Resistance** (armor DR + Wuxing phase resists from materials/effects).
- **Evasion / Dodge**, **Block Value**, **Critical Damage Multiplier**, **Encumbrance Penalties**, **Elemental Resistances** (per phase), **Status Resistance**, **Mana/Stamina Efficiency**.

**Combat & defense** are roll-under: an attacker rolls under their weapon skill to
hit; armor provides a damage-reduction **rating** (and Block can negate a share of a
blow). See [`combat.md`](./combat.md) for the full defense model (evasion → block → armor DR),
[`armor.md`](./armor.md) for the slot/layer DR system, and
[`conditions.md`](./conditions.md) for condition effects. Many secondary stats are
delivered through the shared **effect system**.

## Open items / follow-ups

- **Combat mode** — fully specified in [`combat.md`](./combat.md) with dual-resolution
  (TTRPG turn-based and video game action-combat). Turn-based uses Action Points from AGI;
  action-combat uses a Stamina pool from END. Initiative varies by mode.
- **Cultivation-style breakthroughs** — see the new `docs/modules/advancement.md`.
  Base level-up (skill points, resource scaling, perks) is described here. The
  optional item-boosted "breakthrough" system (monster cores + refined elixirs,
  driven by `phase` + `qualityGrade` and Wuxing cycles) lives in the dedicated
  advancement document. It lets players deliberately delay levels for stronger
  permanent gains.
- **Weapons** — Use two classification axes in `equipment/weapons.json` (full details and material variation rules in `docs/modules/weapons.md`):
  - Damage type (`damage.type`: slashing / bludgeoning / piercing) → maps to the split combat skills (`blades` / `blunt` / `piercing`).
  - Length category (`length`: short / normal / reach) + explicit `attackReach` (1 or 2 units of length). short and normal weapons are typically one-handed; reach weapons are two-handed.
  Base damage uses classic 3.5-style dice. Material-based variations (how each material turns the same base weapon into a meaningfully different version, with attack bonuses in the 0 to +10 range — none of the current materials reach the full +10) are explained in `docs/modules/weapons.md`.
  `category` (simple/martial) and free-form `properties` (finesse, versatile, heavy, two_handed, light, thrown, ...) are retained.
  Armor is still closer to legacy shape.
  (Consumables already use the shared effect pool.)
- **Effect magnitudes for brewing** — the Alchemy quality formula
  (`f(skill, ingredient_qualities, station_tier)`) that turns ingredient effect ids
  into concrete potion/poison magnitudes is not yet specified.
