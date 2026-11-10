# APP-05 Human CLI

Date: 2026-11-11

- **Decision:** `riftweave play` is in `backend/cmd/riftweave/play` and calls `core.Run`, so it holds no rule. `backend/core` and `backend/internal/` stay free of Charm, and a test fails if they import it. Spec: `docs/backend/cli.md#human-mode`. **[V 2026-11-11, `go test ./cmd/riftweave`]**
- **Decision:** Bubble Tea 1.3.10, Bubbles 0.21.0, and Lip Gloss 1.1.0, so the module now has a `go.sum`. Only the human CLI uses them.
- **Why:** a person needs a readable way to play the rules without writing JSON.
- **Difference from the issue:** one binary links Charm for every path, so the agent CLI cannot be free of it at link time. `play` is its only caller, and `riftweave list` went from about 1.8 ms to 2.3 ms a call. **[V 2026-11-11, 200 runs]**
- **Open:** the tests drive the model with key messages, and the program was only started in a shell to see its menu; no person has played a full session in a terminal. **[U]**
