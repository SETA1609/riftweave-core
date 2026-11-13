#### APP-13 A session step applies the actor's perks, held modifiers, and gear
- Status: todo
- Priority: high
- Needs: APP-04, APP-12
- Do: Found by the M5-18 agent runs: a session fills a command's input from the actor's attributes and pools only. It never adds the modifiers of the actor's perks, its held modifiers, or its race and background, and `combat` does not read the attacker's skill or the armor DR of equipped items. So a perk changed no command result in any run, Toughness included: `derived` through a session ignores `+15 hp`. Fill `modifiers` from the stored actor (perks at their rank, held modifiers by origin), the attacker's skill, and the worn armor's DR, by the one emitter the kernel already has (docs/core/modifier.md#origin).
- Files: ~6 (backend/core/session.go, backend/core/session_test.go, docs/backend/sessions.md, docs/core/state.md, ruleset/tests/vectors/playthrough.json, ruleset/scripts/test_playthrough.py)
- See: docs/decisions/M5-18.md, docs/core/actor-sheet.md
- Done when: a session's `derived` step for an actor with Toughness rank 2 reads 30 more HP; a `combat` step fills the attacker's skill and the defender's armor DR from the actor; a playthrough vector pins both.
