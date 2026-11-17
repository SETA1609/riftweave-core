### APP-25 One settings file that both apps read, export, and import
- Added: the settings file format, `backend/contract/settings.schema.json`: locale, theme, enabled modules, network, and the bestiary setting; unknown keys survive a round trip.
- Added: `riftweave settings export [file]` and `riftweave settings import <file>`, and `exportSettings()` and `importSettings()` on the frontend's `AppContext`; a file from either imports into the other and exports back byte for byte.
