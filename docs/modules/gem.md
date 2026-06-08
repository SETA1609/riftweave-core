# Gems & Engraving

**Status:** Core (vocabulary + data) · engraving resolution lives in the crafting modules

Gems are the inputs to **engraving** — a process distinct from enchanting. This
document defines what a gem is and what engraving does. Data lives in
[`ruleset/data/gems/core.json`](../../ruleset/data/gems/core.json) (schema
`gem.schema.json`); the rarity numbers come from
[`ruleset/data/tiers/core.json`](../../ruleset/data/tiers/core.json).

---

## 1. Engraving vs. enchanting

| | Enchanting (magic-crafting) | **Engraving (gems)** |
| --- | --- | --- |
| Adds effects? | **Yes** — binds effects from the pool onto an item | **No** — engraving grants no effects |
| Touches element? | no | **Yes** — overwrites the item's `phase` |
| Bonus | the effects themselves | a magnitude bonus to *later* enchantments |

Engraving never gives an item a new spell or power. What it does is **two things**:
re-element the item, and make any enchantments it later carries land harder.

## 2. Engraving overwrites the element

A gem carries a `phase` (its element). When an item is engraved, the **dominant gem**
— the highest `quality`, or decided by gem quantity — **overwrites the item's element**:

> Engrave a **wooden** sword (element Wood) with a **ruby** (Fire) → the sword's
> element becomes **Fire**.

The new element then resolves through the five-phase cycles in
[`data/wuxing/core.json`](../../ruleset/data/wuxing/core.json) exactly like any other
phased thing (see [`magic.md`](./magic.md) § Elemental interaction). Re-elementing a
weapon is how you turn it from weak-into-a-foe (overcome) to strong-against-it.

> The exact rule for *which* gem dominates when several are set — top quality vs. a
> majority by count — and how socket count works are **item/recipe concerns defined
> later** by the crafting modules, not here.

## 3. The enchantment bonus comes from tier

An engraved item gains a bonus to the **magnitude of effects later enchanted onto
it**. That bonus is **not** stored on the gem — it comes from the gem's **`tier`**, the
single source of truth (`data/tiers`):

| Tier | Name | effectMagnitude | valueFactor |
| --- | --- | --- | --- |
| 1 | Common | +8% | ×1 |
| 2 | Uncommon | +12% | ×3 |
| 3 | Rare | +20% | ×6 |
| 4 | Epic | +25% | ×10 |
| 5 | Legendary | +35% | ×20 |

How the bonus combines across multiple gems and is applied by a recipe is **deferred**
to the crafting / magic-crafting modules.

## 4. What a gem type defines (and what it doesn't)

A gem **type** is just two things:

- **`phase`** — the element it imparts on engraving.
- **`tier`** — Common→Legendary rarity, a number referencing `data/tiers`. The source
  of truth for the engraving bonus and value.

A gem carries **no effects** and no per-gem numbers — only `phase` + `tier`.

**Quality is not part of the type.** An individual stone's quality varies (one ruby
may be *petty*, another *grand*), so it is a **per-instance attribute** assigned to a
specific gem when it is found or cut, drawn from the shared `qualityGrade` ladder:
`petty · minor · lesser · common · major · greater · grand · legendary` (lowest→highest,
*common* is the baseline). It is what decides engraving dominance (§2) and is not
stored in the gem definitions here.

## 5. The gem list

Colors map to phases on the classic five-element associations (red→Fire, green→Wood,
blue→Water, yellow→Earth, white→Metal):

| Gem | Element | Tier |
| --- | --- | --- |
| Ruby | fire | Rare |
| Garnet | fire | Common |
| Emerald | wood | Rare |
| Jade | wood | Common |
| Sapphire | water | Rare |
| Aquamarine | water | Common |
| Topaz | earth | Uncommon |
| Citrine | earth | Common |
| Diamond | metal | Epic |
| Moonstone | metal | Common |

Each element has a higher-tier stone and a humble Common one. Quality is independent —
any of these stones can turn up at any quality from petty to legendary.

---

## Data & schema map

| Concern | Schema | Data |
| --- | --- | --- |
| Gems | `gem.schema.json` | `data/gems/core.json` |
| Rarity numbers | `tier.schema.json` | `data/tiers/core.json` |
| Quality ladder | `schema.json#/definitions/qualityGrade` | — |
| Element vocabulary | `schema.json#/definitions/phase` | — |
| Interaction cycles | `wuxing.schema.json` | `data/wuxing/core.json` |

## Open items

- **Engraving resolution.** Which gem dominates (quality vs. quantity), socket count,
  and how the tier bonus and quality feed a recipe are specified by the crafting /
  magic-crafting modules, not here.
- **Per-gem overrides.** Numbers come from `tier`; a uniquely prized stone that beats
  its tier would need an optional per-gem override field (not yet added).
- **More stones.** No gem is Legendary tier yet; the ladder leaves headroom.
