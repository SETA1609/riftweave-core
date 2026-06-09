# Character Advancement & Breakthroughs

**Status:** In design (core level-up formulas exist in `progression.md`; item-boosted cultivation-style breakthroughs are new).

This document describes how characters grow in power over time, with a deliberate Wuxia-inspired twist: while base progression uses earned experience, a cultivator can **delay and empower** a level-up ("breakthrough") by offering high-quality monster cores or alchemically refined elixirs. The five-phase (Wuxing) system and `qualityGrade` ladder become central to long-term power.

## Core Philosophy

Riftweave is classless and point-buy at heart, but advancement should feel **weighty and strategic**. Players are encouraged to:

- Hunt or clear monsters (especially via quests) whose phase matches a desired growth direction in order to obtain high-quality cores.
- Use Alchemy to refine raw cores into superior elixirs (concentrating essence, blending compatible phases, or mitigating clashes).
- **Wait** before spending experience if better materials are within reach — a classic cultivation trope.

Base gains on level-up remain reliable. The breakthrough system provides **meaningful extra power** (attributes, resources, skill points, or foundation) at the cost of preparation and risk/reward decisions.

## Base Level-Up (Always Available)

See `docs/modules/progression.md` and `docs/modules/attributes.md` for the reference formulas. In summary:

- Experience is earned almost exclusively from event-driven sources (quests, radiant missions, story milestones, exploration, etc.). Direct experience from killing monsters is extremely limited — only the very first monster a character ever kills grants any direct XP (see `progression.md` § Experience for the full rule and `progression.md` § Experience Progression and Thresholds for the exact formula and event award guidance). Skills do not train by use.
- On reaching the required experience for the next level, the character follows the
  full level-up procedure (see `progression.md` § Full Level-Up Procedure). This always
  includes the base package (one perk, skill points at the correct ratios, resource
  growth, and TTRPG AP rules). A breakthrough can be performed as part of this process
  (or as tied downtime) to gain additional permanent bonuses from offered cores/elixirs.
- Level thresholds themselves are not yet rigidly tabulated in the ruleset (see open items below); consuming games may use fixed tables, milestone awards, or any curve.

These gains happen even if no special items are offered.

## Breakthroughs — Item-Boosted Advancement

At the moment a character is eligible to level (or as a distinct "cultivation / refinement / breakthrough" downtime activity tied to a level-up), they may **offer** one or more **essence items**. These items are consumed and convert into permanent growth.

### Essence Item Types

1. **Monster Cores** (raw essence)
   - Only certain monsters can drop cores. Mundane or spiritually empty creatures such as bandits and zombies never drop cores. Only monsters with meaningful elemental or spiritual essence are eligible.
   - Eligible monsters have a base percentage drop chance.
   - The final drop chance is improved by two factors:
     - The monster’s **level** (higher-level monsters are significantly more likely to yield a core).
     - The killing character’s **Luck (LCK)** attribute.
   - **Proposed reference formula** (engines may tune the constants):

     ```
     finalChance = clamp(
         baseChance × (1 + (mobLevel − playerLevel) × 0.05 + LCK × 0.06),
         0.02, 0.75
     )
     ```

     - `baseChance` comes from the monster’s `coreDrop.baseChance` (0.0–1.0).
     - `(mobLevel − playerLevel)` gives a bonus for hunting stronger prey (negative if the monster is much weaker).
     - Luck provides a steady improvement to essence farming.
     - The result is clamped so eligible monsters always have at least a small chance, but never a guaranteed drop.

   - Core acquisition uses **two independent rolls**:

     1. **Drop Roll** — Use the formula above to determine whether a core drops at all.
     2. **Grade Roll** — If the drop succeeds, roll for the `qualityGrade` within the monster’s `minGrade`–`maxGrade` range. Lower grades always have significantly higher probability than higher grades, even for powerful monsters.

   - **Grade Roll (proposed reference method)**

     Given the ordered list of possible grades from `minGrade` to `maxGrade`:

     - Assign decreasing weights favoring the bottom of the range. Example weights for a 4-grade range (lesser → major):

       | Grade   | Weight |
       |---------|--------|
       | lesser  | 8      |
       | common  | 4      |
       | major   | 2      |
       | greater | 1      |

     - Pick a grade randomly according to these weights (or generalize with a power curve such as `index = floor(random() ^ 1.7 * numGrades)` for a smooth bias toward lower grades).

     - Result: Even a grand-tier monster will most often drop lesser or common cores, with grand/legendary cores being rare jackpot results.

   - **Worked Examples**

     **Example 1: Early-game monster**
     - Monster: level 5, `baseChance = 0.08`, minGrade=lesser, maxGrade=common
     - Player: level 4, LCK = 3
     - Drop roll: `0.08 × (1 + (5-4)×0.05 + 3×0.06) = 0.08 × 1.23 = 0.0984` (≈ 9.8% chance)
     - If it drops, grade roll (using the weight table above for lesser→common): ~67% lesser, ~33% common.

     **Example 2: Hunting above your level**
     - Monster: level 12, `baseChance = 0.12`, minGrade=common, maxGrade=greater
     - Player: level 8, LCK = 5
     - Drop roll: `0.12 × (1 + (12-8)×0.05 + 5×0.06) = 0.12 × 1.50 = 0.18` (18% chance)
     - Grade roll (common→greater weights 8/4/2/1): heavily biased toward common, with greater being rare.

     **Example 3: Lucky high-level hunter**
     - Monster: level 18, `baseChance = 0.15`, minGrade=major, maxGrade=grand
     - Player: level 15, LCK = 9
     - Drop roll: `0.15 × (1 + (18-15)×0.05 + 9×0.06) = 0.15 × 1.69 = 0.2535` (≈25.4% chance)
     - Even with good Luck and level advantage, the grade roll will still favor major/common over grand because of the lower-tier bias.

   - On a successful drop + grade roll, award the matching spirit core ingredient (e.g. `common_earthen_spirit_core` or `greater_blazing_spirit_core`). The awarded core uses the monster’s `phase`.
   - Spirit cores are represented as special ingredients (see `data/ingredients/core.json`). Examples of the expected naming pattern include `greater_blazing_spirit_core`, `major_verdant_spirit_core`, etc.
   - Cores can also rarely appear as loot or treasure.

2. **Refined Elixirs / Essence Pills / Spirit Elixirs** (processed)
   - Created through Alchemy by refining one or more cores, often with stabilizing or amplifying ingredients.
   - Can concentrate a single phase, blend compatible phases (via generating cycle), or create "balanced" multi-phase elixirs.
   - Higher Alchemy skill + better stations produce higher-grade results or purer essence.
   - These are typically modeled as special consumables (new `delivery` value such as `"refine"` or `"breakthrough"`) or as a distinct class of advancement reagent that references the shared effect pool with a new channel.

### How a Breakthrough Works (Reference Procedure)

1. Character has accumulated enough experience for a new level (or chooses to attempt a breakthrough at a narrative milestone).
2. Player decides whether to perform a **standard level-up** or a **breakthrough**.
3. For a breakthrough, the player may select up to **3** essence items (cores and/or elixirs) to offer. These are removed from inventory. (Maximum of 3 per breakthrough.)
4. The system resolves the offered essence (one resolution per item, up to the maximum of 3):
   - Each item is governed by its `phase`. The phase determines a set of possible targets the player can choose from (see Phase → Growth Affinity table below):
     - Attributes (e.g. STR, END, WIL, etc.) — **only selectable if the item's qualityGrade is greater than common**. Attributes can only increase by +1 total from any number of breakthroughs at this level.
     - Resource pools (HP growth, Mana, Stamina)
     - Skill areas or specific skills (e.g. combat skills, magic schools, survival, etc.) — bonus amounts fall in a **0–10** range overall. Lower grades have a lower maximum; higher grades raise both the floor and ceiling (see Grade Scaling section for example ranges per grade). Higher grades also improve success chance on the luck roll.
   - The player **selects one target** from the governed options for that specific item.
   - Breakthrough success is **luck-based** (not guaranteed):
     - Calculate a success chance for this item using this proposed reference formula:
       ```
       success = base_grade + phase_synergy + (LCK × 2.5) + target_weight
       final_success_chance = clamp(success, 5, 95)
       ```
       Where:
       - `base_grade` is the core chance from the item's `qualityGrade` (example values: petty=15%, minor=25%, lesser=35%, common=45%, major=55%, greater=65%, grand=75%, legendary=85%).
       - `phase_synergy` is a Wuxing modifier between the item's phase and the character's foundation or chosen target (+15% generating, 0 neutral, -10% overcoming/weakening).
       - `LCK × 2.5` is the direct Luck contribution.
       - `target_weight` favors easier targets: +10 for skills, 0 for resources, -15 for attributes.
     - Roll for success (d100). Even a refined legendary essence can fail. A petty core has a much lower (but real) chance.
   - On success:
     - **Attributes**: +1 (only if grade > common; max +1 total per level from breakthroughs).
     - **Skills**: bonus in the **0–5** range. Grade controls floor and ceiling (lower grades have lower max e.g. petty 0–1; higher grades raise both e.g. greater 4–5, grand/legendary 4–5). See Grade Scaling section for full per-grade examples.
     - **Resource pools** (HP, Mana, or Stamina): the normal level-up growth for that resource is rolled first. Then a second "growth of a level" roll is performed for the breakthrough (using the same growth formula as a normal level for that resource). The higher of the two rolls is used for the final growth applied at this level-up.
     - Wuxing cycle interactions can further amplify or reduce the final bonus value.
   - On failure: no bonus is granted from this item.
   - Final bonuses from all successful items are granted as permanent increases on top of the normal level-up package. Wuxing interactions from `data/wuxing/core.json` are applied across the offered items and (optionally) the character's current "foundation phase" or dominant attributes/race phase (Generating ×1.5, Overcoming/Weakening reduced, Insulting backlash possible).
5. The level (or "realm") is gained. Future levels continue to use the improved foundation.

### Phase → Growth Affinity (Suggested Mapping)

These are **design reference** values. Engines may tune numbers; the important part is that phase gives identity and strategic choice.

| Phase  | Favored Attributes | Favored Resources       | Favored Skill / Playstyle Areas          | Flavor |
|--------|--------------------|-------------------------|------------------------------------------|--------|
| Wood   | END, WIL, (INT)   | HP growth, healing      | Survival, nature, poison resistance, regeneration | Vitality, growth, endurance |
| Fire   | STR, AGI, (PER)   | Attack power, crits     | Combat skills, fire/heat magic, burst    | Power, passion, transformation |
| Earth  | END, STR, (CHA)   | Stamina, carry, defense | Heavy armor, fortitude, crafting, stability | Solidity, protection, rooted power |
| Metal  | STR, AGI, (LCK)   | Precision, penetration  | Blades, finesse, anti-magic or cutting effects | Sharpness, discipline, conduction |
| Water  | WIL, INT, (PER)   | Mana, recovery, flow    | Magic schools, adaptability, stealth, movement | Fluidity, wisdom, evasion |

A single core or elixir is tied to one chosen target from its phase's governed options. Up to 3 items may be offered per breakthrough. Each has its own independent luck-based success roll (base from grade + phase synergy + LCK, plus target weighting: skills easiest, resources neutral, attributes hardest). On success: attributes +1 (high-grade only), skills 0–10 range (grade sets floor/ceiling), resources = higher of two level-growth rolls. Wuxing can still modify the final bonus. Failures possible for any item.

### Grade Scaling, Target Selection, and Luck-Based Success

`qualityGrade` (from `schema.json#/definitions/qualityGrade`) primarily determines both the success chance *and* the range/strength of the bonus to the player-selected target (see the resolution rules above). Higher grades give better base success odds and better bonus ranges.

**Proposed reference success formula** (for each item):
```
success = base_grade + phase_synergy + (LCK × 2.5) + target_weight
final_success_chance = clamp(success, 5, 95)
```
- `base_grade`: petty=15%, minor=25%, lesser=35%, common=45%, major=55%, greater=65%, grand=75%, legendary=85%.
- `phase_synergy`: +15% (generating), 0 (neutral), -10% (overcoming/weakening) based on Wuxing between item phase and foundation/chosen target.
- `LCK × 2.5`: direct Luck contribution.
- `target_weight` (skills easiest → resources → attributes hardest): +10 for skills, 0 for resources, -15 for attributes (attributes also restricted to grades > common).

**Bonus ranges on success (skills now 0–10 overall):**
Grade controls both floor and ceiling for the 0–10 skill bonus (lower grades have a lower maximum; higher grades raise the minimum as well as the maximum). Higher grades also improve the base success chance on the luck roll. Example ranges:
- petty: 0–2
- minor: 1–3
- lesser: 2–4
- common: 3–5
- major: 4–6
- greater: 6–8
- grand: 7–9
- legendary: 8–10

**Resource pools**: second "level growth" roll; take the higher of normal level-up growth and breakthrough growth.

**Attributes**: +1 only (if grade > common; hard cap +1 total from breakthroughs this level).

Exact base success percentages and precise bonus values within each range are left to the consuming game, but the structure (luck-based per-item roll with target weighting, grade-controlled 0–10 skill ranges, second-roll max for resources) is the reference.

A single core or elixir is tied to one chosen target from its phase's governed options (with restrictions: attributes only for grades > common, and only +1 total per level from breakthroughs). Up to 3 items may be offered per breakthrough. Each has its own independent luck-based success roll (influenced by grade, phase synergy, and LCK). On success, Wuxing interactions can amplify or reduce the bonus. Failures are possible even with excellent items.

### Wuxing Interaction During Breakthrough

The same cycles defined in `data/wuxing/core.json` that govern spell/poison interactions and material–effect resonance now also govern essence refinement and breakthrough potency:

- Using a generating-phase elixir with a core (or with the cultivator's current foundation) **amplifies** the granted bonuses.
- Clashing phases may require an Alchemy "harmonizing" step or a high-skill check, or they simply deliver reduced value.
- Severe insulting-cycle use without mitigation could produce **impure foundation** (smaller gains + a temporary or permanent minor flaw until corrected by later, purer breakthroughs).

This makes the Wuxing data a first-class part of long-term character building, not just combat.

## Data & Schema Implications

- **Ingredients**: `ingredient.schema.json` has been extended to support optional `phase` and `qualityGrade`. Spirit cores are implemented as ingredients (see current examples in `data/ingredients/core.json` using the `greater_blazing_spirit_core` / `major_verdant_spirit_core` naming pattern).
- **Effects**: Marker effects (`essence_fire`, `essence_wood`, etc.) have been added to `data/effects/core.json` using the new `"essence"` channel. These identify cores for the breakthrough and refining systems.
- **Monsters**: The monster schema (`monster.schema.json`) has been extended with an optional `coreDrop` object (`baseChance`, `minGrade`, `maxGrade`). Only monsters that include this field can drop spirit cores. Mundane creatures (bandits, zombies, etc.) must omit the field entirely. The final drop chance is calculated at runtime using the reference formula above (`baseChance` modified by monster level and the killer’s Luck (LCK)). When a core drops, a matching spirit core ingredient (e.g. `greater_blazing_spirit_core`) is awarded using the monster’s phase and a grade within the declared range.
- **Features / Perks**: Perks can gate, enhance, or mitigate breakthrough results (e.g. "Stable Foundation", "Heavenly Refiner", "Phase Harmony").
- **Alchemy module**: Gains a distinct "refining / pill crafting" sub-system on top of normal potion brewing. Refining recipes consume cores + catalysts and output higher-grade or multi-phase elixirs. Quality formula (still TBD in alchemy.md) becomes especially important here.

Until a full module-loading system exists, these additions live in the core collections (`ingredients`, `effects`, `equipment/consumables`) with clear tagging and documentation.

## Example Breakthrough (Narrative + Mechanical Sketch)

A level 7 cultivator with a strong Wood foundation has accumulated enough XP for level 8. They have been holding off.

They offer (max 3 items allowed):
- 1× `greater_blazing_spirit_core` (Fire phase, greater grade) — player selects "combat skills" as the target.
- 1× `refined_wood_spirit_elixir` (Wood phase, major grade) — player selects "HP growth" as the target.

**Live example resolution** (player level 8, LCK 6, current foundation Wood):

For the greater Fire core (target: combat skills):
- Base success from grade (greater) = 65%
- Phase synergy (Fire vs Wood foundation) = -10% (overcoming)
- LCK bonus = +15% (6 × 2.5)
- Target type weighting (skill) = +10%
- Final success chance = 65 - 10 + 15 + 10 = **80%**
- Roll: 72 → Success!
- Skill bonus range for greater = 6–8. Player rolls a 7 → +7 to chosen combat skill.
- Wuxing: no further interaction this time.

For the major Wood elixir (target: HP growth):
- Base success from grade (major) = 55%
- Phase synergy (Wood vs Wood foundation) = +15% (generating)
- LCK bonus = +15%
- Target type weighting (resource) = 0%
- Final success chance = 55 + 15 + 15 = **85%**
- Roll: 91 → Failure. No extra HP growth from this item.

Normal level-up HP growth was rolled as + (END × 8 + 4) = +36.
No second growth roll occurred because the breakthrough item failed.

Net result: +4 to a combat skill from the Fire core. The Wood elixir failed its luck roll, so only the normal level growth applied. The character still gets their normal level-8 skill points, perk, and resource increases, plus the one successful breakthrough bonus.

Had both items succeeded, the player would have received the skill bonus *and* a second HP growth roll (taking the higher of the normal +36 and the new roll).

Players who rush every level with whatever petty cores they have on hand will have a noticeably weaker long-term foundation than those who cultivate deliberately.

## Open Questions & Future Work

- Exact numeric tables for success % (base + LCK + phase synergy) and bonus ranges per grade/target (the mechanism is defined; numbers are left for campaigns or future reference).
- Whether the +1 attribute cap per level (only for grades > common) feels right long-term, or if it needs further restrictions.
- How a character's "foundation phase" or dominant element is tracked (race phase? highest attribute phase? chosen at key breakthroughs? cumulative?).
- Risk / backlash mechanics for poor-phase or low-grade breakthroughs (purely smaller gains, or actual drawbacks?).
- Carry-over "unused essence" or partial foundation that improves future breakthroughs even if not spent immediately.
- Integration with the eventual module system (advancement as a core rule with optional "cultivation" depth module?).
- XP / level threshold tables (still missing from base progression).
- How many essence items can be offered per breakthrough, and whether there is a "purity" or "balance" cap.

## Related Documents & Data

- `docs/modules/progression.md` — base level-up, skill points, perks, derived stats.
- `docs/modules/attributes.md` — detailed attribute and resource formulas.
- `docs/modules/alchemy.md` — brewing model; will be extended for refining.
- `data/wuxing/core.json` — the interaction matrix used for essence synergy.
- `data/monsters/core.json` — source of raw phased cores.
- `schema.json#/definitions/qualityGrade` and `#/definitions/phase` — shared vocabulary.
- `issues.md` — high-priority gap: "Full character advancement / XP procedure".

## Next Steps (Non-Exhaustive)

- Flesh out concrete reference tables or formulas for grade + phase + Luck modifiers once stakeholder feedback is gathered.
- Add `coreDrop` data to appropriate monsters in `data/monsters/core.json` (only those with spiritual/elemental essence).
- Add growth/essence effects and a refinement delivery type (partially done for essence markers).
- Write example refining recipes (once a recipe shape exists or as prose in alchemy).
- Tune and finalize the runtime formula for core drop chance (a reference proposal has been added to this document; constants are open to adjustment).
- Update the validator / cross-reference checks if new channels or special advancement effects are introduced.
- Provide sample "before and after" character sheets showing the power delta of a good vs. rushed breakthrough.

Contributions to this design (especially numeric tuning, phase affinity suggestions, and risk/reward ideas) are welcome via updates to this document.

---

*This system keeps the reliable, data-driven point-buy heart of Riftweave while adding flavorful, Wuxing-driven strategic depth to long-term progression.*