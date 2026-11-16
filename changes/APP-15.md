### APP-15 The frontend is published on GitHub Pages
- Added: `.github/workflows/pages.yml` publishes the app with the core and its ruleset built in, on a release tag, through the official Pages actions pinned by commit.
- Added: the `pages` build configuration with a service worker, a web manifest, and an icon, so the app works offline after one visit and can be installed.
- Added: the footer shows the rules version (the tag, or `dev`); the guide says how a player opens and installs the app.
