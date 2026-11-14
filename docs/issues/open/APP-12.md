#### APP-12 Sessions record skill purchases, perks, and learned spells
- Status: review
- Priority: high
- Needs: APP-04
- Do: Found by the M5-18 agent runs: a session cannot record a level-up's choices. `progression spend` returns ranks but writes nothing back, and no command takes a perk or learns a spell, so all eight runs bought every skill rank and took every perk through dev mode (419 `dev skill`, 138 `dev perk`, and 81 `dev learn` steps), which marks a played run as a dev run. Add session adapters, or commands, that write a bought rank to `skills`, a perk to `perks` (checking prerequisites and `ranks`, as the human CLI would), and a learned spell to `spells`, so a run to level 30 needs no dev step.
- Files: ~6 (backend/core/session.go, backend/contract/progression.json, docs/backend/sessions.md, docs/core/progression.md, ruleset/scripts/playtest_driver.py, backend/core/session_test.go)
- See: docs/decisions/M5-18.md, docs/playtests/README.md, docs/backend/sessions.md
- Done when: the playtest driver plays a build to level 30 with no `dev` step, and the session replays to the same actor; a perk whose prerequisites are not met is refused.
