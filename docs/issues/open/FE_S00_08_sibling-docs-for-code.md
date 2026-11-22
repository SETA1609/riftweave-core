#### FE_S00_08_sibling-docs-for-code Move the frontend's code comments into sibling Markdown files
- Status: todo
- Priority: low
- Needs: FE_S11_01_frontend-i18next
- Modified: 2026-11-22
- Do: Owner request, 2026-11-22: a `.ts` file under `frontend/src/` keeps one pointer line (`// See <name>.md.`) and no other comments; its explanation, the reasons, and the upstream references go in a sibling `<name>.md` with a `Modified:` line ([frontend.md#conventions](../../backend/frontend.md#conventions)). FE_S11_01 did it for `i18n.ts`, `translate-pipe.ts`, `ui/language-select.ts`, and `i18n.spec.ts`. Do the rest: `context.ts`, `api.ts`, `wasm.ts`, `saves.ts`, `characters.ts`, `files.ts`, `settings-file.ts`, `dev-console.ts`, `app.routes.ts`, every screen, every `ui/` component, and the specs. A refactor: no behavior changes, and no comment is dropped without its content moving to the `.md`. Splits by folder if it passes 20 counted files (screens, `ui/`, services and specs).
- Keys: one `<name>.md` beside each `<name>.ts`
- Files: ~40, split in two or three pull requests (src/app/*.ts and their .md, src/app/ui/*.ts and their .md, the specs)
- See: docs/backend/frontend.md#conventions, frontend/src/app/i18n.md
- Done when: no `.ts` file under `frontend/src/` has a comment but its pointer line (a `@vite-ignore` and similar tool directives stay); every `.ts` file with an explanation has its sibling `.md`; `ng build`, the Vitest specs, and the Playwright specs pass unchanged.
