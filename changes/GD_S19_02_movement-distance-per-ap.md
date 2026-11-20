### GD_S19_02_movement-distance-per-ap How far one move goes: squares per AP from the race's speed
- Added: `speed_divisor` on an action row; `move` has 2, so one move covers `floor(speed / 2)` squares (2 or 3 at today's speeds).
- Added: the `actions` command takes `speed` and returns `squares`; modifiers on `move` change its AP, not its squares.
- Changed: a race's `speed` is no longer a distance per round; [actions.md](docs/core/actions.md#how-far-a-move-goes) has the rule.
