# Roles

Every issue belongs to one role: the person (or agent) who can finish it alone. The role is the
first part of the issue id ([format](../stories/README.md#ids)). Work that needs two roles is two
issues in the same story, the second one listing the first under `Needs:`.

| Code | Role | Works on | Typical paths |
| --- | --- | --- | --- |
| `PO` | Product owner | Stories and their scope, priorities, acceptance criteria, the release gates, licensing, and the issue conventions | `ROADMAP.md`, `docs/stories/`, `docs/issues/`, `LICENSING.md`, `NOTICE` |
| `GD` | Game designer | Rules, perks, builds, spells, monsters, items, balance, modules, and the SRD text; a rule issue follows the five steps of the [workflow](../../CONTRIBUTING.md#workflow), vectors included | `docs/core/`, `docs/modules/`, `docs/guide/`, `ruleset/data/`, `ruleset/modules/`, `ruleset/tests/vectors/` |
| `DE` | Data engineer | Schemas, normal forms, naming and casing, integrity gates in `validate.py`, and the diagrams generated from the data | `ruleset/schemas/`, `ruleset/scripts/validate.py`, `ruleset/scripts/test_integrity.py`, `docs/core/data-model*.md` |
| `BE` | Backend dev | The Go core, the command contract, the agent and human CLIs and the TUI (and their UX), the HTTP API, MCP, WebAssembly, and the Supabase side of rooms | `backend/`, `supabase/`, `docs/backend/` |
| `FE` | Frontend dev | The Angular app and its UX: screens, routes, components, locales, and the browser dev console | `frontend/`, `docs/backend/frontend.md` |
| `QA` | QA | Playtests, test vectors that check other work, end-to-end and Playwright suites, drift tests, and validation runs | `docs/playtests/`, `frontend/e2e/`, `ruleset/scripts/test_*.py` |
| `OPS` | DevOps and tooling | CI, workflows, formatting, versioning and releases, containers, and the observability stack | `.github/`, `scripts/`, `tools/`, `container/` |

There is no UX role: **FE and BE each own the UX of their own surface**, FE the browser and BE the
CLIs and the TUI.

## Which role takes an issue

| The issue mainly changes | Role |
| --- | --- |
| A rule, a perk, a spell, a monster, a build, or balance numbers | `GD` |
| A schema, a cross-reference, a normal form, or an integrity check | `DE` |
| A contract command, Go code, a CLI or TUI screen | `BE` |
| An Angular screen, route, or component | `FE` |
| A playtest, or a test suite that checks work other issues did | `QA` |
| A workflow, a script, a release, or a container | `OPS` |
| Scope, order, priorities, a gate, or these conventions | `PO` |

A rule issue that also needs contract and Go changes is two issues: `GD` (spec, schema, vectors,
data) and then `BE` (contract and Go core), in the same story. This replaces the old
`<issue-id>-ruleset-<slug>` and `<issue-id>-backend-<slug>` split for new issues.
