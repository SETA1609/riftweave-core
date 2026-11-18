#### FE_S10_04_gm-approval-table Rooms: the GM approves steps, and the table sees who is there (frontend)
- Status: todo
- Priority: low
- Needs: BE_S10_03_gm-approval-supabase, FE_S10_02_room-log-client
- Was: APP-18
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S10_03_gm-approval-supabase` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S10_03_gm-approval-supabase`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: ~3 (frontend/src/app/room.ts, frontend/src/app/table.ts, frontend/src/app/table.spec.ts)
- See: docs/issues/open/BE_S10_03_gm-approval-supabase.md
- Done when: the GM's client shows a player's request with its previewed result and accepts or declines it; presence shows who is in the room; the approval setting is in the room screen; the specs pass.
