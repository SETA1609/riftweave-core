# BE_S09_01_spell-author-wasm A spell editor page over the author command (backend)

Date: 2026-11-21

- **Decision:** `author` gets an `effects` operation that lists a color's effect pool, and the browser core embeds `ruleset/schemas` and opens with them; both are in [authoring.md](../backend/authoring.md).
- **The pool's rule** is the one magic.md already states for the core spell list: an effect on the spell's color or on none, and it must list the `spell` channel, as a check of a draft requires. [V 2026-11-21, `author_test.go`]
- **A row, not a projection:** each effect comes back as effects/core.json writes it, so the editor shows whatever fields it needs (label, magnitude unit, parameter category) without a contract change per field.
- **Valid colors** are the ones a `<color>_magic` skill governs, read from the data, not a list in Go.
- **One opener:** `openRuleset` in `open.go` is shared by the browser entry point and its test, so the test opens the same tree layout the build embeds.
- **Open:** a module's effects are not in the pool yet; the check of a draft accepts `module:` refs unchecked, as `validate.py` does. [A]
- **Risk:** the WebAssembly binary grows by the schemas (a few hundred kilobytes before gzip). [U]
- **Risk:** `contract_diagram.py`'s Modified line and `docs/backend/contract-diagram.md` also change in BE_S08_10 and BE_S09_03; whichever merges later regenerates the page.
