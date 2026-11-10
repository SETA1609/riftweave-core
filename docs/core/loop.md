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

## Reference playthrough

The chain in `ruleset/tests/vectors/playthrough.json` extends this loop into a whole session on core rows only. Every step is a vector, and each step's input is the state the step before left. `test_playthrough.py` checks those links. [V 2026-11-11, Python and Go]

| Step | Function | What happens |
| --- | --- | --- |
| Create | `derived` | A level-1 character gets 67 HP and 42 MP. |
| Buy | `economy` | A shortsword, at the general store's price. |
| Fight | `combat` | A hit on a zombie, its slam back, then a crit that drops it. |
| Loot | `loot` | The zombie's grave goods give 2 sp. |
| Rest | `rest` | A short rest restores the lost HP. |
| Level | `progression` | 900 XP reaches level 2, then the level-up pool. |
| Cast | `casting` | Fire Bolt spends mana. |
| Condition | `conditions` | A blinding effect sets `blinded`. |
| Recover | `rest` | A long rest clears it and refills the pools. |
