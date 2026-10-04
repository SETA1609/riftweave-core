# Kernel contract

Author: SETA1609

This is the only sentence a system is allowed to say to another system. Later modules (magic, alchemy, Wuxing, armor layering, breakthroughs) must speak it. They must not invent a second one.

A system may read an actor snapshot. It may emit modifiers, packets, or items. It may not call another system's procedure.

Mike Acton's rule is the reason: the purpose of a program, and of every part of it, is to transform data from one form to another. If magic calls combat's formula, a change in combat wakes magic. Hunt and Thomas call the opposite property orthogonality: a change in one thing does not force a change in the other.

Fallout stores primaries and lets the engine recompute derived stats. Morrowind spells, enchantments, and potions are lists of the same effect records. Elin puts attributes, skills, feats, and spells on one sheet, split by a group tag. This contract is that pattern. D&D 3.5's typed-bonus table is the failure mode: every chapter wrote into the same number, so the rules grew a stacking taxonomy. We write the stacking rule once, here.

## Actor snapshot

Read-only to everyone but the actor's owner. Current values are base plus modifiers. Derived stats (HP, MP, AP, carry, initiative, speed) are functions of this snapshot, not stored truth. That function lives in a later change; this document only names the fields.

```
Actor
  id
  attrs[8]          STR PER END INT WIL AGI CHA LCK
                    current = base + sum(mods)
  skills[id]        points, tag
  resources         hp, mp, stamina, ap
  flags[]           prone, grappled, ...
  inventory[]       item instances (base + material + mods)
```

## Modifier

The only shared sentence.

```
Modifier
  target            attr | skill | resource | flag | damage_tag
  op                add | mul | set | grant | resist
  value             number
  duration          turns | uses | while_equipped | permanent
  source            perk:12 | item:44 | effect:19
```

Stacking, written once:

- Same `source` does not stack. The latest application wins.
- Different sources add.
- `mul` applies after `add`.
- `set` replaces the current value for that source's duration.
- `grant` adds a flag or a tag. It does not add a number.
- `resist` subtracts from a packet whose tag matches `target`.

There is no bonus taxonomy. No dodge-stacks-except-circumstance. A perk that wants a second copy uses a different source id.

## Event

The only shared verb. A producer emits one. A consumer reads the result. Neither imports the other's formula.

```
roll_request      { skill, mods[] }
roll_result       { roll, target, margin, tags[] }
packet            { amount, tags[], source }
spawn_item        { base, material, mods[] }
```

Resolution is one function, named here so nobody writes a second one: `d100 <= target`, `margin = target - roll`. Crit and fumble are tags on `roll_result`, not a second roll system. The windows live in data, in a later change.

```
cast(actor, spell):
  require resource(actor, spell.cost)
  result = resolution.roll(actor, spell.skill)
  if result.margin >= 0:
    effects.apply(spell.effects, result.margin)
  # magic does not know DR, alchemy, or factions
```

## What a later module is forbidden to do

- Call another system's procedure. Alchemy does not call advancement. Magic does not call defense.
- Require another module to be loaded for a core action to resolve. Level-up must work with no ingredients directory. An attack must resolve with no phase table.
- Store a derived stat as authority. HP is a function of END plus modifiers.
- Define a second resolution function, a second stacking rule, or a TTRPG-versus-video column inside the kernel.
- Override a core row. Modules add rows. Duplicate ids are conflicts, not patches.
- Invent a private effect language. A condition is a named bundle of modifiers plus flags. A spell is a skill roll plus an effect list. A brew is a crafting profile whose output emits the same modifiers.

Unknown opcodes are ignored by a consumer. They are not rejected by this contract. Extending the opcode table is a later change; until then, new modules still speak only the ops listed above.

## Out of scope for this document

This file does not change JSON, schemas, combat resolution, armor, magic, alchemy, or advancement. Those are separate pull requests.
