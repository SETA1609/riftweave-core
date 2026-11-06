# Hazards

A trap or an environmental hazard, as data. The rows are in `ruleset/data/hazards/core.json` (`hazard.schema.json`), the vectors in `ruleset/tests/vectors/hazard.json`. A hazard is a trigger plus a damage packet and effects, with up to three checks. Every check is one call of the [resolution function](resolution.md) through a [check row](checks.md); a hazard adds only a difficulty, an `add` on the target. Nothing here needs another module. [V 2026-11-06]

## Row

```
key, label, description, level
kind        trap | environment
trigger     pressure_plate | tripwire | touch | open | proximity | enter
detect      { check, difficulty }      notice it before it triggers
disarm      { check, difficulty }      traps only; an environment hazard has none
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

## Trigger

The hazard emits one packet, `{ amount, type }`, and its effect list, as the contract has it ([contract.md](contract.md)). The consumer applies DR, resist, and the effects' conditions; the hazard does not. A failed detect does not trigger the hazard by itself; the trigger is the caller's event (a step on the plate, a door opened). [A]

The `level` places a hazard on the same level scale as monsters and loot. A balance band for hazard damage by level is M4-05's. [U]
