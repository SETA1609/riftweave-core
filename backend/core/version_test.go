package core

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestTheVersionIsTheRepoVersion: VERSION at the repo root is the one version (M5-62); the
// core's embedded copy and frontend/package.json must say the same.
func TestTheVersionIsTheRepoVersion(t *testing.T) {
	content, err := os.ReadFile("../../VERSION")
	if err != nil {
		t.Fatal(err)
	}
	var want string = strings.TrimSpace(string(content))
	if RulesVersion != want {
		t.Errorf("the core embeds %q, VERSION says %q: run go generate ./core", RulesVersion, want)
	}
	var frontend struct{ Version string }
	if content, err = os.ReadFile("../../frontend/package.json"); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(content, &frontend); err != nil {
		t.Fatal(err)
	}
	if frontend.Version != want {
		t.Errorf("frontend/package.json says %q, VERSION says %q", frontend.Version, want)
	}
}

// TestTheAuthorManifestStartsAtTheVersion: an exported module manifest carries the project
// version, and so does the fixture that pins the export.
func TestTheAuthorManifestStartsAtTheVersion(t *testing.T) {
	opened := authorCore(t)
	result := author(t, opened, map[string]any{"kind": "spell", "spells": []any{draftSpell()}, "export": map[string]any{"shape": "module", "module": "authored"}})
	manifest, _ := result["export"].(map[string]any)["manifest"].(map[string]any)
	if manifest["version"] != RulesVersion {
		t.Errorf("the manifest says %v, the core %q", manifest["version"], RulesVersion)
	}
	var fixture struct{ Manifest struct{ Version string } }
	content, err := os.ReadFile("../../ruleset/tests/fixtures/author/module.json")
	if err == nil {
		err = json.Unmarshal(content, &fixture)
	}
	if err != nil || fixture.Manifest.Version != RulesVersion {
		t.Errorf("ruleset/tests/fixtures/author/module.json says %q (%v): run go test ./core -run TestTheExportsMatchTheirFixtures -update", fixture.Manifest.Version, err)
	}
}
