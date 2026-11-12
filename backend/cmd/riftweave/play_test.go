package main

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/core"
)

// press sends keys to the model: "enter", "down", "clear", or text to type.
func press(model tea.Model, keys ...string) tea.Model {
	for _, key := range keys {
		var message tea.KeyMsg
		switch key {
		case "enter":
			message = tea.KeyMsg{Type: tea.KeyEnter}
		case "down":
			message = tea.KeyMsg{Type: tea.KeyDown}
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

func TestPlayCreatesACharacterLikeTheAgentCLI(t *testing.T) {
	var model tea.Model = press(started(t), "enter", "clear", "9")
	for range 9 {
		model = press(model, "enter")
	}
	matchesTheAgentCLI(t, model, "derived")
	if !strings.Contains(model.View(), "created: level 1") {
		t.Errorf("no log line: %s", model.View())
	}
}

func TestPlayResolvesAnAttackLikeTheAgentCLI(t *testing.T) {
	var model tea.Model = press(started(t), "down", "enter", "clear", "60", "enter", "clear", "33")
	for range 7 {
		model = press(model, "enter")
	}
	matchesTheAgentCLI(t, model, "combat")
	if !strings.Contains(model.View(), "attack: rolled 33 against 60") {
		t.Errorf("no log line: %s", model.View())
	}
}

func TestPlayUsesTheCreatedCharactersLuckForAnAttack(t *testing.T) {
	var model tea.Model = press(started(t), "enter")
	for range 7 {
		model = press(model, "enter")
	}
	model = press(model, "clear", "8", "enter", "enter", "enter", "down", "enter")
	for range 8 {
		model = press(model, "enter")
	}
	input, _ := model.(play.Model).LastCall()
	if !strings.Contains(string(input), `"lck":8`) {
		t.Errorf("attack input %s", input)
	}
}

func TestPlayRefusesANumberThatIsNotOne(t *testing.T) {
	var model tea.Model = press(started(t), "down", "enter", "clear", "x")
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
	var human map[string]any = decode(t, string(model.(play.Model).Session()))
	if human["dev"] != true {
		t.Fatalf("the panel's session is not marked dev: %v", human)
	}
	start, _ := json.Marshal(map[string]any{"operation": "start", "start": human["start"]})
	code, stdout, stderr := call(t, string(start), "--data", dataDirectory, "session")
	if code != exitOK {
		t.Fatalf("session start: exit %d: %s", code, stderr)
	}
	var session any = decode(t, stdout)["session"]
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
		session = decode(t, stdout)["session"]
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
	model = press(model, "down", "enter")
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
