### BE_S09_03_print-character-cli Print the current character from both apps (backend)
- Added: `print` sheet takes `state`: `current` (the default) prints the HP, mana, and stamina left, the conditions, the held modifiers, and the gear; `fresh` prints full resources with no condition or held modifier.
- Added: the human CLI prints a saved character from the character list with `w`, to an HTML file, `--fresh` for a fresh sheet; the browser half is FE_S09_04.
