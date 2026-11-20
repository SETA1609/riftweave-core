package play

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/characters"
)

// The Characters screen (BE_S07_03, docs/backend/app-layout.md#characters) is the character
// list: each saved character with its level, its last change, and who made it, then Import,
// Export the party, and Back. On a character, enter opens it as the character a session starts
// from, e exports it, d duplicates it, x deletes it, and p adds it to the open session's party.

const (
	importRow = "Import a character file"
	partyRow  = "Export the party of the open session"
)

// charactersShown lists the saved characters again; a list that cannot be read is the problem.
func (model Model) charactersShown() Model {
	model.saved = nil
	if model.library == nil {
		model.problem = "no character list on this device"
		return model
	}
	entries, err := model.library.List()
	if err != nil {
		model.problem = err.Error()
		return model
	}
	model.saved = entries
	model.cursor = min(model.cursor, len(model.characterRows())-1)
	return model
}

// characterRows are the screen's rows: one per saved character, then the actions and Back.
func (model Model) characterRows() []string {
	var rows []string
	for _, entry := range model.saved {
		if entry.Problem != "" {
			rows = append(rows, fmt.Sprintf("%-24s cannot be read: %s (x to delete)", entry.ID, entry.Problem))
			continue
		}
		var when string = entry.ChangedAt
		if parsed, err := time.Parse(time.RFC3339, entry.ChangedAt); err == nil {
			when = parsed.Local().Format("2006-01-02 15:04")
		}
		rows = append(rows, fmt.Sprintf("%-24s level %-3d %s by %s (revision %d)", entry.Name, entry.Level, when, entry.ChangedBy, entry.Revision))
	}
	return append(rows, importRow, partyRow, "Back")
}

// selected is the character under the cursor, if the cursor is on one.
func (model Model) selected() (characters.Entry, bool) {
	if model.cursor < len(model.saved) {
		return model.saved[model.cursor], true
	}
	return characters.Entry{}, false
}

// noted adds a line to the log, keeping the last few, and clears the problem.
func (model Model) noted(line string) Model {
	model.log = append(model.log, line)
	if len(model.log) > maximumLogLines {
		model.log = model.log[len(model.log)-maximumLogLines:]
	}
	model.problem = ""
	return model
}

// charactersKey is one key on the Characters screen.
func (model Model) charactersKey(key string) (Model, tea.Cmd) {
	var count int = len(model.characterRows())
	entry, onCharacter := model.selected()
	switch key {
	case "up", "k":
		model.cursor, model.problem = (model.cursor+count-1)%count, ""
	case "down", "j":
		model.cursor, model.problem = (model.cursor+1)%count, ""
	case "esc":
		model.mode, model.cursor, model.problem = modeMenu, model.returnCursor, ""
	case "enter":
		switch {
		case onCharacter:
			return model.openCharacter(entry), nil
		case model.cursor == len(model.saved):
			return model.importCharacter()
		case model.cursor == len(model.saved)+1:
			return model.exportParty()
		default:
			model.mode, model.cursor, model.problem = modeMenu, model.returnCursor, ""
		}
	case "e":
		if onCharacter {
			return model.prompt("Export "+entry.Name+" to", characters.FileName(entry.Character), func(model Model, path string) Model {
				encoded, err := model.library.Export(entry.ID)
				if err == nil {
					err = os.WriteFile(path, encoded, 0o644)
				}
				if err != nil {
					model.problem = err.Error()
					return model
				}
				return model.noted(fmt.Sprintf("exported %s (revision %d) to %s", entry.Name, entry.Revision, path))
			})
		}
	case "d":
		if onCharacter {
			copied, err := model.library.Duplicate(entry.ID, "player")
			if err != nil {
				model.problem = err.Error()
				return model, nil
			}
			return model.noted("duplicated " + entry.Name + " as " + copied.Name).charactersShown(), nil
		}
	case "x":
		if onCharacter {
			return model.prompt("Delete "+entry.Name+"? Type yes", "", func(model Model, answer string) Model {
				if answer != "yes" {
					return model.noted("kept " + entry.Name)
				}
				if err := model.library.Delete(entry.ID); err != nil {
					model.problem = err.Error()
					return model
				}
				return model.noted("deleted " + entry.Name).charactersShown()
			})
		}
	case "p":
		if onCharacter {
			return model.addToParty(entry), nil
		}
	}
	return model, nil
}

// openCharacter makes the character the one a new session starts from, as Import does.
func (model Model) openCharacter(entry characters.Entry) Model {
	saved, err := model.library.Open(entry.ID)
	if err != nil {
		model.problem = err.Error()
		return model
	}
	var actor map[string]any
	if json.Unmarshal(saved.Data, &actor) == nil {
		model.character = map[string]any{"attributes": actor["attributes"], "level": actor["level"]}
		if attributes, ok := actor["attributes"].(map[string]any); ok && attributes["lck"] != nil {
			model.luck = fmt.Sprint(attributes["lck"])
		}
	}
	return model.noted(fmt.Sprintf("opened %s, level %d, revision %d", entry.Name, entry.Level, entry.Revision))
}

// importCharacter asks for a file and applies it by the sync rule; a divergence asks which
// side to keep, and nothing changes until the person answers.
func (model Model) importCharacter() (Model, tea.Cmd) {
	if model.pending != nil {
		return model.prompt("Both sides changed: keep saved or incoming", "", func(model Model, side string) Model {
			return model.keepSide(side)
		})
	}
	return model.prompt("Import from (add --accept-loss to import without a missing module)", "", func(model Model, text string) Model {
		var path, accept = strings.CutSuffix(text, acceptLoss)
		content, err := os.ReadFile(path)
		if err != nil {
			model.problem = err.Error() + "; nothing changed"
			return model
		}
		var options = characters.ImportOptions{Name: strings.TrimSuffix(filepath.Base(path), ".riftweave.json"), ChangedBy: "player", AcceptLoss: accept}
		report, err := model.library.Import(content, options)
		if err != nil {
			model.problem = err.Error() + "; nothing changed"
			return model
		}
		if report.Outcome == characters.Diverged {
			model.pending, model.pendingOptions = content, options
			model.problem = fmt.Sprintf("%s: %s (saved: changed %s by %s; incoming: changed %s by %s); enter on Import to choose",
				report.Character.Name, report.Note, report.Saved.Character.ChangedAt, report.Saved.Character.ChangedBy, report.Incoming.Character.ChangedAt, report.Incoming.Character.ChangedBy)
			return model.charactersShown()
		}
		return model.reported(report)
	})
}

// reported logs what an import did and lists the characters again.
func (model Model) reported(report characters.Report) Model {
	var line string = fmt.Sprintf("%s: %s, revision %d", report.Outcome, report.Character.Name, report.Character.Revision)
	if report.Note != "" {
		line += " (" + report.Note + ")"
	}
	if len(report.Warnings) > 0 {
		line += " (" + strings.Join(report.Warnings, "; ") + ")"
	}
	model = model.noted(line).charactersShown()
	if report.Outcome == characters.Refused {
		model.problem = report.Note
	}
	return model
}

// keepSide answers a divergence with saved or incoming; anything else leaves both as they are.
func (model Model) keepSide(side string) Model {
	var content = model.pending
	model.pending = nil
	if side != "saved" && side != "incoming" {
		return model.noted("kept both sides as they were; nothing changed")
	}
	var options = model.pendingOptions
	options.Keep = side
	report, err := model.library.Import(content, options)
	if err != nil {
		model.problem = err.Error() + "; nothing changed"
		return model
	}
	return model.reported(report)
}

// addToParty loads the character into the open session's party (docs/backend/sessions.md#the-party).
func (model Model) addToParty(entry characters.Entry) Model {
	if model.session == nil {
		model.problem = "no session is open: import a session file, or start one in dev mode"
		return model
	}
	updated, err := model.library.PartyAdd(model.session, entry.ID)
	if err != nil {
		model.problem = err.Error() + "; nothing changed"
		return model
	}
	model.session = updated
	return model.noted(fmt.Sprintf("added %s (revision %d) to the party", entry.Name, entry.Revision))
}

// exportParty writes the open session's party, one file per player, into a directory.
func (model Model) exportParty() (Model, tea.Cmd) {
	if model.session == nil {
		model.problem = "no session is open: import a session file, or start one in dev mode"
		return model, nil
	}
	return model.prompt("Export the party into the directory", ".", func(model Model, directory string) Model {
		exported, updated, err := model.library.PartyExport(model.session)
		if err != nil {
			model.problem = err.Error() + "; nothing written"
			return model
		}
		var names []string
		for name, content := range exported {
			if err := os.WriteFile(filepath.Join(directory, name), content, 0o644); err != nil {
				model.problem = err.Error()
				return model
			}
			names = append(names, name)
		}
		// The session records the export, so the next one counts on from it.
		model.session = updated
		return model.noted(fmt.Sprintf("exported %d party files into %s", len(names), directory))
	})
}

// charactersView is the Characters screen.
func (model Model) charactersView() string {
	var view strings.Builder
	view.WriteString(titleStyle.Render("Characters") + "\n\n")
	if len(model.saved) == 0 {
		view.WriteString(dimStyle.Render("  no saved characters yet") + "\n")
	}
	for index, row := range model.characterRows() {
		var marker string = "  "
		if index == model.cursor {
			marker = "> "
		}
		view.WriteString(marker + row + "\n")
	}
	if model.problem != "" {
		view.WriteString("\n" + errorStyle.Render(model.problem) + "\n")
	}
	view.WriteString(model.logView())
	view.WriteString("\n" + hintStyle.Render("up/down; on a character: enter to open, e export, d duplicate, x delete, p add to the party; esc to go back") + "\n")
	return view.String()
}
