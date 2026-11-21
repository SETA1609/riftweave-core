package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/core"
	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

// printedNumber is the number a data- attribute names on a printed page.
func printedNumber(t *testing.T, page, attribute, name string) string {
	t.Helper()
	match := regexp.MustCompile(`data-` + attribute + `="` + regexp.QuoteMeta(name) + `"[^>]*>\s*(-?\d+)`).FindStringSubmatch(page)
	if match == nil {
		t.Fatalf("no data-%s=%q on the page", attribute, name)
	}
	return match[1]
}

// TestPlayPrintsASavedCharacter is BE_S09_03 in the human CLI: w on a character writes its sheet
// as HTML to a file and says where; the sheet is the character as saved (HP 41 of the fixture,
// burning), or with --fresh at full resources and with no condition.
func TestPlayPrintsASavedCharacter(t *testing.T) {
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	library, err := libraryFor(checkerFor(opened.Run, dataDirectory), dataDirectory, settings.Default(), fileEnvironment(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	var model tea.Model = play.NewWith(opened.Run, play.Options{Characters: &library})
	model = press(model, "down", "enter", "enter", syncedFixture, "enter", "up")
	var directory string = t.TempDir()
	var current string = filepath.Join(directory, "mira.html")
	model = press(model, "w", "clear", current, "enter")
	if view := model.View(); !strings.Contains(view, "printed Mira Vell (current) to "+current) {
		t.Fatalf("print: %s", view)
	}
	page, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if got := printedNumber(t, string(page), "current", "hp"); got != "41" || !strings.Contains(string(page), `data-condition="burning"`) {
		t.Errorf("the current sheet: hp %s, %s", got, page)
	}
	matchesTheAgentCLI(t, model, "print")
	var fresh string = filepath.Join(directory, "mira-fresh.html")
	model = press(model, "w", "clear", fresh+" --fresh", "enter")
	if view := model.View(); !strings.Contains(view, "printed Mira Vell (fresh) to "+fresh) {
		t.Fatalf("print --fresh: %s", view)
	}
	page, _ = os.ReadFile(fresh)
	if current, maximum := printedNumber(t, string(page), "current", "hp"), printedNumber(t, string(page), "stat", "hp"); current != maximum || strings.Contains(string(page), "data-condition=") {
		t.Errorf("the fresh sheet: hp %s of %s, %s", current, maximum, page)
	}
}
