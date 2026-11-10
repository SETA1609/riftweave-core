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
| `speed` | Walking speed in abstract units (e.g. 6 for a human). |
| `size` | `tiny` · `small` · `medium` · `large`. |
| `traits` | Array of `{ id, label }` references to `racial_trait` rows in the perk table (`data/features/core.json`). See §5. |
| `lineage` | How this race relates to others — see §2. **The defining structural field.** |

Every race also carries the standard `id` (integer primary key), `key` (stable symbolic identifier), `label` (human-readable name), `description`, and optional `source`.

---

## 2. The lineage system

Every race declares a `lineage` object: a `role` and, when relevant, a `parentRace`.
There are **two distinct relationships**, and keeping them separate is the point.

```json
"lineage": { "role": "subrace", "parentRace": "core:race/elf" }
```

| Role | Playable? | `parentRace` | Relationship |
| --- | --- | --- | --- |
| `standalone` | yes | — | A complete, unrelated race (e.g. `human`). |
| `parent` | no (abstract) | — | An ancestor that subraces inherit from (e.g. the elf parent). |
| `subrace` | yes | required (string key ref) | **Inherits** from its parent; overrides freely. |
| `template` | no (abstract) | — | A **shape** for sister races (`beastman`); provides structure, not values. |
| `kin` | yes | required (string key ref) | A sister conforming to a template; shares nothing mechanical with siblings. |

### 2a. Parent → subrace = inheritance

A `parent` race (elf, dwarf) is abstract — you don't play "a generic elf," you play a
green elf or a mountain dwarf. The parent holds the **shared defaults**; each
`subrace` **descends from it** and may keep or override any of them:

- **Inherit** — omit a field and the parent's value applies. A subrace may drop
  `speed`/`size` to take the parent's (e.g. `green_elf` inherits the elf's 6 / medium).
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

All roles require `id` (integer primary key), `key` (symbolic), `label` (display name), `description`, `lineage`. Beyond that:

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
engine's job. The reference rule for a `subrace` whose `parentRace` is `P` (where `P`
is a string key, e.g. `"core:race/elf"` or `"elf"`):

1. Start from the parent race's fields (looked up by the string key `parentRace`).
2. `phase`, `speed`, `size`: the subrace's value **replaces** the parent's if present,
   otherwise the parent's carries over.
3. `abilityModifiers`: **sum** per attribute (parent `{end:2}` + subrace `{wil:1}`
   → `{end:2, wil:1}`).
4. `traits`: the array of `{ id, label }` references is the list of racial traits
   this race entry specifically contributes, added to the parent's. The definitions
   are `racial_trait` rows in `data/features/core.json` (§5).

A `kin` performs **no** such merge — it is read exactly as written. A `parent`/
`template` is never instantiated as a playable character on its own.

---

## 4. Element (phase) & affinity

Each playable race has exactly one `phase` — `wood`, `fire`, `earth`, `metal`, or
`water`. This is the **same five-phase vocabulary** as effects and monsters, so a
character's race plugs directly into the interaction cycles in
[`modules/wuxing/data/wuxing.json`](../../ruleset/modules/wuxing/data/wuxing.json): incoming phased
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

Racial traits are perks. They live in the single perk table,
[`data/features/core.json`](../../ruleset/data/features/core.json) (schema
`feature.schema.json`), as rows with `"type": "racial_trait"`. There is no separate
traits table: it was folded into `features` in M1-06 (see
[decisions](../decisions.md)).

A `racial_trait` row has the same shape as any perk: `id`, `key`, `label`,
`description`, `type`, and optionally `category`, `effects`, `source`. It is never chosen
or bought; a character has it because its race (or the race's parent) lists it.

In a race, `traits` is an array of lightweight references:

```json
"traits": [
  { "id": "core:feature/elf_darkvision", "label": "Darkvision" },
  { "id": "core:feature/elf_fey_ancestry", "label": "Fey Ancestry" }
]
```

The `id` is a namespaced ref `core:feature/<key>` (or a `module:` ref). The `label` is
duplicated for readability when looking at race data; the full description lives in
the feature row.

Rules, checked by `validate.py`:

- A race's `traits[].id` must be a `core:feature/<key>` ref (or a `module:` ref), and
  the `core:` feature must exist and have `type: racial_trait` **[V 2026-11-02]**.
- Every core `racial_trait` row is listed by at least one core race, so the type
  cannot hold orphan rows **[V 2026-11-02]**.
- Perks of type `perk`, `creation`, or `universal` are not listed in a race's
  `traits`; a race grants only racial traits **[V 2026-11-02]**.

Most racial traits are still prose today. Because they are features, they can carry
`effects[]` like any perk, so a trait can compose through the shared effect pool
without a schema change **[V 2026-11-02]** (`feature.schema.json` `effects`).

**Why `features` and not a traits table.** The project rule is one perk table
(`AGENTS.md`), and `feature.schema.json` already had the `racial_trait` type for this.
A separate table duplicated the row shape, had no `effects`, and would have added a
second ref category (`core:trait/`) for the same concept. Keeping one table means one
ref format, one prerequisite and effect shape, and one place an engine looks for
"what does this character have" **[A]**.

---

## 5a. What a race emits

A race and its traits reach the actor only as contract modifiers ([modifier.md](../core/modifier.md)); no rule lives only in a description. `ruleset/scripts/ancestry_reference.py` reads them, and `test_ancestry.py` pins the reading. [V 2026-11-06]

- **`abilityModifiers`** is the compact form of `add` modifiers: `{ "agi": 1 }` is `add agi 1`, source `race:<key>`, origin `innate`.
- **Traits** carry their own `modifiers` (source `feature:<key>`, origin `innate`): a number is an `add` on a skill or a named target (`darkvision_range`, `tag_skills`, `hp_per_level`), a sense or a proficiency is a `grant` flag (`darkvision`, `proficient_medium_armor`). A trait with weapon dice carries `natural_weapon` instead (catman claws, 1d6 slashing).
- **Lineage.** A subrace emits its parent's modifiers and traits, then its own; the two stack because their sources differ. A kin emits only its own: nothing comes from the template or from sibling kin. A parent or a standalone race emits its own. A template is never played.
- **Situational bonuses** (shadowmeld in darkness, pack tactics against an engaged foe, a scent-based detection) are `add` on a named situational target the caller reads when the situation holds. [A]

## 6. Authoring guide

**Add a subrace to elf or dwarf**

```json
{
  "id": 13,
  "key": "high_elf",
  "label": "High Elf",
  "description": "Cloister-trained elves steeped in the arcane.",
  "lineage": { "role": "subrace", "parentRace": "core:race/elf" },
  "phase": "wood",
  "abilityModifiers": { "int": 1 },
  "traits": [ { "id": "core:feature/high_elf_innate_cantrip", "label": "Innate Cantrip" } ]
}
```

Add the trait itself as a `racial_trait` row in `data/features/core.json` first.

Omit `speed`/`size` to inherit the elf's; set `phase` to keep `wood` or override it.
`abilityModifiers` and `traits` here **stack** on the elf base.

**Author your own beastfolk (a new kin)**

```json
{
  "id": 19,
  "key": "ratman",
  "label": "Ratman",
  "description": "Quick, cunning rodent-folk of the under-city.",
  "lineage": { "role": "kin", "parentRace": "core:race/beastman" },
  "phase": "water",
  "speed": 30,
  "size": "small",
  "abilityModifiers": { "agi": 2, "lck": 1, "str": -1 },
  "traits": [ { "id": "core:feature/ratman_squeeze", "label": "Squeeze" } ]
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
| Racial traits (`type: racial_trait`) | `feature.schema.json` | `data/features/core.json` |
| Attribute ids (`abilityModifiers` keys) | `schema.json#/definitions/ability` | `data/abilities/core.json` |
| Element vocabulary (`phase`) | `schema.json#/definitions/phase` | — |
| Interaction cycles | `wuxing.schema.json` | `modules/wuxing/data/wuxing.json` |

---

## 8. Open items / follow-ups

- **Engine inheritance.** The subrace-merge rules in §3 are a reference contract; the
  consuming engine implements them. The ruleset stores relationships, not resolved
  characters.
- **Trait mechanics.** Racial traits are `racial_trait` features (§5), so they
  accept `effects[]`, but most are still prose. Encoding them as effects is
  content work for a later issue.
- **Prereq readability.** Feature prerequisites for skills now use numeric keys
  (e.g. `"skills": { "1": 40 }`). This works but is less readable than the old
  symbolic form. Consider using the stable `key` string in prereq maps or
  changing the shape to an array of objects.
- **Referential integrity.** The validator only performs JSON Schema checks.
  There is still no automatic cross-file validation that numeric IDs (effects,
  traits, parentRace, materials, etc.) actually exist in their target
  collections. A separate integrity pass (or script) would be valuable now that
  IDs are integers.
- **Documentation maintenance.** Major drifting docs (especially race.md and
  backgrounds.md) have been refreshed for the new `id`/`key`/`label` model and
  extracted traits, but scattered examples and explanatory text across the
  modules will need ongoing care as the data evolves.
- **Playable parents.** Parents and templates are non-playable scaffolding; the
  "generic" version of each family is its baseline subrace (`common_elf`,
  `common_dwarf`) rather than the parent itself.
- **More elements.** No current race is `fire` — an obvious slot for a future
  ancestry (e.g. a fire-touched beastfolk kin or salamander-kin).
