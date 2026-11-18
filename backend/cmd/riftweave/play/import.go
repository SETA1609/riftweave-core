package play

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Imported is a file the Import entry read and checked (APP-30, docs/backend/files.md): its kind,
// its data, what to tell the person, and, when a module it names is not loaded here, the alert.
type Imported struct {
	Kind     string
	Data     json.RawMessage
	Warnings []string
	Alert    string
}

// Importer reads a file at a path; with acceptLoss it goes on past missing modules.
type Importer func(path string, acceptLoss bool) (Imported, error)

// acceptLoss is what the person adds to the path to import a file without its modules.
const acceptLoss = " --accept-loss"

// importFile opens the prompt for a file, then reads it: a character becomes the character a
// new dev session starts from, a session the dev panel's session, and a spell is checked only
// (its screen is APP-22). A missing module stops it with the module and the rows at stake.
func (model Model) importFile() (Model, tea.Cmd) {
	return model.prompt("Import from (add --accept-loss to import without a missing module)", "", func(model Model, text string) Model {
		var path, accept = strings.CutSuffix(text, acceptLoss)
		if model.importer == nil || path == "" {
			model.problem = "no file to import"
			return model
		}
		read, err := model.importer(path, accept)
		if err != nil {
			model.problem = err.Error() + "; nothing changed"
			return model
		}
		if read.Alert != "" {
			model.problem = read.Alert + "; nothing changed (import again with --accept-loss to go on)"
			return model
		}
		switch read.Kind {
		case "character":
			var actor map[string]any
			if json.Unmarshal(read.Data, &actor) == nil {
				model.character = map[string]any{"attributes": actor["attributes"], "level": actor["level"]}
				if attributes, ok := actor["attributes"].(map[string]any); ok && attributes["lck"] != nil {
					model.luck = fmt.Sprint(attributes["lck"])
				}
			}
		case "session":
			model.session = read.Data
		}
		model.problem = ""
		var line string = "imported a " + read.Kind + " file"
		if len(read.Warnings) > 0 {
			line += " (" + strings.Join(read.Warnings, "; ") + ")"
		}
		model.log = append(model.log, line)
		return model
	})
}
