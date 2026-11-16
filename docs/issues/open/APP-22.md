#### APP-22 A spell editor page over the author command
- Status: todo
- Priority: low
- Needs: APP-10, APP-14
- Modified: 2026-11-16
- Do: Split from APP-10 on 2026-11-15. A spell editor screen in the frontend over the `author` command (docs/backend/authoring.md), built from the reusable components (docs/backend/frontend.md#conventions). You pick a color, then effects from the shared pool, filtered by that color and by each effect's `spell` channel; set magnitude, duration, target, and radius; and see the formula's cost and thresholds and the problems as you type (a debounced call, RxJS). It can load an existing spell (`operation: load`) to edit or copy, and it downloads the core or module export. The pool needs a way to list effects: an `author` operation that returns the effects a color may use, added here [A]. The browser core (APP-14) must embed `ruleset/schemas` too and open with `core.NewWithSchemas`, or `author` refuses there for want of the schemas.
- Keys: the spell editor screen; an `author` operation listing the effect pool [A]
- Files: ~12 (backend/cmd/riftweave-wasm/main_js.go, scripts/build-wasm.sh, frontend/src/app/spells.ts, spells.html, spells.css, app.ts, app.html, api.ts, app.spec.ts, e2e/app.spec.ts, backend/core/author.go, docs/backend/frontend.md)
- See: docs/backend/authoring.md, docs/issues/open/APP-10.md (or done/ once merged)
- Done when: the page loads fire_bolt, shows the same cost and no problems the `author` command gives for it, shows a problem for a broken ref as you type, and downloads a core export that passes `validate.py`; the frontend specs and the Playwright tests pass over both transports.
