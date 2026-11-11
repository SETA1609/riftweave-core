# APP-04 Session state and the play chain

Date: 2026-11-11

- **Decision:** one `session` command holds a session as `docs/core/state.md` defines it. Spec: `docs/backend/sessions.md`. **[V 2026-11-11, `go test ./core`: the 12-step playthrough matches its vectors and replays]**
- **Decision:** a step's named fields win, and the session fills the rest from the actor. The log keeps the input as given, so a replay fills it the same way. **[A]**
- **Decision:** a combat step names the actor's side with `as`, because a combat input holds two actors and the stored actor is one of them. **[A]**
- **Decision:** the field that picks the action is `operation`, spelled out as `CONTRIBUTING.md#naming` asks; the older `op` elsewhere is M5-46. **[V 2026-11-11]**
- **Why:** commands were stateless, so nothing carried a character from one to the next.
- **Difference from the issue:** this is Go only. It adds no rule, so it needs no Python reference, and its tests are in `backend/core`, not `ruleset/tests/vectors/`.
- **Open:** the stored actor has no coins and loot adds nothing to the inventory, so `economy` and `loot` steps are logged and change nothing. The ruleset version is not recorded, because the data has no version yet; only the modules are. Rolls are still inputs. **[U]**
