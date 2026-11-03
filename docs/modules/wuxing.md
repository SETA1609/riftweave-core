# Wuxing

Author: SETA1609

Optional tag reader. The matrix lives in `ruleset/data/wuxing/core.json`.

A fire bolt is an effect list. A phase is a tag. This module may read it. Core must resolve an attack and a level-up with this module unloaded. Unknown phase tags are ignored.

## The interaction function

`wuxing` takes two phased participants and names the one cycle that fires between them.

| Input | Meaning |
| --- | --- |
| `actor_phase` | the phase of the incoming effect (the edge's `from`) |
| `recipient_phase` | the phase it meets: a standing effect, a monster, a race (the edge's `to`) |
| `actor_magnitude`, `recipient_magnitude` | the two magnitudes; read only by a cycle with a `condition` |

| Result | Meaning |
| --- | --- |
| `cycle` | the cycle id that fires, or `null` |
| `interaction` | `amplify`, `suppress`, `drain`, `backlash`, or `none` |
| `affects` | `actor` or `recipient`: whose magnitude the multiplier scales; `null` when nothing fires |
| `multiplier` | the cycle's `magnitudeMultiplier`, or `1` when nothing fires |

The function does not apply the multiplier. Which number it scales (damage, duration, a dispel check) belongs to the caller. **[A]**

## Rules

1. **No phase, same phase.** If either phase is missing, or both are the same, nothing fires: `cycle: null`, `multiplier: 1`. **[A]** [magic.md](magic.md#6-elemental-interaction--the-five-phase-layer) says a same-element attack lands "at affinity (reduced)" but gives no number; until one is decided it is ×1.
2. **One edge per pair.** Each cycle lists its edges, or names another cycle in `reverse_of` and uses that cycle's edges read backwards. Weakening is `reverse_of: generating`; insulting is `reverse_of: overcoming`. The reverse is computed in one place, from the data; no cycle stores a reverse copy. **[V 2026-11-03]** the four cycles cover each of the 20 ordered pairs of two different phases exactly once (`wuxing.json` has one vector per pair).
3. **Conditional cycles.** A cycle with `condition: actor_magnitude_exceeds_recipient` fires only when `actor_magnitude > recipient_magnitude`. Otherwise its `reverse_of` cycle fires on the same pair with the roles read backwards: the recipient is that cycle's actor, so `affects` flips. For insulting: a Fire effect stronger than the Water it meets backlashes onto the Water (×1.25, recipient); a Fire effect equal or weaker is suppressed by the Water (overcoming ×0.5, actor). Equal magnitudes do not insult. **[A]**
4. **Missing magnitudes.** A conditional cycle with either magnitude missing returns `error`. A cycle without a condition ignores magnitudes.

## Differences from the earlier data

- The insulting trigger was an English sentence in `condition`. It is now an enum value, so a second implementation reads it instead of parsing prose (review DOD-13).
- Weakening and insulting listed their edges by hand. They now name `reverse_of`, so the canonical cycles exist once.
- The demo scripts `combat_reference.py` and `crafting_reference.py` each had a `1.0 / multiplier` fallback for an unlisted pair. It divided by zero at multiplier 0, gave ×0.8 for reverse insulting, and never fired, because every pair has an edge (review M8). Both now call `wuxing_reference.py`.

## Inspiration

The five phases and four cycles mirror the classical Chinese wuxing: generating (sheng), overcoming (ke), and their reverses, weakening (xie) and insulting (wu). The multipliers are this ruleset's own. **[U]** the cycle names and orders are from memory, not from a cited source.
