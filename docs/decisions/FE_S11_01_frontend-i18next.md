# FE_S11_01_frontend-i18next Localize the frontend with i18next

Date: 2026-11-21

- **Decision:** the app's own text is localized with i18next, one instance in `I18n` (`frontend/src/app/i18n.ts`) that follows `AppContext.locale`; the rules live in [frontend.md](../backend/frontend.md#languages).
- **Version:** i18next 26.7.2 (published 2026-11-05), the newest release at least two weeks old on 2026-11-21; the issue named 26.6.0 as the newest when it was filed. It has no dependencies. [V 2026-11-21, `npm view i18next time`]
- **Lockfile:** npm 11.5 dropped the optional `@emnapi/*` entries when it added i18next, which other platforms need; the lockfile keeps them, with only the i18next entry added. [V 2026-11-21, `npm ci --dry-run`]
- **One select, in the header:** the Settings screen's free-text Language field is gone; the header's `rw-language-select` is on every screen, and a second one on Settings would only repeat it.
- **Pipe file name:** `translate-pipe.ts`, as `ng generate pipe translate` names it, not the `translate.pipe.ts` the issue listed; the pipe's name in templates is `t`.
- **Split:** the whole issue passed 20 counted files, so the play and character screens moved to `FE_S11_04_frontend-screen-strings`; `PENDING` in `i18n.spec.ts` lists them, and that issue empties it.
- **The spec reads the templates with node:fs** loaded at run time: the app has no Node types, and Vitest runs on Node. [V 2026-11-21, `ng test`]
- **Open:** the settings file's rule messages stay in English, because they are the Go CLI's words; BE_S11_02 shares the locale files with the CLI and can move them. [A]
- **Risk:** the German and Spanish text is a first draft for the owner's review.
