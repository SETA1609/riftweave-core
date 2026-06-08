# Races & Ancestry

**Status:** Core (implemented as data + schema)

This document defines how races work in Riftweave: what a race contributes to a
character, the **lineage system** that relates races to one another (parent →
subrace inheritance vs. template → kin sisterhood), and how to author new ones. For
the character-build context (attributes, skills, perks) see
[`progression.md`](./progression.md); for how a race's element feeds combat see
[`magic.md`](./magic.md) § Elemental interaction.

Data lives in [`ruleset/data/races/core.json`](../../ruleset/data/races/core.json),
governed by `race.schema.json`.

---

## 1. What a race provides

A race is a small bundle of character-creation modifiers, applied **on top of** the
base attribute spread (start each attribute at 4, distribute 10 points, then apply
the racial `abilityModifiers` — see [`progression.md`](./progression.md)
§ Attributes).

| Field | Meaning |
| --- | --- |
| `abilityModifiers` | Integer deltas to the eight attributes, e.g. `{ "end": 2, "str": 1, "agi": -1 }`. |
| `phase` | The race's single **element** (Wuxing five-phase). Drives elemental affinity — see §4. |
| `speed` | Walking speed (feet per round). |
| `size` | `tiny` · `small` · `medium` · `large`. |
| `traits` | Named, described special abilities (darkvision, lucky, claws, …). Mechanically resolved by the engine; the ruleset stores their text. |
| `lineage` | How this race relates to others — see §2. **The defining structural field.** |

Every race also carries the standard `id`, `name`, `description`, and optional
`source`.

---

## 2. The lineage system

Every race declares a `lineage` object: a `role` and, when relevant, a `parentRace`.
There are **two distinct relationships**, and keeping them separate is the point.

```json
"lineage": { "role": "subrace", "parentRace": "elf" }
```

| Role | Playable? | `parentRace` | Relationship |
| --- | --- | --- | --- |
| `standalone` | yes | — | A complete, unrelated race (e.g. `human`). |
| `parent` | no (abstract) | — | An ancestor that subraces inherit from (`elf`, `dwarf`). |
| `subrace` | yes | required | **Inherits** from its parent; overrides freely (`green_elf`, `mountain_dwarf`). |
| `template` | no (abstract) | — | A **shape** for sister races (`beastman`); provides structure, not values. |
| `kin` | yes | required | A sister conforming to a template (`catman`, `wolfman`); shares nothing mechanical with siblings. |

### 2a. Parent → subrace = inheritance

A `parent` race (elf, dwarf) is abstract — you don't play "a generic elf," you play a
green elf or a mountain dwarf. The parent holds the **shared defaults**; each
`subrace` **descends from it** and may keep or override any of them:

- **Inherit** — omit a field and the parent's value applies. A subrace may drop
  `speed`/`size` to take the parent's (e.g. `green_elf` inherits the elf's 30 / medium).
- **Override** — declare a field to replace the parent's. A subrace may even change
  its **element**: `black_elf` overrides the elf's `wood` to `water`; `hill_dwarf`
  overrides the dwarf's `metal` to `earth`.
- **Stack** — `abilityModifiers` and `traits` add to the parent's rather than
  replacing them (see §3 resolution).

Because a parent is abstract, each family ships a **baseline subrace** that overrides
nothing — `common_elf` and `common_dwarf`. These declare only the required `phase`
(equal to the parent's) and inherit speed, size, modifiers, and traits unchanged, so
a player who just wants a "plain" elf or dwarf has a playable entry.

### 2b. Template → kin = sisterhood (NOT inheritance)

A `template` (beastman) is also abstract, but it works the opposite way. It defines
only the **shape** — "beastfolk are anthropomorphic animals with their own element,
natural weapons, and heightened senses." Its `kin` (catman, wolfman, ratman, …) are
**sisters**: each declares *everything* itself and they **need share nothing
mechanical**. The template imposes a contract, not content.

> The practical difference: change the `elf` parent and every elf subrace shifts with
> it. Change the `beastman` template and **no** kin changes — they only ever borrowed
> its shape. This is exactly why beastfolk authored by different users can look
> wildly different while still being "beastfolk."

---

## 3. Field requirements & inheritance resolution

### Required fields by role (enforced by schema)

All roles require `id`, `name`, `description`, `lineage`. Beyond that:

| Role | `phase` | `speed` + `size` |
| --- | --- | --- |
| `standalone` | required | required |
| `parent` | required (it's the inheritable default) | required |
| `subrace` | **required** (keep or override) | optional (inherit if omitted) |
| `kin` | required | required (kin are self-contained) |
| `template` | omitted | omitted |

(These conditionals are enforced in `race.schema.json` via `if/then`; a subrace
missing `parentRace`, or a standalone missing `phase`, fails validation.)

### How an engine resolves a subrace (reference)

References across files are **not** schema-validated — resolution is the consuming
engine's job. The reference rule for a `subrace` whose `parentRace` is `P`:

1. Start from `P`'s fields.
2. `phase`, `speed`, `size`: the subrace's value **replaces** the parent's if present,
   otherwise the parent's carries over.
3. `abilityModifiers`: **sum** per attribute (parent `{end:2}` + subrace `{wil:1}`
   → `{end:2, wil:1}`).
4. `traits`: **union** (parent traits + subrace traits).

A `kin` performs **no** such merge — it is read exactly as written. A `parent`/
`template` is never instantiated as a playable character on its own.

---

## 4. Element (phase) & affinity

Each playable race has exactly one `phase` — `wood`, `fire`, `earth`, `metal`, or
`water`. This is the **same five-phase vocabulary** as effects and monsters, so a
character's race plugs directly into the interaction cycles in
[`data/wuxing/core.json`](../../ruleset/data/wuxing/core.json): incoming phased
effects resolve against the character's element via Generating (amplify), Overcoming
(suppress), Weakening (drain), and Insulting (backlash). See
[`magic.md`](./magic.md) § Elemental interaction for the full matrix.

Practical read: a `water` black elf shrugs off fire (Water overcomes Fire) but is
vulnerable where Water is itself overcome (by Earth). Phase is an **affinity**, not a
resistance list — the cycles compute the relationship.

Current assignments: human `earth`, elf `wood` (green elf `wood`, black elf
`water`), dwarf `metal` (mountain dwarf `metal`, hill dwarf `earth`), and beastfolk
kin choose their own (catman `metal`, wolfman `wood`).

---

## 5. Traits

`traits` is an array of `{ id, name, description }`. They are descriptive hooks the
engine interprets (e.g. *Darkvision*, *Pack Tactics*, *Retractable Claws*). The
ruleset deliberately stores their **text**, not a mechanical encoding — anything that
needs hard rules (a natural-weapon attack, a resistance) is expected to be expressed
through the shared systems (effects, skills) by the consuming game. Trait `id`s are
namespaced by race (`elf_darkvision`, `catman_claws`) to stay unique across the
union produced by inheritance.

---

## 6. Authoring guide

**Add a subrace to elf or dwarf**

```json
{
  "id": "high_elf",
  "name": "High Elf",
  "description": "Cloister-trained elves steeped in the arcane.",
  "lineage": { "role": "subrace", "parentRace": "elf" },
  "phase": "wood",
  "abilityModifiers": { "int": 1 },
  "traits": [ { "id": "high_elf_cantrip", "name": "Innate Cantrip", "description": "You know one cantrip of any color." } ]
}
```

Omit `speed`/`size` to inherit the elf's; set `phase` to keep `wood` or override it.
`abilityModifiers` and `traits` here **stack** on the elf base.

**Author your own beastfolk (a new kin)**

```json
{
  "id": "ratman",
  "name": "Ratman",
  "description": "Quick, cunning rodent-folk of the under-city.",
  "lineage": { "role": "kin", "parentRace": "beastman" },
  "phase": "water",
  "speed": 30,
  "size": "small",
  "abilityModifiers": { "agi": 2, "lck": 1, "str": -1 },
  "traits": [ { "id": "ratman_squeeze", "name": "Squeeze", "description": "You can move through gaps a small creature normally could not." } ]
}
```

A kin is **self-contained**: declare `phase`, `speed`, `size`, modifiers, and traits
in full. It owes nothing to catman or wolfman beyond also being beastfolk.

**Add a whole new lineage**

- A new freestanding race → `role: standalone`.
- A new family with shared defaults → add a `role: parent` ancestor, then
  `role: subrace` children pointing at it via `parentRace`.
- A new open-ended family where members differ wildly → add a `role: template`, then
  `role: kin` members pointing at it.

---

## 7. Data & schema map

| Concern | Schema | Data |
| --- | --- | --- |
| Race entries & lineage | `race.schema.json` | `data/races/core.json` |
| Attribute ids (`abilityModifiers` keys) | `schema.json#/definitions/ability` | `data/abilities/core.json` |
| Element vocabulary (`phase`) | `schema.json#/definitions/phase` | — |
| Interaction cycles | `wuxing.schema.json` | `data/wuxing/core.json` |

---

## 8. Open items / follow-ups

- **Engine inheritance.** The subrace-merge rules in §3 are a reference contract; the
  consuming engine implements them. The ruleset stores relationships, not resolved
  characters.
- **Trait mechanics.** Traits are text today. A future pass could let traits carry
  `effects[]` (like perks in `features/core.json`) so racial abilities compose
  through the shared effect pool instead of prose.
- **Playable parents.** Parents and templates are non-playable scaffolding; the
  "generic" version of each family is its baseline subrace (`common_elf`,
  `common_dwarf`) rather than the parent itself.
- **More elements.** No current race is `fire` — an obvious slot for a future
  ancestry (e.g. a fire-touched beastfolk kin or salamander-kin).
