# APP-23 An app context with a light and dark mode toggle

Date: 2026-11-16
Modified: 2026-11-16

- **Decision:** `AppContext` is a root service of signals (`theme`, `locale`, `systemDark`, the
  derived `shown` and `next`), generated with `npx ng generate service context --skip-tests`; the
  toggle is `rw-theme-toggle`, generated with `npx ng generate component ui/theme-toggle --flat`
  and renamed to the `rw-` prefix. Angular 22's CLI writes the service with the new `@Service()`
  decorator, which is kept. **[V 2026-11-16, Vitest and Playwright over HTTP and WebAssembly]**
- **`system` sets no `data-theme`.** The attribute is `light` or `dark` only when chosen, so the
  `prefers-color-scheme` block decides on its own, guarded by `:root:not([data-theme="light"])`. The
  dark tokens are written twice in `styles.css`, under the attribute and under the media query,
  as plain CSS has no way to share one block between them. **[A]**
- **The class is `AppContext`,** as the issue names it, not the generated `Context`; the file stays
  `context.ts`.
- **`locale` is saved but read by nothing yet;** APP-24 wires it to i18next. It defaults to `en`.
- **Difference from the issue:** the `ng generate` rule in `docs/backend/frontend.md#conventions`
  uses `--flat` for shared components, so they sit beside the others in `ui/` as `name.ts`,
  `name.html`, `name.css`, matching the existing layout instead of a folder each. The generated
  `theme-toggle.spec.ts` held only a "should create" check and was removed; the specs are in
  `app.spec.ts`.
- **Risk:** none known; the toggle is in the header of every screen.
