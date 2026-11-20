package play

import (
	"encoding/json"
	"io/fs"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/contract"
)

// mainMenu is the parent key of the main menu's entries in backend/contract/menu.json.
const mainMenu = "main"

// menuEntry is one row of backend/contract/menu.json (docs/backend/app-layout.md): its key,
// the menu it sits in, its label, and the issue that adds it while that issue is not merged.
type menuEntry struct {
	Key    string   `json:"key"`
	Parent string   `json:"parent"`
	Label  string   `json:"label"`
	Issue  string   `json:"issue"`
	Apps   []string `json:"apps"` // the apps that show it, cli or web; both when empty
	// IssueApps are the apps the issue is pending in, when it is one app's half; both when empty.
	IssueApps []string `json:"issue_apps"`
}

// backKey is the entry the app adds at the end of every menu but the main one.
const backKey = "back"

// menuTree reads the menu tree both apps show from the embedded contract.
func menuTree() []menuEntry {
	content, err := fs.ReadFile(contract.Files, "menu.json")
	if err != nil {
		panic("backend/contract/menu.json is not embedded: " + err.Error())
	}
	var file struct {
		Entries []menuEntry `json:"entries"`
	}
	if err := json.Unmarshal(content, &file); err != nil {
		panic("backend/contract/menu.json: " + err.Error())
	}
	return file.Entries
}

var tree = menuTree()

// entriesOf is the menu's entries in file order, with Back added below a submenu.
func entriesOf(parent string) []menuEntry {
	var entries []menuEntry
	for _, entry := range tree {
		if entry.Parent == parent && (len(entry.Apps) == 0 || slices.Contains(entry.Apps, "cli")) {
			if len(entry.IssueApps) > 0 && !slices.Contains(entry.IssueApps, "cli") {
				entry.Issue = "" // pending in the other app only
			}
			entries = append(entries, entry)
		}
	}
	if parent != mainMenu {
		entries = append(entries, menuEntry{Key: backKey, Parent: parent, Label: "Back"})
	}
	return entries
}

// parentOf is the menu that holds the menu, the main menu for the main menu itself.
func parentOf(menu string) string {
	for _, entry := range tree {
		if entry.Key == menu {
			return entry.Parent
		}
	}
	return mainMenu
}

// titleOf is the label of the menu, "Riftweave" for the main menu.
func titleOf(menu string) string {
	for _, entry := range tree {
		if entry.Key == menu {
			return entry.Label
		}
	}
	return "Riftweave"
}
