# GD_S19_06_combat-start-ap A fight starts on a partial pool

Date: 2026-11-23

- **Decision:** a fight's first turn has `pool - 1` AP, at least 2 and at most the pool, with nothing carried; an ambusher's surprise round has the full pool, also with nothing carried ([actions.md#pool](../core/actions.md#pool)). [A]
  - **Data:** `pool.combat_start` in `actions/core.json`: `less` 1, `minimum` 2.
  - **Command:** the `actions` command's `start`: `combat` or `surprise`. Carried AP with either is an error. Pools 3, 4, and 5 start on 2, 3, and 4; pool 2 stays at 2. [V 2026-11-23, actions vectors in Python and Go]
- **The end of a fight refills nothing:** AP is per turn, and the next fight starts on the partial pool again; a short rest refills the pool as before. [A]
- **The encounter tracker** (`BE_S08_07`) names the rule for each actor's first turn, so its `encounter_start` step has it to read.
- **Not done here:** a perk that changes the start (it would need a modifier target), and who acts in a surprise round (the table decides).
