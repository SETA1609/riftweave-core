package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const settingsFixture = "../../../ruleset/tests/fixtures/settings/settings.json"

func settingsEnvironment(directory string) func(string) string {
	return func(name string) string {
		if name == "RIFTWEAVE_CONFIG_DIR" {
			return directory
		}
		return ""
	}
}

func TestSettingsImportThenExportGivesTheFileBack(t *testing.T) {
	var directory string = t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"settings", "import", settingsFixture}, nil, &stdout, &stderr, settingsEnvironment(directory)); code != exitOK {
		t.Fatalf("import: %d %s", code, stderr.String())
	}
	stdout.Reset()
	if code := run([]string{"settings", "export"}, nil, &stdout, &stderr, settingsEnvironment(directory)); code != exitOK {
		t.Fatalf("export: %d %s", code, stderr.String())
	}
	want, _ := os.ReadFile(settingsFixture)
	if stdout.String() != string(want) {
		t.Errorf("export differs from the imported file, unknown keys included:\n%s", stdout.String())
	}
}

func TestSettingsExportWithNoFileIsTheDefault(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"settings", "export"}, nil, &stdout, &stderr, settingsEnvironment(t.TempDir())); code != exitOK {
		t.Fatalf("%d %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"theme": "system"`) || !strings.Contains(stdout.String(), `"mode": "local"`) {
		t.Errorf("not the defaults: %s", stdout.String())
	}
}

func TestABadSettingsFileIsRefusedAndChangesNothing(t *testing.T) {
	var directory string = t.TempDir()
	var stdout, stderr bytes.Buffer
	run([]string{"settings", "import", settingsFixture}, nil, &stdout, &stderr, settingsEnvironment(directory))
	before, _ := os.ReadFile(filepath.Join(directory, "settings.json"))
	var bad string = filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(bad, []byte(`{"riftweave": "settings", "version": 9}`), 0o644)
	stderr.Reset()
	if code := run([]string{"settings", "import", bad}, nil, &stdout, &stderr, settingsEnvironment(directory)); code != exitInput {
		t.Fatalf("want exit %d, got %d", exitInput, code)
	}
	if !strings.Contains(stderr.String(), "version 9") || !strings.Contains(stderr.String(), "nothing changed") {
		t.Errorf("message: %s", stderr.String())
	}
	after, _ := os.ReadFile(filepath.Join(directory, "settings.json"))
	if string(before) != string(after) {
		t.Error("a refused import changed the saved settings")
	}
}
