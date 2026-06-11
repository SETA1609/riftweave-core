# Riftweave Issues & Missing Components

This file tracks gaps, stubs, planned features, and incomplete areas in the Riftweave ruleset.

Riftweave is a **classless, data-driven d100 roll-under cRPG ruleset**. It has a strong foundation in attributes, skills, perks, races (with lineage), a unique 9-color + 5-phase magic system, and early crafting primitives (materials, gems, tiers).

However, many areas are either:
- High-level only (formulas without full procedures)
- Design documents without implemented data
- Stubs or explicitly marked as "follow-up" / "planned"
- Missing entirely when compared to inspirations (D&D 5e/3.5, Fallout SPECIAL, Elder Scrolls)

This document is a living backlog. Items are grouped by category with current state, why it matters, and suggested next steps.

---

## 1. Core Mechanics (High Priority)

### Combat Resolution
**Current state:** Significantly advanced.  
`docs/modules/combat.md` is a comprehensive 590+ line document covering:
- Full action economy (actions, bonus actions, reactions, AP variant with cost table)
- Opportunity attacks (TTRPG + video game dual resolution)
- Defense model (physical + magic trees with 4/5 layers each)
- Critical hits with confirmation roll, Luck interaction table, and d100 effects table (11 entries)
- Cover, flanking, difficult terrain, prone
- Called / targeted shots (5 locations with penalties and effects)
- Initiative (TTRPG d20+PER, video game aggro/proximity)
- Mode comparison table (17 aspects)
- Encumbrance in combat (TTRPG + video game tables)
- 3 full worked examples (knight vs bandit, crit effects, party combat)
- Combat flow checklists (TTRPG, video game, cRPG)
- Edge cases (prone, invisible, shooting into melee, stacks, mounted stub, underwater stub)
- Balance guidelines (damage by tier, DR by tier, HP by role, hits-to-kill, skill benchmarks)
- Video game implementation notes (timed block/parry, auto-dodge formula, aggro/threat, damage popup UI, animation timing)

`docs/modules/conditions.md` and `data/conditions/core.json` define 24 conditions with effect-driven application/removal. `docs/modules/armor.md` defines the full DR-based slot+layer armor system with encumbrance. `data/equipment/armor.json` has 21 base types by slot/layer.

**Still missing / lower priority:**
- Mounted combat rules (stub exists in edge cases)
- Underwater combat rules (stub exists)
- Vehicle combat
- Mass combat
- More perks that interact with combat-specific mechanics (e.g. timed block bonuses, OA improvements)

**Why it matters:** Combat was the #1 gap. It is now documented well enough to run a game or implement in an engine.

### Character Advancement & Leveling
**Current state:** Base formulas exist in `progression.md` (skill points = `5 + INT × 2 + random(0…LCK)`, HP/Mana/Stamina scaling with level, perk cadence ~every 3 levels). A full Wuxia-inspired "breakthrough" system using phased monster cores and alchemically refined elixirs (leveraging `qualityGrade` + Wuxing cycles for bonus growth) is now documented.

See `docs/modules/advancement.md` for the cultivation-style item-boosted advancement design (players can delay level-ups to farm and refine better essence for stronger permanent gains to attributes, resources, or skill points).

**Still missing / thin (broader, affecting creation & long-term progression):**
- Refining path details in alchemy (how cores become higher-grade or multi-phase elixirs).
- Any requirements for performing a breakthrough (e.g. safe location, time investment, special catalysts) — currently none defined; can be added per campaign.
- Data expansion for creation/progression: more low-level perks, more spells (especially for new magic users), more equipment variety for starting characters.
- Full multi-level example characters that walk through creation → several levels with breakthroughs, using the new rules end-to-end (single-level creation examples now exist and are consistent).
- Data support for creation: more creation perks, more low-level / ability-gated perks, curated example starting gear/spells lists.

**Recently addressed:**
- Full numeric tables and resolution formulas for breakthrough bonuses (success % per grade, 0–10 skill bonus ranges with per-grade floor/ceiling, resource second-growth "higher of two", attribute +1 with restrictions). See `docs/modules/advancement.md` § Grade Scaling, Target Selection, and Luck-Based Success.
- XP thresholds: Formula **Total XP to reach level N = 450 × N × (N − 1)** (with difficulty tuning constants) + full sample table and event award guidance in `docs/modules/progression.md` § Experience Progression and Thresholds.
- XP awarding rules: Strictly event/quest/radiant driven after the very first monster kill. Meaningful skill challenges (non-crafting) award XP. Crafting does not grant XP via skill rolls (uses separate quality/material/station mechanics). See `docs/modules/progression.md` § Experience.
- Core acquisition: Only certain monsters can drop cores (no bandits/zombies). % drop chance based on monster level + killer’s Luck, using a two-roll process (drop roll → biased grade roll favoring lower tiers). See `docs/modules/advancement.md`.
- Character creation reference procedure: full consistency pass on `docs/modules/character-creation.md` against the finalized per-level gains, 2:1/1:1 tag skill ratios, AP carry rules, and breakthrough positioning language. Added explicit "Creation choices that shape long-term advancement" section (phase as foundation, INT/LCK compounding, permanent tag efficiency), strengthened the 100 gp + kit as the recommended reference baseline, added "advancement outlook" paragraphs to both worked examples (Borin metal-phase tank and Lira water-phase scout), and improved cross-references to `progression.md` (Base Gains on a Normal Level + Full Level-Up Procedure + level-30 skill cap note) and `advancement.md`. Single-level creation examples are now solid and demonstrate the new systems. Full multi-level (creation through several breakthroughs) examples remain a high-value follow-up item.
- Backgrounds: New `backgrounds` collection + `background.schema.json`. Optional one-time creation packages giving skill bonuses (additive to normal `5 + ability×2` base), starting equipment, known spells, wealth bonuses, and suggested tags. 8 varied examples implemented. Full design and many additional suggestions documented in `docs/modules/backgrounds.md`. Integrated into the official character creation sequence and both worked examples in `character-creation.md`.

**Why it matters:** Players need to know how characters grow beyond "spend skill points." The breakthrough system makes Wuxing, Alchemy, and monster hunting matter for long-term power.

**Related files:**
- `docs/modules/advancement.md` (new — breakthrough design)
- `docs/modules/progression.md`
- `docs/modules/alchemy.md` (refining path needed)
- `data/monsters/core.json` (already phased; cores will be derived or added as ingredients)
- `data/wuxing/core.json` (used for essence synergy)

### Economy, Merchants & Loot
**Current state:** Very thin. Equipment has `cost`, materials have `valueFactor`. Some price examples exist.

**Missing:**
- Merchant tables / availability
- Bartering / haggling rules
- Loot tables / treasure generation
- Currency denominations beyond gp/sp/cp
- Economic simulation (inflation, supply/demand)

**Why it matters:** Core loop in D&D, Fallout, and Elder Scrolls.

**Suggested files/steps:**
- `docs/modules/economy.md`
- `ruleset/data/loot/` or `treasure/` collection
- `ruleset/data/merchants/` or shop stock data

---

## 2. Crafting & Alchemy Modules (Mostly Stubs)

### General Crafting
**Current state:** `docs/modules/crafting.md` is a design document only.  
Materials, gems, and tiers data exist and are well-structured. Weapons now support `material` integration in concept.

**Missing:**
- Actual `recipes` data
- Crafting check rules, time, tools, stations
- Quality / masterwork system (beyond tier data)
- Downtime / activity rules

**Why it matters:** One of the major planned extensions. Currently you can only buy equipment.

**Suggested files/steps:**
- Implement minimal `recipes` collection (see modules/README.md)
- Flesh out recipe schema
- Create example data for blacksmithing, woodworking, etc.

### Alchemy
**Current state:** Good data shapes (`ingredients`, `consumables`, effects with polarity/channels). Design doc exists.

**Missing:**
- Exact quality/magnitude formula (`f(Alchemy skill, ingredient qualities, station tier)`) — explicitly TBD
- Discovery system details (how players learn ingredient effects)
- Station/tool gating rules
- Multi-effect brew resolution

**Why it matters:** Alchemy is one of the more complete-feeling planned systems but still not playable.

### Magic Crafting / Enchanting
**Current state:** `docs/modules/magic-crafting.md` is pure design document.

**Missing:**
- Any recipe/formula data
- Item affix vs full item model decision
- Charges, attunement, rarity modeling
- Rune / inscription subsystem

**Why it matters:** Critical for high-level play and magic item economy (D&D, ES, Fallout all have versions of this).

---

## 3. Social & Faction Systems

### Factions & Reputation
**Current state:** Almost nonexistent. Basic persuasion/barter skills exist.

**Missing:**
- Reputation tracking
- Faction ranks / membership benefits
- Consequences for reputation (hostile NPCs, quest locks, etc.)
- Karma / alignment equivalent (Fallout) or guild progression (Elder Scrolls)

**Why it matters:** Core to Fallout and Elder Scrolls gameplay loops.

**Suggested files/steps:**
- `docs/modules/factions.md`
- Possible `data/factions/` collection or reputation effects

### Companions & Followers
**Current state:** Completely absent.

**Missing:**
- Companion stats, loyalty, equipment
- Recruitment / dismissal rules
- Companion-specific perks or quests

**Why it matters:** Major feature in Fallout and Elder Scrolls.

---

## 4. Magic System Completeness

**Current state:** Very good on the unique parts (color schools + five-phase interaction, effect composition, spell crafting).

**Missing / Thin:**
- Components (verbal, somatic, material)
- Concentration rules
- Counterspelling / dispelling
- Ritual / extended casting
- Spell failure (armor, encumbrance, etc.)
- Metamagic or spell modification options
- More complete spell list (current `spells/core.json` is small)

**Why it matters:** Magic is one of Riftweave's strongest differentiators, but the "how do I actually cast this in play" layer is incomplete.

---

## 5. Data Completeness

| Collection     | Current Size          | Assessment                          | Needed |
|----------------|-----------------------|-------------------------------------|--------|
| monsters       | ~ dozen entries      | Too small for a usable bestiary    | High |
| spells         | Small list           | Needs many more examples           | Medium-High |
| features       | Decent but limited   | Needs many more perks for real progression | Medium |
| equipment      | Basic weapons + armor + consumables | Missing tools, kits, vehicles, more armor types | Medium |
| effects        | Good foundation      | Could use more variety | Medium |
| backgrounds    | 8 entries (new)      | Solid starter set + design doc           | Low (more examples always welcome) |

**Other missing data collections:**
- Recipes (crafting)
- Loot / treasure tables
- Factions / reputation
- Traps / hazards
- Companions

---

## 6. Documentation & Examples

**Missing or Stubs:**
- `docs/modules/factions.md`
- `docs/modules/companions.md`
- `docs/modules/economy.md`
- `docs/modules/quests.md` or adventure design guidelines
- `docs/modules/optional-rules.md`
- Full example character builds (multiple levels, with equipment and spells)
- Balance / power level guidelines
- "How to run a game" or GM section

**Recently addressed (no longer stubs):**
- `docs/modules/combat.md` (comprehensive — action economy, defense, crits, cover, called shots, examples, balance, edge cases, video game notes)
- `docs/modules/conditions.md` (24 conditions with effect-driven application/removal)
- `docs/modules/armor.md` (DR model, slots, layers, shields, encumbrance)
- `docs/modules/weapons.md` (good recent addition)

**Partially addressed:**
- Most planned modules have design docs but no implementation

---

## 7. Architecture & Tooling

**Current state:**
- Everything is monolithic under `ruleset/data/` and `ruleset/schemas/`.
- No actual module loading system (see `docs/modules/README.md`).

**Missing:**
- Module metadata and dependency system
- Module-aware validator
- Optional module loading in examples/CI
- Clear extension points for new `equipment.type`, `feature.type`, effect categories, etc.

**Why it matters:** The project explicitly wants to support optional, composable modules, but the infrastructure isn't there yet.

---

## 8. Other / Nice-to-Have

- Alignment or morality system (D&D-style)
- Birthsigns / additional creation packages (Elder Scrolls style)
- Diseases, radiation, addiction mechanics (beyond basic effects)
- Vehicles, mounts, and ship combat (listed as future module)
- Psionics (listed as future)
- Detailed encumbrance effects beyond raw carry weight
- Hunger, thirst, sleep / survival needs (some cRPGs)
- Traps and environmental hazards (detailed rules)
- Optional rules for different power levels or "gritty" play

---

## Priority Summary (Suggested Order)

**High (blocks playability):**
- ~~Combat system + conditions~~ **⦿ DONE** — see `combat.md`, `conditions.md`, `armor.md`
- Actual crafting recipes + quality rules (start with one module)
- Full character advancement / XP procedure (base + breakthrough system — design doc started in `advancement.md`)

**Medium-High:**
- Economy / loot
- Factions / reputation
- Companions
- Magic casting details

**Medium:**
- Data expansion (monsters, spells, features)
- Module system infrastructure
- More documentation

**Lower / Future:**
- Psionics, vehicles, optional subsystems
- Full alignment / morality

---

**How to use this file:**  
Update this document as items are completed or new gaps are discovered. When starting work on a module or system, create a dedicated design doc in `docs/modules/` first (as recommended in `docs/modules/README.md`), then link it here.

Last major review: Combat resolution system completed — `docs/modules/combat.md`, `docs/modules/conditions.md`, `docs/modules/armor.md` all finalized with dual-resolution mechanics, data, examples, edge cases, and balance guidelines. Phase 3 documentation and future-proofing (flow checklist, edge cases, balance numbers, video game notes) also finished.