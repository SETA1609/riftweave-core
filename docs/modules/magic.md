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
`skills/core.json`. Functional parallels are commentary only:

`red` (offense) · `orange` (conjuration) · `yellow` (illusion) · `green`
(restoration) · `blue` (alteration) · `indigo` (force/levitation) · `violet`
(mysticism) · `white` (holy) · `black` (necromancy).

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
  "id": "fire_bolt",
  "name": "Fire Bolt",
  "color": "red",
  "effects": [ { "effect": "damage_fire", "magnitude": 8, "target": "ranged" } ],
  "cost": { "mana": 5 }
}
```

Each entry under `effects` is a `schema.json#/definitions/appliedEffect`: an effect
id plus its `magnitude`, `duration`, `target` shape (`self`/`touch`/`ranged`/`area`/
`world`), and `radius` for areas. A spell's **phase(s)** are read from its composed
effects — there is no separate phase field on the spell. A multi-effect spell can
therefore carry multiple phases.

`cost`, `cooldown`, and the `custom` flag round out the spell. `cost` is the
authored baseline; the engine derives it (§5).

---

## 4. Casting resolution (d100 roll-under)

Casting uses the standard resolution from [`progression.md`](./progression.md):

1. **Cost** is paid from the mana pool (`INT × 8 + level × 2`; regenerates with WIL).
   Insufficient mana → no cast.
2. **Success** — roll `1d100`, succeed if `≤ target`, where
   `target = <color>_magic + floor(LCK/2) + situational`. Crit on `01–05`, fumble on
   `96–00`.
3. **Potency** — `margin = target − roll` scales the realized magnitude (a wide
   margin overcharges the effect; a narrow one underdelivers).
4. **Elemental interaction** (§6) then adjusts the realized magnitude/duration based
   on the target's element and any standing effects.

Casting **never grants skill XP** — magic skills rise by point-buy like every other
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

### Reference cost formula

These are **reference values the engine tunes**, in the spirit of the rest of the
ruleset. Per applied effect:

```
effectCost = magnitude × unitWeight(magnitudeUnit)
           + durationSeconds × 0.1
effectCost ×= targetMultiplier      // self 1.0, touch 1.0, ranged 1.25, area 1.5 (+0.1 per metre radius), world 3.0

spellComplexity = sum(effectCost over all effects)
manaCost        = ceil( spellComplexity × (1 − skillDiscount) )   // skillDiscount = <color>_magic / 200, capped at 0.45
castThreshold   = round( spellComplexity )                        // min <color>_magic to cast at all
craftThreshold  = round( spellComplexity × 1.25 )                 // min <color>_magic to author the recipe
```

`unitWeight` example weights: `points` 1.0, `percent` 0.5, `dice` 1.0, `count` 8.0,
`seconds` 0.25, `level` 6.0. Higher skill lowers mana cost but never the thresholds.

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
