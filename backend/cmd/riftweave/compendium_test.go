package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
)

// toCompendium is Start, then Compendium (BE_S08_10).
var toCompendium = []string{"enter", "down", "down", "down", "down", "down", "enter"}

// The Compendium lists the collections, opens a table, filters it, opens a row, and Esc goes
// back one step at a time to the menu it came from.
func TestTheCompendiumListsFiltersAndOpensARow(t *testing.T) {
	var model tea.Model = press(started(t), toCompendium...)
	var view string = model.View()
	for _, label := range []string{"Compendium", "Weapons (", "Perks and traits (", "Monsters (", "Effects ("} {
		if !strings.Contains(view, label) {
			t.Fatalf("collections have no %q: %s", label, view)
		}
	}
	model = press(model, "down", "down", "down", "down", "down", "down", "down", "down", "enter") // Spells
	if view = model.View(); !strings.Contains(view, "Fire Bolt") || !strings.Contains(view, "Color") {
		t.Fatalf("the spell table: %s", view)
	}
	matchesTheAgentCLI(t, model, "catalog")
	model = press(model, "/", "fire bolt", "enter")
	if view = model.View(); !strings.Contains(view, `Spells matching "fire bolt"`) || !strings.Contains(view, "Fire Bolt") || strings.Contains(view, "Frost") {
		t.Fatalf("the filtered table: %s", view)
	}
	model = press(model, "enter")
	if view = model.View(); !strings.Contains(view, "core:spell/fire_bolt from core") || !strings.Contains(view, `"color": "red"`) {
		t.Fatalf("the opened row: %s", view)
	}
	model = press(model, "esc", "esc", "esc")
	if menu := model.(play.Model).Menu(); menu != "start" || !strings.Contains(model.View(), "Compendium") {
		t.Errorf("Esc goes back to Start: menu %q, %s", menu, model.View())
	}
}
