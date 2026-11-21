# GD_S19_02_movement-distance-per-ap How far one move goes: squares per AP from the race's speed

Date: 2026-11-21

- **Decision:** one `move` covers `floor(speed / speed_divisor)` squares, with `speed_divisor` 2 on the `move` row of `ruleset/data/actions/core.json`. A race's `speed` stops being a distance per round. The `actions` command takes `speed` and returns `squares`; modifiers on `move` (encumbrance, armor, terrain) change its AP, not its squares. The rule lives in [actions.md](../core/actions.md#how-far-a-move-goes). [V 2026-11-21, `actions` vectors in Python and Go]
- **Why:** `move` cost 1 AP and covered no set distance, while `speed` was per round, so the battle grid (`BE_S10_05_battle-grid-steps`) had no unit to draw. Divinity: Original Sin 1 and 2 buy a distance per AP that grows with speed.
- **Key:** the issue offered `move.squares_per_speed` or a `distance` formula row; `speed_divisor` on the action row was chosen because it is one integer the existing row can hold, and any action may take one (a module's crawl or swim). [A]
- **Open:** the real-time pace, `floor(speed / 2)` squares per `seconds_per_turn / pool` seconds, is [A]: no real-time client checks it yet. The `speed` modifier target is still `pending`. [A]
- **Risk:** odd speeds lose half a square (7 covers 3 like 6), so a speed bonus of 1 only counts every other point. [A]
