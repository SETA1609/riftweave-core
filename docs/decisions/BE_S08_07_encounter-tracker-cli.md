# BE_S08_07_encounter-tracker-cli An offline encounter tracker: initiative and whose turn it is (backend)

Date: 2026-11-24

- **Decision:** a kernel command `encounter` (operations `start` and `turn_end`) holds the initiative order, whose turn it is, the round, who has acted, and the ambushers ([sessions.md#encounters](../backend/sessions.md#encounters)). As a session step it is campaign state, a `sessionStateAdapters` entry beside `bestiary`: `turn_end` is filled from `session.encounter`, the result is written back, and a replay rebuilds it. Its MCP tool and its HTTP route come from the contract like every command's. [V 2026-11-24, `encounter.json` vectors in Python and through the Go package; `TestAnEncounterRunsAsStepsAndReplays` runs in CI]
  - **The issue's `encounter_start` and `turn_end` steps** are the command's two operations, so a session log reads `{ command: encounter, input: { operation: start } }`. [A]
  - **Inputs, not dice:** each combatant brings its derived `initiative`, its AGI and LCK after modifiers, and the d100 `rolls` of the roll-off, so the step holds the roll and replays it. A tie the rolls do not break is refused, so the caller rolls again. The rolls compare first to first, then second to second. [A]
  - **The surprise round** is round 0, in which only the ambushers act, each on `start: surprise`. An actor's first turn of round 1 or later that it has not acted in reads `start: combat`, which the caller passes to `actions`. [A]
  - **Damage, AP, and conditions** are the existing steps on each actor; the encounter holds ids only, so the party and the bestiary's monsters are the caller's to name. [A]
- **Differences from the issue:** `test_initiative_ties.py` named the open issue file as a page that must link the tie rule; it now names `docs/backend/sessions.md`, which does, so the test survives the issue moving to `done/`.
- **Not done here:** the page (`FE_S08_08_encounter-tracker-page`); a combatant that joins mid-fight (it would roll against only those it ties with, per derived.md) [U].
