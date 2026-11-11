# Riftweave: one-page design

> **One roll. One rule set. Any table.**

Riftweave is a classless, d100 roll-under cRPG ruleset, written as data. The same rows run a
computer game, a tabletop session, and an agent, and the golden vectors prove they agree.

## Pillars

1. **Data, not code.** Every rule is JSON rows, a schema, and a formula tree. A new mechanic is
   new data, not a special case. ([data model](core/data-model.md))
2. **One roll.** Skill plus modifiers is the target; roll d100 at or under it. The margin says how
   well, crit and fumble are tags, and there is no second roll. ([resolution](core/resolution.md))
3. **Grow by choice, not class.** Eight attributes, point-buy skills with three tag skills, and a
   perk at every level gained, up to 30. ([progression](modules/progression.md))
4. **One sentence between systems.** Systems speak only in modifiers; none calls another. A
   module adds rows and never breaks core. ([contract](core/contract.md))
5. **One rule set for every mode.** A turn is 5 seconds and AP pays for actions, so a table and a
   real-time client run the same numbers. Rolls are inputs, so any result can be replayed.
   ([actions](core/actions.md))

## The loop

Make a character → explore → fight or cast → loot → rest → level up → repeat.
The [reference playthrough](core/loop.md) is that loop as a chain of vectors.

## Systems

| System | In short |
| --- | --- |
| Characters | Race, background, eight attributes from 1 to 10 at creation, skills from 0 to 100, perks |
| Combat | An AP budget, one reaction (evade or block), delivery decides the defense, then armor and resistance |
| Magic | Nine color schools, spells built from one shared pool of effects, mana and a skill check |
| Bestiary | Monsters from creature types and templates, leveled, with loot on the row |
| World | Economy, loot tables, hazards, rest |
| Modules | Optional layers such as Wuxing (the five phases) and factions; crafting and companions come later |

## What it is not

No classes. No engine or tabletop app in this repo. No separate rules for turn-based and
real-time play. No feature that exists only as prose.

## Who it is for

Game developers who want rules they can load, and game masters who want rules they can run.
Start at the [player and GM guide](guide/README.md); scope and the definition of finished are
in the [roadmap](../ROADMAP.md).
