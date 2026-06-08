# Materials

**Status:** Core (vocabulary + data) · crafting resolution lives in the crafting modules

Materials are the substances a **weapon, armor, or accessory** is crafted from. Each
material carries an element and the bonuses it confers on the finished item. Data
lives in [`ruleset/data/materials/core.json`](../../ruleset/data/materials/core.json)
(schema `material.schema.json`).

Materials are distinct from **gems** (see [`gem.md`](./gem.md)): a material is the
*body* of the item and sets its **base** element and combat/conduction stats;
engraving with a gem can later **overwrite** that element.

---

## 1. What a material confers

| Field | Meaning |
| --- | --- |
| `phase` | The material's element. The crafted item carries it (until engraving overwrites it) and it resolves through the five-phase cycles in [`data/wuxing`](../../ruleset/data/wuxing/core.json). |
| `modifiers.attack` | Bonus to the item's attack/damage rating (weapons). |
| `modifiers.defense` | Bonus to the item's defense/armor rating (armor, shields). |
| `modifiers.effectMagnitude` | Magic-conduction bonus: a fractional boost to the magnitude of effects enchanted onto the item (e.g. `0.15` = +15%). |
| `appliesTo` | Which item kinds the material can make: `weapon`, `armor`, `accessory`. |
| `tier` | Rarity/quality tier (aligns with the `data/tiers` ladder). |
| `weightFactor` / `valueFactor` | Multipliers on the item's base weight and value (1.0 = neutral). |

A material declares **at least one** modifier. The three modifier types map directly
to the three things a material can improve: **attack**, **defense**, and the
**magnitude of enchanted effects**.

## 2. Element interaction

Because the crafted item inherits the material's `phase`, the choice of material is
also an elemental choice. A **Metal** blade, for instance, amplifies a **Water**
enchantment set into it (Metal generates Water in the generating cycle), while it is
soft against Fire (Fire melts Metal). See [`magic.md`](./magic.md) § Elemental
interaction for the full matrix. Engraving (gems) can later re-element the item,
changing these matchups.

## 3. The material list

| Material | Element | Attack | Defense | effectMagnitude | Makes | Tier |
| --- | --- | --- | --- | --- | --- | --- |
| Copper | metal | +1 | +1 | +10% | weapon, armor, accessory | 1 |
| Bronze | metal | +2 | +2 | — | weapon, armor | 1 |
| Iron | metal | +3 | +3 | — | weapon, armor | 2 |
| Steel | metal | +5 | +5 | — | weapon, armor | 3 |
| Silver | metal | +2 | — | +15% | weapon, accessory | 2 |
| Gold | metal | — | — | +30% | accessory | 2 |
| Wood | wood | 0 | 0 | — | weapon, armor, accessory | 1 |
| Elven Wood | wood | +3 | +3 | +20% | weapon, armor, accessory | 3 |
| Obsidian | fire | +6 | — | — | weapon | 2 |

Reading the table: workhorse metals (bronze→iron→steel) climb in attack/defense;
wood provides 0 combat bonuses (it's for hafts, bows, and shields — light and cheap);
**conductive** materials (copper, silver, gold, elven wood) trade combat stats for
`effectMagnitude` — gold is useless as a blade but the best enchanting metal; obsidian
is wickedly sharp but brittle (weapon only, no defense).

## 4. Materials vs. gems vs. tier

- **Material** — the item's body: sets **base** element + attack/defense/conduction.
- **Gem** ([`gem.md`](./gem.md)) — engraving: **overwrites** element, adds a tier-based
  enchant bonus, no combat stats.
- **Tier** ([`data/tiers`](../../ruleset/data/tiers/core.json)) — the shared rarity
  ladder; a material's `tier` places it on that ladder.

Unlike gems (whose numbers derive wholly from `tier`), a material's combat modifiers
are **intrinsic** — iron is iron regardless of rarity — so they are stated explicitly
on each material rather than read from its tier.

---

## Data & schema map

| Concern | Schema | Data |
| --- | --- | --- |
| Materials | `material.schema.json` | `data/materials/core.json` |
| Element vocabulary | `schema.json#/definitions/phase` | — |
| Rarity ladder | `tier.schema.json` | `data/tiers/core.json` |
| Interaction cycles | `wuxing.schema.json` | `data/wuxing/core.json` |

## Open items

- **Recipe resolution.** How a recipe consumes materials and combines their modifiers
  (and a gem's engraving) onto the finished item is specified by the crafting /
  magic-crafting modules, not here.
- **Phase coverage.** Materials currently cover Metal, Wood, and Fire; Earth and Water
  materials (e.g. stone, a watery alloy) are open slots.
- **Quality axis.** Materials use `tier` only; if material craftsmanship should vary
  by the `qualityGrade` ladder (as gems do), that field can be added later.
