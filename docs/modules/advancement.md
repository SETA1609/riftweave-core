# Character Advancement & Breakthroughs

**Status:** In design (core level-up formulas exist in `progression.md`; item-boosted cultivation-style breakthroughs are new).

This document describes how characters grow in power over time, with a deliberate Wuxia-inspired twist: while base progression uses earned experience, a cultivator can **delay and empower** a level-up ("breakthrough") by offering high-quality monster cores or alchemically refined elixirs. The five-phase (Wuxing) system and `qualityGrade` ladder become central to long-term power.

## Core Philosophy

Riftweave is classless and point-buy at heart, but advancement should feel **weighty and strategic**. Players are encouraged to:

- Farm or hunt monsters whose phase matches a desired growth direction.
- Use Alchemy to refine raw cores into superior elixirs (concentrating essence, blending compatible phases, or mitigating clashes).
- **Wait** before spending experience if better materials are within reach — a classic cultivation trope.

Base gains on level-up remain reliable. The breakthrough system provides **meaningful extra power** (attributes, resources, skill points, or foundation) at the cost of preparation and risk/reward decisions.

## Base Level-Up (Always Available)

See `docs/modules/progression.md` and `docs/modules/attributes.md` for the reference formulas. In summary:

- Experience is earned from quests, kills, exploration, and major events (event-driven; skills do not train by use).
- On reaching the required experience for the next level, the character gains:
  - A pool of **skill points**: `5 + INT × 2 + random(0…LCK)`.
  - Tag skills (up to 3 chosen at creation) receive +2 per point spent instead of +1.
  - Resource maximums scale with level via the standard formulas (HP from END, Mana from INT, Stamina from END).
  - Periodic access to **perks** from `features/core.json` (suggested: one every 3 levels).
- Level thresholds themselves are not yet rigidly tabulated in the ruleset (see open items below); consuming games may use fixed tables, milestone awards, or any curve.

These gains happen even if no special items are offered.

## Breakthroughs — Item-Boosted Advancement

At the moment a character is eligible to level (or as a distinct "cultivation / refinement / breakthrough" downtime activity tied to a level-up), they may **offer** one or more **essence items**. These items are consumed and convert into permanent growth.

### Essence Item Types

1. **Monster Cores** (raw essence)
   - Harvested from defeated monsters (or rarely found as loot/treasure).
   - Inherit the monster's `phase` (see `data/monsters/core.json` — every monster entry already carries a phase).
   - Carry a `qualityGrade` (petty through legendary) based on the monster's power, rarity, and condition of harvest.
   - Example future data shape (as special ingredients):
     ```json
     {
       "id": "ember_drake_core",
       "name": "Ember Drake Core",
       "description": "A fist-sized crystal of condensed flame essence pulsing with inner heat.",
       "phase": "fire",
       "qualityGrade": "greater",
       "effects": ["essence_fire"],   // special channel or tag for advancement use
       "tags": ["core", "monster", "fire"]
     }
     ```

2. **Refined Elixirs / Essence Pills / Spirit Elixirs** (processed)
   - Created through Alchemy by refining one or more cores, often with stabilizing or amplifying ingredients.
   - Can concentrate a single phase, blend compatible phases (via generating cycle), or create "balanced" multi-phase elixirs.
   - Higher Alchemy skill + better stations produce higher-grade results or purer essence.
   - These are typically modeled as special consumables (new `delivery` value such as `"refine"` or `"breakthrough"`) or as a distinct class of advancement reagent that references the shared effect pool with a new channel.

### How a Breakthrough Works (Reference Procedure)

1. Character has accumulated enough experience for a new level (or chooses to attempt a breakthrough at a narrative milestone).
2. Player decides whether to perform a **standard level-up** or a **breakthrough**.
3. For a breakthrough, the player selects 1–N essence items to offer (cores and/or elixirs). These are removed from inventory.
4. The system resolves the offered essence:
   - Base essence value is derived from each item's `qualityGrade` (higher grades give exponentially or tiered larger bonuses).
   - Each item's `phase` maps to favored growth vectors (see table below).
   - Wuxing cycle interactions from `data/wuxing/core.json` are applied between the offered items and (optionally) the character's current "foundation phase" or dominant attributes/race phase.
     - Generating (×1.5): harmonious growth.
     - Overcoming / weakening: reduced efficiency.
     - Insulting (×1.25 backlash when overpowering): possible waste, instability, or temporary drawbacks unless mitigated by Alchemy or perks.
   - Final bonuses are granted as permanent increases on top of the normal level-up package.
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

A single core or elixir can contribute to **multiple** categories at reduced strength, or heavily to one primary vector.

### Grade Scaling

`qualityGrade` (from `schema.json#/definitions/qualityGrade`) is the primary lever:

- petty / minor: small foundation boost (good for early levels or "filler" breakthroughs).
- lesser / common: solid, reliable gains.
- major / greater: significant — the kind of treasure worth hunting or crafting toward.
- grand / legendary: realm-defining. May grant rare secondary effects (e.g. +1 to a hard cap, new perk-like option, or permanent minor passive).

Exact numeric tables are left to the consuming game, but the ruleset should eventually provide reference values or formulas (e.g. "a greater Fire core grants the equivalent of +2–3 effective levels of HP growth plus a Fire-aspected combat edge").

### Wuxing Interaction During Breakthrough

The same cycles defined in `data/wuxing/core.json` that govern spell/poison interactions and material–effect resonance now also govern essence refinement and breakthrough potency:

- Using a generating-phase elixir with a core (or with the cultivator's current foundation) **amplifies** the granted bonuses.
- Clashing phases may require an Alchemy "harmonizing" step or a high-skill check, or they simply deliver reduced value.
- Severe insulting-cycle use without mitigation could produce **impure foundation** (smaller gains + a temporary or permanent minor flaw until corrected by later, purer breakthroughs).

This makes the Wuxing data a first-class part of long-term character building, not just combat.

## Data & Schema Implications

- **Ingredients**: Extend `ingredient.schema.json` to optionally carry `phase` and `qualityGrade` for cores and other high-essence natural treasures. Add a conventional `tags` value or new field to mark "advancement" / "core" / "essence" use. Existing mundane ingredients remain unchanged.
- **Effects**: New effects (or a new `category` / channel) for permanent growth / essence. Examples to consider:
  - `essence_wood`, `essence_fire`, etc. (abstract units that the breakthrough procedure converts).
  - Or direct growth effects: `permanent_increase_hp`, `permanent_fortify_str`, `foundation_skill_points`, etc. (these would be special-cased as "only usable in advancement context").
- **Consumables / Equipment**: Add a delivery option (e.g. `"refine"`, `"breakthrough"`, `"cultivate"`) for refined elixirs that are not meant to be drunk for temporary effects. These may carry `appliedEffect` entries with a special "advancement" interpretation.
- **Monsters**: Already carry `phase`. Future work can add explicit `core` drop data or generation rules (grade derived from monster level + rarity tags).
- **Features / Perks**: Perks can gate, enhance, or mitigate breakthrough results (e.g. "Stable Foundation", "Heavenly Refiner", "Phase Harmony").
- **Alchemy module**: Gains a distinct "refining / pill crafting" sub-system on top of normal potion brewing. Refining recipes consume cores + catalysts and output higher-grade or multi-phase elixirs. Quality formula (still TBD in alchemy.md) becomes especially important here.

Until a full module-loading system exists, these additions live in the core collections (`ingredients`, `effects`, `equipment/consumables`) with clear tagging and documentation.

## Example Breakthrough (Narrative + Mechanical Sketch)

A level 7 cultivator with a strong Wood foundation has accumulated enough XP for level 8. They have been holding off.

They offer:
- 1× `greater_fire_core` (from a powerful ember drake)
- 1× `refined_wood_spirit_elixir` (alchemically purified from several wood-phase ingredients + a minor core; grade: major)

Resolution (illustrative):
- The Fire core clashes with their Wood foundation (overcoming cycle → suppressed). Base value reduced.
- The refined Wood elixir is same-phase and high grade → strong generating synergy with the cultivator.
- Net result: solid HP and END growth (Wood), plus a smaller but still valuable Fire-aspected combat or crit bonus (the clash was not total waste because of the elixir's purity).
- They gain the normal level-8 skill points + resources, **plus** the breakthrough extras.
- Had they waited for a legendary Wood-aligned treasure or a harmonizing pill, the gains would have been substantially larger.

Players who rush every level with whatever petty cores they have on hand will have a noticeably weaker long-term foundation than those who cultivate deliberately.

## Open Questions & Future Work

- Exact numeric formulas or lookup tables for grade → bonus (and how they compound with normal level scaling).
- Whether attribute increases are allowed at all (the 1–10 scale is tight; many breakthroughs may focus on resources, skill points, and "foundation" multipliers instead of raw attributes).
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

- Flesh out concrete reference tables or formulas once stakeholder feedback is gathered.
- Extend `ingredient.schema.json` + add example cores to `ingredients/core.json`.
- Add growth/essence effects and a refinement delivery type.
- Write example refining recipes (once a recipe shape exists or as prose in alchemy).
- Update the validator / cross-reference checks if new channels or special advancement effects are introduced.
- Provide sample "before and after" character sheets showing the power delta of a good vs. rushed breakthrough.

Contributions to this design (especially numeric tuning, phase affinity suggestions, and risk/reward ideas) are welcome via updates to this document.

---

*This system keeps the reliable, data-driven point-buy heart of Riftweave while adding flavorful, Wuxing-driven strategic depth to long-term progression.*