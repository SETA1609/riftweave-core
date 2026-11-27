# navbar.ts

Modified: 2026-11-27

`rw-navbar` (FE_S08_17): the app's header as a navbar, pinned to the top of the window, laid out after the owner's Stitch mockups (docs/design/mockups/, 2026-11-23). The shell shows it above every screen.

- **Pinned:** the host is `position: sticky; top: 0`, so it stays in view while a long screen scrolls and, unlike `fixed`, takes its own room, so no content hides under it. It reaches the page's edges by a negative margin the size of the page margin (`--page-margin`, or `--space-3` before the design tokens), and pads the notch of a phone with `env(safe-area-inset-*)` (`viewport-fit=cover` in `index.html`).
- **One row, by flex, mobile first.** The base is the smallest phone (320px); wider screens add to it:
  - **Every size:** the brand (an icon tile, `faCube`, and the game's name as the link home, the page's `h1`), the language select in its compact form (flag and code, the flag alone under 360px), the theme toggle, and the burger. The brand takes the room that is left and gives it up first, so its name shrinks (`clamp`) before anything wraps. The controls are 44px touch targets (the `rw-navbar .controls .btn` rule in `styles.css`, since the buttons are the language select's and the theme toggle's own).
  - **From 768px** (a tablet upright): a 64px bar, the rules version under the name, and a chip with the game in play, by its save's name (`Saves.current`), absent before a game starts.
  - **From 1024px** (a tablet on its side, a desktop): the main menu inline, GitHub as an icon, no burger, and 36px controls. The brand never shrinks here; the playing chip gives up its label and its width first, so the menu shows whole in every language.
- **The burger** opens the main menu, the entries of `backend/contract/menu.json` under `main` (`pathOf` from `menu-grid.ts`, the same tree the home page shows), then GitHub. Following a link closes it; so do a click outside and Escape, and while it was open Escape stops there, so it does not also go back a screen (the screen header's Esc, FE_S08_12). The button carries `aria-expanded` and `aria-controls` (`NAV_MENU_ID`).
- **GitHub** (`REPOSITORY_URL`, `faGithub` from `@fortawesome/free-brands-svg-icons`) opens in a new tab with `rel="noopener"` and the localized `nav.github` label; in the menu it reads `nav.github_short`.
- **Not here yet:** the side navigation and the bottom tab bar (FE_S22_05), and the profile avatar, which needs a profile (FE_S25_07). The mockups' status chip ("offline / WASM") was dropped (owner, 2026-11-27).
- **Words** are `nav.*` locale keys (`label`, `menu`, `github`, `github_short`, `rules`, `playing`) in en, de, and es; the game's name is not translated.

References: [frontend.md#conventions](../../../../docs/backend/frontend.md#conventions), [language-select.md](language-select.md), [theme-toggle.md](theme-toggle.md), [MDN position: sticky](https://developer.mozilla.org/en-US/docs/Web/CSS/position#sticky_positioning), [MDN env()](https://developer.mozilla.org/en-US/docs/Web/CSS/env). [U]
