# navbar.ts

Modified: 2026-11-27

`rw-navbar` (FE_S08_17): the app's header as a navbar, pinned to the top of the window. The shell shows it above every screen.

- **Pinned:** the host is `position: sticky; top: 0`, so it stays in view while a long screen scrolls and, unlike `fixed`, takes its own room, so no content hides under it. It reaches the page's edges by a negative margin the size of the body's padding.
- **What it holds:** the game's name as the link home (the page's `h1`), the language select (`rw-language-select`), the theme toggle (`rw-theme-toggle`), and a link to the repository (`REPOSITORY_URL`) with the GitHub icon (`faGithub`, from `@fortawesome/free-brands-svg-icons`), opening in a new tab with `rel="noopener"` and the localized `nav.github` label.
- **On a narrow screen** (under 768px, Bootstrap's `md`), the three items fold into a burger button (`faBars`, labelled `nav.menu`) that opens them as a dropdown. The button carries `aria-expanded` and `aria-controls` (`NAV_ITEMS_ID`). A click outside closes it, as the language menu does, and so does Escape; while the dropdown was open, Escape stops there, so it does not also go back a screen (the screen header's Esc, FE_S08_12). On a wide screen the items show inline and the burger is hidden.
- **Words** are `nav.*` locale keys (`label`, `menu`, `github`) in en, de, and es; the game's name is not translated.

References: [frontend.md#conventions](../../../../docs/backend/frontend.md#conventions), [language-select.md](language-select.md), [theme-toggle.md](theme-toggle.md), [MDN position: sticky](https://developer.mozilla.org/en-US/docs/Web/CSS/position#sticky_positioning). [U]
