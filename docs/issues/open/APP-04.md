#### APP-04 Session state and the play chain
- Status: doing
- Priority: medium
- Needs: APP-02, M5-25, M5-38
- Do: Commands are stateless. Add a session as `docs/core/state.md` defines it, a starting actor and a log of commands, with the current actor snapshot (HP, MP, AP, XP, level, inventory, conditions) that commands read and update, so a chain of commands (create, buy, fight, loot, rest, level, cast) runs as calls. The chain grows with the rules; M5-04's playthrough runs through it once it lands. Sessions serialize to JSON so a run can be saved, replayed, or turned into vectors.
- Keys: session object and its schema in `backend/contract/`; `session` input on commands
- Files: ~6 (backend/, backend/contract/, docs/backend/sessions.md)
- See: docs/core/loop.md, docs/stories/05-release.md (M5-04)
- Done when: a chain of the commands that exist (at least derived, combat, rest, progression) runs as backend calls and matches their vectors; a saved session replays to the same result.
