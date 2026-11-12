### APP-08 Dev mode: inflict state on a session for testing
- Added: `riftweave --dev dev <operation>`, which changes a session's actor directly for testing (resources, XP and level, attributes, skills, conditions, effects with an origin, items, turns), logs each operation, and marks the session `dev: true`; without `--dev` it is refused.
- Added: `dev` on the session schema.
- Added: the `rest`, `perk`, `learn`, and `gold` dev operations; `perks`, `spells`, `known_effects`, and `coins` on the stored actor.
- Added: the human CLI's dev panel (`riftweave --dev play`) with `:give`, `:inflict`, `:level`, and `:roll`, which queues the next roll prompts.
