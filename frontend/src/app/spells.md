# spells.ts

Modified: 2026-11-23

`rw-spells` (FE_S09_02), the spell editor at `/play/spell_creation`: the browser half of BE_S09_01, over the `author` command ([authoring.md](../../../docs/backend/authoring.md)). It adds no rule; every number and problem is the command's.

- **The draft** is one spell row as `spells/core.json` writes it, a signal. The form edits its key, label, description, color, level, delivery, and each effect (ref, magnitude, duration, target, radius); every other field of a loaded row (`template`, `cost`, `save`) stays in the draft as loaded and goes into the check and the export.
- **Load** (`operation: load`) puts an existing spell into the draft, to edit or copy.
- **The effect pool** (`operation: effects`) is what the draft's color may use; it is read on a load and on a color change, and its first ref is picked to add.
- **Checked as you type:** every change of the draft is checked (`operation: check`) after `CHECK_DELAY_MS` (300 ms) with no further change, through RxJS (`toObservable`, `debounceTime`, `switchMap`), so a slow answer to an old draft never overwrites a newer one. The cost, the problems, and the warnings show what the command returns; with no problem the card says so.
- **Download** checks the draft with `export: { shape: core }` and saves the export's `file` as `spells.json`, for `ruleset/data/spells/`. An export with problems is still written, as the command does.
- **Words** are `spells.*` locale keys; the colors and targets show as the data writes them.

References: [frontend.md#conventions](../../../docs/backend/frontend.md#conventions), [RxJS debounceTime](https://rxjs.dev/api/operators/debounceTime), [Angular toObservable](https://angular.dev/ecosystem/rxjs-interop). [U]
