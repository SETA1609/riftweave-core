# Character Creation

**Status:** Core (reference procedure)

This document is the **step-by-step sequence** for building a 1st-level character from
the existing data. It assembles the rules defined in detail elsewhere —
[`progression.md`](./progression.md) (attributes, skills, perks, base level gains, full level-up procedure),
[`advancement.md`](./advancement.md) (breakthroughs with cores/elixirs),
[`backgrounds.md`](./backgrounds.md) (prior life, training packages, starting gear & spells),
[`race.md`](./race.md) (ancestry & lineage), and [`magic.md`](./magic.md) (spells) —
into one ordered checklist. Numbers here are the **ruleset reference**; a consuming
game may override them.

Creation choices have long-term weight: your race phase becomes your Wuxing foundation for breakthroughs, INT and LCK compound through every future skill-point pool and success chance, and your three tag skills lock in permanent 1:1 efficiency for the rest of the character's career.

> All of creation draws only on existing collections: `abilities`, `races`, `skills`,
> `backgrounds`, `features` (traits/perks), `spells`, `equipment`. Nothing here needs new schema.

---

## The sequence at a glance

1. [Choose a race](#1-choose-a-race) → ability modifiers, element, speed, size, traits
2. [Assign attributes](#2-assign-attributes) → 8 attributes, 1–10
3. [Choose a background](#3-choose-a-background) → prior life, training, starting skill bonuses + gear + spells
4. [Choose creation perks](#4-choose-creation-perks-optional) → creation-only perks, with tradeoffs
5. [Pick 3 tag skills](#5-pick-3-tag-skills)
6. [Compute starting skills & spend the level-1 pool](#6-starting-skills--the-level-1-pool)
7. [Take starting perks](#7-starting-perks) → usually none at level 1
8. [Derive statistics](#8-derive-statistics) → HP, mana, carry, crit, initiative
9. [Learn spells & take equipment](#9-spells--equipment)
10. [Record your element](#10-record-your-element)

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
- **`phase`** — your **element / foundation**. This is your Wuxing affinity for elemental interactions (see `magic.md`) and becomes the reference for phase synergy on future breakthroughs (see `advancement.md` § Phase → Growth Affinity and the success formula).
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

## 3. Choose a background

A **background** represents your character's prior life, profession, or origin. It grants a one-time package of skill bonuses (added on top of the normal attribute-seeded base), starting equipment, a small number of known spells (for appropriate concepts), and occasionally a minor feature or extra wealth.

Backgrounds are **optional**. Choose one from [`backgrounds/core.json`](../../ruleset/data/backgrounds/core.json) or skip this step for a blank-slate character.

The bonuses are flat additions at creation only — they do **not** consume your level-1 skill point pool and do not change how you spend points at 2:1 (normal) or 1:1 (tagged) ratios. See [`backgrounds.md`](./backgrounds.md) for the full design and many more suggestions.

Record what the background grants:
- **Skill bonuses** — add these directly to the affected skills after calculating their normal starting value (`5 + associatedAbility × 2`).
- **Starting equipment** — add the listed items (and quantities) to your inventory.
- **Starting spells** — add these to your known spells (you still need the governing magic skill to cast them well).
- **Wealth bonus** — add to your starting gold (the reference baseline is 100 gp + basic kit).
- **Suggested tags** — purely advisory; helps you pick coherent tag skills in the next step.

---

## 4. Choose creation perks (optional)

**Creation perks** (the former "traits") are perks in the perk table flagged
`type: "creation"` — selectable **only at creation**, each a double-edged deal.
These are the main way to get strong boons at the cost of meaningful drawbacks (in the style of classic advantages/disadvantages systems).

Currently available:

| Creation perk | Upside | Downside |
| --- | --- | --- |
| **Gifted** | +1 to every attribute | −10% skill points per level |
| **Fast Shot** | ranged attacks 20% faster | no aimed/targeted shots |
| **Small Frame** | +1 Agility | −25% carrying capacity |
| **Genius** | +1 tag skill | −15 starting skill points |
| **Undead Phobia** | +2 WIL + 25% disease/poison resistance (via innate `resist` effects) | −20% attack & damage vs undead |
| **Beast Phobia** | +1 STR + 10 starting ranks in Animal Handling | Major combat & survival penalties vs beasts |
| **Leather Allergy** | +1 END + 15% defense when wearing metal armor | Cannot equip leather/hide armor (or severe penalties) |
| **Iron Allergy** | +10% magic effect magnitude + strong exotic material affinity | −25% effectiveness with iron/steel weapons & armor |
| **Silver Sensitivity** | +30% resistance to undead/unholy + chance to absorb hostile magic (via `resist` + `spell_absorption`) | Extra damage from silver weapons; cannot benefit from silver gear |
| **Arcane Phobia** | +20% resistance to hostile magic + improved spell absorption (via `resist` + `spell_absorption`) | −15% effectiveness vs constructs and magically-animated beings |

Take **0–2** (suggested cap: 2). Their attribute bonuses apply now, alongside the
racial modifiers, and may exceed the normal caps. (*Genius* adds to your tag-skill
count in step 5 and reduces the pool in step 6.)

Some of these traits include hard restrictions (e.g. "cannot equip leather"). The engine or GM is expected to enforce these. They are balanced by strong, synergistic upsides that reward playing into the drawback (e.g. Leather Allergy pushes you toward metal armor playstyles).

The mechanical effects for these traits (resistances, phobias, allergies, magic absorption, etc.) are now defined as first-class entries in the central effect registry (`data/effects/core.json`) using the new `innate` channel. The creation trait entries reference them by ID in their effect `target` fields for better data consistency and potential reuse elsewhere. See `features/core.json` and the new innate effects for details.

---

## 5. Pick 3 tag skills

Choose **3 tag skills** from [`skills/core.json`](../../ruleset/data/skills/core.json)
(any skill with `"taggable": true` — currently all of them). Humans pick a **4th**
(Versatility).

A tag skill costs **half as many** skill points for the same gain — normal skills
are bought at a 2:1 ratio (2 points for +1 rating), while tagged skills are bought at
a 1:1 ratio (1 point for +1 rating). Tagging is therefore about long-term efficiency
in your focus skills (an ongoing discount, *not* a flat starting bonus).

The skills span seven categories — combat (`blades`, `blunt`, `piercing`, `unarmed`, `block`, `bows`, `crossbows`, `guns`, `throwing_weapons`), the nine `<color>_magic` schools, stealth, social, knowledge, survival, and utility (`alchemy`, `repair`, `athletics`, plus crafting disciplines).

---

## 6. Starting skills & the level-1 pool

**Base rating** for every skill (before spending points):

```
start = 5 + associatedAbility × 2
```

Luck is **not** baked in here — it is applied at check time as `+floor(LCK/2)`.

Then spend your **1st-level skill-point pool**:

```
points = 5 + INT × 2 + random(0 … LCK)         // Gifted: ×0.9
spend:  normal skill 2:1 (2 points per +1) · tagged skill 1:1 (1 point per +1)
```

Skills cap at **100**. (Whether a fresh character receives the level-1 pool or starts
at bare base values is a per-game call; the reference assumes you get and spend it.)

---

## 7. Starting perks

Perks (`features` with `type: "perk"`) are the character-defining progression, gated
by `prerequisite` (`level` / `abilities` / `skills` / `perks`). Every normal level grants
**one perk**. A 1st-level character normally has **none** — their skill prerequisites
(e.g. *Pyromancer* needs `red_magic ≥ 50`) are out of reach at creation anyway.

Exception: **Humans** *may* take **one perk at creation** (their *Ambition* racial feature) —
it is an option, not a requirement. Any perk whose prerequisites they already meet
qualifies; at level 1 that means an ability-only perk such as *Toughness* (`end ≥ 5`)
or *Alert* (`per ≥ 6`). A human who has no perk worth taking yet may decline and save
it for later.

Note: creation perks (step 3) and level-up perks come from the **same table** —
they differ only in *when* you may take them (`type: "creation"` is creation-only and
not offered on level-up).

---

## 8. Derive statistics

From your final attributes (`progression.md` § Derived statistics), at `level = 1`:

| Stat | Formula | At level 1 |
| --- | --- | --- |
| Hit points | `15 + END × 8 + level × 4` | `19 + END × 8` |
| Mana | `INT × 8 + level × 2` (regen via WIL) | `2 + INT × 8` |
| Stamina *(action combat)* | `15 + END × 5 + level × 2` (regen via WIL) | `17 + END × 5` |
| Action Points *(turn-based)* | Max 10 pool; carries over between turns but cannot exceed 10. Starting pool is typically 2 + floor(AGI / 3) or set by the GM at the start of play. | — |

**Resistance, Disease & Poison (new full systems — see `attributes.md` for complete rules):**

- **Base Disease/Poison Resistance**: `END × 5 + bonuses` (as % reduction or d100 roll bonus). Add flat bonuses from race traits, backgrounds, and creation perks (implemented as innate `resist` effects with parameter "poison" or "disease" and value as %).
  - Example: Dwarf (END 8 base + racial) starts with ~40 + 25% = 65% base disease/poison resistance.
  - Magic/Status Resistance: `WIL × 4 + bonuses` (for magic, paralyze, fear, etc.). Per-type via `resist` effects (parameter "magic", "undead", phases, etc.).
- **Disease System**: Exposure (failed resistance roll vs disease effect) leads to incubation (duration), symptoms (applied as damage/drain/control effects), and possible progression/contagion. Periodic resistance rolls (e.g. daily) to fight it off. Cure via `cure` effect (parameter "disease" or specific like "blight").
  - Samples in effects: Blight (wood-phase vitality drain), Fever (stamina/END drain), Plague (severe contagious damage).
  - At creation: Calculate base from END + bonuses. Some creation perks (e.g. Undead Phobia, Silver Sensitivity) or racial traits grant starting % resistance or vulnerabilities.
- **Poison System**: Delivered via "poison" channel (coatings, consumables, monster attacks). Resistance roll on application to negate/reduce. Poisons are effects tagged ["poison"] (or the generic `poison` effect id 46) that can generically apply other effects via their `sub_effects` array (the symptoms: damage, drain, control like paralyze). This supports recursive application of non-poison effects as payload. However, sub_effects of a poison must not include any effect with "poison" tag (to prevent recursion). 
  - Samples: poison_paralytic (sub_effects: [11] paralyze), poison_weakness (sub_effects: [6] damage_stamina), poison_health_drain (sub_effects: [5] damage_health). The generic poison (id 46) provides the mechanism for custom/generic poisons.
  - At creation: Same base resistance calc as disease. Alchemy skill and poison kits from backgrounds enable brewing/preparation. Creation perks can provide resistance or poison-related benefits/drawbacks (granted as resist effects with parameter "poison").

These systems make END and certain creation choices (race, perks) matter for survival and exploration, not just combat. Full resolution (d100 under resistance vs potency/DC from effect magnitude) and Wuxing interactions are detailed in `attributes.md`. Add your final resistance values to the character sheet as part of derived stats. 

(The other 3 points in character creation — full derived stats formulas beyond the basics above, encumbrance rules, and light/vision systems — will be expanded after these three systems are complete, per your direction.)
| Carry weight | `25 + STR × 10` | — |
| Critical chance | `1% + LCK%` | — |
| Initiative | `PER` | — |

These are your **three resources** — HP, Mana, and a third that depends on combat
mode (Action Points in turn-based play, a Stamina pool in action combat). Luck also
adds **+floor(LCK/2)%** to every check at resolution time.

---

## 9. Spells & equipment

- **Spells.** If you tagged or raised a `<color>_magic` skill, you can learn spells of
  that color from [`spells/core.json`](../../ruleset/data/spells/core.json) (e.g.
  `fire_bolt`, `mend_wounds`). **All nine color schools are INT-seeded** — INT sets
  the mana pool and spell power, while WIL regenerates mana and resists magic. Casting
  uses d100 roll-under against the color skill; see [`magic.md`](./magic.md).
- **Equipment & wealth.** The reference baseline for a solid starting character is:
  - Starting wealth: **100 gp** (or roll 3d6×10 gp for variety).
  - Must acquire at least: one weapon (or rely on unarmed), basic armor if desired for the concept, and a simple kit (bedroll, 7 rations, waterskin, backpack, 50 ft rope, flint & steel, 10 torches or a lantern + oil).
  - Spend the rest on personal choices (extra weapons, tools, consumables, trade goods, or saved for later).
  This 100 gp + kit package is the recommended reference starting loadout. Full item lists and costs live in the equipment collections. GMs and consuming games may raise or lower the budget for desired power level or campaign tone.

---

## 10. Record your element

Your race's **`phase`** (`wood` / `fire` / `earth` / `metal` / `water`) is your
character's **foundation element**. It feeds the five-phase interaction cycles in
[`data/wuxing/core.json`](../../ruleset/data/wuxing/core.json): incoming phased
effects resolve against it via Generating / Overcoming / Weakening / Insulting (see
[`magic.md`](./magic.md) § Elemental interaction). It also determines phase synergy
when you later offer cores or elixirs during breakthroughs (see `advancement.md`).

Treat it as an affinity, not a flat resistance — the cycles compute each matchup.

---

## Creation choices that shape long-term advancement

Your starting build directly influences how strong and reliable your **breakthroughs** (see `advancement.md`) will be across 30 levels:

- **Race phase (foundation)**: Determines which cores/elixirs give you phase synergy bonuses (+15% generating) or penalties. A metal-phase dwarf naturally favors metal and earth essence for tanky resource or attribute breakthroughs.
- **INT**: Raises every future skill-point pool (`5 + INT × 2 + random(0…LCK)`) and seeds all magic. High INT at creation is the strongest lever on total ranks earned the "normal" way.
- **LCK**: Adds variance + average bonus to every skill-point pool and is the single biggest controllable factor in breakthrough success chance (`+ LCK × 2.5` in the reference formula). It also improves core drop rates while hunting.
- **Tag skills (the permanent 1:1)**: The three (or four for humans) skills you tag at creation stay at 1:1 efficiency forever. Choose them for the skills you actually want to push hard via both normal points and breakthrough skill bonuses (0–10 ranges, easier success weighting).
- **Creation perks**: Powerful but permanent tradeoffs. *Gifted* gives early power at the cost of −10% points every level. *Genius* gives an extra tag but starves your starting pool. Many players take zero or one.
- **Attribute spread**: While attributes rarely rise after creation (only high-grade breakthroughs can grant +1, and only once per level), a high END or INT at creation pays dividends in HP/mana growth and point income for decades.

A "deliberate cultivator" often prioritizes LCK and a clear phase identity at creation, accepts a slightly slower normal progression, and then waits for good cores rather than rushing every level.

## Worked example — "Borin", a dwarven shield-bearer (level 1)

1. **Race** — `common_dwarf` (baseline subrace, inherits the dwarf parent): phase
   **metal**, speed 25, medium; modifiers `{ end: +2, str: +1, agi: −1 }`; traits
   *Darkvision*, *Dwarven Resilience*, *Stoneworker*.
2. **Attributes** — start all at 4; distribute +10 as
   str +2, end +2, per +1, int +1, agi +1, lck +3; then apply racial mods. Final:
   **STR 7 · PER 5 · END 8 · INT 5 · WIL 4 · AGI 4 · CHA 4 · LCK 7**.
3. **Background** — `former_soldier`. Adds Block +12, Blades +10; starts with chain shirt, shield, longsword, mace, and +25 gp. This immediately makes Borin feel like a trained shield-bearer.
4. **Creation perks** — none.
5. **Tag skills** — `blunt`, `block`, `athletics`.
5. **Skills** — bases `5 + abil×2`: blunt 19, block 21, athletics 19. Apply background: block 21+12=**33**, blades (from background) base 19 +10 = **29**. Level-1 pool
   `5 + 5×2 + random(0…7)` → say a roll of 4 = **19 points**, all on tagged skills
   (1:1): blunt +8 → **27**, block +6 → **39**, athletics +5 → **24**.
6. **Perks** — none (not human).
7. **Derived** — HP `15 + 8×8 + 4` = **83**; Mana `5×8 + 2` = **42**; third resource:
   Stamina `15 + 8×5 + 2` = **57** (action combat) *or* Action Points pool (max 10, carries over, typical start 2 + floor(AGI/3) or GM-set at start of play); Carry `25 + 7×10` = **95**; Crit `1 + 7` = **8%**;
   Initiative **5**.
8. **Spells / gear** — no color skill tagged → no spells. Background already gave chain shirt + shield + longsword + mace. Starting wealth 100 + 25 = 125 gp. Adds basic kit (~20 gp) and a few extra rations or a backup dagger. Plenty left for sundries.
9. **Element** — **metal**: strong against Wood-phase foes (Metal overcomes Wood),
   soft against Fire (Fire melts Metal). **Advancement outlook**: High LCK (7) + solid END (8) + metal phase makes Borin excellent at farming metal/earth cores later. He will be very strong at resource (HP/Stamina) and attribute breakthroughs once he starts finding greater+ cores. His three tagged combat skills will benefit from both 1:1 point spending and the easiest success weighting on skill-target breakthroughs. The former_soldier background already gives him a strong combat foundation.

## Another quick example — "Lira", a human scout (level 1)

1. **Race** — human: phase **any** (player choice; here water), speed 30, medium; no ability modifiers; traits *Versatility* (+1 tag skill) and *Ambition* (option for +1 creation perk).
2. **Attributes** — start all at 4; distribute +10 as per +3, agi +3, lck +2, int +1, end +1. Final: **STR 4 · PER 7 · END 5 · INT 5 · WIL 4 · AGI 7 · CHA 4 · LCK 6**.
3. **Background** — `scout`. Adds Survival +12, Stealth +10, Bows +8; starts with shortbow, leather armor, dagger, and some rations. Perfect thematic fit for a human scout.
4. **Creation perks** — via *Ambition*, takes *Fast Shot* (ranged 20% faster, but no aimed shots). (Chosen in step 4; recorded here for the sheet.)
5. **Tag skills** — `bows`, `stealth`, `survival`, `investigation` (the extra from Versatility).
6. **Skills** — bases: bows 19, stealth 19, survival 19, investigation 15. Apply background: bows 19+8=**27**, stealth 19+10=**29**, survival 19+12=**31**. Level-1 pool `5 + 5×2 + random(0…6)` → roll 3 = **18 points**.
   Spend on tagged (1:1): bows +3 → 30, stealth +3 → 32, survival +3 → 34, investigation +3 → 18. (She already has a very strong wilderness base from the background.)
7. **Perks** — *Fast Shot* (the creation perk taken via human Ambition). No normal level-up perk yet.
8. **Derived** — HP 15+5×8+4=59; Mana 5×8+2=42; Action Points pool (max 10, carry, typical start ~4); Carry 25+4×10=65; Crit 1+6=7%; Initiative 7.
9. **Spells / gear** — no magic tagged. Background already supplied shortbow + leather armor + dagger. Starting wealth 100 gp. Adds basic kit, 20–30 arrows, thieves' tools or a few potions, and keeps the rest for roleplay or future breakthroughs.
10. **Element** — **water** (player choice): good synergy with flow/adaptability themes. **Advancement outlook**: LCK 6 + chosen water phase positions Lira well for mana/recovery or magic-school breakthroughs later (water favors WIL/INT/Mana and magic skills). Her four tag skills (thanks to Versatility) give her unusually broad 1:1 efficiency. The *Fast Shot* tradeoff is permanent, so she will lean into burst ranged play rather than precision sniping. The scout background gives her an excellent head start on the exact skills (now including the split bows skill) she will want to push further with both normal points and future breakthroughs.

These examples show the 2:1 vs 1:1 spending, tag skill efficiency, creation perk option for humans (including ability-gated perks via Ambition), the recommended 100 gp + kit baseline, and how starting choices (especially backgrounds) set up both immediate competence and the long-term breakthrough system. You can (and should) play with the attribute spreads, background, creation perks, and gear choices to fit the concept and your intended cultivation path.

---

## Data & schema map

| Step | Data | Schema |
| --- | --- | --- |
| Race | `data/races/core.json` | `race.schema.json` |
| Attributes | `data/abilities/core.json` | `ability.schema.json` |
| Background | `data/backgrounds/core.json` (new) | `background.schema.json` (new) |
| Skills / tags | `data/skills/core.json` | `skill.schema.json` |
| Perks (level-up / creation / racial) | `data/features/core.json` | `feature.schema.json` |
| Spells | `data/spells/core.json` | `spell.schema.json` |
| Equipment | `data/equipment/*.json` | `equipment.schema.json` |
| Element cycles & breakthrough affinities | `data/wuxing/core.json` | `wuxing.schema.json` |

## Open items

- **Level-1 skill pool.** Whether creation grants the first level's skill points (as
  assumed here) or starts at bare base ratings is a tunable per-game choice.
- **Starting loadout variety.** The 100 gp + basic kit is the recommended reference
  baseline. Backgrounds now supply many concrete starting packages; more "kit" style
  multi-item equipment entries (thieves' tools, healer's satchel, etc.) would still be
  valuable.
- **Point-buy spread.** The "start at 4, distribute 10, max 10" spread is the
  reference; alternatives (rolled, fixed arrays) are easy variants.
- **More creation support data.** Additional backgrounds, more creation perks (with
  interesting tradeoffs), and more low-level / ability-only perks reachable by
  ambitious humans at level 1 would make creation richer. More starter-viable spells
  for the nine colors would also help new magic users.
