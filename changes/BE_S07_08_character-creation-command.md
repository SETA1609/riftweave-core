### BE_S07_08_character-creation-command Character creation as one command that checks a draft step by step (backend)
- Added: the `creation` command: `rules`, `check` (each step's problems, the points left, the attributes, skills, and derived stats), and `finish` (a stored actor); the `creation_draft` shape in `common.json`.
- Added: races and backgrounds in the Go core, for creation.
- Changed: the human CLI's Character creation is a screen over `check` and `finish` (name, race, background, points, creation perks, tag skills), not eight numbers through `derived`.
