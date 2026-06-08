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
| `end` | Endurance | hit points, fatigue, poison/disease resistance |
| `int` | Intelligence | skill-point rate, mana pool, knowledge/tech skills |
| `wil` | Willpower | magicka regen, magic resistance, willpower magic |
| `agi` | Agility | action economy, stealth, dexterous skills |
| `cha` | Charisma | persuasion, barter, leadership |
| `lck` | Luck | crits, loot, a flat bonus to every check, bonus skill points |

This is the Fallout **SPECIAL** set plus **Willpower** (to give magic a home
without overloading another stat). Suggested creation: each attribute starts at
**4**, distribute **10** points (max 10 each), then apply racial modifiers from
`races/core.json`.

## Skills (0–100, point-buy)

Defined in `ruleset/data/skills/core.json`, each `associatedAbility`-seeded and
tagged with a `category` (combat, magic, stealth, social, knowledge, survival,
utility).

- **Starting value** = `5 + associatedAbility × 2` (Luck is applied at check time,
  not stored in the rating).
- **XP is event-driven** — quests, kills, exploration milestones grant XP via the
  event bus. *Using* a skill does **not** train it. (This is the Fallout 1/2 model,
  explicitly **not** Morrowind/Skyrim use-leveling.)
- On **level-up** you get a pool of skill points to spend:
  - **skill points/level = `5 + INT × 2 + random(0…LCK)`** — Luck adds a random
    bonus, so a luckier character gains more points per level on average, but
    unpredictably.
- **Tag skills** — choose **3** at creation. Spending 1 point on a tagged skill
  grants **+2**; an untagged skill grants **+1**. Same pool, double yield (this is
  the F1/2 discount, **not** New Vegas's +15-at-start). Mark a skill non-taggable
  with `"taggable": false`.
- Skills cap at **100** in v1 (above-100 cost escalation is a per-game extension).

## Luck

Luck is deliberately diffuse so it always helps but never dominates:

- **+`floor(LCK / 2)`%** to every check's target number.
- **+1% critical-hit chance** per point of Luck.
- A **random bonus to skill points** each level (see formula above).
- Better loot quality and more favorable random outcomes.

## Perks (and traits)

Perks are the **character-defining** progression in this classless system, defined
as `features` in `ruleset/data/features/core.json` (`type: "perk"`). They are gated
by `prerequisite` — any of `level`, `abilities` (min attribute values), `skills`
(min skill values), and `perks` (prerequisite perk ids) — and may have multiple
`ranks`. Their `effects` compose into the character's active modifier set.

- **Perks** are chosen on level-up (suggested cadence: one every 3 levels).
- **Traits** (`type: "trait"`) are chosen at creation and usually carry a tradeoff
  (e.g. *Gifted*: +1 to all attributes but −10% skill points).
- **Racial features** come from `races/core.json`.

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
- **Spellmaking**: players combine effects into custom spells; cost derives from the
  sum of effects (`custom: true`).

## Effect registry — the shared pool

Effects are **atomic, reusable** building blocks defined once in
`ruleset/data/effects/core.json` (schema `effect.schema.json`) and drawn from by
every system that applies an effect — **magic, potions, poisons, coatings, oils,
food, ingredients, and enchantments**. This is the single source of truth (much
like the Elder Scrolls shared magic-effect table).

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

| Stat | Formula |
| --- | --- |
| Hit points | `15 + END × 8 + level × 4` |
| Mana | `INT × 5 + WIL × 3` (regeneration scales with WIL) |
| Carry weight | `25 + STR × 10` |
| Critical chance | `1% + LCK%` (+ perks) |
| Initiative | `PER` (+ perks such as *Alert*) |

**Combat & defense** are roll-under: an attacker rolls under their weapon skill to
hit; armor provides a damage-reduction **rating** (and Block can negate a share of a
blow). A full defense model (dodge vs. block vs. armor) is a follow-up.

## Open items / follow-ups

- **Combat mode** — turn-based (with Action Points from AGI) vs. real-time is
  undecided; the AP formula and initiative detail depend on it.
- **Weapons & armor** — `equipment/weapons.json` and `armor.json` are still on the
  original D&D shape (`acBase`, `acDexBonus`, weapon categories). Aligning them to
  the d100 model (armor as a defense/DR %, weapons declaring their governing skill)
  is pending. (Consumables already use the effect pool.)
- **Effect magnitudes for brewing** — the Alchemy quality formula
  (`f(skill, ingredient_qualities, station_tier)`) that turns ingredient effect ids
  into concrete potion/poison magnitudes is not yet specified.
