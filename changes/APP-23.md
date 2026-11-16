### APP-23 An app context with a light and dark mode toggle
- Added: `AppContext` (`frontend/src/app/context.ts`), the viewer's mode and language as signals, saved in `localStorage` with every access guarded.
- Added: `rw-theme-toggle` in the header cycles light, dark, and the system's mode; the mode is `data-theme` on `<html>`.
- Changed: `docs/backend/frontend.md#conventions` says to generate components and services with `ng generate`, and how.
