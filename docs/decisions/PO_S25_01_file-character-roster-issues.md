# PO_S25_01_file-character-roster-issues File the character roster, profile, and creation flow issues

Date: 2026-11-24

- **Decision** (owner, 2026-11-24): story S25 for the character roster, the profile, and a step-by-step creation flow.
  - **GD:** `GD_S25_02` adds sex and a starting perk to the creation rules.
  - **BE:** `BE_S25_03` puts the order of the flow in the `creation` command, and `BE_S25_04` adds the profile to the agent CLI and the MCP server.
  - **FE:** `FE_S25_05` is the flow page, `FE_S25_06` the roster, and `FE_S25_07` playing as the profile.
- **What exists** [V 2026-11-24, read in this session]: the `creation` command (`rules`, `check`, `finish`) takes a draft's steps in any order (docs/backend/creation.md), and the frontend's character list keeps saved characters in IndexedDB with the sync rule (FE_S07_04, done). Missing are a sex step, a starting-perk step, the order of the steps, a roster view with filters, and the profile.
- **The order is a rule, so it is the core's** [A]: who (name, sex); race; background; traits; attributes; tag skills; perks; review. Each step comes after what it depends on:
  - Race sets ability modifiers, traits, and the counts.
  - Background gives skill bonuses before tag skills.
  - Traits move attributes and limits before the point buy.
  - The starting perk's prerequisites read attributes and skills.
  - The page reads `rules.flow` and holds no order of its own.
- **FE_S07_09 dropped:** its one-page CV form is replaced by the flow. Its point stepper, its no-JSON rule, and its kept draft move to FE_S25_05, and FE_S07_06 now needs FE_S25_05.
- **Ambition** (`human_ambition`) adds to `creation_perks` (traits), but its description and docs/modules/character-creation.md say a bonus normal perk. GD_S25_02 moves it to the new `starting_perks` count [A, the owner confirms].
- **The profile is per device, beside the character list**, not in the settings file: settings travel between devices, and a character id means nothing on another one. [A]
- **Sex** (owner, 2026-11-24): male or female, recorded and read by no rule. An artificial or undead being (a golem, an undead, an animated object) has none, shown as "it": the creature types `construct`, `undead`, `elemental`, `ooze`, and `plant` read `none`, and a template of one of them lets the player switch to it by choice, never automatically (owner, 2026-11-24).
- **GitHub issues:** filed for each S25 issue.
