# GD_S06_03_commander-summoner-perks Commander summoner perks: the caster spends turns to drive their summons

Date: 2026-11-19

- **Decision:** a new kernel function, `command_summons` (docs/core/commanding.md), works out what each summon gets from its commander, as `share` does for allies (M3-14): an **order** (the new `command` action, 2 AP) gives the summons it reaches AP for their next turn, and a **held command** (Steady the Host, a concentration spell with the new `command_summons` effect) gives every summon +5 to attack while held. A perk's `command` object (`ap`, `reach`, `turns`) defines its orders. [A]
- **AP to another actor** is a modifier on the summon's `ap` with `duration: turns`, `turns: 1`, so the pool's cap and the real-time rate (`stamina_per_ap`) apply unchanged (M3-11). [A]
- **Morale origin:** orders and the held command use `core:origin/morale`, which stacks highest, so two orders on one summon give the larger. [A]
- **Concentration:** the one-held-spell rule is kept, so holding a command ends a held Summon Elemental; the issue left it to the spec. A commander who holds a command leads familiars or the undead. [A, for the owner to confirm]
- **The ladder:** Field Command (6: one summon, +2 AP), Sharp Orders (10: the order costs 1 AP less), Voice of the Host (14: every summon), Iron Drill (18: +3 AP). Numbers are [A] until a playtest.
- **`reach` is `"one"` or `"all"`,** not a number or `"all"`: the first-normal-form gate (one kind of value per field) refused the mixed field. [V 2026-11-19, `validate.py`]
- **Authored only:** `command_summons` has `craftable: false`, like `bound_weapon`, since the held command belongs to the ladder, not to a recipe; `test_spells.py` and the uncraftable table in docs/modules/magic.md list it. [A]
- **Verified:** the 10 `command_summons` vectors pass in Python (420 tests) and Go (`go test ./...`); `validate.py` passes 52 of 52 files. [V 2026-11-19]
- **Differences from the issue:** a new command (`backend/contract/command_summons.json`, `backend/internal/commandsummons/`, `backend/core/core.go`, `kernel_reference.py`) and its page (`docs/core/commanding.md`), the effect row, `feature.schema.json`, `docs/core/vectors.md`, `test_spells.py`, and the regenerated diagrams are files the issue did not list. `frontend/src/api/openapi.d.ts` is not regenerated (as in GD_S06_01).
- **Branch:** made from `development`, on its own: it needs neither GD_S06_02 nor GD_S06_04, and merges in any order with them.
