# Kernel contract

Author: SETA1609

This is the only sentence a system is allowed to say to another system. Later modules (magic, alchemy, Wuxing, armor layering, breakthroughs) must speak it. They must not invent a second one.

A system may read an actor snapshot. It may emit modifiers, packets, or items. It may not call another system's procedure.

Mike Acton's rule is the reason: the purpose of a program, and of every part of it, is to transform data from one form to another. If magic calls combat's formula, a change in combat wakes magic. Hunt and Thomas call the opposite property orthogonality: a change in one thing does not force a change in the other.

Fallout stores primaries and lets the engine recompute derived stats. Morrowind spells, enchantments, and potions are lists of the same effect records. Elin puts attributes, skills, feats, and spells on one sheet, split by a group tag. This contract is that pattern. D&D 3.5's typed-bonus table is the failure mode: every chapter wrote into the same number, so the rules grew a stacking taxonomy. We write the stacking rule once, here.

## Actor snapshot

Read-only to everyone but the actor's owner. Current values are base plus modifiers. Derived stats (`hp`, `mp`, `sp`, `ap`, `carry`, `initiative`) are functions of this snapshot, not stored truth; [derived.md](derived.md) has the function. [V 2026-11-11]

```
Actor             stored_actor in backend/contract/common.json; see state.md
  attributes[8]     STR PER END INT WIL AGI CHA LCK
                    current = base + sum(mods)
  level, xp
  skills[key]       value; tags[] names the tag skills
  growth            hp, mp, sp from level-up rolls
  resources         hp, mp, sp, ap
  conditions[]      key, stacks
  modifiers[]       an active modifier and its turns_elapsed
  flags[]           prone, grappled, ...
  inventory[]       item instances (id, base, material, mods)
  equipped[slot]    an inventory id
```

## Modifier

The only shared sentence.

```
Modifier
  target            a key naming one row (modifier.md#target)
  op                add | mul | set | grant | resist
  value             number
  duration          turns | uses | while_equipped | while_applied | permanent
  source            feature:toughness | item:lockpicks | condition:blinded (a kind and a key, never a number)
  origin            core:origin/alchemy (optional; the kind of source)
```

Stacking, written once:

- Same `source` does not stack. The strongest application wins.
- Different sources add, unless they share an `origin` whose policy is `highest`: then the strongest bonus and the strongest penalty of that origin count.
- `mul` applies after `add`.
- `set` replaces the current value for that source's duration.
- `grant` adds a flag or a tag. It does not add a number.
- `resist` subtracts from a packet whose tag matches `target`.

The target, op, and duration values are the `modifier` definition in `schema.json`, and no modifier in the data has a numeric source. [V 2026-11-11] The order of these steps for one target, expiry, and how a condition is applied and removed are in [modifier.md](modifier.md).

The origins (arcane, divine, alchemy, trained, ...) are one short table in data, each with one policy, `highest`, `add`, or `replace`; [modifier.md](modifier.md#origin) lists them. There is no per-chapter exception. No dodge-stacks-except-circumstance. A perk that wants a second copy uses a different source id.

## Event

The only shared verb. A producer emits one. A consumer reads the result. Neither imports the other's formula.

```
roll_request      { skill, mods[] }
roll_result       { roll, target, margin, tags[] }
packet            { amount, tags[], source }
spawn_item        { base, material, mods[] }
```

Resolution is one function, named here so nobody writes a second one: `d100 <= target`, `margin = target - roll`. Crit and fumble are tags on `roll_result`, not a second roll system. The windows live in `ruleset/data/resolution/core.json`. [V 2026-11-11]

```
cast(actor, spell):
  require resource(actor, spell.cost)
  result = resolution.roll(actor, spell.skill)
  if result.success:
    effects.apply(spell.effects)
  # magic does not know DR, alchemy, or factions
```

## Distance

Every distance in the data and the docs is in **units**. One unit is one square of a grid.
Walking speed (`speed`, units per round), weapon range (`rangeNormal`, `rangeLong`), melee
reach (`attackReach`), area radius (`radius`), and distances in descriptions (darkvision,
falls, rope) all count squares. No definition is meant to use feet or metres; the exceptions left are listed in
[data-model.md](data-model.md#known-exceptions). [V 2026-11-11]

For tabletop play a square is 1.5 m by 1.5 m: a human walks 6 units (9 m) a round, and a
longbow reaches 30 units (45 m) at normal range. That conversion is for reading aloud only;
nothing stores it. [A]

## Weight

Every weight in the data and the docs is in **weight units**: an item's `weight`, a
material's `weightFactor` applied to it, the `carry` limit (`derived/core.json`), and the
encumbrance thresholds ([armor.md](../modules/armor.md#9-encumbrance)). No definition is
meant to use pounds or kilograms. [V 2026-11-11]

For tabletop play one weight unit is 0.5 kg (500 g): a longsword of 3 weighs 1.5 kg, and a
character with STR 7 carries 95 units (47.5 kg) before encumbrance. That conversion is for
reading aloud only; nothing stores it. Today's numbers were written as pounds (0.454 kg), so
reading them as weight units moves them by about 10% and changes no row. [A]

## What a later module is forbidden to do

- Call another system's procedure. Alchemy does not call advancement. Magic does not call defense.
- Read a phase in a core rule. `phase` on a core row is an optional tag, and the Wuxing module reads it ([wuxing.md](../modules/wuxing.md)).
- Require another module to be loaded for a core action to resolve. Level-up must work with no ingredients directory. An attack must resolve with no phase table.
- Store a derived stat as authority. HP is a function of END plus modifiers.
- Define a second resolution function, a second stacking rule, or a TTRPG-versus-video column inside the kernel.
- Override a core row without listing its key in `overrides` ([overrides.md](overrides.md)). A duplicate key with no such entry is a conflict, not a patch.
- Invent a private effect language. A condition is a named bundle of modifiers plus flags. A spell is a skill roll plus an effect list. A brew is a crafting profile whose output emits the same modifiers.

Unknown opcodes are ignored by a consumer. They are not rejected by this contract. Extending the opcode table is a later change; until then, new modules still speak only the ops listed above.

## Out of scope for this document

This file does not change JSON, schemas, combat resolution, armor, magic, alchemy, or advancement. Those are separate pull requests.
