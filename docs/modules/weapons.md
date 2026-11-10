# Weapons

Weapons in Riftweave are defined as base equipment entries in `ruleset/data/equipment/weapons.json` (governed by `equipment.schema.json`). They are deliberately simple and use **classic D&D 3.5-style dice** for damage. Material-based variations and crafting options are **not** encoded as data variations in the schema or base entries. Instead, they are explained here for use by the crafting and magic-crafting modules.

## Core Classification Axes

Weapons use two primary orthogonal classification systems:

1. **Damage Type** (maps directly to combat skills)
   - `slashing` → `blades` skill
   - `bludgeoning` → `blunt` skill
   - `piercing` → `piercing` skill
   - Ranged weapons use one of the split archery skills: `bows`, `crossbows`, `guns`, or `throwing_weapons`.

   This split replaced the old unified `marksmanship` skill (and the earlier single `melee_weapons`). A weapon's `damage.type` (for melee) or explicit `skill` (for ranged) determines the governing skill for the attack roll.

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
| Shortbow        | 1d6        | piercing  | -      | -            | bows         | ammunition, two_handed   |
| Longbow         | 1d8        | piercing  | -      | -            | bows         | ammunition, heavy, two_handed |
| Greatsword      | 2d6        | slashing  | normal | 1            | blades     | heavy, two_handed          |
| Glaive          | 1d10       | slashing  | reach  | 2            | blades     | heavy, two_handed          |
| Warhammer       | 1d10       | bludgeoning | reach | 2            | blunt      | heavy, two_handed          |
| Halberd         | 1d10       | slashing  | reach  | 2            | blades     | heavy, two_handed          |
| Improvised Bow  | 1d6        | piercing  | -      | -            | bows       | ammunition, two_handed     |
| Composite Bow   | 1d8        | piercing  | -      | -            | bows       | ammunition, two_handed     |
| Hand Crossbow   | 1d6        | piercing  | -      | -            | crossbows  | ammunition, light, loading |
| Crossbow        | 1d8        | piercing  | -      | -            | crossbows  | ammunition, two_handed     |
| Javelin         | 1d6        | piercing  | -      | -            | throwing_weapons | thrown               |
| Throwing Knife  | 1d4        | piercing  | -      | -            | throwing_weapons | thrown, light, finesse |
| Throwing Axe    | 1d6        | slashing  | -      | -            | throwing_weapons | thrown               |
| Brass Knuckles  | 1d4        | bludgeoning | short | 1           | unarmed    | light                      |
| Cestus          | 1d6        | bludgeoning | short | 1           | unarmed    | light                      |
| Spiked Gauntlet | 1d6        | piercing  | short  | 1            | unarmed    | light                      |
| Flintlock Pistol | 1d10      | piercing  | -      | -            | guns       | ammunition, loading, light |
| Arquebus        | 1d12       | piercing  | -      | -            | guns       | ammunition, loading, two_handed |
| Blunderbuss     | 2d6        | piercing  | -      | -            | guns       | ammunition, loading, two_handed |

These are the **base** dice for the weapon type. They assume a baseline construction (historically often iron or steel quality). The actual performance changes with the material used.

## Damage Dice by Handling and Size

The dice in the table above are for a normal-sized weapon. Two things set a weapon's dice:
how it is handled, and how big it is built. The handling ranges are a guide for new rows;
the size step is data. [A]

### The dice ladder

One step is one die size. Moving up or down the ladder is the only way dice grow or shrink.

| Step | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Dice | 1d2 | 1d3 | 1d4 | 1d6 | 1d8 | 1d10 | 2d6 | 2d8 | 3d6 | 3d8 | 4d8 |
| Mean | 1.5 | 2 | 2.5 | 3.5 | 4.5 | 5.5 | 7 | 9 | 10.5 | 13.5 | 18 |

`1d12` (mean 6.5) sits on step 7 with `2d6`: same step, a swingier roll. [A]

### By handling (normal size)

| Handling | Properties | Dice | Examples |
| --- | --- | --- | --- |
| Light | `light` | 1d4 to 1d6 | dagger, shortsword, brass knuckles, throwing knife |
| One-handed | none, or `versatile` | 1d6 to 1d8 | mace, longsword, battleaxe, javelin |
| Two-handed | `two_handed` | 1d6 to 1d10 | spear, shortbow, composite bow, crossbow |
| Heavy two-handed | `heavy`, `two_handed` | 1d10 to 2d6 | glaive, halberd, greatsword |

A weapon one step above its handling's range pays for it elsewhere: `loading` (the
flintlock pistol is light at 1d10, the arquebus and blunderbuss two-handed at 1d12 and 2d6),
or a high price. A bow trades dice for range: the longbow is `heavy` for its draw and rolls
1d8, and pays back in range (30 / 120 units). A weapon below the range is cheap or
concealable; the balance gate lists it as an accepted finding (dagger, sling, throwing
knife, brass knuckles). The balance band `weapon_dice` (`balance/core.json`, 1d6 to 2d10)
is checked after the size step. [V 2026-11-07, `balance_report.py --gate`]

### By size

Size is how a weapon is built, not who wields it: anyone can find an undersized shortbow or
an oversized longbow. A row's `size` is `undersized`, `normal`, or `oversized`; omitted
means `normal`, so every row without it keeps its dice. Undersized moves the dice one rung
down the ladder, oversized one rung up, never more. The ladder and the steps are data:
`ruleset/data/combat/dice_ladder.json`. A row keeps its base dice; the consumer applies the
step (`balance_report.sized_dice`). [V 2026-11-07, `test_equipment.py`]

Size also changes how easy the weapon is to use: an undersized weapon gives +10 to the
attack, an oversized one −10. The weapon emits it as an `add` on `attack`, source
`item:<key>`, origin `core:origin/equipment`, while it is used. Small dice, easy to land;
big dice, hard to land. [A] Size sets the price too: an undersized weapon costs 75% of its
base, an oversized one 150% (`value_percent`), multiplied with the material's
`valueFactor`. [A]

| Normal | Undersized (−1, +10 attack, 75% price) | Oversized (+1, −10 attack, 150% price) | Normal examples |
| --- | --- | --- | --- |
| 1d4 | 1d3 | 1d6 | dagger, sling |
| 1d6 | 1d4 | 1d8 | mace, shortsword, shortbow |
| 1d8 | 1d6 | 1d10 | longsword, longbow, crossbow |
| 1d10 | 1d8 | 2d6 | glaive, halberd, flintlock pistol |
| 1d12, 2d6 | 1d10 | 2d8 | greatsword, arquebus, blunderbuss |

No row is sized. A sized weapon is composed where it is used, like its material: a base
plus `size` (an undersized shortbow is the shortbow at 1d4). Size is not an
enchantment: it takes no enchantment slot, does not stack under the `enchantment` origin,
and a dispel does not remove it. Size does not touch range, handling, or who can use the
weapon. [A] Monster natural attacks keep their own flat `attack.damage.amount`
([monsters.md](monsters.md)); the ladder is for weapons. [A]

## Rows and Composed Items

A weapon is a base plus a material, a size, enchantments, and coatings, and the rules
build any combination. `weapons.json` lists only what the rules cannot build: each
distinct base weapon (the catalogue entry that loot, shops, and prices name as
`core:weapon/<key>`), and named items that differ from the rules on purpose, with their
own name, story, or price (`elven_wood_longbow`, `cursed_iron_longsword`). Generic
variants are not rows: a loot row or a crafted item names the base and its `material`
(and, from M4-07, its `size`). [V 2026-11-07, owner decision]

## Materials and Weapon Variations

**Default materials, not valid materials.** Like Elin, any material can go into any base:
a granite bow, a gold dagger, a wooden greatsword. A base's `default_materials` names the
materials it is found and sold in; loot and shops should use only that list, so odd
combinations come from player crafting alone. Not yet enforced: a loot row's `material`
applies to every weapon and armor it yields, so the stone golem's steel hoard can drop a
steel bow. [U] A poor combination is weak, not forbidden:
the material's own numbers (attack bonus, weight, tags such as `brittle`) are the cost.
[V 2026-11-07, owner decision] `validate.py` no longer rejects a `crafted_items` row whose
material is outside the base's list. [V 2026-11-07]

Materials (see `data/materials/core.json` and `docs/modules/materials.md`) are the primary way to create **variations** of the base weapons.

When a weapon is crafted from a material:
- Start with the base dice listed above (the "classic 3.5" profile for that weapon type).
- Apply the material's `weapon.attack` bonus.
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
  - Default materials are wood and elven wood: wood gives the flexibility and light weight, and elven wood versions are superior in power and magic conduction. A granite or iron bow is possible, but only a player crafts one.

### Design Notes for Crafting Modules

- The base weapon entry gives the **identity and dice profile**.
- The chosen material supplies the **stat variation** (attack bonus in the 0 to +10 range, phase, weight/value factors, special tags).
- Brittle materials (obsidian) or soft materials (gold) are self-limiting via low or negative suitability for certain weapon types.
- Training versions (wooden swords) are intentionally weaker but safe and cheap.
- Future recipes can combine a weapon type + material + optional gem engraving to produce the final item.
- The five-phase system (via material `phase`) means a metal greatsword behaves differently against fire- or wood-aligned foes than a wooden training sword would.

## Full Current Weapon List (Base Stats)

See the table in the "Base Damage" section above for the complete list with dice, length, reach, skill, and key properties.

All entries live in `ruleset/data/equipment/weapons.json`. The schema only enforces the core structure (category, range, length, attackReach, skill, damage dice+type, properties). Material application lives in documentation and the crafting modules, not as schema constraints on the base data.

## Related Systems

- **Skills**: See `data/skills/core.json` and `docs/modules/progression.md`.
- **Materials**: `data/materials/core.json`, `docs/modules/materials.md`.
- **Crafting & Magic Crafting**: `docs/modules/crafting.md`, `docs/modules/magic-crafting.md` (these modules will consume weapon types + materials to produce actual items).
- **Five-Phase Interactions**: Material phase + weapon damage type interact via `modules/wuxing/data/wuxing.json`.

This design keeps the core data clean and declarative while giving the crafting systems rich, explainable ways to generate weapon variations.