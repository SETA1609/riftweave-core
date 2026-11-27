# screen-header.ts

Modified: 2026-11-27

`rw-screen-header` (FE_S08_12): the screen's title and **Back**, shown once by the app shell above the router outlet on every route but `/`, so every screen has a way back without its own markup ([app-layout.md](../../../../docs/backend/app-layout.md#the-menu-tree)).

- **Where Back goes** (`parentPath`): to the menu the screen's entry sits in, from `backend/contract/menu.json` through `pathOf`, as the CLI's Esc does, not to the browser's history. So a screen opened by its URL still has a way back. The key is the path's last segment (`/play` is Start); a path with no entry goes to the menu its parent path names.
- **The title** is the entry's locale key, `menu.<key>`, the same key the route's title uses.
- **Esc** does what Back does, from anywhere in the document, unless a field (`input`, `textarea`, `select`, or an editable element) has focus or a `dialog` is open (`escapeIsTaken`), so Esc still clears a field or closes a dialog first.
- **The menus** no longer end with a Back card; the header is their Back too, so a screen never shows two.
- **Leaving with unsaved work** is the route's guard, [unsaved-work-guard.md](../unsaved-work-guard.md); Back and Esc are ordinary navigations, so the guard asks for both.
- The URL is read from the router's `NavigationEnd` events as a signal (`toSignal`).

References: [Angular toSignal](https://angular.dev/ecosystem/rxjs-interop), [Angular host bindings](https://angular.dev/guide/components/host-elements). [U]
