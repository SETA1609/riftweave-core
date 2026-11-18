package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/internal/settings"
)

// settingsPath is where the human CLI keeps its settings (docs/backend/settings.md):
// os.UserConfigDir()/riftweave/settings.json, or RIFTWEAVE_CONFIG_DIR/settings.json when set.
func settingsPath(getenv func(string) string) (string, error) {
	if directory := getenv("RIFTWEAVE_CONFIG_DIR"); directory != "" {
		return filepath.Join(directory, "settings.json"), nil
	}
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "riftweave", "settings.json"), nil
}

// savedSettings is the CLI's settings file, or the defaults when there is none yet.
func savedSettings(path string) (settings.File, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return settings.Default(), nil
	}
	if err != nil {
		return nil, err
	}
	return settings.Parse(content)
}

// settingsCommand is `riftweave settings export [file]` and `riftweave settings import <file>`
// (APP-25): the same file the Angular app exports and imports.
func settingsCommand(args []string, stdout, stderr io.Writer, getenv func(string) string) int {
	path, err := settingsPath(getenv)
	if err != nil {
		return fail(stderr, exitInput, "no settings directory: %v", err)
	}
	switch {
	case len(args) >= 1 && len(args) <= 2 && args[0] == "export":
		file, err := savedSettings(path)
		if err != nil {
			return fail(stderr, exitInput, "%s: %v", path, err)
		}
		if len(args) == 1 {
			_, _ = stdout.Write(file.Encode())
			return exitOK
		}
		if err := os.WriteFile(args[1], file.Encode(), 0o644); err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
	case len(args) == 2 && args[0] == "import":
		content, err := os.ReadFile(args[1])
		if err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		file, err := settings.Parse(content)
		if err != nil {
			return fail(stderr, exitInput, "%s: %v; nothing changed", args[1], err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
		if err := os.WriteFile(path, file.Encode(), 0o644); err != nil {
			return fail(stderr, exitInput, "%v", err)
		}
	default:
		return fail(stderr, exitInput, "settings takes export [file] or import <file>")
	}
	body, _ := json.Marshal(map[string]string{"settings": path})
	fmt.Fprintf(stdout, "%s\n", body)
	return exitOK
}

// playOptions are the human CLI's settings file, how to save it, and the modules beside the
// ruleset directory (docs/backend/app-layout.md).
func playOptions(directory string, getenv func(string) string) (play.Options, error) {
	path, err := settingsPath(getenv)
	if err != nil {
		return play.Options{}, fmt.Errorf("no settings directory: %v", err)
	}
	file, err := savedSettings(path)
	if err != nil {
		return play.Options{}, fmt.Errorf("%s: %v", path, err)
	}
	return play.Options{
		Settings: file,
		Save: func(file settings.File) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return os.WriteFile(path, file.Encode(), 0o644)
		},
		Modules: moduleIDs(filepath.Join(directory, "..", "modules")),
	}, nil
}

// moduleIDs are the ids in the manifests of the module directories, sorted.
func moduleIDs(parent string) []string {
	var ids []string
	entries, _ := os.ReadDir(parent)
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(parent, entry.Name(), "manifest.json"))
		if err != nil {
			continue
		}
		var manifest struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(content, &manifest) == nil && manifest.ID != "" {
			ids = append(ids, manifest.ID)
		}
	}
	slices.Sort(ids)
	return ids
}
