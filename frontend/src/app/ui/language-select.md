# language-select.ts

Modified: 2026-11-27

`rw-language-select` (FE_S11_01): the offered languages, each with its flag and its own name (`LANGUAGES` in `context.ts`), bound to `AppContext.locale`.

- **A button and a menu**, not a native `<select>`: an `<option>` holds only text, so it cannot show a flag. The button shows the current language's flag and name, with `aria-label` the `language.label` key and `aria-haspopup="menu"`; a click opens a `role="menu"` list of `menuitemradio` buttons, the current one `aria-checked`. Picking one sets the locale and closes the menu; Escape or a click outside closes it.
- **The shown language** is `I18n.language()`, not the raw locale, so a code the app does not offer shows English.
- **The flags** are SVGs from [flag-icons](https://github.com/lipis/flag-icons) (`flag-icons` 7.6.0, MIT), 4:3. Only the three in use are copied into the build, as `flags/<code>.svg` (the `assets` glob in `angular.json`), and the service worker prefetches them (`ngsw-config.json`), so they show offline. English uses the US flag (`us`, owner, 2026-11-22); a flag stands for a country, not a language, and the name beside it is what the picker means.
- **Compact** (`[compact]="true"`, FE_S08_17): the button shows the flag and the language's code (`EN`) instead of its name, and the flag alone under 360px; the menu still lists the names. The navbar uses it.
- It sits in the navbar (`rw-navbar`), so every screen has it.

References: [frontend.md#languages](../../../../docs/backend/frontend.md#languages), [i18n.md](../i18n.md), the [WAI-ARIA menu button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/menu-button/). [U]
