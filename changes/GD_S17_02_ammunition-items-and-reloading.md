### GD_S17_02_ammunition-items-and-reloading Ammunition items, the reload action, and recovering spent ammunition
- Added: the `ammoKind` vocabulary and `weapon.ammo` on every weapon base with the `ammunition` property; five ammunition bundles in `equipment/ammunition.json` (`core:ammunition/<key>`).
- Added: `combat`'s optional `ammunition` input: a shot spends a round, and a weapon with none left, or a `loading` weapon not loaded, cannot attack.
- Added: the `reload` action at 2 AP, and the `ammunition` command's `recover`: half the arrows and bolts back, rounded down, no bullets, scatter, or stones, every thrown weapon.
- Changed: Fast Reload emits `add reload -1` on the `reload` action. Removed: the pending `reload_cost` modifier target (breaking).
