package play

import (
	"fmt"
	"os"
	"slices"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

// The settings screens (docs/backend/app-layout.md#settings-modules-and-network) edit the
// APP-25 file. Every change is checked by the settings rules and saved at once; a change the
// rules refuse is shown and nothing is saved.

var themes = []string{"system", "light", "dark"}

// copyOf is a deep copy of the file, so an edit the rules refuse leaves the saved one as it was.
func copyOf(file settings.File) settings.File {
	copied, err := settings.Parse(file.Encode())
	if err != nil {
		return settings.Default()
	}
	return copied
}

func objectOf(file settings.File, key string) map[string]any {
	object, _ := file[key].(map[string]any)
	if object == nil {
		object = map[string]any{}
		file[key] = object
	}
	return object
}

func textOf(file settings.File, key, fallback string) string {
	if text, ok := file[key].(string); ok {
		return text
	}
	return fallback
}

// networkMode is local, server, or room; local when the file names none.
func networkMode(file settings.File) string {
	mode, _ := objectOf(copyOf(file), "network")["mode"].(string)
	if mode == "" {
		return "local"
	}
	return mode
}

func serverURL(file settings.File) string {
	url, _ := objectOf(copyOf(file), "network")["server_url"].(string)
	return url
}

// enabledModules is settings.modules as text.
func enabledModules(file settings.File) []string {
	var keys []string
	list, _ := file["modules"].([]any)
	for _, module := range list {
		if text, ok := module.(string); ok {
			keys = append(keys, text)
		}
	}
	return keys
}

// apply checks the edited file, saves it, and makes it the model's settings: the runner follows
// the network mode and the colours follow the theme.
func (model Model) apply(edited settings.File) Model {
	checked, err := settings.Parse(edited.Encode())
	if err != nil {
		model.problem = err.Error() + "; nothing changed"
		return model
	}
	if model.save != nil {
		if err := model.save(checked); err != nil {
			model.problem = "saving the settings: " + err.Error()
			return model
		}
	}
	model.settings, model.problem = checked, ""
	return model.follow()
}

// follow sets the runner from the network mode and the colours from the theme.
func (model Model) follow() Model {
	model.run = model.local
	if networkMode(model.settings) == "server" && serverURL(model.settings) != "" && model.remote != nil {
		model.run = model.remote(serverURL(model.settings))
	}
	switch textOf(model.settings, "theme", "system") {
	case "light":
		lipgloss.SetHasDarkBackground(false)
	case "dark":
		lipgloss.SetHasDarkBackground(true)
	}
	return model
}

// settingsRows are the settings screen's rows.
func (model Model) settingsRows() []string {
	var known bool
	if bestiary, ok := model.settings["bestiary"].(map[string]any); ok {
		known, _ = bestiary["players_see_only_known"].(bool)
	}
	return []string{
		"Language: " + textOf(model.settings, "locale", "en"),
		"Theme: " + textOf(model.settings, "theme", "system"),
		"Players see only the monsters they know: " + yesNo(known),
		"Export the settings to a file",
		"Import the settings from a file",
		"Back",
	}
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// moduleRows are core, locked, and each module in ruleset/modules with its check box.
func (model Model) moduleRows() []string {
	var rows = []string{"[x] core (always loaded)"}
	enabled := enabledModules(model.settings)
	for _, module := range model.modules {
		box := "[ ]"
		if slices.Contains(enabled, module) {
			box = "[x]"
		}
		rows = append(rows, box+" "+module)
	}
	return append(rows, "Back")
}

func (model Model) networkRows() []string {
	var mode = networkMode(model.settings)
	var radio = func(name string) string {
		if mode == name {
			return "(o)"
		}
		return "( )"
	}
	url := serverURL(model.settings)
	if url == "" {
		url = "not set"
	}
	return []string{
		radio("local") + " Local: the core in this program, offline",
		radio("server") + " Server: the HTTP API at the server URL",
		"Server URL: " + url,
		dimStyle.Render("( ) Room (APP-17)"),
		"Back",
	}
}

func (model Model) rows() []string {
	switch model.mode {
	case modeSettings:
		return model.settingsRows()
	case modeModules:
		return model.moduleRows()
	case modeNetwork:
		return model.networkRows()
	}
	return nil
}

// prompt opens a one-line input for a setting; done applies the typed text.
func (model Model) prompt(label, initial string, done func(Model, string) Model) (Model, tea.Cmd) {
	model.line = textinput.New()
	model.line.Prompt = label + ": "
	model.line.SetValue(initial)
	model.line.Focus()
	model.promptReturn, model.promptDone, model.mode = model.mode, done, modePrompt
	return model, textinput.Blink
}

// choose acts on the row under the cursor of a settings screen.
func (model Model) choose() (Model, tea.Cmd) {
	var rows = model.rows()
	if model.cursor == len(rows)-1 {
		model.mode, model.cursor, model.problem = modeMenu, model.returnCursor, ""
		return model, nil
	}
	edited := copyOf(model.settings)
	switch model.mode {
	case modeSettings:
		switch model.cursor {
		case 0:
			return model.prompt("Language code", textOf(model.settings, "locale", "en"), func(model Model, text string) Model {
				edited := copyOf(model.settings)
				edited["locale"] = text
				return model.apply(edited)
			})
		case 1:
			current := slices.Index(themes, textOf(model.settings, "theme", "system"))
			edited["theme"] = themes[(current+1)%len(themes)]
		case 2:
			bestiary := objectOf(edited, "bestiary")
			known, _ := bestiary["players_see_only_known"].(bool)
			bestiary["players_see_only_known"] = !known
		case 3:
			return model.prompt("Export to", "riftweave-settings.json", func(model Model, path string) Model {
				if err := os.WriteFile(path, model.settings.Encode(), 0o644); err != nil {
					model.problem = err.Error()
					return model
				}
				return model.logged("settings exported to " + path)
			})
		case 4:
			return model.prompt("Import from", "riftweave-settings.json", func(model Model, path string) Model {
				content, err := os.ReadFile(path)
				if err != nil {
					model.problem = err.Error()
					return model
				}
				imported, err := settings.Parse(content)
				if err != nil {
					model.problem = fmt.Sprintf("%s: %v; nothing changed", path, err)
					return model
				}
				model = model.apply(imported)
				if model.problem != "" {
					return model
				}
				return model.logged("settings imported from " + path)
			})
		}
	case modeModules:
		if model.cursor == 0 {
			model.problem = "core is always loaded"
			return model, nil
		}
		module := model.modules[model.cursor-1]
		enabled := enabledModules(edited)
		if index := slices.Index(enabled, module); index >= 0 {
			enabled = slices.Delete(enabled, index, index+1)
		} else {
			enabled = append(enabled, module)
		}
		list := make([]any, 0, len(enabled))
		for _, key := range enabled {
			list = append(list, key)
		}
		edited["modules"] = list
	case modeNetwork:
		network := objectOf(edited, "network")
		switch model.cursor {
		case 0:
			network["mode"] = "local"
		case 1:
			if serverURL(model.settings) == "" {
				model.problem = "set the server URL first"
				return model, nil
			}
			network["mode"] = "server"
		case 2:
			return model.prompt("Server URL", serverURL(model.settings), func(model Model, url string) Model {
				edited := copyOf(model.settings)
				objectOf(edited, "network")["server_url"] = url
				return model.apply(edited)
			})
		case 3:
			model.problem = "a shared room comes with APP-17"
			return model, nil
		}
	}
	return model.apply(edited), nil
}

// screenView is a settings screen: its rows with the cursor.
func (model Model) screenView(title string) string {
	var view = titleStyle.Render(title) + "\n\n"
	for index, row := range model.rows() {
		marker := "  "
		if index == model.cursor {
			marker = "> "
		}
		view += marker + row + "\n"
	}
	return view
}
