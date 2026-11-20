package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/files"
	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/core"
	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

// moduleVersions are the enabled modules of the settings file, each with the version its
// manifest beside the ruleset names (docs/backend/files.md#modules).
func moduleVersions(directory string, file settings.File) []files.Module {
	var enabled []string
	if list, ok := file["modules"].([]any); ok {
		for _, entry := range list {
			if id, ok := entry.(string); ok {
				enabled = append(enabled, id)
			}
		}
	}
	var modules []files.Module = []files.Module{}
	entries, _ := os.ReadDir(filepath.Join(directory, "..", "modules"))
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(directory, "..", "modules", entry.Name(), "manifest.json"))
		if err != nil {
			continue
		}
		var manifest files.Module
		if json.Unmarshal(content, &manifest) == nil && slices.Contains(enabled, manifest.ID) {
			modules = append(modules, manifest)
		}
	}
	return modules
}

// checkerFor is the check over the core and the ruleset's schemas beside its data directory.
func checkerFor(run files.Runner, directory string) files.Checker {
	return files.Checker{Run: run, Schemas: os.DirFS(filepath.Join(directory, "..", "schemas"))}
}

// fileContext is what export and import need: the core's runner, the settings, and the modules.
func fileContext(dataFlag string, getenv func(string) string) (files.Checker, string, settings.File, string, error) {
	directory, err := rulesetDirectory(dataFlag, getenv)
	if err != nil {
		return files.Checker{}, "", nil, "", err
	}
	opened, err := core.Open(directory)
	if err != nil {
		return files.Checker{}, "", nil, "", fmt.Errorf("loading %s: %v", directory, err)
	}
	path, err := settingsPath(getenv)
	if err != nil {
		return files.Checker{}, "", nil, "", fmt.Errorf("no settings directory: %v", err)
	}
	file, err := savedSettings(path)
	if err != nil {
		return files.Checker{}, "", nil, "", fmt.Errorf("%s: %v", path, err)
	}
	return checkerFor(opened.Run, directory), directory, file, path, nil
}

// exportCommand is `riftweave export <kind> [input]` (APP-30): the data from the input file or
// stdin, checked, in the envelope with these rules and the enabled modules, on stdout.
func exportCommand(args []string, dataFlag string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string) int {
	if len(args) < 1 || len(args) > 2 || !slices.Contains(files.Kinds, args[0]) {
		return fail(stderr, exitInput, "export takes a kind (character, session, spell) and an optional input file")
	}
	var data []byte
	var err error
	if len(args) == 2 {
		data, err = os.ReadFile(args[1])
	} else {
		data, err = io.ReadAll(stdin)
	}
	if err != nil {
		return fail(stderr, exitInput, "%v", err)
	}
	checker, directory, file, _, err := fileContext(dataFlag, getenv)
	if err != nil {
		return fail(stderr, exitRuleset, "%v", err)
	}
	encoded, err := files.Wrap(args[0], data, core.RulesVersion, moduleVersions(directory, file), checker)
	if err != nil {
		return fail(stderr, exitInput, "%v; nothing written", err)
	}
	_, _ = stdout.Write(encoded)
	return exitOK
}

// importCommand is `riftweave import <file> [--enable-modules | --accept-loss]` (APP-30): the
// file is checked before anything applies. A settings file goes to `settings import`. A file
// that names a module this device does not load stops with the module and the rows at stake,
// unless --enable-modules turns them on in the settings file or --accept-loss confirms the loss.
// It prints {kind, data, warnings} for the next command.
func importCommand(args []string, dataFlag string, stdout, stderr io.Writer, getenv func(string) string) int {
	if len(args) < 1 || len(args) > 2 || (len(args) == 2 && args[1] != "--enable-modules" && args[1] != "--accept-loss") {
		return fail(stderr, exitInput, "import takes a file and optionally --enable-modules or --accept-loss")
	}
	content, err := os.ReadFile(args[0])
	if err != nil {
		return fail(stderr, exitInput, "%v", err)
	}
	var kind struct {
		Riftweave string `json:"riftweave"`
	}
	if json.Unmarshal(content, &kind) == nil && kind.Riftweave == "settings" {
		return settingsCommand([]string{"import", args[0]}, stdout, stderr, getenv)
	}
	checker, directory, file, path, err := fileContext(dataFlag, getenv)
	if err != nil {
		return fail(stderr, exitRuleset, "%v", err)
	}
	read, err := files.Open(content, core.RulesVersion, moduleVersions(directory, file), checker)
	if err != nil {
		return fail(stderr, exitInput, "%s: %v; nothing changed", args[0], err)
	}
	if len(read.Missing) > 0 {
		switch {
		case len(args) == 2 && args[1] == "--enable-modules":
			if err := enableModules(read.Missing, directory, file, path); err != nil {
				return fail(stderr, exitInput, "%s: %v; nothing changed", args[0], err)
			}
			read.Warnings = append(read.Warnings, "enabled in the settings file: "+files.Alert(read.Missing))
		case len(args) == 2 && args[1] == "--accept-loss":
			read.Warnings = append(read.Warnings, "imported without its modules: "+files.Alert(read.Missing))
		default:
			body, _ := json.Marshal(map[string]any{"error": fmt.Sprintf("%s: %s; nothing changed", args[0], files.Alert(read.Missing)), "missing": read.Missing})
			fmt.Fprintf(stderr, "%s\n", body)
			return exitInput
		}
	}
	var printed = map[string]any{"kind": read.Kind, "data": read.Data, "warnings": read.Warnings}
	if read.Character != nil {
		printed["character"] = read.Character
	}
	body, _ := json.Marshal(printed)
	fmt.Fprintf(stdout, "%s\n", body)
	return exitOK
}

// enableModules turns on the missing modules in the settings file at path, when this ruleset
// has them (docs/backend/files.md#rules-both-apps-keep).
func enableModules(missing []files.Missing, directory string, file settings.File, path string) error {
	list, _ := file["modules"].([]any)
	for _, module := range missing {
		if !slices.Contains(moduleIDs(filepath.Join(directory, "..", "modules")), module.Module) {
			return fmt.Errorf("module %s is not in this ruleset", module.Module)
		}
		list = append(list, module.Module)
	}
	file["modules"] = list
	if _, err := settings.Parse(file.Encode()); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, file.Encode(), 0o644)
}

// playImport is the human CLI's Import (APP-30): the same check as `riftweave import`, and the
// alert when a module the file names is not loaded here.
func playImport(checker files.Checker, directory string, file settings.File) play.Importer {
	return func(path string, acceptLoss bool) (play.Imported, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return play.Imported{}, err
		}
		read, err := files.Open(content, core.RulesVersion, moduleVersions(directory, file), checker)
		if err != nil {
			return play.Imported{}, err
		}
		var imported = play.Imported{Kind: read.Kind, Data: read.Data, Warnings: read.Warnings}
		if len(read.Missing) > 0 && !acceptLoss {
			imported.Alert = files.Alert(read.Missing)
		}
		return imported, nil
	}
}
