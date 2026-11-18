package core

import (
	_ "embed"
	"strings"
)

// The project's one version lives in VERSION at the repo root (docs/core/versioning.md). A Go
// embed cannot reach outside its module, so `go generate ./core` copies it here, and
// TestTheVersionIsTheRepoVersion fails when the copy, frontend/package.json, or the author
// manifest disagree with it (M5-62).
//
//go:generate cp ../../VERSION VERSION
//go:embed VERSION
var embeddedVersion string

// RulesVersion names the rules this core runs: the project version. A session records it at
// start, and a replay stamps the rebuilt session with it, so a save from other rules shows
// (APP-16). An exported module manifest starts at it too.
var RulesVersion string = strings.TrimSpace(embeddedVersion)
