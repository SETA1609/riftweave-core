# FE_S07_04_character-list-page A character list, and characters travel between players and the GM as files (frontend)

Date: 2026-11-21

- **Decision:** the Characters page (`/characters`) and the `Characters` service keep the browser's character list and import by the CLI's sync rule, with its outcomes and notes word for word; a refused import shows its note, and a diverged one shows both sides and keeps the one picked. No contract or core change: the data is checked by `stored_actor`, as the Import page does. [V 2026-11-21, `characters.spec.ts`, `e2e/characters.spec.ts` on the API and WebAssembly builds]
- **Storage:** a database of its own, `riftweave-characters`, rather than a second store in `riftweave`, so the saves' database needs no version upgrade and an old browser's saves are untouched. [A]
- **Between the apps:** the page's export of a saved character is its file byte for byte, so the fixture the Go tests import (`character-synced.riftweave.json`) is what the browser writes; a GM's higher revision replaces it, which the e2e spec runs. [V 2026-11-21]
- **Open:** opening a character to play it, and adding one to a session's party from the browser, are not on the page yet; the GM's party lives in the CLI (`BE_S07_03`) and the browser's GM tools come with the encounter tracker (`FE_S08_08_encounter-tracker-page`). [A] The page sorts by name; the CLI lists by its directory order. [A]
