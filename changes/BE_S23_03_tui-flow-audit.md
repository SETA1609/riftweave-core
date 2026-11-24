### BE_S23_03_tui-flow-audit Move every rule step that lives only in the terminal app into the core
- Added: `stored_actor`'s `new`: a whole new stored actor from attributes (and an optional level and tags), with xp 0, nothing held, and every pool full at the derived maximum.
- Changed: the web app's new character and the terminal app's dev panel get their new actor from the core instead of building it by hand; the audit of the terminal app found no other rule step.
