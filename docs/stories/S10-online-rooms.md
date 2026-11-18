# S10 Online rooms

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S10 Online rooms

**Goal:** A GM and up to four players share one session log online, with GM approval and a shared battle grid.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `BE_S10_01_room-log-supabase` | BE | Rooms: a GM and up to four players share one session log (backend) | APP-17 | APP-15, APP-16 |
| `FE_S10_02_room-log-client` | FE | Rooms: a GM and up to four players share one session log (frontend) | APP-17 | BE_S10_01_room-log-supabase, APP-15, APP-16 |
| `BE_S10_03_gm-approval-supabase` | BE | Rooms: the GM approves steps, and the table sees who is there (backend) | APP-18 | BE_S10_01_room-log-supabase |
| `FE_S10_04_gm-approval-table` | FE | Rooms: the GM approves steps, and the table sees who is there (frontend) | APP-18 | BE_S10_03_gm-approval-supabase, FE_S10_02_room-log-client |
| `BE_S10_05_battle-grid-steps` | BE | Online table: a shared battle grid, and the initiative list in sync (backend) | APP-35 | BE_S10_03_gm-approval-supabase, BE_S08_07_encounter-tracker-cli |
| `FE_S10_06_battle-grid-page` | FE | Online table: a shared battle grid, and the initiative list in sync (frontend) | APP-35 | BE_S10_05_battle-grid-steps, FE_S10_04_gm-approval-table, FE_S08_08_encounter-tracker-page |

**Done when:** every issue above is done, and five browsers in one room share the log, the GM approves steps, and the grid stays in sync.
