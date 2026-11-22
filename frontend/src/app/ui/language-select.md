# language-select.ts

Modified: 2026-11-22

`rw-language-select` (FE_S11_01): a select of the offered languages, each by its own name (`LANGUAGES` in `context.ts`), bound to `AppContext.locale`.

- It is built on `rw-select-field`; its label is the `language.label` key.
- The shown value is `I18n.language()`, not the raw locale, so a code the app does not offer shows English selected.
- It sits in the app header (`app.html`), so every screen has it; the Settings screen has none of its own.

References: [frontend.md#languages](../../../../docs/backend/frontend.md#languages), [i18n.md](../i18n.md).
