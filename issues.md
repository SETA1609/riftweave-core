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
**Current state:** High-level only.  
`docs/modules/progression.md` covers d100 roll-under, margin of success, criticals (01–05 success / 96–00 fumble), and basic derived stats (HP, Stamina/Mana, carry weight, initiative). Weapons now have good classification (length/reach, damage type → skill). Armor is described vaguely as "damage-reduction rating".

**Missing:**
- Full action economy (actions, bonus actions, reactions, movement)
- Opportunity attacks / attacks of opportunity
- Detailed conditions (blinded, grappled, prone, poisoned, etc.) with mechanical effects
- Cover, flanking, difficult terrain
- Called shots / limb damage (Fallout-style)
- Critical hit tables or special effects
- Full defense model (dodge vs block vs armor) — explicitly called a "follow-up"
- Turn-based vs action-combat resolution details (still undecided per progression.md)

**Why it matters:** Combat is the most visible part of any cRPG/TTRPG. Current state is too abstract to run a real game.

**Suggested files/steps:**
- Create `docs/modules/combat.md`
- Create `docs/modules/conditions.md` (or `data/conditions/core.json` + schema)
- Expand `equipment/armor.json` with more detailed DR properties
- Decide on action points vs stamina model

### Character Advancement & Leveling
**Current state:** Base formulas exist in `progression.md` (skill points = `5 + INT × 2 + random(0…LCK)`, HP/Mana/Stamina scaling with level, perk cadence ~every 3 levels). A full Wuxia-inspired "breakthrough" system using phased monster cores and alchemically refined elixirs (leveraging `qualityGrade` + Wuxing cycles for bonus growth) is now documented.

See `docs/modules/advancement.md` for the cultivation-style item-boosted advancement design (players can delay level-ups to farm and refine better essence for stronger permanent gains to attributes, resources, or skill points).

**Still missing / thin:**
- Full numeric tables or resolution formulas for grade/phase breakthrough bonuses (exact gains from offered cores/elixirs — affinity table, drop rules, and grade bias exist, but concrete bonus application is still high-level).
- Refining path details in alchemy (how cores become higher-grade or multi-phase elixirs).

**Drafted (needs review/expansion):**
- Concrete "what you get at each level" full procedure + checklist — A complete step-by-step level-up procedure (base package + optional breakthrough) with player checklist and GM notes has been drafted in `docs/modules/progression.md` § Full Level-Up Procedure. It integrates the new XP table, per-level gains (1 perk, skill points at 2:1/1:1, resource growth, TTRPG AP carry-over rules), breakthrough option (max 3 items; player selects one governed target per item from its phase — attributes only if grade > common and +1 max total this level; resources or skills otherwise), luck-based success per item using the proposed formula (base from grade + phase synergy + LCK × 2.5, with target-type weighting: skills easiest → resources → attributes hardest; even legendary items can fail), and the two-roll core drop system. On success: attributes +1 (high-grade only), skills 0–5 range (grade sets floor/ceiling), resources = higher of two level-growth rolls. Still needs concrete numeric success % tables and bonus ranges per grade/target, plus any additional breakthrough requirements (time, safety, etc.).

**Recently addressed:**
- XP thresholds: Formula **Total XP to reach level N = 450 × N × (N − 1)** (with difficulty tuning constants) + full sample table and event award guidance in `docs/modules/progression.md` § Experience Progression and Thresholds.
- XP awarding rules: Strictly event/quest/radiant driven after the very first monster kill. Meaningful skill challenges (non-crafting) award XP. Crafting does not grant XP via skill rolls (uses separate quality/material/station mechanics). See `docs/modules/progression.md` § Experience.
- Core acquisition: Only certain monsters can drop cores (no bandits/zombies). % drop chance based on monster level + killer’s Luck, using a two-roll process (drop roll → biased grade roll favoring lower tiers). See `docs/modules/advancement.md`.

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
| effects        | Good foundation      | Needs many more condition-style effects | Medium |

**Other missing data collections:**
- Conditions / status effects
- Recipes (crafting)
- Loot / treasure tables
- Factions / reputation
- Traps / hazards
- Companions

---

## 6. Documentation & Examples

**Missing or Stubs:**
- `docs/modules/combat.md` (critical)
- `docs/modules/conditions.md`
- `docs/modules/factions.md`
- `docs/modules/companions.md`
- `docs/modules/economy.md`
- `docs/modules/quests.md` or adventure design guidelines
- `docs/modules/optional-rules.md`
- Full example character builds (multiple levels, with equipment and spells)
- Balance / power level guidelines
- "How to run a game" or GM section

**Partially addressed:**
- `docs/modules/weapons.md` (good recent addition)
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
- Combat system + conditions
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

Last major review: Added `docs/modules/advancement.md` (Wuxia-style item-boosted breakthroughs with monster cores + alchemical elixirs using Wuxing phases and qualityGrade) (current session).