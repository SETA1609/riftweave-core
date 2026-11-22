# theme-toggle.ts

Modified: 2026-11-22

`rw-theme-toggle`: one button in the header that cycles light, dark, and the system's mode, through `AppContext` ([frontend.md#the-app-context](../../../../docs/backend/frontend.md#the-app-context)).

- **The icon** is the mode shown now, from Font Awesome's free solid set: `faSun` for light, `faMoon` for dark, `faCircleHalfStroke` for the system's mode (`ICONS`, one entry per `Theme`). It is drawn by `fa-icon` (`@fortawesome/angular-fontawesome`) as inline SVG, so it needs no font file and works offline.
- **The words** are in `aria-label` and `title`: the mode it switches to, as `theme.switch_to` with `theme.mode_<next>` ([i18n.md](../i18n.md)), so a screen reader and a tooltip say what a click does.

References: [Font Awesome for Angular](https://docs.fontawesome.com/web/use-with/angular) and its [free solid icons](https://fontawesome.com/search?ic=free&s=solid). [U]
