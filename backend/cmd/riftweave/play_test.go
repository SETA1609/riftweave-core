package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/files"
	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/core"
	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

// press sends keys to the model: "enter", "down", "up", "left", "right", "esc", "clear", or text to type.
func press(model tea.Model, keys ...string) tea.Model {
	for _, key := range keys {
		var message tea.KeyMsg
		switch key {
		case "enter":
			message = tea.KeyMsg{Type: tea.KeyEnter}
		case "down":
			message = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			message = tea.KeyMsg{Type: tea.KeyUp}
		case "left":
			message = tea.KeyMsg{Type: tea.KeyLeft}
		case "right":
			message = tea.KeyMsg{Type: tea.KeyRight}
		case "esc":
			message = tea.KeyMsg{Type: tea.KeyEsc}
		case "clear":
			message = tea.KeyMsg{Type: tea.KeyCtrlU}
		default:
			message = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		model, _ = model.Update(message)
	}
	return model
}

func started(t *testing.T) tea.Model {
	t.Helper()
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	return play.New(opened.Run)
}

// matchesTheAgentCLI runs the model's last input through the agent CLI and compares.
func matchesTheAgentCLI(t *testing.T, model tea.Model, command string) {
	t.Helper()
	input, output := model.(play.Model).LastCall()
	if len(input) == 0 {
		t.Fatalf("no command ran: %s", model.View())
	}
	code, stdout, stderr := call(t, string(input), "--data", dataDirectory, command)
	if code != exitOK {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if !reflect.DeepEqual(decode(t, stdout), decode(t, string(output))) {
		t.Errorf("agent CLI %s, human CLI %s", stdout, output)
	}
}

// The paths through the menus of backend/contract/menu.json: Start, then an entry.
var (
	toCreation   = []string{"enter", "down", "enter"}                 // Start, Character creation
	toAttack     = []string{"enter", "enter", "enter"}                // Start, GM tools, Resolve an attack
	toResolution = []string{"enter", "down", "down", "down", "enter"} // Start, Skill resolution
)

// createBrunna walks the creation screen (BE_S07_08): a hill dwarf scout with every point spent,
// Small Frame, and three tag skills. Rows: name, race, background, the eight attributes, creation
// perks, tag skills, finish.
func createBrunna(model tea.Model) tea.Model {
	model = press(press(model, toCreation...), "Brunna", "down")
	for range 8 { // none, then human ... hill dwarf, the eighth playable race
		model = press(model, "right")
	}
	model = press(model, "down", "right", "down") // background: scout
	for _, spent := range []int{2, 1, 2, 0, 0, 3, 0, 2} {
		for range spent {
			model = press(model, "+")
		}
		model = press(model, "down")
	}
	return press(model, "small_frame", "down", "survival stealth bows", "down", "enter")
}

func TestPlayCreatesACharacterOverCheckAndFinish(t *testing.T) {
	var model tea.Model = createBrunna(started(t))
	matchesTheAgentCLI(t, model, "creation")
	input, _ := model.(play.Model).LastCall()
	if !strings.Contains(string(input), `"operation":"finish"`) {
		t.Errorf("the last call is not a finish: %s", input)
	}
	if !strings.Contains(model.View(), "created: Brunna, level 1, 83 HP") {
		t.Errorf("no log line: %s", model.View())
	}
}

func TestPlayCreationShowsThePointsLeftAndTheProblems(t *testing.T) {
	var model tea.Model = press(started(t), toCreation...)
	if !strings.Contains(model.View(), "10 points left") {
		t.Errorf("no points left shown: %s", model.View())
	}
	model = press(model, "down", "down", "down", "down", "down", "down") // the INT row
	for range 7 {
		model = press(model, "+")
	}
	if !strings.Contains(model.View(), "int is 11, over the maximum of 10") {
		t.Errorf("no problem shown: %s", model.View())
	}
	model = press(model, "-", "-", "-")
	if !strings.Contains(model.View(), "6 points left") {
		t.Errorf("a removed point is not given back: %s", model.View())
	}
}

func TestPlayUsesTheCreatedCharactersLuckForAnAttack(t *testing.T) {
	// Brunna's LCK is 6 (4 + 2 points); back to the Start menu, then GM tools and the attack.
	var model tea.Model = press(createBrunna(started(t)), "up", "enter", "enter")
	for range 8 {
		model = press(model, "enter")
	}
	input, _ := model.(play.Model).LastCall()
	if !strings.Contains(string(input), `"lck":6`) {
		t.Errorf("attack input %s", input)
	}
}

func TestPlayRefusesANumberThatIsNotOne(t *testing.T) {
	var model tea.Model = press(press(started(t), toAttack...), "clear", "x")
	for range 8 {
		model = press(model, "enter")
	}
	if !strings.Contains(model.View(), "Attacker skill must be a whole number") {
		t.Errorf("no problem shown: %s", model.View())
	}
}

func TestPlayNeedsATerminal(t *testing.T) {
	code, stdout, stderr := call(t, "", "--data", dataDirectory, "play")
	if code != exitInput || stdout != "" || !strings.Contains(stderr, "needs a terminal") {
		t.Errorf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestCharmStaysOutOfTheCore(t *testing.T) {
	output, err := exec.Command("go", "list", "-deps", "../../core", "../../internal/...").CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	if strings.Contains(string(output), "charmbracelet") {
		t.Error("the core imports Charm")
	}
}

// TestTelemetryStaysOutOfTheCore keeps the core pure (APP-43, docs/backend/observability.md): the
// adapters observe the calls they make, and no OpenTelemetry package reaches the core or
// internal/.
func TestTelemetryStaysOutOfTheCore(t *testing.T) {
	output, err := exec.Command("go", "list", "-deps", "../../core", "../../internal/...").CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	if strings.Contains(string(output), "go.opentelemetry.io") {
		t.Error("the core imports an OpenTelemetry package")
	}
}

func startedDev(t *testing.T) tea.Model {
	t.Helper()
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	opened.EnableDev()
	return play.NewDev(opened.Run)
}

// devLine opens the command line with ":", types the command, and runs it.
func devLine(model tea.Model, line string) tea.Model {
	return press(model, ":", line, "enter")
}

func TestPlayDevCommandsMatchTheAgentCLI(t *testing.T) {
	var model tea.Model = startedDev(t)
	for _, line := range []string{
		"give core:consumable/potion_healing 2",
		"inflict blinded 1",
		"inflict core:effect/fortify_attribute 2 core:origin/arcane 3",
		"level 4",
	} {
		model = devLine(model, line)
	}
	var human map[string]any = decode(t, string(model.(play.Model).Session())).(map[string]any)
	if human["dev"] != true {
		t.Fatalf("the panel's session is not marked dev: %v", human)
	}
	start, _ := json.Marshal(map[string]any{"operation": "start", "start": human["start"]})
	code, stdout, stderr := call(t, string(start), "--data", dataDirectory, "session")
	if code != exitOK {
		t.Fatalf("session start: exit %d: %s", code, stderr)
	}
	var session any = decode(t, stdout).(map[string]any)["session"]
	for _, operation := range []map[string]any{
		{"operation": "item", "base": "core:consumable/potion_healing", "quantity": 2},
		{"operation": "condition", "key": "blinded", "stacks": 1},
		{"operation": "effect", "effect": "core:effect/fortify_attribute", "magnitude": 2, "origin": "core:origin/arcane", "turns": 3},
		{"operation": "level", "set": 4},
	} {
		operation["session"] = session
		body, _ := json.Marshal(operation)
		code, stdout, stderr = call(t, string(body), "--data", dataDirectory, "--dev", "dev")
		if code != exitOK {
			t.Fatalf("dev %v: exit %d: %s", operation["operation"], code, stderr)
		}
		session = decode(t, stdout).(map[string]any)["session"]
	}
	if !reflect.DeepEqual(human, session) {
		t.Errorf("human CLI session %v\nagent CLI session %v", human, session)
	}
	if view := model.View(); !strings.Contains(view, "Dev") || !strings.Contains(view, "level 4") {
		t.Errorf("no dev panel with the level: %s", view)
	}
}

func TestPlayRollQueueFillsTheNextRollPrompts(t *testing.T) {
	var model tea.Model = devLine(startedDev(t), "roll 77 3")
	model = press(model, toAttack...)
	for range 8 {
		model = press(model, "enter")
	}
	input, _ := model.(play.Model).LastCall()
	if !strings.Contains(string(input), `"roll":77`) {
		t.Errorf("first attack after :roll 77 3: %s", input)
	}
	model = press(model, "enter", "enter") // back to the menu, where the cursor is still on the attack
	for range 8 {
		model = press(model, "enter")
	}
	if input, _ = model.(play.Model).LastCall(); !strings.Contains(string(input), `"roll":3`) {
		t.Errorf("second attack takes the second queued roll: %s", input)
	}
}

func TestPlayDevRefusesABadCommandAndHasNoPanelWithoutDev(t *testing.T) {
	var model tea.Model = devLine(startedDev(t), "inflict no_such_condition 1")
	if !strings.Contains(model.View(), "condition") {
		t.Errorf("no refusal shown: %s", model.View())
	}
	if view := press(started(t), ":").View(); strings.Contains(view, "Dev") {
		t.Errorf("a dev panel without --dev: %s", view)
	}
}

// withSettings is the model with a settings file in a temporary directory, saved on each change.
func withSettings(t *testing.T, remote func(string) play.Runner) (tea.Model, string) {
	t.Helper()
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	var path string = filepath.Join(t.TempDir(), "settings.json")
	return play.NewWith(opened.Run, play.Options{
		Settings: settings.Default(),
		Save:     func(file settings.File) error { return os.WriteFile(path, file.Encode(), 0o644) },
		Modules:  moduleIDs(filepath.Join(dataDirectory, "..", "modules")),
		Remote:   remote,
	}), path
}

func saved(t *testing.T, path string) settings.File {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, err := settings.Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestPlayOpensOnTheMainMenu(t *testing.T) {
	var view string = started(t).View()
	for _, label := range []string{"Start", "Characters", "Settings", "Modules", "Network", "Quit"} {
		if !strings.Contains(view, label) {
			t.Errorf("main menu has no %q: %s", label, view)
		}
	}
}

func TestPlayMenusShowEveryEntryAndPendingIssues(t *testing.T) {
	var tools string = press(started(t), "enter").View()
	for _, label := range []string{"GM tools", "Character creation", "Spell creation (APP-22)", "Skill resolution", "Bestiary", "Back"} {
		if !strings.Contains(tools, label) {
			t.Errorf("tools menu has no %q: %s", label, tools)
		}
	}
	var gm string = press(started(t), "enter", "enter").View()
	for _, label := range []string{"Resolve an attack", "Encounter tracker (APP-34)", "Back"} {
		if !strings.Contains(gm, label) {
			t.Errorf("GM tools has no %q: %s", label, gm)
		}
	}
	for keys, issue := range map[string]string{"enter down down enter": "APP-22", "enter enter down enter": "APP-34"} {
		if view := press(started(t), strings.Fields(keys)...).View(); !strings.Contains(view, "comes with "+issue) {
			t.Errorf("%s: no pending issue %s shown: %s", keys, issue, view)
		}
	}
}

func TestPlayEscGoesBackOneMenu(t *testing.T) {
	var model tea.Model = press(started(t), "enter", "enter")
	if menu := model.(play.Model).Menu(); menu != "gm_tools" {
		t.Fatalf("menu %s, want gm_tools", menu)
	}
	model = press(model, "esc")
	if menu := model.(play.Model).Menu(); menu != "start" || !strings.Contains(model.View(), "> GM tools") {
		t.Errorf("one esc: menu %s, %s", menu, model.View())
	}
	model = press(model, "esc")
	if menu := model.(play.Model).Menu(); menu != "main" || !strings.Contains(model.View(), "> Start") {
		t.Errorf("two esc: menu %s, %s", menu, model.View())
	}
	model = press(model, "down", "down", "enter", "esc") // into Settings and out
	if !strings.Contains(model.View(), "> Settings") {
		t.Errorf("esc from Settings: %s", model.View())
	}
}

func TestPlaySettingsAreSavedInTheSettingsFile(t *testing.T) {
	model, path := withSettings(t, nil)
	model = press(model, "down", "down", "enter")         // Settings
	model = press(model, "enter", "clear", "de", "enter") // Language
	model = press(model, "down", "enter")                 // Theme: system to light
	model = press(model, "down", "enter")                 // Bestiary: players see only known
	var file settings.File = saved(t, path)
	if file["locale"] != "de" || file["theme"] != "light" || file["bestiary"].(map[string]any)["players_see_only_known"] != true {
		t.Errorf("saved %s", file.Encode())
	}
	model = press(model, "up", "up", "enter", "clear", "not a code", "enter")
	if !strings.Contains(model.View(), "not a language code") || saved(t, path)["locale"] != "de" {
		t.Errorf("a bad locale was not refused: %s", model.View())
	}
}

func TestPlaySettingsExportAndImport(t *testing.T) {
	model, path := withSettings(t, nil)
	var exported string = filepath.Join(t.TempDir(), "exported.json")
	model = press(model, "down", "down", "enter", "down", "enter") // Settings, Theme to light
	model = press(model, "down", "down", "enter", "clear", exported, "enter")
	content, err := os.ReadFile(exported)
	if err != nil || !strings.Contains(string(content), `"theme": "light"`) {
		t.Fatalf("export: %v %s", err, content)
	}
	var other string = filepath.Join(t.TempDir(), "other.json")
	_ = os.WriteFile(other, []byte(`{"riftweave": "settings", "version": 1, "theme": "dark", "kept": 1}`+"\n"), 0o644)
	model = press(model, "down", "enter", "clear", other, "enter")
	if file := saved(t, path); file["theme"] != "dark" || file["kept"] == nil {
		t.Errorf("import: saved %s, view %s", file.Encode(), model.View())
	}
}

func TestPlayModulesAreCheckBoxesWithCoreLocked(t *testing.T) {
	model, path := withSettings(t, nil)
	model = press(model, "down", "down", "down", "enter") // Modules
	if view := model.View(); !strings.Contains(view, "[x] core (always loaded)") || !strings.Contains(view, "[ ] factions") {
		t.Fatalf("modules screen: %s", view)
	}
	model = press(model, "enter")
	if !strings.Contains(model.View(), "core is always loaded") {
		t.Errorf("core could be changed: %s", model.View())
	}
	model = press(model, "down", "down", "enter") // example_module, then factions
	if modules := saved(t, path)["modules"]; !reflect.DeepEqual(modules, []any{"factions"}) {
		t.Errorf("saved modules %v", modules)
	}
	if !strings.Contains(model.View(), "[x] factions") {
		t.Errorf("no check on factions: %s", model.View())
	}
}

func TestPlayServerModeSendsCommandsToTheServer(t *testing.T) {
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	var served []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var name string = strings.TrimPrefix(request.URL.Path, "/v1/commands/")
		served = append(served, request.Method+" "+name)
		body, _ := io.ReadAll(request.Body)
		output, err := opened.Run(name, body)
		if err != nil {
			writer.WriteHeader(http.StatusUnprocessableEntity)
		}
		_, _ = writer.Write(output)
	}))
	defer server.Close()
	model, path := withSettings(t, func(url string) play.Runner { return play.Remote(url, server.Client()) })
	model = press(model, "down", "down", "down", "down", "enter") // Network
	model = press(model, "down", "enter")
	if !strings.Contains(model.View(), "set the server URL first") {
		t.Errorf("server mode without a URL: %s", model.View())
	}
	model = press(model, "down", "enter", "clear", server.URL, "enter", "up", "enter", "esc")
	if network := saved(t, path)["network"].(map[string]any); network["mode"] != "server" || network["server_url"] != server.URL {
		t.Fatalf("saved network %v", network)
	}
	model = press(model, "up", "up", "up", "up")
	model = press(model, toResolution...)
	model = press(model, "enter", "enter", "enter", "enter", "clear", "30", "enter")
	if !reflect.DeepEqual(served, []string{"POST resolution"}) {
		t.Fatalf("the server got %v: %s", served, model.View())
	}
	matchesTheAgentCLI(t, model, "resolution")
}

func TestPlaySkillResolutionMatchesTheAgentCLI(t *testing.T) {
	var model tea.Model = press(press(started(t), toResolution...), "clear", "blades", "enter", "clear", "55", "enter", "enter", "clear", "10", "enter", "clear", "61", "enter")
	matchesTheAgentCLI(t, model, "resolution")
	input, _ := model.(play.Model).LastCall()
	if !strings.Contains(string(input), `"value":10`) || !strings.Contains(string(input), `"roll":61`) {
		t.Errorf("input %s", input)
	}
	if !strings.Contains(model.View(), "blades check: rolled 61 against 65, success") {
		t.Errorf("no log line: %s", model.View())
	}
}

func TestPlaySkillResolutionRollsABlankRoll(t *testing.T) {
	var model tea.Model = press(started(t), toResolution...)
	for range 5 {
		model = press(model, "enter")
	}
	input, _ := model.(play.Model).LastCall()
	var fields struct {
		Roll int64 `json:"roll"`
	}
	if err := json.Unmarshal(input, &fields); err != nil || fields.Roll < 1 || fields.Roll > 100 {
		t.Errorf("a blank roll gave %s", input)
	}
	matchesTheAgentCLI(t, model, "resolution")
}

// TestPlayImportsAFileFromTheMainMenu is APP-30 in the human CLI: Import, a path, and the
// file's character becomes the character a dev session starts from; a missing module stops it.
func TestPlayImportsAFileFromTheMainMenu(t *testing.T) {
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	var options = play.Options{Import: playImport(checkerFor(opened.Run, dataDirectory), dataDirectory, settings.Default())}
	var model tea.Model = play.NewWith(opened.Run, options)
	model = press(model, "down", "down", "down", "down", "down", "enter", "../../../ruleset/tests/fixtures/files/character.riftweave.json", "enter")
	if view := model.View(); !strings.Contains(view, "imported a character file") {
		t.Errorf("no import line: %s", view)
	}
	// The cursor is still on Import.
	model = press(model, "enter", "../../../ruleset/tests/fixtures/files/missing-module.riftweave.json", "enter")
	if view := model.View(); !strings.Contains(view, "module factions is not enabled here") {
		t.Errorf("no module alert: %s", view)
	}
}

// TestPlayListsImportsAndAddsCharactersToTheParty is BE_S07_03 in the human CLI: Characters lists
// the saved ones with level, last change, and who made it; Import applies the sync rule and asks
// on a divergence; p adds a character to the open session's party; d duplicates and x deletes.
func TestPlayListsImportsAndAddsCharactersToTheParty(t *testing.T) {
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	var environment = fileEnvironment(t.TempDir())
	library, err := libraryFor(checkerFor(opened.Run, dataDirectory), dataDirectory, settings.Default(), environment)
	if err != nil {
		t.Fatal(err)
	}
	var options = play.Options{Import: playImport(checkerFor(opened.Run, dataDirectory), dataDirectory, settings.Default()), Characters: &library}
	var model tea.Model = play.NewWith(opened.Run, options)
	model = press(model, "down", "enter")
	if view := model.View(); !strings.Contains(view, "no saved characters yet") || !strings.Contains(view, "Import a character file") {
		t.Fatalf("an empty list: %s", view)
	}
	// Import the player's file: it is listed with its level, its last change, and who made it.
	model = press(model, "enter", syncedFixture, "enter")
	if view := model.View(); !strings.Contains(view, "added: Mira Vell, revision 3") || !strings.Contains(view, "level 3") || !strings.Contains(view, "by player (revision 3)") {
		t.Fatalf("after import: %s", view)
	}
	// A diverged file (same revision, other content) asks which side to keep.
	content, _ := os.ReadFile(syncedFixture)
	var diverged map[string]any
	_ = json.Unmarshal(content, &diverged)
	diverged["data"].(map[string]any)["xp"] = float64(900)
	var divergedFile string = filepath.Join(t.TempDir(), "diverged.riftweave.json")
	_ = os.WriteFile(divergedFile, files.Encode(diverged), 0o644)
	model = press(model, "down", "enter", divergedFile, "enter")
	if view := model.View(); !strings.Contains(view, "both sides changed revision 3") {
		t.Fatalf("no divergence shown: %s", view)
	}
	model = press(model, "enter", "incoming", "enter")
	if view := model.View(); !strings.Contains(view, "replaced: Mira Vell, revision 3") {
		t.Fatalf("keeping the incoming side: %s", view)
	}
	// Open it: it becomes the character a session starts from.
	model = press(model, "up", "enter")
	if view := model.View(); !strings.Contains(view, "opened Mira Vell, level 3, revision 3") {
		t.Fatalf("open: %s", view)
	}
	// p needs an open session; one imported from a file takes the character into its party.
	model = press(model, "p")
	if view := model.View(); !strings.Contains(view, "no session is open") {
		t.Errorf("p without a session: %s", view)
	}
	model = press(model, "esc", "down", "down", "down", "down", "enter", filesFixtures+"session.riftweave.json", "enter", "up", "up", "up", "up", "enter", "p")
	if view := model.View(); !strings.Contains(view, "added Mira Vell (revision 3) to the party") {
		t.Fatalf("party add: %s", view)
	}
	if !strings.Contains(string(model.(play.Model).Session()), `"party"`) {
		t.Errorf("the session has no party: %s", model.(play.Model).Session())
	}
	// Duplicate, then delete the copy.
	model = press(model, "d")
	if view := model.View(); !strings.Contains(view, "Mira Vell (copy)") {
		t.Fatalf("duplicate: %s", view)
	}
	model = press(model, "down", "x", "yes", "enter")
	if view := model.View(); !strings.Contains(view, "deleted Mira Vell (copy)") {
		t.Errorf("delete: %s", view)
	}
	if entries, _ := library.List(); len(entries) != 1 {
		t.Errorf("after delete the list holds %d", len(entries))
	}
}
