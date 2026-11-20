# GD_S19_03_swap-draw-costs Drawing, sheathing, and swapping a loadout cost AP

Date: 2026-11-21

- **Decision:** `actions/core.json` gets `draw` 1, `sheathe` 1, `drop` 0, and `swap` 1, each with `changes_loadout`. A `stance` loadout event names the action that paid in `by`, and the step costs that action's AP after the held perks' modifiers; without `by` the change is free. The rule lives in [actions.md](../core/actions.md#changing-the-loadout). [V 2026-11-21, `actions` and `stance` vectors in Python and Go]
- **Why:** the loadout decides what an attack costs and strikes (GD_S05_01), and a change ends a stance that no longer fits (GD_S05_02), but no action priced it.
- **Perk name:** the issue suggested Quick Hands, which is a Fallout 4 perk, so the perk is **Drilled Hands** (AGI 5, ungated, `add swap -1`), an original name.
- **Open:** the secondary loadout on a stored actor (the issue's last key) is not added: the `stance` command takes the loadout it changes to, and storing a second set of equipped slots belongs with the session's equip step and the action menu (`GD_S08_14_action-menu-rules`). [A] The numbers, 1 AP against Fallout Tactics' 2, are [A] until a playtest at pool 3 and 4.
- **Risk:** drop at 0 AP and draw at 1 makes drop-then-draw (1 AP) cheaper than a swap without the perk when one hand changes; that is intended, since the dropped weapon stays on the ground. [A]
