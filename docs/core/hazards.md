# Hazards

A trap or an environmental hazard, as data. The rows are in `ruleset/data/hazards/core.json` (`hazard.schema.json`), the vectors in `ruleset/tests/vectors/hazard.json`. A hazard is a trigger plus a damage packet and effects, with up to three checks. Every check is one call of the [resolution function](resolution.md) through a [check row](checks.md); a hazard adds only a difficulty, an `add` on the target. Nothing here needs another module. [V 2026-11-06]

## Row

```
key, label, description, level
kind        trap | environment
nature      mechanical | natural | arcane
color       arcane only: the color school the hazard was woven in
trigger     pressure_plate | tripwire | touch | open | proximity | enter
detect      { check, difficulty }      notice it before it triggers; arcane: { difficulty }
disarm      { check, difficulty }      traps only; an environment hazard has none; arcane: { difficulty }
avoid       { check, difficulty, on_avoid: none | half }   optional
damage      { dice, type }             optional; dice is the row's, the roll is the caller's
effects     [ core:effect/<key> ]      optional; applied when the hazard is not avoided
```

## The three checks

```
target = the check's skill + difficulty + modifiers (stacked by modifier.md)
```

- **Detect.** A success means the hazard is noticed and does not trigger by surprise. A fumble is only a failure. [A]
- **Disarm.** A success makes the trap safe. A failure leaves it armed. A fumble triggers it on the one disarming. A hazard with no `disarm` cannot be disarmed: the result is an error. [A]
- **Avoid.** Rolled when the hazard triggers. A success with `on_avoid: none` avoids the hazard whole: no damage, no effects. With `on_avoid: half`, the damage is halved (rounded down) and the effects are avoided. A hazard with no `avoid`, or a trigger with no avoid roll in the input (the victim had no chance: asleep, surprised), always lands. [A]

## Nature

The `nature` says what the hazard is made of, and so which skill finds and undoes it. [V 2026-11-08]

| Nature | Detect | Disarm | Undone by a spell |
| --- | --- | --- | --- |
| `mechanical` | the row's check (`search`) | the row's check (`disarm_trap`, lockpick) | `disarm_hazard` of magnitude ≥ its level |
| `natural` | the row's check (`navigate`) | none: a natural hazard is a place, not a device | none |
| `arcane` | `<color>_magic` | `<color>_magic` | `dispel` of magnitude ≥ its level |

An arcane hazard names no check row for detect and disarm: it rolls its `color`'s magic skill, with the row's difficulty, so a fire rune woven in red is found and unwoven with `red_magic`, never with lockpick. A buff on the skill (a lantern on investigation, a perk on lockpick) works as before, through the modifiers on the check's skill. Avoiding a triggered hazard is the same for every nature.

## Spells on a hazard

`operation: spell` applies a cast that succeeded (the cast itself is [casting.md](casting.md)) to one hazard. The input names the `spell`; the result is `{ disarmed, effect }`, the effect that undid it, or `null`. [V 2026-11-08]

- A `disarm_hazard` effect of magnitude N makes a mechanical trap of level N or below safe.
- A `dispel` effect of magnitude N ends an arcane hazard of level N or below; an area dispel reaches every arcane hazard in its radius, the caller giving one call each.
- Nothing undoes a natural hazard, and a spell with neither effect does nothing: `disarmed` is false, `effect` null.

Unlocking is the same rule on a lock: an `unlock` effect of magnitude N opens a lock of level N or below. Locks are not rows yet, so the cast emits the effect and the consumer compares it to the lock's level. [A]

## Trigger

The hazard emits one packet, `{ amount, type }`, and its effect list, as the contract has it ([contract.md](contract.md)). The consumer applies DR, resist, and the effects' conditions; the hazard does not. A failed detect does not trigger the hazard by itself; the trigger is the caller's event (a step on the plate, a door opened). [A]

The `level` places a hazard on the same level scale as monsters and loot. A balance band for hazard damage by level is M4-05's. [U]
