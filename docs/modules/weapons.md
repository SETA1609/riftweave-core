# Weapons

Weapons in Riftweave are defined as base equipment entries in `ruleset/data/equipment/weapons.json` (governed by `equipment.schema.json`). They are deliberately simple and use **classic D&D 3.5-style dice** for damage. Material-based variations, crafting options, and restrictions are **not** encoded as data variations in the schema or base entries. Instead, they are explained here for use by the crafting and magic-crafting modules.

## Core Classification Axes

Weapons use two primary orthogonal classification systems:

1. **Damage Type** (maps directly to combat skills)
   - `slashing` → `blades` skill
   - `bludgeoning` → `blunt` skill
   - `piercing` → `piercing` skill
   - Ranged weapons use `marksmanship`.

   This split replaced the old single `melee_weapons` skill. A weapon's `damage.type` determines which skill is used for the attack roll.

2. **Length / Reach Category** (determines engagement distance and handedness)
   - `short`: Close-quarters weapons (daggers, handaxes, small training tools). Typically one-handed.
   - `normal`: Standard melee weapons (swords, maces, axes). Typically one-handed (or versatile).
   - `reach`: Extended-reach polearms and large two-handed weapons. Can attack **2 units of length** (most weapons attack 1 unit).

   Each weapon also stores an explicit `attackReach` (1 or 2) for clarity in rules resolution.

Additional properties (in the free-form `properties` array) handle details like:
- `light`, `finesse`, `thrown`
- `versatile` (can be used one- or two-handed)
- `heavy`, `two_handed`
- `reach` (for some polearms)

**Handedness rule of thumb**: short and normal weapons are one-handed (or versatile). Reach weapons are two-handed.

## Base Damage — Classic 3.5 Dice

All weapons use standard D&D 3.5 / Pathfinder-style dice expressions for their base damage (before any material or enchantment modifiers). The dice are chosen for balance and familiarity:

| Weapon          | Damage     | Type      | Length | Attack Reach | Skill      | Notes / Properties          |
|-----------------|------------|-----------|--------|--------------|------------|-----------------------------|
| Dagger          | 1d4        | piercing  | short  | 1            | piercing   | finesse, light, thrown     |
| Handaxe         | 1d6        | slashing  | short  | 1            | blades     | light, thrown              |
| Quarterstaff    | 1d6        | bludgeoning | normal | 1          | blunt      | versatile                  |
| Mace            | 1d6        | bludgeoning | normal | 1          | blunt      | -                          |
| Spear           | 1d6        | piercing  | reach  | 2            | piercing   | thrown, two_handed         |
| Longsword       | 1d8        | slashing  | normal | 1            | blades     | versatile                  |
| Rapier          | 1d6        | piercing  | normal | 1            | piercing   | finesse, light             |
| Shortbow        | 1d6        | piercing  | -      | -            | marksmanship | ammunition, two_handed   |
| Longbow         | 1d8        | piercing  | -      | -            | marksmanship | ammunition, heavy, two_handed |
| Greatsword      | 2d6        | slashing  | normal | 1            | blades     | heavy, two_handed          |
| Glaive          | 1d10       | slashing  | reach  | 2            | blades     | heavy, two_handed          |
| Warhammer       | 1d10       | bludgeoning | reach | 2            | blunt      | heavy, two_handed          |
| Halberd         | 1d10       | slashing  | reach  | 2            | blades     | heavy, two_handed          |

These are the **base** dice for the weapon type. They assume a baseline construction (historically often iron or steel quality). The actual performance changes with the material used.

## Materials and Weapon Variations

Materials (see `data/materials/core.json` and `docs/modules/materials.md`) are the primary way to create **variations** of the base weapons.

When a weapon is crafted from a material:
- Start with the base dice listed above (the "classic 3.5" profile for that weapon type).
- Apply the material's `modifiers.attack` bonus.
- The material also sets the item's `phase` (for five-phase elemental interactions) and can provide other effects (weight factors, value, special tags like "brittle" or "conductive", and `effectMagnitude` for enchantments).
- Different materials turn the *same* base weapon profile into meaningfully different tools.

### Material Attack Bonuses (Current Range)

Materials provide an attack bonus in the range of **0 to +10** depending on the material's quality and purpose (0 being the floor for materials that provide no combat attack bonus, such as gold for arms). 

**Important**: None of the current materials in `data/materials/core.json` give the maximum +10. The highest are:
- Obsidian: +6 (wickedly sharp but brittle — best for small blades, poor for heavy impact weapons).
- Steel: +5 (the premium "standard" for arms).
- Elven Wood: +3 (light and magically attuned).
- Iron: +3 (reliable baseline).
- Bronze: +2.
- Wood: 0 (no direct combat bonus; light and cheap for hafts, training versions, bows, and shields); Silver: +2 (with special undead properties); Copper: +1 with trade-offs.

Gold gives 0 attack (too soft for arms) but excels at `effectMagnitude` for magical items.

### How Materials Create Variations (Examples)

The same base weapon type becomes a different effective weapon depending on material:

- **Longsword (base 1d8 slashing)**:
  - Iron: solid, reliable 1d8 +3. Everyday soldier's weapon.
  - Steel: 1d8 +5. Sharper, tougher, the preferred martial blade.
  - Wood: 1d8 +0 (no combat bonus). Classic training waster / practice sword. Very light (weightFactor 0.5) and cheap, safe for sparring, but poor against real armor. The "variation" comes from low weight and cost rather than stat bonuses.
  - Obsidian (material): high +6 attack but brittle — risky on a large blade (prone to chipping).
  - Gold: useless for combat (0 attack) but excellent conduction if you want to enchant it heavily.

- **Greatsword (base 2d6 slashing)**:
  - Steel: devastating 2d6 +5 heavy blade.
  - Wood: possible as an oversized training or ceremonial piece (very light, 0 combat bonus).
  - Obsidian (material) or copper: generally avoided (brittle or soft for such a large, high-force weapon).

- **Mace (base 1d6 bludgeoning)**:
  - Iron/Steel: good crushing power.
  - Wood: possible but weak (0 combat bonus, more like a heavy club). Better represented by the quarterstaff entry.
  - Obsidian (material): poor choice — the material is optimized for cutting, not impact.

- **Spear / Glaive / Halberd (reach weapons, base 1d6 or 1d10)**:
  - Wood (for spear): natural for the haft; head can still benefit from metal modifiers in a real crafted item.
  - Steel: excellent reach weapon.
  - Pure wood versions of glaive/halberd are rare and weak for the specialized blade/axe head.

- **Dagger (base 1d4 piercing)**:
  - Iron/Steel: standard sidearm.
  - **Obsidian (base dagger + obsidian material)**: exceptionally sharp (+6 attack bonus) — excellent for a small blade. This is the profile that was previously listed as a separate "Obsidian Dagger" base entry; it is now correctly represented as the generic dagger form modified by the obsidian material at crafting time.
  - Wood: almost never (a wooden dagger is barely a weapon).

- **Bows (shortbow 1d6, longbow 1d8 piercing)**:
  - Heavily restricted to wood or elven wood. Metal bows don't make sense. Wood provides the necessary flexibility and light weight. Elven wood versions are superior in power and magic conduction.

### Design Notes for Crafting Modules

- The base weapon entry gives the **identity and dice profile**.
- The chosen material supplies the **stat variation** (attack bonus in the 0 to +10 range, phase, weight/value factors, special tags).
- Brittle materials (obsidian) or soft materials (gold) are self-limiting via low or negative suitability for certain weapon types.
- Training versions (wooden swords) are intentionally weaker but safe and cheap.
- Future recipes can combine a weapon type + material + optional gem engraving to produce the final item.
- The five-phase system (via material `phase`) means a metal greatsword behaves differently against fire- or wood-aligned foes than a wooden training sword would.

## Full Current Weapon List (Base Stats)

See the table in the "Base Damage" section above for the complete list with dice, length, reach, skill, and key properties.

All entries live in `ruleset/data/equipment/weapons.json`. The schema only enforces the core structure (category, range, length, attackReach, skill, damage dice+type, properties). Material application and any "this weapon can't be made from X" logic live in documentation and the crafting modules, not as hard schema constraints on the base data.

## Related Systems

- **Skills**: See `data/skills/core.json` and `docs/modules/progression.md`.
- **Materials**: `data/materials/core.json`, `docs/modules/materials.md`.
- **Crafting & Magic Crafting**: `docs/modules/crafting.md`, `docs/modules/magic-crafting.md` (these modules will consume weapon types + materials to produce actual items).
- **Five-Phase Interactions**: Material phase + weapon damage type interact via `data/wuxing/core.json`.

This design keeps the core data clean and declarative while giving the crafting systems rich, explainable ways to generate weapon variations.