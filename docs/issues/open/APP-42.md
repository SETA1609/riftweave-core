#### APP-42 File the observability and dev console issues, and raise the data-integrity priorities
- Status: review
- Priority: medium
- Needs: —
- Modified: 2026-11-18
- Do: Owner request, 2026-11-18: every component (server, agent CLI, human CLI, frontend) should be observable, in the OpenTelemetry format; the frontend should expose its objects and functions in the browser console in a dev build; and the normal-form issues go first for data integrity. File APP-43 (the telemetry format and setting), APP-44 (the server), APP-45 (the CLIs), APP-46 (the frontend), APP-47 (a local stack to look at it), and APP-48 (the dev console). Raise M5-70, M5-71, and M5-74 to high, and M5-75, M5-58, and M5-59 to medium.
- Keys: none
- Files: ~13 (docs/issues/open/APP-42.md to APP-48.md, the six raised issues, docs/decisions/APP-42.md, changes/APP-42.md)
- See: docs/backend/README.md, docs/backend/settings.md, docs/theory/normal-forms.md
- Done when: the six issues exist with their Needs; the six priorities are raised; `roadmap-order show` lists the new issues.
