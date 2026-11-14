# APP-19 The frontend follows its conventions

Date: 2026-11-14

- **Decision:** the conventions are written once, in `docs/backend/frontend.md#conventions`, and `AGENTS.md` points to them, so every contributor and agent sees the same rules: Angular 22, signals first and RxJS only where signals do not fit, standalone components, three files per component, Bootstrap's look in our own CSS, small reusable components, Vitest for unit tests, Playwright for end-to-end tests. [V 2026-11-14]
- **Decision:** six reusable components in `src/app/ui/`: `rw-number-field`, `rw-text-field`, and `rw-select-field` (a label and a `form-control`, the value a `model` signal, so a screen binds `[(value)]="signal"`), `rw-card` (a title and projected content), `rw-alert` (shown only with a message), and `rw-key-value-list` (a two-column table). The screens are compositions of them: the sheet has nine fields, the attack form eleven. [V 2026-11-14, `app.spec.ts`]
- **Decision:** the attack form's state is one signal per field, replacing a mutable object, and the result is a `computed` in pipeline order; the screens send exactly the input they sent before, which `app.spec.ts` checks against a fake API and the 14 API specs check against a server. [V 2026-11-14]
- **Decision:** `styles.css` holds the design tokens (custom properties for color, spacing, radius, and font, with a dark scheme under `prefers-color-scheme`) and the shared Bootstrap-like classes (`.btn`, `.btn-primary`, `.btn-outline`, `.form-control`, `.form-label`, `.card`, `.alert`, `.table`, `.nav`); each component's own layout is in its `.css`. No Bootstrap package is installed. [V 2026-11-14]
- **Decision:** the screens keep their three files flat in `src/app/` rather than a folder each: moving them made git count each as a deleted and a new file, and the pull request past the 20-file limit. The `API` token moved to `api.ts`, which also removes the import cycle between the shell and the screens. [A]
- **Why:** owner rules for every frontend change, 2026-11-14.
- **Open:** the Playwright end-to-end tests are APP-20, filed in its own change. The new look was checked by the build and the specs, not by eye in a browser. [U]
