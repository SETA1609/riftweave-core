# FE_S08_06_bestiary-page A bestiary, and the GM decides what the players know (frontend)

Date: 2026-11-27

- **Decision:** the page lives at `/play/bestiary`, the Start menu's Bestiary entry, which is no longer pending in the web app. It holds no rule: the list is `catalog`, every view is `bestiary view`, and a full entry's stats are `monster_sheet` [V 2026-11-27].
- **Where the marks live:** with a game in play, in its session, as `bestiary` steps, autosaved like any other step, as the CLI keeps them in its session [V 2026-11-27, unit spec]. With no game, in the page until it closes, with a notice; a mark is still `bestiary reveal`, so it is the command's rule either way [A]. A shared room (BE_S10_01) is where a player's own view will come from; until then, the GM switches **view as** to show the table a player's view [A].
- **Viewers:** the GM, `table` (what the whole table was shown, as the CLI's players' view), and each player named in a mark. A player is named by typing them in the form; there is no player list before rooms [A].
- **A Lore check** uses the session's character as the checking actor, so it needs a game in play and a player name [A]. The roll is the app's d100 (`rollD100`).
- **Filters:** the CLI's field words (`type:`, `phase:`, `level:`) as form fields over the same catalog filters; a level bound of 0 is no bound, since the number field has no empty value [A].
- **Open:** a player list from the party (`session.party`) instead of typed names, once the party is shown in the app [U].
