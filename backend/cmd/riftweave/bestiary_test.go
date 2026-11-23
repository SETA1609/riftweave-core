package main

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/core"
	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

// toBestiary is Start, then Bestiary (BE_S08_05).
var toBestiary = []string{"enter", "down", "down", "down", "down", "enter"}

// bestiaryWith opens the human CLI with the "Players see only what they know" box set.
func bestiaryWith(t *testing.T, onlyKnown bool) tea.Model {
	t.Helper()
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	var file settings.File = settings.Default()
	file["bestiary"] = map[string]any{"players_see_only_known": onlyKnown}
	return press(play.NewWith(opened.Run, play.Options{Settings: file}), toBestiary...)
}

// The GM's bestiary lists every monster with the table's mark, filters it by type, level, and
// phase, and marks a monster seen for the table as a session step.
func TestTheBestiaryListsFiltersAndMarks(t *testing.T) {
	var model tea.Model = bestiaryWith(t, true)
	var view string = model.View()
	for _, text := range []string{"Bestiary", "GM view", "Skeleton", "Bandit", "Table", "unknown"} {
		if !strings.Contains(view, text) {
			t.Fatalf("the GM's bestiary has no %q: %s", text, view)
		}
	}
	matchesTheAgentCLI(t, model, "bestiary")
	model = press(model, "/", "type:undead level:1 phase:water", "enter")
	if view = model.View(); !strings.Contains(view, "Skeleton") || strings.Contains(view, "Bandit") {
		t.Fatalf("filtered to level 1 water undead: %s", view)
	}
	model = press(model, "s")
	if view = model.View(); !strings.Contains(view, "seen") {
		t.Fatalf("the skeleton marked seen: %s", view)
	}
	var session struct {
		Bestiary struct {
			Knowledge []map[string]any `json:"knowledge"`
		} `json:"bestiary"`
		Log []map[string]any `json:"log"`
	}
	if err := json.Unmarshal(model.(play.Model).Session(), &session); err != nil {
		t.Fatalf("no session: %v", err)
	}
	if len(session.Log) != 1 || session.Log[0]["command"] != "bestiary" || len(session.Bestiary.Knowledge) != 1 ||
		session.Bestiary.Knowledge[0]["monster"] != "core:monster/skeleton" || session.Bestiary.Knowledge[0]["level"] != "seen" {
		t.Errorf("the mark is a session step: %+v", session)
	}
	model = press(model, "esc", "esc")
	if menu := model.(play.Model).Menu(); menu != "start" {
		t.Errorf("Esc clears the filter, then goes back to Start: menu %q, %s", menu, model.View())
	}
}

// The players' view follows the box: off, every monster shows; on, only what the table was
// shown, a seen monster in short.
func TestThePlayersViewFollowsTheBox(t *testing.T) {
	var on tea.Model = press(bestiaryWith(t, true), "s", "p")
	var view string = on.View()
	if !strings.Contains(view, "Players' view") || !strings.Contains(view, "Skeleton") || strings.Contains(view, "Bandit") {
		t.Fatalf("with the box on, players see only the skeleton: %s", view)
	}
	matchesTheAgentCLI(t, on, "bestiary")
	var off tea.Model = press(bestiaryWith(t, false), "p")
	if view = off.View(); !strings.Contains(view, "Skeleton") || !strings.Contains(view, "Bandit") {
		t.Errorf("with the box off, players see every monster: %s", view)
	}
}
