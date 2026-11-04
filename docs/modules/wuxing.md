# Wuxing

Author: SETA1609

Optional tag reader. The matrix in `ruleset/data/wuxing/core.json` is unchanged.

A fire bolt is an effect list. A phase is a tag. This module may read it. Core must resolve an attack and a level-up with this module unloaded. Unknown phase tags are ignored.

## Interaction function

One function reads the matrix: given the phase of the acting effect, the phase of the effect it meets, and optionally both magnitudes, it returns which cycle fires, the interaction, the multiplier, and which side the multiplier modifies. It lives once, in `ruleset/scripts/wuxing_reference.py`; the combat and crafting references call it. Vectors: `ruleset/tests/vectors/wuxing.json`.

```
interaction(actor, recipient, actor_magnitude?, recipient_magnitude?):
  no phase on either side       -> none, x1.0
  actor == recipient            -> same_phase, identity, x1.0
  directed edge actor -> recipient in a cycle:
        that cycle's interaction, multiplier, and affects
        (an insulting edge is gated, below)
  otherwise                     -> unrelated, none, x1.0
```

- **Direction matters.** The matrix holds all 20 ordered pairs of distinct phases, each in exactly one cycle. Fire acting on Wood is weakening (x0.75), not the reverse of Wood acting on Fire (generating, x1.5). There is no `1 / multiplier` fallback: the earlier "reverse cycle" lookup could never run on this data, and on an incomplete override matrix it would have invented a number. An unlisted pair is `unrelated`. [V 2026-11-02, all 20 pairs checked against `core.json` in this change]
- **Insulting is gated.** Backlash (x1.25) fires only when the actor's magnitude is greater than the recipient's. Otherwise the controlling cycle applies in its own direction: the recipient controls the actor, so the result is overcoming (suppress, x0.5) and it modifies the actor. If a caller passes no magnitudes the gate is not checked and backlash fires, as the simulators did before. [A]
- A cycle with a `gate` names the cycle to fall back to (`otherwise_cycle`); the reference reads it from the data.
