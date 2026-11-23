# FE_S09_02_spell-editor-page A spell editor page over the author command (frontend)

Date: 2026-11-23

- **Decision:** `rw-spells` at `/play/spell_creation` is the browser's spell editor over the `author` command of BE_S09_01: `load`, `effects`, and `check`, with the core export downloaded as `spells.json` ([spells.md](../../frontend/src/app/spells.md)). It adds no contract or core change.
- **The draft is the whole row:** the form edits key, label, description, color, level, delivery, and each effect; a loaded row's other fields (`template`, `cost`, `save`) stay in the draft, so a loaded spell checks and exports as written. [V 2026-11-23, `app.spec.ts`]
- **Checked as you type:** a debounced check (300 ms, `switchMap`) after every change, so an answer to an older draft never overwrites a newer one. [V 2026-11-23, the Playwright test flags `core:effect/no_such_effect` and clears it, over HTTP and WebAssembly]
- **The menu reads `issue_apps`:** Spell creation stays pending in the CLI (APP-22, `issue_apps: ["cli"]`) and is a link in the browser; the web menu now drops an issue pending only in the CLI, as the CLI already dropped one pending only in the web.
- **Locale keys:** the page's words are `spells.*` keys; the German and Spanish are mine, for the owner to review. The missing `menu.compendium` key is added, as in BE_S08_05.
- **Verified:** fire_bolt loads at mana 6 with no problems in both transports, and a core export of a renamed copy passes `validate.py`. [V 2026-11-23]
- **Open:** a module export from the page (`shape: module`); parameters and duration types on an effect; picking an effect from the pool by its label rather than its ref.
