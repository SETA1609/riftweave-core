# Character Creation

**Status:** Core (reference procedure)

This document is the **step-by-step sequence** for building a 1st-level character from
the existing data. It assembles the rules defined in detail elsewhere —
[`progression.md`](./progression.md) (attributes, skills, perks, formulas),
[`race.md`](./race.md) (ancestry & lineage), and [`magic.md`](./magic.md) (spells) —
into one ordered checklist. Numbers here are the **ruleset reference**; a consuming
game may override them.

> All of creation draws only on existing collections: `abilities`, `races`, `skills`,
> `features` (traits/perks), `spells`, `equipment`. Nothing here needs new schema.

---

## The sequence at a glance

1. [Choose a race](#1-choose-a-race) → ability modifiers, element, speed, size, traits
2. [Assign attributes](#2-assign-attributes) → 8 attributes, 1–10
3. [Choose traits](#3-choose-traits-optional) → creation-only, with tradeoffs
4. [Pick 3 tag skills](#4-pick-3-tag-skills)
5. [Compute starting skills & spend the level-1 pool](#5-starting-skills--the-level-1-pool)
6. [Take starting perks](#6-starting-perks) → usually none at level 1
7. [Derive statistics](#7-derive-statistics) → HP, mana, carry, crit, initiative
8. [Learn spells & take equipment](#8-spells--equipment)
9. [Record your element](#9-record-your-element)

---

## 1. Choose a race

Pick a **playable** race from [`races/core.json`](../../ruleset/data/races/core.json)
— one whose `lineage.role` is `standalone`, `subrace`, or `kin`. The abstract
`parent` (elf, dwarf) and `template` (beastman) entries are **not** played directly;
their playable forms are the baseline subraces `common_elf` / `common_dwarf`, the
differentiated subraces (green/black elf, hill/mountain dwarf), and the beastfolk kin
(catman, wolfman, …). See [`race.md`](./race.md) for the lineage model.

For a **subrace**, resolve inheritance against its parent first (see
[`race.md`](./race.md) §3): the parent's `speed`/`size` carry over unless overridden,
`abilityModifiers` **sum**, and `traits` form the **union**.

Record from the resolved race:

- **`abilityModifiers`** — applied in step 2.
- **`phase`** — your element (step 9).
- **`speed`**, **`size`**.
- **`traits`** — e.g. an elf's *Darkvision* / *Fey Ancestry*, a catman's *Retractable
  Claws*. (Human's *Versatility* grants **+1 tag skill** in step 4, and *Ambition*
  gives the **option of +1 perk** in step 6.)

---

## 2. Assign attributes

The eight attributes (`str per end int wil agi cha lck`, defined in
[`abilities/core.json`](../../ruleset/data/abilities/core.json)) run 1–10.

1. **Start** every attribute at **4**.
2. **Distribute 10 points** among them, to a maximum of **10** in any one.
3. **Apply racial `abilityModifiers`** from step 1 (these can push a stat above the
   point-buy cap or below 4).

Order matters: spend your 10 points first, then layer the racial deltas on top.

---

## 3. Choose traits (optional)

Traits are `features` with `type: "trait"` — chosen **only at creation**, each a
double-edged deal. Currently available:

| Trait | Upside | Downside |
| --- | --- | --- |
| **Gifted** | +1 to every attribute | −10% skill points per level |
| **Fast Shot** | ranged attacks 20% faster | no aimed/targeted shots |
| **Small Frame** | +1 Agility | −25% carrying capacity |

Take **0–2** (suggested cap: 2). Trait attribute bonuses apply now, alongside the
racial modifiers, and may exceed the normal caps.

---

## 4. Pick 3 tag skills

Choose **3 tag skills** from [`skills/core.json`](../../ruleset/data/skills/core.json)
(any skill with `"taggable": true` — currently all of them). Humans pick a **4th**
(Versatility).

A tag skill costs the **same** skill point but yields **+2 per point** instead of
**+1** — an ongoing discount, *not* a flat starting bonus. Tagging is therefore about
long-term efficiency in your focus skills.

The 25 skills span seven categories — combat (`melee_weapons`, `marksmanship`,
`block`, `unarmed`), the nine `<color>_magic` schools, stealth, social, knowledge,
survival, and utility (`alchemy`, `repair`, `athletics`).

---

## 5. Starting skills & the level-1 pool

**Base rating** for every skill (before spending points):

```
start = 5 + associatedAbility × 2
```

Luck is **not** baked in here — it is applied at check time as `+floor(LCK/2)`.

Then spend your **1st-level skill-point pool**:

```
points = 5 + INT × 2 + random(0 … LCK)         // Gifted: ×0.9
spend:  tagged skill +2 per point · untagged +1 per point
```

Skills cap at **100**. (Whether a fresh character receives the level-1 pool or starts
at bare base values is a per-game call; the reference assumes you get and spend it.)

---

## 6. Starting perks

Perks (`features` with `type: "perk"`) are the character-defining progression, gated
by `prerequisite` (`level` / `abilities` / `skills` / `perks`). The suggested cadence
is **one perk every 3 levels**, so a 1st-level character normally has **none** —
their skill prerequisites (e.g. *Pyromancer* needs `red_magic ≥ 50`) are out of reach
at creation anyway.

Exception: **Humans** *may* take **one perk at creation** (their *Ambition* trait) —
it is an option, not a requirement. Any perk whose prerequisites they already meet
qualifies; at level 1 that means an ability-only perk such as *Toughness* (`end ≥ 5`)
or *Alert* (`per ≥ 6`). A human who has no perk worth taking yet may decline and save
it for later.

Note: **traits ≠ perks**. Traits are the creation-only choices in step 3; perks come
later.

---

## 7. Derive statistics

From your final attributes (`progression.md` § Derived statistics), at `level = 1`:

| Stat | Formula | At level 1 |
| --- | --- | --- |
| Hit points | `15 + END × 8 + level × 4` | `19 + END × 8` |
| Mana | `INT × 8 + level × 2` (regen via WIL) | `2 + INT × 8` |
| Stamina *(action combat)* | `15 + END × 5 + level × 2` (regen via WIL) | `17 + END × 5` |
| Action Points *(turn-based)* | `2 + floor(AGI / 3)` per turn | — |
| Carry weight | `25 + STR × 10` | — |
| Critical chance | `1% + LCK%` | — |
| Initiative | `PER` | — |

These are your **three resources** — HP, Mana, and a third that depends on combat
mode (Action Points in turn-based play, a Stamina pool in action combat). Luck also
adds **+floor(LCK/2)%** to every check at resolution time.

---

## 8. Spells & equipment

- **Spells.** If you tagged or raised a `<color>_magic` skill, you can learn spells of
  that color from [`spells/core.json`](../../ruleset/data/spells/core.json) (e.g.
  `fire_bolt`, `mend_wounds`). **All nine color schools are INT-seeded** — INT sets
  the mana pool and spell power, while WIL regenerates mana and resists magic. Casting
  uses d100 roll-under against the color skill; see [`magic.md`](./magic.md).
- **Equipment.** Take starting gear from
  [`equipment/`](../../ruleset/data/equipment/) (weapons, armor, consumables). The
  specific starting loadout is left to the game.

---

## 9. Record your element

Your race's **`phase`** (`wood` / `fire` / `earth` / `metal` / `water`) is your
character's element. It feeds the five-phase interaction cycles in
[`data/wuxing/core.json`](../../ruleset/data/wuxing/core.json): incoming phased
effects resolve against it via Generating / Overcoming / Weakening / Insulting (see
[`magic.md`](./magic.md) § Elemental interaction). Treat it as an affinity, not a
flat resistance — the cycles compute each matchup.

---

## Worked example — "Borin", a dwarven shield-bearer

1. **Race** — `common_dwarf` (baseline subrace, inherits the dwarf parent): phase
   **metal**, speed 25, medium; modifiers `{ end: +2, str: +1, agi: −1 }`; traits
   *Darkvision*, *Dwarven Resilience*, *Stoneworker*.
2. **Attributes** — start all at 4; distribute +10 as
   str +2, end +2, per +1, int +1, agi +1, lck +3; then apply racial mods. Final:
   **STR 7 · PER 5 · END 8 · INT 5 · WIL 4 · AGI 4 · CHA 4 · LCK 7**.
3. **Traits** — none.
4. **Tag skills** — `melee_weapons`, `block`, `athletics`.
5. **Skills** — bases `5 + abil×2`: melee 19, block 21, athletics 19. Level-1 pool
   `5 + 5×2 + random(0…7)` → say a roll of 4 = **19 points**, all on tagged skills
   (+2/pt): melee +16 → **35**, block +12 → **33**, athletics +10 → **29**.
6. **Perks** — none (not human).
7. **Derived** — HP `15 + 8×8 + 4` = **83**; Mana `5×8 + 2` = **42**; third resource:
   Stamina `15 + 8×5 + 2` = **57** (action combat) *or* Action Points `2 + floor(4/3)`
   = **3** per turn (turn-based); Carry `25 + 7×10` = **95**; Crit `1 + 7` = **8%**;
   Initiative **5**.
8. **Spells / gear** — no color skill tagged → no spells; takes a warhammer, shield,
   and mail.
9. **Element** — **metal**: strong against Wood-phase foes (Metal overcomes Wood),
   soft against Fire (Fire melts Metal).

---

## Data & schema map

| Step | Data | Schema |
| --- | --- | --- |
| Race | `data/races/core.json` | `race.schema.json` |
| Attributes | `data/abilities/core.json` | `ability.schema.json` |
| Skills / tags | `data/skills/core.json` | `skill.schema.json` |
| Traits & perks | `data/features/core.json` | `feature.schema.json` |
| Spells | `data/spells/core.json` | `spell.schema.json` |
| Equipment | `data/equipment/*.json` | `equipment.schema.json` |
| Element cycles | `data/wuxing/core.json` | `wuxing.schema.json` |

## Open items

- **Level-1 skill pool.** Whether creation grants the first level's skill points (as
  assumed here) or starts at bare base ratings is a tunable per-game choice.
- **Starting loadout.** No canonical starting-equipment table exists yet; games
  define their own from the equipment collections.
- **Point-buy spread.** The "start at 4, distribute 10, max 10" spread is the
  reference; alternatives (rolled, fixed arrays) are easy variants.
