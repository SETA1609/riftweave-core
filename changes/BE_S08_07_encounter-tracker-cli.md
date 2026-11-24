### BE_S08_07_encounter-tracker-cli An offline encounter tracker: initiative and whose turn it is (backend)
- Added: the `encounter` command (`start`, `turn_end`): the initiative order with the tie rule's roll-off as input, the surprise round, whose turn it is with its `start` for `actions`, and the round.
- Added: a session keeps its `encounter`, filled into `turn_end` steps and rebuilt on replay.
