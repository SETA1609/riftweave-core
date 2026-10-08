# Vertical loop

Author: SETA1609

One actor, existing rows, no magic, alchemy, wuxing, or layering. The record is `ruleset/data/examples/loop.json`.

1. Create a human scout. Background `core:background/scout`.
2. Equip `core:weapon/shortsword` stamped with material `iron`.
3. Attack. d100 <= blades. Margin is target minus roll.
4. DR comes from the worn piece's drBase. No layer required.
5. Spend 900 XP. Level becomes 2. No ingredient.
6. Take a perk that emits `add` on a skill. Same source does not stack.

Example. Blades 40. Roll 25. Hit, margin 15. Iron does not change the roll. A perk `source: perk:example` with `add blades 5` makes the next target 45.
