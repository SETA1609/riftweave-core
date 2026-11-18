# APP-36 The command contract drawn in Mermaid

Date: 2026-11-18

- **Decision:** `ruleset/scripts/contract_diagram.py` writes `docs/backend/contract-diagram.md` from `backend/contract/`: one `classDiagram` per command (its input and result with every field and its type), the object types of `common.json` drawn once with their links, and a `flowchart` of the session steps. A Python script beside `data_model_diagram.py`, as the issue allowed [A]. [V 2026-11-18, `test_contract_diagram.py`]
- **Links only to object types:** a field typed by a shared object (`actor`, `stored_actor`, `session`) links to its class; a scalar shared type (`key`, `d100`, `tags`) is shown as the field's type, without an arrow, so each diagram stays readable [A].
- **The session flowchart is read from the table** in docs/backend/sessions.md#what-a-step-reads-and-writes, so it says what the table says; a test checks that the table's commands are the ones `sessionAdapters` in `backend/core/session.go` adapts (`dev`, which runs only with `--dev`, is the one row it does not).
- **Open:** the blocks were not rendered by Mermaid in this change [U].
