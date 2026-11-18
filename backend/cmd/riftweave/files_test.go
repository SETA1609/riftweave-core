package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const filesFixtures = "../../../ruleset/tests/fixtures/files/"

// fileEnvironment is the settings directory and the ruleset for export and import.
func fileEnvironment(directory string) func(string) string {
	return func(name string) string {
		switch name {
		case "RIFTWEAVE_CONFIG_DIR":
			return directory
		case "RIFTWEAVE_DATA":
			return dataDirectory
		}
		return ""
	}
}

// TestAFileOfEachKindImportsAndExportsBackByteForByte is APP-30's round trip: the shared
// fixture of each kind imports, and its data exports back as the same bytes.
func TestAFileOfEachKindImportsAndExportsBackByteForByte(t *testing.T) {
	for _, kind := range []string{"character", "session", "spell"} {
		t.Run(kind, func(t *testing.T) {
			var environment = fileEnvironment(t.TempDir())
			var fixture string = filesFixtures + kind + ".riftweave.json"
			var stdout, stderr bytes.Buffer
			if code := run([]string{"import", fixture}, nil, &stdout, &stderr, environment); code != exitOK {
				t.Fatalf("import: %d %s", code, stderr.String())
			}
			var read struct {
				Kind     string          `json:"kind"`
				Data     json.RawMessage `json:"data"`
				Warnings []string        `json:"warnings"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &read); err != nil || read.Kind != kind {
				t.Fatalf("import printed %s (%v)", stdout.String(), err)
			}
			if len(read.Warnings) != 0 {
				t.Errorf("warnings for a file of these rules: %v", read.Warnings)
			}
			var exported, exportErrors bytes.Buffer
			if code := run([]string{"export", kind}, bytes.NewReader(read.Data), &exported, &exportErrors, environment); code != exitOK {
				t.Fatalf("export: %d %s", code, exportErrors.String())
			}
			want, _ := os.ReadFile(fixture)
			if exported.String() != string(want) {
				t.Errorf("export differs from the fixture:\n%s", exported.String())
			}
		})
	}
}

func TestASettingsFileImportsThroughImport(t *testing.T) {
	var directory string = t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"import", settingsFixture}, nil, &stdout, &stderr, fileEnvironment(directory)); code != exitOK {
		t.Fatalf("%d %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(directory, "settings.json")); err != nil {
		t.Errorf("the settings were not written: %v", err)
	}
}

func TestAnInvalidFileIsRefusedAndChangesNothing(t *testing.T) {
	for _, case_ := range []struct{ body, says string }{
		{"[]", "not a JSON object"},
		{`{"riftweave": "save", "rules_version": "x", "modules": [], "data": {}}`, "not one of"},
		{`{"riftweave": "character", "data": {}}`, "rules_version, modules, and data"},
	} {
		var path string = filepath.Join(t.TempDir(), "bad.json")
		_ = os.WriteFile(path, []byte(case_.body), 0o644)
		var stdout, stderr bytes.Buffer
		if code := run([]string{"import", path}, nil, &stdout, &stderr, fileEnvironment(t.TempDir())); code != exitInput || !strings.Contains(stderr.String(), case_.says) {
			t.Errorf("%s: exit %d, %s", case_.body, code, stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"import", filesFixtures + "invalid.riftweave.json"}, nil, &stdout, &stderr, fileEnvironment(t.TempDir())); code != exitInput || !strings.Contains(stderr.String(), "does not match the contract") {
		t.Errorf("the invalid fixture: exit %d, %s", code, stderr.String())
	}
}

func TestAFileWithAModuleThisDeviceDoesNotLoadStopsWithTheRowsAtStake(t *testing.T) {
	var directory string = t.TempDir()
	var fixture string = filesFixtures + "missing-module.riftweave.json"
	var stdout, stderr bytes.Buffer
	if code := run([]string{"import", fixture}, nil, &stdout, &stderr, fileEnvironment(directory)); code != exitInput {
		t.Fatalf("want exit %d, got %d: %s", exitInput, code, stdout.String())
	}
	for _, says := range []string{"module factions is not enabled here", "module:factions/accessory/warden_badge", "nothing changed"} {
		if !strings.Contains(stderr.String(), says) {
			t.Errorf("the alert does not say %q: %s", says, stderr.String())
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"import", fixture, "--accept-loss"}, nil, &stdout, &stderr, fileEnvironment(directory)); code != exitOK || !strings.Contains(stdout.String(), "imported without its modules") {
		t.Errorf("--accept-loss: exit %d, %s %s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"import", fixture, "--enable-modules"}, nil, &stdout, &stderr, fileEnvironment(directory)); code != exitOK {
		t.Fatalf("--enable-modules: exit %d, %s", code, stderr.String())
	}
	saved, _ := os.ReadFile(filepath.Join(directory, "settings.json"))
	if !strings.Contains(string(saved), `"factions"`) {
		t.Errorf("factions is not enabled in the settings file: %s", saved)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"import", fixture}, nil, &stdout, &stderr, fileEnvironment(directory)); code != exitOK {
		t.Errorf("with factions enabled: exit %d, %s", code, stderr.String())
	}
}

func TestAFileFromOtherRulesWarnsAndStillOpens(t *testing.T) {
	content, _ := os.ReadFile(filesFixtures + "character.riftweave.json")
	var path string = filepath.Join(t.TempDir(), "old.json")
	_ = os.WriteFile(path, bytes.Replace(content, []byte(`"rules_version": "`), []byte(`"rules_version": "0.0.1-`), 1), 0o644)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"import", path}, nil, &stdout, &stderr, fileEnvironment(t.TempDir())); code != exitOK || !strings.Contains(stdout.String(), "made with rules 0.0.1-") {
		t.Errorf("exit %d, %s %s", code, stdout.String(), stderr.String())
	}
}

func TestExportRefusesDataThatDoesNotMatchTheContract(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"export", "character"}, strings.NewReader(`{"level": 1}`), &stdout, &stderr, fileEnvironment(t.TempDir())); code != exitInput || !strings.Contains(stderr.String(), "nothing written") {
		t.Errorf("exit %d, %s", code, stderr.String())
	}
}
