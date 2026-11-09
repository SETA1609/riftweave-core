# Magic & Spell Crafting

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine

This document defines how magic works in Riftweave: how spells are built from
effects, how casting resolves, how players craft their own spells, and how the
**five-phase elemental layer** makes effects interact with each other, with
monsters, and with races. For the character-progression view (skills, point-buy,
attributes) see [`progression.md`](./progression.md). For *permanent magic items*
(enchanting), see [`magic-crafting.md`](./magic-crafting.md) — that is a separate
module.

> **Naming rule.** Engine vocabulary is **color-named** (`red_magic`, `blue_magic`,
> …) and **English-named** for phases (`wood`, `fire`, `earth`, `metal`, `water`).
> Do not introduce school names like "destruction" or non-English phase labels in
> ids or keys.

---

## 1. Two orthogonal axes

Every piece of magic is described by **two independent axes**. Keeping them separate
is the whole design:

| Axis | Field | Answers | Governs |
| --- | --- | --- | --- |
| **Color school** | `color` | *Which skill casts it?* | cost, success chance, power scaling |
| **Element (phase)** | `phase` | *How does it interact?* | amplify / suppress / drain / backlash vs other phased things |

A Red-school **Fire Bolt** (`color: red`, effect phase `fire`) and a Red-school
**Frost Bolt** (`color: red`, effect phase `water`) are trained by the *same*
`red_magic` skill but behave completely differently when they meet other elements.
Color never changes how elements interact; phase never changes which skill you
train.

### The nine color schools

Each color is a normal point-buy skill (`<color>_magic`, 0–100) in
`skills/core.json`. A school is defined by the kind of change its magic makes; the element
of that change is the other axis. The word in brackets is a functional parallel,
commentary only. Spells are listed by level, from `spells/core.json`. "Not in the data
yet" is what the skill's description promises and no effect or spell provides.
[V 2026-11-09]

**`red` (offense).** Harms at range with raw energy: fire, frost, shock, and force. Its
variety is shape, from one bolt to a whole battlefield, and it has no utility spell.
- Spells: Fire Bolt, Force Dart, Frost Lance, Fire Burst, Flame Shield, Chain Lightning,
  Ice Storm, Meteor, Inferno.

**`orange` (conjuration).** Makes what was not there, for as long as the spell lasts:
spirit creatures, elementals, weapons of light, and barriers.
- Spells: Summon Familiar, Bound Blade, Conjured Barrier, Conjured Blades, Bound Bow,
  Summon Elemental, Conjured Legion.

**`yellow` (illusion).** Changes what others see, hear, and believe: light and darkness of
the senses, false friendship, invisibility, and pain that exists only in the mind.
- Spells: Light, Blinding Flash, Blur, Mind Spike, Charm Person, Silence, Veil, Psychic
  Scream, Mass Charm.
- Not in the data yet: calm.

**`green` (restoration).** Mends the mortal body: health, stamina, poison and its cures. It
deals no damage.
- Spells: Mend Wounds, Cure Poison, Vigor, Regeneration, Purity Ward, Greater Heal, Mass
  Heal, Renewal.

**`blue` (alteration).** Changes the physical state of a body or thing: how heavy, how
hard, how fast, or what it is made of. Its damage is matter altered into acid; its
strongest spell turns flesh to stone.
- Spells: Feather Fall, Ward, Corrode, Slow, Stoneskin, Paralyze, Acid Rain, Petrify.
- Not in the data yet: water-walking, transmutation.

**`indigo` (force).** Pushes, blocks, and shakes with unseen force and sound, and works
locks and mechanical traps from a distance.
- Spells: Unlock, Disarm Trap, Thunderclap, Force Push, Force Wall, Greater Unlock, Greater
  Disarm, Sonic Boom.
- Not in the data yet: levitation.

**`violet` (mysticism).** Works on magic itself and on what magic connects: unravels and
absorbs spells, drains mana and will, moves objects by thought.
- Spells: Telekinesis, Dispel Magic, Sap Will, Spell Absorption, Dispel Ward, Mana Drain,
  Greater Dispel.
- Not in the data yet: Mark and Recall (planned effects, M3-06), soul detection and storage.

**`white` (holy).** Divine power that shelters allies and burns the unholy: radiant damage,
blessings, wards, and the cure of disease. Opposed by black.
- Spells: Holy Light, Blessing, Smite, Holy Ward, Cleanse, Sunburst, Divine Shield.
- Not in the data yet: banishing undead and demons, divine-grade healing, sanctify.

**`black` (necromancy).** Death and what it leaves behind: necrotic damage, curses, stolen
strength, armor of bone, and corpses raised to fight. Opposed by white.
- Spells: Drain Life, Curse, Bone Armor, Raise Thrall, Wither, Raise Horde, Death Grip.
- Not in the data yet: soul-corruption.

Where two schools come close, this is the line between them:

| Both touch | Who does what |
| --- | --- |
| Force | Red strikes with it (Force Dart); indigo pushes and blocks with it. |
| Summoning | Orange calls spirits and elementals; black raises the dead. |
| Moving things | Indigo shoves with force; violet lifts by thought. |
| Drains | Violet takes mana and will; black takes life and strength. |
| Cures | Green cures poison; white cures disease. |

### The five phases

`wood · fire · earth · metal · water` — defined in `schema.json#/definitions/phase`,
related through the cycles in [`data/wuxing/core.json`](../../ruleset/data/wuxing/core.json).
`phase` is **optional** on effects: non-elemental effects (invisibility,
telekinesis, summons, pure attribute buffs) carry none and sit out the cycles.

---

## 2. Effects — the atomic pool

A spell does not contain mechanics directly; it **references effects** from the
shared registry [`data/effects/core.json`](../../ruleset/data/effects/core.json)
(schema `effect.schema.json`). The same pool feeds potions, poisons, coatings,
ingredients, food, and enchantments — see [`progression.md`](./progression.md)
§ Effect registry.

An effect declares: `category`, `polarity`, an optional default `color`, an optional
`phase`, a `magnitudeUnit`, and the `channels` allowed to draw it. A spell may only
use an effect whose `channels` include `"spell"`.

---

## 3. Spells — composition

A spell (`spells/core.json`, schema `spell.schema.json`) is **one `color` + one or
more applied effects**:

```json
{
  "id": 1,
  "key": "fire_bolt",
  "label": "Fire Bolt",
  "color": "red",
  "effects": [ { "effect": 1, "magnitude": 8, "target": "ranged" } ],
  "cost": { "mana": 5 }
}
```

Each entry under `effects` is a `schema.json#/definitions/appliedEffect`: an effect
id plus its `magnitude`, `duration`, `target` shape (`self`/`touch`/`ranged`/`area`/
`world`), and `radius` for areas. A spell's **phase(s)** are read from its composed
effects — there is no separate phase field on the spell. A multi-effect spell can
therefore carry multiple phases.

`level` (1–30) places the spell on the level scale; dispel reads it. `concentration`
marks a spell the caster holds. `cost`, `cooldown`, and the `custom` flag round out the
spell. `cost` is the
authored baseline; the engine derives it (§5).

### The core spell list

`spells/core.json` holds at least five spells per color, 45 or more in all, from level 1 to
30. An authored spell draws only effects on its own color or with no color, so the list
teaches which school casts what; a crafted spell may be off-color only where a game allows it
(§5). [V 2026-11-08]

Every color has a damage, a utility, and a defensive spell, or a reason it has none:

| Color | Damage | Utility | Defensive |
| --- | --- | --- | --- |
| red | fire, frost, shock, force | none: red is the offense school; its range is shape (bolt, burst, storm) | flame shield |
| orange | conjured blades (bleed) | familiar, bound weapons, summoned elemental | conjured barrier |
| yellow | psychic | light, blinding flash, charm, silence, veil | blur |
| green | none: restoration does not harm; life taken is black's | cures, vigor | regeneration, purity ward, heals |
| blue | acid (corrosion) | feather fall, slow, paralyze, petrify | ward, stoneskin |
| indigo | thunder | unlock, disarm trap, force push | force wall |
| violet | mana and will drain | dispel, dispel ward | spell absorption |
| white | radiant | blessing, cleanse | holy ward, divine shield |
| black | necrotic | curse, wither, raise thrall | bone armor |

The damage effects `damage_radiant` (white), `damage_necrotic` (black), `damage_psychic`
(yellow), `damage_thunder` (indigo), and `damage_acid` (blue) join the pool so that every
color but green deals damage of its own; each matches a key in `combat/damage_types.json`.
`disarm_hazard` and `unlock` (indigo) are the spell side of traps and locks
([hazards.md](../core/hazards.md#spells-on-a-hazard)). `test_spells.py` checks the floor,
the colors, the channels, and this table.

Every spell, authored or crafted, costs what §5's formula gives for its effects. A spell's `level`
(1–30) is its power band, which dispel reads; it does not set the cost.

---

## 4. Casting resolution (d100 roll-under)

The procedure is [docs/core/casting.md](../core/casting.md): one resolution request on
the spell's `<color>_magic` skill, after the cast spends its mana and reagents.

1. **Gate.** A `silenced` caster cannot cast. Too little mana, or a missing reagent,
   means no cast and nothing spent.
2. **Cost.** Mana (`INT × 8 + level × 2` is the pool; it regenerates with WIL) and the
   spell's reagents are spent before the roll, so a failed cast still costs them.
3. **Success.** The target is the skill, plus the armor penalty of the heaviest armor
   worn, plus the modifiers on the skill. Crit and fumble are tags from
   [docs/core/resolution.md](../core/resolution.md). A phase is a tag. It is not
   required to cast.
4. **Effects.** A success emits the spell's effects at their authored magnitudes. The
   margin is in the result; scaling a magnitude by it is not core.
5. **Concentration and dispel.** A `concentration` spell is held, one at a time. A
   `dispel` effect of magnitude N ends the target's standing spells of level N or
   below.
6. **Elemental interaction** (§6) then adjusts the realized magnitude/duration based
   on the target's element and any standing effects.

Counterspell is cut (there is no reaction slot), and ritual casting is a module after
1.0. Casting **never grants skill XP** — magic skills rise by point-buy like every other
skill (advancement, not use-training).

---

## 5. Spell crafting (spellmaking)

Players assemble custom spells by combining effects. A crafted spell is an ordinary
spell entry with **`custom: true`**; it is subject to the same schema and the same
casting rules.

### Rules

- **Single color.** Every effect in a custom spell must be castable under one chosen
  color. (An effect's default `color` is the natural fit; a game may allow
  off-color use at a skill penalty.)
- **Channel gate.** Each chosen effect must list `"spell"` in its `channels`.
- **Skill gate.** The spell is only castable once the matching `<color>_magic` skill
  meets the spell's **complexity threshold** (below). Spellmaking the recipe may
  require a higher threshold than casting an authored equivalent.
- **No new mechanics.** Crafting composes *existing* effects at chosen magnitudes,
  durations, and targets. Inventing a new effect is a content/authoring change to the
  pool, not a player action.

- **No saves.** A crafted spell cannot carry a save: saves are for authored instant spells
  (M3-05). A crafted spell is avoided only by a dodge or a block.

### Cost formula

A spell's mana cost comes from its effects alone: what they do, for how long, and how far
they reach. The character's level never enters it; a level-1 and a level-30 caster pay the
same for the same spell, before their own skill and perks (below). The coefficients are in
`ruleset/data/casting/core.json` (`spell_cost`). [V 2026-11-08]

```
per effect
  m          = magnitude, or 1 for an effect whose magnitude is not a strength
               (magnitudeUnit seconds, or none: a stun's strength is its duration)
  cost       = 0.1 × baseCost × m × duration          duration in seconds, 1 = instant
  cost      += 0.05 × baseCost × radius               area only
  cost      ×= target multiplier                       self 1, touch 1, ranged 1.5, area 2
spell cost   = ceil(sum over the effects)              cost.mana on the row
```

- **`baseCost`** is on each effect row: how much one point of it for one second is worth.
  It carries the difference between a point of fire and a second of paralysis, so no unit
  weights are needed. Light is 0.2, a point of fire 5, a second of paralysis 40.
- **Duration** multiplies, so a pool effect is priced by its total (10 health a second for
  10 seconds costs what 100 at once does) and a buff by how long it holds.
- **Riders** cost what they do: a 6-second blind on a burst is priced like a blind spell.

**Sources.** The shape is Morrowind's: OpenMW's `calcEffectCost` multiplies the mean
magnitude by `0.1 × baseCost` and by the duration, adds `0.05 × area × baseCost`, and
multiplies an effect cast on a target by 1.5; an effect with no magnitude counts it as 1
(checked against OpenMW/openmw `apps/openmw/mwmechanics/spellutil.cpp` at `b8285cc95d19`,
2026-11-08). The per-effect base costs follow its ratios (UESP, Morrowind:Magic Effects,
2026-11-08: Light 0.2, Fire Damage 5, Damage Health 8, Paralyze and Silence 40), scaled to
our magnitudes. The area multiplier is Daggerfall's: Daggerfall Unity's
`ApplyTargetCostMultiplier` is caster and touch 1.0, single target at range 1.5, area around
the caster 2.0, area at range 2.5; we have one area shape and take 2.0 (Interkarma/
daggerfall-unity `Assets/Scripts/Game/Formulas/FormulaHelper.cs` at `4e78e1412b17`,
2026-11-08). Not taken: Daggerfall's magnitude and duration that grow per caster level
(`plus` per `perLevel` levels), Morrowind's `fEffectCostMult`, and its offset of one second
on a player-made spell's duration. [V 2026-11-08]

### The caster's side

What a caster pays is the spell cost after their skill and perks, never their level.
Daggerfall lowers spell points as the skill rises (`gold × (110 − skill) / 400`); here the
skill discount is `<color>_magic / 200`, capped at 0.45, and perks emit modifiers on
`spell_cost`. Applying them in the cast is M3-07; until then the cast spends `cost.mana` as
authored. The cast threshold and the craft threshold read the spell cost before any
discount:

```
castThreshold   = round( spell cost )           // min <color>_magic to cast at all
craftThreshold  = round( spell cost × 1.25 )    // min <color>_magic to author the recipe
```

Uncraftable effects (Divine Intervention, Mark and Recall, M3-06) have high base costs
in Morrowind (150, 350); here they are authored only and not priced for crafting.

---

## 6. Elemental interaction — the five-phase layer

This is the part that makes effects "interact with each other." It is fully defined
in [`data/wuxing/core.json`](../../ruleset/data/wuxing/core.json) (the single source
of truth) and summarised in [`progression.md`](./progression.md) § Elemental
interaction. When a **phased** effect resolves, the engine compares its phase against
the relevant other phase and applies the matching cycle's multiplier:

| Cycle | When it fires | Multiplier (reference) |
| --- | --- | --- |
| **Generating** | acting phase nourishes the other (mother → child) | amplify ×1.5 |
| **Overcoming** | acting phase controls the other (e.g. Water → Fire) | suppress ×0.5 (dispel if overwhelming) |
| **Weakening** | a child phase drains its mother | drain ×0.75 |
| **Insulting** | the normally-controlled phase is over-strong and rebels | backlash ×1.25 (conditional) |

The "other phase" can come from three places:

1. **Effect vs standing effect.** Cast `damage_fire` on a target already burning with
   a Wood effect → Generating amplifies the fire. Cast a Water effect on a standing
   Fire effect → Overcoming suppresses/dispels it.
2. **Effect vs monster element.** Every monster has exactly one `phase`
   (`monster.schema.json`). A Water spell against an `ember_drake` (Fire) lands under
   Overcoming → boosted; a Fire spell against it falls under Weakening/affinity →
   reduced. This is the core of elemental counterplay.
3. **Effect vs race element.** Every playable race has one `phase`
   (`race.schema.json`) — the same rules give player characters innate elemental
   strengths and soft spots. (Abstract `parent`/`template` race entries are not
   played and may omit it; subraces inherit or override their parent's phase, and
   beastfolk kin each declare their own.)

> **Direction convention.** In the cycle data an edge `from → to` is the canonical
> reading "*from* acts upon *to*". The engine maps the **incoming** effect and the
> **standing** effect/target onto `from`/`to` by matching the edge, then applies the
> multiplier to the participant named by the cycle's `affects` field.

### Worked examples

- **Frost Bolt vs Ember Drake.** Frost = `water`, drake = `fire`. Water overcomes
  Fire → Overcoming ×0.5 *suppression of the drake's element*: the bolt bites deep.
- **Fire Bolt vs Ember Drake.** Fire vs Fire, and Fire is *fed by* Wood / *drains
  into* Earth — same-element attacks land at affinity (reduced). Bring Water instead.
- **Combo setup.** Apply a Wood damage-over-time, then a Fire spell: Generating
  ×1.5 — the mother phase feeds the child, and the fire flares.
- **Insult backlash.** Hit a Water-phase `frost_revenant` with an overwhelming Fire
  spell whose magnitude exceeds the target's: the controlled Water *insults* back
  ×1.25 — high-risk, high-reward against a resistant element.

---

## 7. Data & schema map

| Concern | Schema | Data |
| --- | --- | --- |
| Atomic effects (+ `phase`) | `effect.schema.json` | `data/effects/core.json` |
| Spells (+ spellmaking `custom`) | `spell.schema.json` | `data/spells/core.json` |
| Phase vocabulary | `schema.json#/definitions/phase` | — |
| Interaction matrix | `wuxing.schema.json` | `data/wuxing/core.json` |
| Monster elements | `monster.schema.json` | `data/monsters/core.json` |
| Race elements | `race.schema.json` | `data/races/core.json` |
| Color skills | `skill.schema.json` | `data/skills/core.json` |

---

## 8. Open items / follow-ups

- **Engine resolution.** Phases, the matrix, and the elements on every monster/race
  live here; the consuming engine has no five-phase concept yet and must implement
  the resolution in §6. This data **leads** that design.
- **Tuning.** All multipliers, weights, and thresholds above are reference values,
  not balanced numbers.
- **Off-color casting.** Whether (and how steeply) an effect may be cast under a
  non-default color is left to the game.
- **Phasing the rest of the pool.** Consumables and ingredients can opt into the same
  `phase` tags so alchemy reuses the cycles (Generating → potency, Overcoming →
  cures/antidotes). Not yet tagged.
- **Multi-phase spells.** Resolution order when one spell carries several phases is
  engine-defined (suggested: resolve each effect independently against the target).
