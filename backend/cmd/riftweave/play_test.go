package main

import (
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
