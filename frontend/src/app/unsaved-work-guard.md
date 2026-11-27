# unsaved-work-guard.ts

Modified: 2026-11-27

`unsavedWorkGuard` (FE_S08_12): a `CanDeactivate` guard that asks before leaving a screen with work it has not saved. A screen opts in by implementing `HoldsUnsavedWork` (`hasUnsavedWork()`) and naming the guard in its route's `canDeactivate`.

- **Character creation** opts in: its form holds unsaved work while its attributes or level differ from what they were when the page opened or a character was last created from it (`Sheet.unsaved`).
- **The question** is the `screen.leave_unsaved` locale key, asked with the browser's `confirm`; leaving goes on only when the person agrees. `ask.confirm` is an object property so a spec can replace it.
- Back, Esc, a menu link, and the browser's back all navigate, so the guard asks for every one of them. Closing the tab is not asked (`beforeunload` is not used). [A]

References: [Angular route guards](https://angular.dev/guide/routing/route-guards). [U]
