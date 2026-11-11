# APP-08 Dev mode: inflict state on a session for testing

Date: 2026-11-11

- **Decision:** dev mode is one command, `dev`, on a session: `operation`, `session`, and the operation's fields; the result is the session with the actor changed, the operation in the log as a `dev` step, and `dev: true`. Spec: `docs/backend/dev-mode.md`. **[V 2026-11-11, `go test ./core`]**
- **Decision:** the gate is a flag on the opened core (`Core.EnableDev`, set by `--dev`), never input. Without it `dev`, a session step of `dev`, and a replay that holds one are refused with `{"error": "dev mode is off; run with --dev"}`. **[V 2026-11-11, `go test ./core ./cmd/riftweave`]**
- **Decision:** nine operations in this part: `resource`, `xp`, `level` (XP set from the `progression` command), `attribute`, `skill`, `condition`, `effect`, `item`, and `turns`. Each checks its refs against the data. An inflicted effect is one modifier with source `dev:<effect key>` and the origin given, so it stacks by that origin's policy: a second arcane bonus on STR keeps only the stronger. **[V 2026-11-11]**
- **Decision:** a dev step is a session step, so a replay repeats it and a dev session replays to the same actor and keeps `dev: true`.
- **Difference from the issue:** this is the first of two pull requests, so the issue stays `doing`. Left for the second: perks, learned spells and effects, and gold, because the stored actor (`docs/core/state.md`) has no field for them; advancing by rests; fixed rolls and the seed; and the human CLI's dev panel and its short commands. **[A]**
- **Why:** a tester (and the alpha playtest, M5-18) needs to reach a state without playing to it.
- **Risk:** the Go loader reads the effect registry for dev refs (`EffectRefs`); M3-06 adds an `Effects` table for spellcraft, and whichever merges second should use one of them. **[U]**
