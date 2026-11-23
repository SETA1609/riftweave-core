package play

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/characters"
)

// Print a saved character (BE_S09_03, docs/backend/print.md#printing-a-character): w on the
// character list asks for a file, runs the `print` command on the character as saved, and writes
// the page there. --fresh after the path prints it at full resources with no condition.

// printCommand is the command the action runs.
const printCommand = "print"

// The sheet states the action sends (core.SheetState).
const (
	sheetCurrent = "current"
	sheetFresh   = "fresh"
)

// freshFlag after the path asks for a fresh sheet.
const freshFlag = " --fresh"

// printCharacter asks where to write the character's sheet.
func (model Model) printCharacter(entry characters.Entry) (Model, tea.Cmd) {
	var initial string = strings.TrimSuffix(characters.FileName(entry.Character), ".riftweave.json") + ".html"
	return model.prompt("Print "+entry.Name+" to (add --fresh for full resources)", initial, func(model Model, text string) Model {
		var path, fresh = strings.CutSuffix(text, freshFlag)
		var state string = sheetCurrent
		if fresh {
			state = sheetFresh
		}
		saved, err := model.library.Open(entry.ID)
		if err != nil {
			model.problem = err.Error()
			return model
		}
		input, _ := json.Marshal(map[string]any{"kind": "sheet", "actor": saved.Data, "state": state})
		output, err := model.run(printCommand, input)
		model.lastInput, model.lastOutput = input, output
		if err != nil {
			model.problem = err.Error()
			return model
		}
		var page struct {
			HTML string `json:"html"`
		}
		if err := json.Unmarshal(output, &page); err != nil {
			model.problem = err.Error()
			return model
		}
		if err := os.WriteFile(path, []byte(page.HTML), 0o644); err != nil {
			model.problem = err.Error()
			return model
		}
		return model.noted(fmt.Sprintf("printed %s (%s) to %s", entry.Name, state, path))
	})
}
