# FE_S11_04_frontend-screen-strings Localize the play and character screens

Date: 2026-11-27

- **Decision:** the character sheet, skill resolution, attack, playthrough, import, the character list, and saved games show their words through keys in en, de, and es (`sheet.*`, `resolution.*`, `attack.*`, `playthrough.*`, `import.*`, `characters.*`, `saves.*`), and `PENDING` in `i18n.spec.ts` is empty, so every template is checked [V 2026-11-27, unit spec]. The English values are the old text word for word, so the Playwright specs, which find elements by English labels, did not change [V 2026-11-27, e2e over both transports].
- **Messages a screen sets** are a `Said` (`{ key, values }`), translated where they show, so a notice already on screen follows a language switch [V 2026-11-27, unit spec]. The `t` pipe takes a `Said` as well as a key. The missing-modules alert of the import is a `computed` over the pending file for the same reason; in English it stays word for word the CLI's.
- **Kept as they came:** the reasons the core or a store gives (a refusal, a parse error, the character store's sync notes such as "revision 2 is older than the saved revision 4; nothing changed") go into a message as a value. Localizing the character store's notes and `files.ts`'s English helpers is left for later [U: its own issue, or part of the next screen issue].
- **Abbreviations:** `abbr.*` (`str` … `lck`, `LCK`, `HP`, `d100`) are the same in every language today [A]; `attack.armor_dr` and `attack.shield_dr` are phrases, translated with DR kept as `DR` in German and `RD` in Spanish [A: the owner reviews de and es].
- **Rule keys stay data:** the stance and delivery options (`dodge`, `block`, `single`, `area`, `instant`) show their rule keys; only the empty stance, "none", is a word of the app.
- **Code shown as code:** `riftweave export` and `.riftweave.json` are named constants interpolated into the templates, not words.
- **The spell notice** no longer says Spell creation "comes with APP-22", since the spell editor exists (FE_S09_02).
