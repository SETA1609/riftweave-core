# Backgrounds

**Status:** Core (new collection — design + initial data implemented)

Backgrounds represent a character's life *before* the story begins. In a classless point-buy system they provide an elegant way to give new characters immediate, flavorful competence, concrete starting gear, and a handful of signature spells without forcing them to spend their precious level-1 skill points or tag choices on "basics."

They are chosen once during character creation and are **not** repeatable or leveled.

## Philosophy

Riftweave characters begin relatively competent but not yet specialized heroes. A pure point-buy from zero can feel thin at level 1. Backgrounds solve this by delivering:

- **Concept immediately**: A scout feels like a scout on day one.
- **Concrete gear**: Instead of "you have 100 gp, go shopping," many backgrounds hand the character the tools of their former trade.
- **Minor magical or trained edges**: A hedge witch starts with *Mend Wounds*; a former soldier starts with real armor and a shield.
- **No tax on normal progression**: The skill bonuses are flat one-time additions on top of the normal `5 + ability × 2` base. They do not consume the level-1 skill point pool (`5 + INT × 2 + random(0…LCK)`) and do not interact with the 2:1 / 1:1 tag ratios.

Backgrounds are **optional**. A player who wants a completely blank-slate amnesiac or "raised by wolves with no past" character can simply skip the step.

## Design

A background lives in `data/backgrounds/core.json` and is validated by `background.schema.json`.

### Fields

- `id` (integer), `key`, `label`, `description`, optional `source` — standard (new ID convention).
- `category` — loose string for grouping (wilderness, military, arcane, criminal, noble, religious, scholarly, merchant, etc.). Not mechanically enforced.
- `skill_bonuses` — array of `{ skill, bonus }`. The `bonus` (typically 6–12) is added once at creation after the normal attribute-seeded base is calculated. These are permanent starting ranks.
- `starting_equipment` — array of `{ item, quantity? }`. Item references use the namespaced format `core:<category>/<key>` (e.g. `core:weapons/shortsword`, `core:armors/leather_jerkin`, `core:consumables/potion_minor_healing`). Legacy numeric ids are still accepted during transition. The engine is responsible for actually adding the items to the character's inventory.
- `starting_spells` — array of spell ids the character knows from day one. Useful for low-magic or tradition-based casters. The character still needs the governing `<color>_magic` skill to cast them reliably.
- `granted_features` — array of feature/perk ids granted automatically (in addition to any chosen creation perks or racial traits). Use sparingly; most backgrounds express their benefit through skills + gear rather than full perks.
- `wealth_bonus` — extra gp added to whatever the campaign's normal starting wealth is (the reference baseline is 100 gp + kit).
- `suggested_tag_skills` — purely advisory list. Helps players make coherent tag choices that reinforce the background.

Equipment cross-references prefer namespaced string keys (`core:category/key`). Skills and spells still use numeric `id` for now. These refs are validated by `validate.py` but not fully enforced by JSON Schema alone.

### Interaction with the rest of creation

- Applied **after** race and attribute assignment.
- Applied **before** (or alongside) spending the level-1 skill point pool and choosing tag skills.
- The bonuses are **additive only**; they never replace or override the normal base formula.
- A background does **not** grant extra tag skills (use the human *Versatility* racial or the *Genius* creation perk for that).
- Creation perks (with their tradeoffs) are still chosen separately.

Recommended creation order (see [`character-creation.md`](./character-creation.md)):

1. Choose race (and record phase)
2. Assign attributes (start 4, +10 points, apply racial mods)
3. **Choose background** (new)
4. Choose creation perks (0–2, optional)
5. Pick 3 tag skills (4 for humans)
6. Compute bases + spend level-1 pool (respecting 2:1 / 1:1)
7. ...derive, gear, element, etc.

## Suggestions & Catalog

Here are concrete, ready-to-use backgrounds plus additional ideas for expansion. The initial data set implements the first eight.

### Implemented (in `data/backgrounds/core.json`)

| ID                | Category    | Signature Grants                              | Flavor                              |
|-------------------|-------------|-----------------------------------------------|-------------------------------------|
| `scout`           | wilderness  | Survival +12, Stealth +10, Bows +8 + shortbow + leather + rations | Wilderness guide / ranger archetype |
| `former_soldier`  | military    | Block +12, Blades +10 + chain shirt + shield + longsword + mace + 25gp | Professional fighter / veteran      |
| `hedge_witch`     | arcane      | Green Magic +10, Medicine +10, Lore +8 + *Mend Wounds* + staff + potions | Practical village healer / wise one |
| `street_urchin`   | criminal    | Stealth +12, Lockpick +10, Sleight +8, Deception +6 + 2×dagger + leather + 15gp | Alley rat / cutpurse                |
| `noble_scion`     | noble       | Persuasion +12, Insight +8, Lore +6 + 75gp    | Well-bred, connected, soft hands    |
| `arcane_apprentice`| arcane     | Red Magic +10, Lore +12, Investigation +6 + *Fire Bolt* + *Ward* + staff | Collegiate or tower apprentice      |
| `wilderness_hunter`| wilderness | Survival +10, Bows +12, Stealth +8 + longbow + leather + handaxe | Bounty hunter / trapper             |
| `itinerant_priest`| religious   | White Magic +10, Persuasion +8, Lore +8 + *Mend Wounds* + *Blessing* + staff + potions | Wandering holy person               |

### Additional Strong Suggestions (not yet in data)

**Wilderness / Frontier**
- Trapper / Furrier — Survival, Repair, Animal Handling; traps, skinning knives, cured pelts.
- Mountain Guide — Athletics, Survival, Medicine; climbing gear, sturdy boots, rope.
- Druidic Acolyte — Green Magic + one other survival skill; *Mend Wounds*, a few herbs/ingredients, staff.

**Military / Mercenary**
- City Watch / Guard — Block, Intimidation, Investigation; club or mace, uniform (light/medium armor), whistle or signal horn.
- Skirmisher / Light Infantry — Bows (or Crossbows/Throwing Weapons), Piercing, Stealth, Survival; shortbow or javelins, studded leather.
- Sapper / Engineer (early) — Engineering, Repair, Athletics; tools, some explosives precursors (future).

**Criminal / Underworld**
- Smuggler — Deception, Persuasion, Stealth (or Repair); hidden-compartment clothing, a mule or cart flavor, contacts.
- Assassin in Training — Stealth, Sleight of Hand, one weapon skill; poison kit (future), dark clothing, garrote or good dagger.
- Gambler / Con Artist — Deception, Insight, Persuasion; loaded dice, fine but flashy clothes, small wealth bonus.

**Scholarly / Professional**
- Scribe / Chronicler — Lore, Investigation, one language-adjacent (future); ink, parchment, a few books as trade goods.
- Physician's Apprentice — Medicine, Lore, Investigation; healer's kit, bandages, a few common drugs.
- Traveling Merchant — Persuasion, Deception (or Insight), one crafting skill; scales, trade goods, a wagon or pack animal flavor, decent wealth.

**Arcane / Esoteric**
- Necromancer's Assistant — Black Magic, Lore, Medicine (or Intimidation); *Raise Thrall* later, bones, a veil or mask.
- Illusionist's Protege — Yellow Magic, Deception, Insight; *minor* light/charm spells, colorful props, a mirror or prism.
- Battle Mage Cadet — Red Magic + one combat skill; *Fire Bolt* + *Ward*, a reinforced staff, light armor.

**Religious / Cultural**
- Temple Guardian — White Magic or one weapon skill, Block; holy symbol, medium armor, a relic or vow item.
- Shaman / Spirit Speaker — One magic school + Insight + Animal Handling; drums, fetishes, minor spirit offerings.
- Mendicant Monk — Unarmed or Athletics, one knowledge skill; simple robes, begging bowl, staff.

**Exotic / Rare Flavor**
- Shipwreck Survivor — Survival, Athletics, Repair; a few tools, a "lucky" trinket, soaked but functional clothes.
- Disgraced Knight — Blades or Blunt + Persuasion (or Intimidation); broken or pawned armor pieces, a surcoat with the sigil torn off, lingering honor.
- Fey-Touched — One magic school + Insight or Deception; minor fey mark (flavor trait), strange small item, sensitivity to iron (purely narrative for now).

## Implementation Notes for Engines & GMs

- **When to apply**: After attributes are final and before (or while) the player spends their level-1 skill points. The flat bonuses simply increase the displayed starting ratings.
- **Gear**: The engine must resolve the `starting_equipment` list into actual inventory entries. If an id does not exist yet, the engine can either ignore it or substitute a close equivalent.
- **Spells**: Add the listed spells to the character's known-spell list. No mana or slot cost at acquisition.
- **Stacking**: Background bonuses stack with racial ability modifiers, creation perks (*Gifted*, *Genius*, etc.), and any future "trained" or "practiced" effects. They are just numbers.
- **TTRPG flavor**: GMs should let the player narrate *why* they have this background and what contacts, enemies, or unfinished business it might bring. A noble scion and a street urchin in the same party create instant roleplaying hooks.
- **Multiple backgrounds**: Do not allow by default. If you want "hybrid" concepts, either write a specific hybrid background or let the player take the strongest single one and roleplay the rest.
- **Custom backgrounds**: Players and GMs can easily invent new ones using the same shape. The schema is intentionally permissive on the exact numbers so long as they stay reasonable (skill bonuses 5–15 is the sweet spot for level 1).

## Open Questions & Future Work

- Should any backgrounds grant a **fourth tag skill** (like human Versatility) or a free creation perk slot? Currently no — those remain special racial/creation-perk territory.
- Do we want a small number of **negative** or "troubled" backgrounds (e.g. "Exiled", "Cursed", "Addict") that grant interesting gear/skills but also a flaw or enemy?
- Background-specific **perks** or **destinies** that only become available later if you took the matching background (light "class-like" feeling without actual classes).
- A "Lifestyle" or "Contacts" subsystem that builds on background (reputation, debts, safe houses).
- Better modeling of "kits" and multi-part starting packages (thieves' tools, healer's kit, scholar's satchel) as first-class equipment types.
- Should backgrounds be allowed to influence **phase** or starting **Wuxing affinity**? (Probably not — race is the foundation for that.)

## Data & Schema Map

| Concern              | File                              | Schema                        |
|----------------------|-----------------------------------|-------------------------------|
| Background definitions | `data/backgrounds/core.json`     | `background.schema.json` (new) |
| Skill references     | `data/skills/core.json`           | — (numeric id)                |
| Spell references     | `data/spells/core.json`           | — (numeric id)                |
| Equipment references | `data/equipment/*.json`           | — (numeric id)                |
| Feature references   | `data/features/core.json`         | — (numeric id)                |
| Shared vocabulary    | `schemas/schema.json`             | (sourceRef, etc.)             |

See also:
- [`character-creation.md`](./character-creation.md) — the full creation sequence now includes choosing a background.
- [`progression.md`](./progression.md) — base skill calculation and level-1 pool.
- [`advancement.md`](./advancement.md) — long-term growth (backgrounds only affect the starting point).

---

*Backgrounds give every new character a small but meaningful "this is who you were" without complicating the elegant classless point-buy heart of Riftweave.*