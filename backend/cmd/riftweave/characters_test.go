package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/files"
)

// device is one person's CLI: its settings directory, so its character list, and the ruleset.
type device struct {
	t           *testing.T
	environment func(string) string
}

func newDevice(t *testing.T) device {
	return device{t: t, environment: fileEnvironment(t.TempDir())}
}

// run calls the CLI on this device and returns the exit code, stdout, and stderr.
func (d device) run(stdin string, args ...string) (int, string, string) {
	d.t.Helper()
	var stdout, stderr bytes.Buffer
	var code int = run(args, strings.NewReader(stdin), &stdout, &stderr, d.environment)
	return code, stdout.String(), stderr.String()
}

// must is run that fails the test on a non-zero exit.
func (d device) must(args ...string) string {
	d.t.Helper()
	code, stdout, stderr := d.run("", args...)
	if code != exitOK {
		d.t.Fatalf("%v: exit %d: %s", args, code, stderr)
	}
	return stdout
}

func (d device) list() []map[string]any {
	d.t.Helper()
	var entries []map[string]any
	if err := json.Unmarshal([]byte(d.must("characters", "list")), &entries); err != nil {
		d.t.Fatal(err)
	}
	return entries
}

func reportOf(t *testing.T, text string) map[string]any {
	t.Helper()
	report, ok := decode(t, text).(map[string]any)
	if !ok {
		t.Fatalf("not a report: %s", text)
	}
	return report
}

const syncedFixture = filesFixtures + "character-synced.riftweave.json"
const syncedID = "5f0c7a52-3c1e-4b8e-9d2a-7e4b1c6f9a10"

// TestACharacterTravelsFromAPlayerToTheGMsPartyAndBack is BE_S07_03's hand-off: a player's file
// loads into a GM's session party, the GM's updated export imports back as a higher revision,
// the older file is then refused, and a diverged one asks which side to keep.
func TestACharacterTravelsFromAPlayerToTheGMsPartyAndBack(t *testing.T) {
	var player, gm = newDevice(t), newDevice(t)
	var folder string = t.TempDir()

	// The player's character, as the frontend exports it, goes into the player's list.
	if report := reportOf(t, player.must("characters", "import", syncedFixture)); report["outcome"] != "added" {
		t.Fatalf("player import: %v", report)
	}
	var handed string = filepath.Join(folder, "mira.riftweave.json")
	player.must("characters", "export", syncedID, handed)

	// The GM imports it and loads it into a session's party.
	gm.must("characters", "import", handed)
	var start map[string]any
	content, _ := os.ReadFile(syncedFixture)
	_ = json.Unmarshal(content, &start)
	input, _ := json.Marshal(map[string]any{"operation": "start", "start": start["data"]})
	code, out, stderr := gm.run(string(input), "session")
	if code != exitOK {
		t.Fatalf("session start: %s", stderr)
	}
	var started struct {
		Session map[string]any `json:"session"`
	}
	_ = json.Unmarshal([]byte(out), &started)
	var sessionFile string = filepath.Join(folder, "table.json")
	_ = os.WriteFile(sessionFile, files.Encode(started.Session), 0o644)
	gm.must("characters", "party", "add", sessionFile, syncedID)
	if code, _, stderr := gm.run("", "characters", "party", "add", sessionFile, syncedID); code != exitInput || !strings.Contains(stderr, "already in the party") {
		t.Errorf("a second add: %d %s", code, stderr)
	}

	// A session step carries the party through, and the GM's play changes the character (the
	// encounter tracker of BE_S08_07 will make this change through steps).
	var session map[string]any
	content, _ = os.ReadFile(sessionFile)
	_ = json.Unmarshal(content, &session)
	step, _ := json.Marshal(map[string]any{"operation": "step", "session": session, "step": map[string]any{"command": "derived", "input": map[string]any{}}})
	code, out, stderr = gm.run(string(step), "session")
	if code != exitOK {
		t.Fatalf("session step: %s", stderr)
	}
	var stepped struct {
		Session map[string]any `json:"session"`
	}
	_ = json.Unmarshal([]byte(out), &stepped)
	party, _ := stepped.Session["party"].([]any)
	if len(party) != 1 {
		t.Fatalf("the step dropped the party: %v", stepped.Session["party"])
	}
	var member = party[0].(map[string]any)
	if member["character"].(map[string]any)["revision"] != float64(3) {
		t.Errorf("the party does not keep the loaded revision: %v", member["character"])
	}
	member["actor"].(map[string]any)["resources"].(map[string]any)["hp"] = float64(12)
	_ = os.WriteFile(sessionFile, files.Encode(stepped.Session), 0o644)

	// The GM exports the party, one file per player: one revision higher, changed by the GM.
	var written struct {
		Written []string `json:"written"`
	}
	_ = json.Unmarshal([]byte(gm.must("characters", "party", "export", sessionFile, filepath.Join(folder, "out"))), &written)
	if len(written.Written) != 1 || !strings.HasSuffix(written.Written[0], "mira-vell-5f0c7a52.riftweave.json") {
		t.Fatalf("party export wrote %v", written.Written)
	}

	// The player imports it: a higher revision replaces the saved one.
	if report := reportOf(t, player.must("characters", "import", written.Written[0])); report["outcome"] != "replaced" {
		t.Fatalf("player re-import: %v", report)
	}
	var entries = player.list()
	if len(entries) != 1 || entries[0]["revision"] != float64(4) || entries[0]["changed_by"] != "gm" || entries[0]["name"] != "Mira Vell" || entries[0]["level"] != float64(3) || entries[0]["changed_at"] == "" {
		t.Fatalf("player list after the session: %v", entries)
	}

	// The file from before the session is a lower revision: refused, with a note.
	code, _, stderr = player.run("", "characters", "import", handed)
	if report := reportOf(t, stderr); code != exitRule || report["outcome"] != "refused" || !strings.Contains(report["note"].(string), "older than the saved revision 4") {
		t.Errorf("a lower revision: %d %s", code, stderr)
	}

	// The same revision with other content: both sides are shown, nothing changes until one is kept.
	var diverged map[string]any
	content, _ = os.ReadFile(written.Written[0])
	_ = json.Unmarshal(content, &diverged)
	diverged["data"].(map[string]any)["resources"].(map[string]any)["hp"] = float64(30)
	var divergedFile string = filepath.Join(folder, "diverged.riftweave.json")
	_ = os.WriteFile(divergedFile, files.Encode(diverged), 0o644)
	code, _, stderr = player.run("", "characters", "import", divergedFile)
	report := reportOf(t, stderr)
	if code != exitRule || report["outcome"] != "diverged" || report["saved"] == nil || report["incoming"] == nil {
		t.Fatalf("a diverged revision: %d %s", code, stderr)
	}
	if report := reportOf(t, player.must("characters", "import", divergedFile, "--keep", "saved")); report["outcome"] != "kept" {
		t.Errorf("keep saved: %v", report)
	}
	if report := reportOf(t, player.must("characters", "import", divergedFile, "--keep", "incoming")); report["outcome"] != "replaced" {
		t.Errorf("keep incoming: %v", report)
	}
	if report := reportOf(t, player.must("characters", "import", divergedFile)); report["outcome"] != "unchanged" {
		t.Errorf("the same file again: %v", report)
	}
}

// TestTheCharacterListOpensDuplicatesAndDeletes is the list's other actions, and a character
// file without sync fields (APP-30's) becoming a new character at revision 1.
func TestTheCharacterListOpensDuplicatesAndDeletes(t *testing.T) {
	var player device = newDevice(t)
	if entries := player.list(); len(entries) != 0 {
		t.Fatalf("a new device lists %v", entries)
	}
	var added map[string]any = reportOf(t, player.must("characters", "import", filesFixtures+"character.riftweave.json", "--name", "Tam"))
	var character = added["character"].(map[string]any)
	if added["outcome"] != "added" || character["name"] != "Tam" || character["revision"] != float64(1) || character["changed_by"] != "player" {
		t.Fatalf("a file without sync fields: %v", added)
	}
	var id string = character["id"].(string)
	var shown map[string]any = reportOf(t, player.must("characters", "show", id))
	if shown["data"].(map[string]any)["level"] != float64(3) {
		t.Errorf("show: %v", shown)
	}
	var copied map[string]any = reportOf(t, player.must("characters", "duplicate", id, "--as", "gm"))
	if copied["name"] != "Tam (copy)" || copied["id"] == id || copied["changed_by"] != "gm" {
		t.Errorf("duplicate: %v", copied)
	}
	if entries := player.list(); len(entries) != 2 || entries[0]["name"] != "Tam" || entries[1]["name"] != "Tam (copy)" {
		t.Errorf("list after duplicate: %v", entries)
	}
	player.must("characters", "delete", id)
	if entries := player.list(); len(entries) != 1 || entries[0]["name"] != "Tam (copy)" {
		t.Errorf("list after delete: %v", entries)
	}
	if code, _, stderr := player.run("", "characters", "show", id); code != exitInput || !strings.Contains(stderr, "no saved character") {
		t.Errorf("a deleted character: %d %s", code, stderr)
	}
	if code, _, stderr := player.run("", "characters", "import", filesFixtures+"spell.riftweave.json"); code != exitInput || !strings.Contains(stderr, "not a character") {
		t.Errorf("a spell file: %d %s", code, stderr)
	}
	if code, _, _ := player.run("", "characters", "import", filesFixtures+"missing-module.riftweave.json"); code != exitInput {
		t.Errorf("a missing module did not stop the import")
	}
	if report := reportOf(t, player.must("characters", "import", filesFixtures+"missing-module.riftweave.json", "--accept-loss")); report["outcome"] != "added" {
		t.Errorf("--accept-loss: %v", report)
	}
}

// TestAnUnchangedPartyMemberGoesOutAsLoaded: the GM's export raises only what the session changed.
func TestAnUnchangedPartyMemberGoesOutAsLoaded(t *testing.T) {
	var gm device = newDevice(t)
	var folder string = t.TempDir()
	gm.must("characters", "import", syncedFixture)
	content, _ := os.ReadFile(syncedFixture)
	var fixture map[string]any
	_ = json.Unmarshal(content, &fixture)
	input, _ := json.Marshal(map[string]any{"operation": "start", "start": fixture["data"]})
	_, out, _ := gm.run(string(input), "session")
	var started struct {
		Session map[string]any `json:"session"`
	}
	_ = json.Unmarshal([]byte(out), &started)
	var sessionFile string = filepath.Join(folder, "table.json")
	_ = os.WriteFile(sessionFile, files.Encode(started.Session), 0o644)
	if code, _, stderr := gm.run("", "characters", "party", "export", sessionFile, folder); code != exitInput || !strings.Contains(stderr, "no party") {
		t.Errorf("a session with no party: %d %s", code, stderr)
	}
	gm.must("characters", "party", "add", sessionFile, syncedID)
	var written struct {
		Written []string `json:"written"`
	}
	_ = json.Unmarshal([]byte(gm.must("characters", "party", "export", sessionFile, folder)), &written)
	exported, _ := os.ReadFile(written.Written[0])
	if !bytes.Equal(exported, content) {
		t.Errorf("an unchanged member is not the file it was loaded from:\n%s", exported)
	}
}
