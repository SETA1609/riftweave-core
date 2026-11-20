package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/characters"
	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/files"
	"github.com/SETA1609/riftweave-core/backend/core"
	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

const charactersUsage = `characters list | show <id> | export <id> [file] | duplicate <id> [--as gm|player] | delete <id>
characters import <file> [--name <name>] [--as gm|player] [--keep saved|incoming] [--enable-modules | --accept-loss]
characters party add <session file> <id> | party export <session file> <directory>`

// charactersDirectory is where the human CLI keeps the character list (docs/backend/files.md#the-character-list):
// os.UserConfigDir()/riftweave/characters, or RIFTWEAVE_CONFIG_DIR/characters when set.
func charactersDirectory(getenv func(string) string) (string, error) {
	path, err := settingsPath(getenv)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "characters"), nil
}

// libraryFor is the character list of this device, checked by the core and these rules.
func libraryFor(checker files.Checker, directory string, file settings.File, getenv func(string) string) (characters.Library, error) {
	folder, err := charactersDirectory(getenv)
	if err != nil {
		return characters.Library{}, fmt.Errorf("no settings directory: %v", err)
	}
	return characters.Library{Directory: folder, Checker: checker, RulesVersion: core.RulesVersion, Modules: moduleVersions(directory, file)}, nil
}

// printJSON writes one JSON value and a newline.
func printJSON(stdout io.Writer, value any) {
	body, _ := json.Marshal(value)
	fmt.Fprintf(stdout, "%s\n", body)
}

// flagValues splits the arguments into the positional ones and the --name value pairs and
// --switches the command allows.
func flagValues(args []string, valued, switches []string) ([]string, map[string]string, error) {
	var positional []string
	var values = map[string]string{}
	for index := 0; index < len(args); index++ {
		var argument string = args[index]
		switch {
		case slices.Contains(valued, argument):
			if index+1 >= len(args) {
				return nil, nil, fmt.Errorf("%s needs a value", argument)
			}
			values[argument], index = args[index+1], index+1
		case slices.Contains(switches, argument):
			values[argument] = "yes"
		case len(argument) > 2 && argument[:2] == "--":
			return nil, nil, fmt.Errorf("unknown option %s", argument)
		default:
			positional = append(positional, argument)
		}
	}
	return positional, values, nil
}

// charactersCommand is `riftweave characters ...` (BE_S07_03, docs/backend/files.md#in-the-cli-1):
// the character list, the sync rule on import, and a session's party.
func charactersCommand(args []string, dataFlag string, stdout, stderr io.Writer, getenv func(string) string) int {
	if len(args) == 0 {
		return fail(stderr, exitInput, "usage: riftweave %s", charactersUsage)
	}
	checker, directory, file, settingsFile, err := fileContext(dataFlag, getenv)
	if err != nil {
		return fail(stderr, exitRuleset, "%v", err)
	}
	library, err := libraryFor(checker, directory, file, getenv)
	if err != nil {
		return fail(stderr, exitInput, "%v", err)
	}
	positional, options, err := flagValues(args[1:], []string{"--name", "--as", "--keep"}, []string{"--enable-modules", "--accept-loss"})
	if err != nil {
		return fail(stderr, exitInput, "%v", err)
	}
	if by, given := options["--as"]; given && !slices.Contains(files.ChangedBy, by) {
		return fail(stderr, exitInput, "--as takes gm or player, not %q", by)
	}
	var wants = func(count int) bool { return len(positional) == count }
	switch {
	case args[0] == "list" && wants(0):
		entries, err := library.List()
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		printJSON(stdout, entries)
	case args[0] == "show" && wants(1):
		saved, err := library.Open(positional[0])
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		printJSON(stdout, saved)
	case args[0] == "export" && (wants(1) || wants(2)):
		encoded, err := library.Export(positional[0])
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		if wants(1) {
			_, _ = stdout.Write(encoded)
			return exitOK
		}
		if err := os.WriteFile(positional[1], encoded, 0o644); err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
	case args[0] == "duplicate" && wants(1):
		copied, err := library.Duplicate(positional[0], options["--as"])
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		printJSON(stdout, copied)
	case args[0] == "delete" && wants(1):
		if err := library.Delete(positional[0]); err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		printJSON(stdout, map[string]string{"deleted": positional[0]})
	case args[0] == "import" && wants(1):
		return charactersImport(library, positional[0], options, directory, file, settingsFile, stdout, stderr)
	case args[0] == "party" && len(positional) == 3 && positional[0] == "add":
		session, err := os.ReadFile(positional[1])
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		updated, err := library.PartyAdd(session, positional[2])
		if err != nil {
			return fail(stderr, exitInput, "%v; nothing changed", err)
		}
		if err := os.WriteFile(positional[1], updated, 0o644); err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		printJSON(stdout, map[string]string{"added": positional[2], "session": positional[1]})
	case args[0] == "party" && len(positional) == 3 && positional[0] == "export":
		session, err := os.ReadFile(positional[1])
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		exported, err := library.PartyExport(session)
		if err != nil {
			return fail(stderr, exitInput, "%v; nothing written", err)
		}
		written, err := writeAll(positional[2], exported)
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		printJSON(stdout, map[string][]string{"written": written})
	default:
		return fail(stderr, exitInput, "usage: riftweave %s", charactersUsage)
	}
	return exitOK
}

// writeAll writes each file into the directory and returns the paths, sorted.
func writeAll(directory string, contents map[string][]byte) ([]string, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	var written []string = []string{}
	for name, content := range contents {
		var path string = filepath.Join(directory, name)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return nil, err
		}
		written = append(written, path)
	}
	sort.Strings(written)
	return written, nil
}

// charactersImport is `riftweave characters import`: the file's check, then the sync rule. A
// refused or diverged import exits 1 with the report on stderr and changes nothing; a missing
// module stops it as `riftweave import` does.
func charactersImport(library characters.Library, path string, options map[string]string, directory string, file settings.File, settingsFile string, stdout, stderr io.Writer) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return fail(stderr, exitInput, "%v", err)
	}
	var choices = characters.ImportOptions{Name: options["--name"], ChangedBy: options["--as"], Keep: options["--keep"], AcceptLoss: options["--accept-loss"] != ""}
	report, err := library.Import(content, choices)
	var missing *characters.MissingModulesError
	if errors.As(err, &missing) && options["--enable-modules"] != "" {
		if err := enableModules(missing.Missing, directory, file, settingsFile); err != nil {
			return fail(stderr, exitInput, "%s: %v; nothing changed", path, err)
		}
		library.Modules = moduleVersions(directory, file) // enableModules turned them on in file
		var enabled string = "enabled in the settings file: " + files.Alert(missing.Missing)
		if report, err = library.Import(content, choices); err == nil {
			report.Warnings = append(report.Warnings, enabled)
		}
	}
	switch {
	case errors.As(err, &missing):
		printJSON(stderr, map[string]any{"error": fmt.Sprintf("%s: %s; nothing changed", path, missing.Error()), "missing": missing.Missing})
		return exitInput
	case err != nil:
		return fail(stderr, exitInput, "%s: %v; nothing changed", path, err)
	case report.Outcome == characters.Refused || report.Outcome == characters.Diverged:
		printJSON(stderr, report)
		return exitRule
	}
	printJSON(stdout, report)
	return exitOK
}
