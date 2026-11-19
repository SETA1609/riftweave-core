# PO_S08_13_file-action-menu-issues File the action menu issues

Date: 2026-11-19

- **Decision:** file an action menu in three issues: `GD_S08_14` (the rules: a kernel function `action_menu` in the Python reference, with vectors), `BE_S08_15` (the contract command, the Go core, and the CLI and TUI screen), and `FE_S08_16` (the panel in the app). Owner request, 2026-11-19.
- **Why a GD issue first:** no kernel function lists what an actor can do. Costs are spread over `actions/core.json`, the loadout rows of `attack_sequences/core.json`, perk modifiers on an action's key, maneuvers and stances that perks give, spells, and summon orders. `actions` only sums the cost of actions the caller already chose [V 2026-11-19, `backend/contract/actions.json`, `ruleset/data/actions/core.json`]. A rule issue with contract and Go work is two issues, GD and BE (`docs/issues/ROLES.md`), so the panel holds no rule.
- **In S08, not a new story:** S08 holds the encounter tracker and the player's own session page, where the menu shows on the player's turn.
- **Needs:**
  - `GD_S08_14` needs GD_S05_01 to 04, since the menu lists their loadout attack costs, stances, the shield bash, and maneuvers.
  - `BE_S08_15` needs the encounter tracker's backend (`BE_S08_07`), which holds the AP left and whose turn it is.
  - `FE_S08_16` needs `BE_S08_15` and the tracker page (`FE_S08_08`). It shows in the player session page once that page exists, without needing it.
- **S19:** the turn-economy story adds rows (draw, sheathe, swap, ready) and held AP; the menu lists them as they merge, with no `Needs` edge, so it is not held back by them.
- **Only unlocked actions are listed** [A]. An action the actor holds but cannot take now (too little AP, no mana, the wrong hands, not prone, no summon out) is listed with `available: false` and a reason, so the player sees why.
- **Risk:** Go runs every vector file, so `GD_S08_14`'s vectors fail Go's tests until `BE_S08_15` lands. As in GD_S05_01, the two may land as one pull request.
- **Open:**
  - Should a locked action show, greyed out, with the perk that unlocks it, to help plan a build [U]?
  - Does picking an entry in the menu record the step, or only open the screen that resolves it (the attack screen for attacks) [A]?
