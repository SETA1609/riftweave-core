package settings

import (
	"os"
	"strings"
	"testing"
)

const fixture = "../../../ruleset/tests/fixtures/settings/settings.json"

func TestTheFixtureReadsAndWritesBackUnchanged(t *testing.T) {
	content, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	file, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	if file["locale"] != "de" || file["theme"] != "dark" || file["network"].(map[string]any)["mode"] != "server" || file["telemetry"].(map[string]any)["mode"] != "export" {
		t.Errorf("read %v", file)
	}
	if got := string(file.Encode()); got != string(content) {
		t.Errorf("a round trip changed the file:\n%s", got)
	}
}

func TestABadFileIsRefusedWithAMessage(t *testing.T) {
	for _, case_ := range []struct{ body, says string }{
		{`[]`, "not a JSON object"},
		{`{"riftweave": "save", "version": 1}`, "not a settings file"},
		{`{"riftweave": "settings", "version": 2}`, "version 2"},
		{`{"riftweave": "settings", "version": 1, "theme": "blue"}`, "theme blue"},
		{`{"riftweave": "settings", "version": 1, "modules": ["core"]}`, "lists core"},
		{`{"riftweave": "settings", "version": 1, "network": {"mode": "lan"}}`, "network.mode"},
		{`{"riftweave": "settings", "version": 1, "telemetry": {"mode": "on"}}`, "telemetry.mode"},
		{`{"riftweave": "settings", "version": 1, "telemetry": {"mode": "export", "endpoint": "collector:4318"}}`, "telemetry.endpoint"},
	} {
		if _, err := Parse([]byte(case_.body)); err == nil || !strings.Contains(err.Error(), case_.says) {
			t.Errorf("%s: want an error saying %q, got %v", case_.body, case_.says, err)
		}
	}
}

func TestTheDefaultIsAValidFile(t *testing.T) {
	if _, err := Parse(Default().Encode()); err != nil {
		t.Fatal(err)
	}
}

func TestTheDefaultRecordsNothing(t *testing.T) {
	if mode := Default()["telemetry"].(map[string]any)["mode"]; mode != "off" {
		t.Errorf("telemetry.mode defaults to %v, not off", mode)
	}
}
