// Package settings reads, checks, and writes the settings file both apps trade (APP-25,
// docs/backend/settings.md, backend/contract/settings.schema.json). A file keeps every key it
// holds, the ones this app does not use included, so an import followed by an export gives the
// same file back.
package settings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
)

// Version is the file format this app reads and writes.
const Version = 1

var (
	localePattern   = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
	modulePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	endpointPattern = regexp.MustCompile(`^https?://`)
	themes          = []string{"light", "dark", "system"}
	networkModes    = []string{"local", "server", "room"}
	telemetryModes  = []string{"off", "local", "export"}
)

// File is one settings file, every key as read.
type File map[string]any

// Default is the settings of a fresh install: English, the system's colours, no optional
// modules, offline, and no telemetry.
func Default() File {
	return File{"riftweave": "settings", "version": float64(Version), "locale": "en", "theme": "system", "modules": []any{}, "network": map[string]any{"mode": "local"}, "telemetry": map[string]any{"mode": "off"}}
}

// Parse reads a settings file and checks it against the schema's rules; a file that breaks
// one is refused with a message naming it.
func Parse(content []byte) (File, error) {
	var file File
	decoder := json.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&file); err != nil || file == nil {
		return nil, fmt.Errorf("not a settings file: it is not a JSON object")
	}
	if file["riftweave"] != "settings" {
		return nil, fmt.Errorf(`not a settings file: "riftweave" is %v, not "settings"`, file["riftweave"])
	}
	if version, _ := file["version"].(float64); version != Version {
		return nil, fmt.Errorf("settings version %v is not one this app reads (it reads %d)", file["version"], Version)
	}
	if locale, present := file["locale"]; present {
		if text, ok := locale.(string); !ok || !localePattern.MatchString(text) {
			return nil, fmt.Errorf("locale %v is not a language code", locale)
		}
	}
	if theme, present := file["theme"]; present {
		if text, ok := theme.(string); !ok || !slices.Contains(themes, text) {
			return nil, fmt.Errorf("theme %v is not light, dark, or system", theme)
		}
	}
	if modules, present := file["modules"]; present {
		list, ok := modules.([]any)
		if !ok {
			return nil, fmt.Errorf("modules is not a list")
		}
		var seen []string
		for _, module := range list {
			text, ok := module.(string)
			switch {
			case !ok || !modulePattern.MatchString(text):
				return nil, fmt.Errorf("module %v is not a module id", module)
			case text == "core":
				return nil, fmt.Errorf("modules lists core, which is always loaded and cannot be turned off")
			case slices.Contains(seen, text):
				return nil, fmt.Errorf("modules lists %s twice", text)
			}
			seen = append(seen, text)
		}
	}
	if network, present := file["network"]; present {
		object, ok := network.(map[string]any)
		mode, _ := object["mode"].(string)
		if !ok || !slices.Contains(networkModes, mode) {
			return nil, fmt.Errorf("network.mode is not local, server, or room")
		}
		for _, field := range []string{"server_url", "room_code"} {
			if value, present := object[field]; present {
				if _, ok := value.(string); !ok {
					return nil, fmt.Errorf("network.%s is not text", field)
				}
			}
		}
	}
	if bestiary, present := file["bestiary"]; present {
		object, ok := bestiary.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("bestiary is not an object")
		}
		if value, present := object["players_see_only_known"]; present {
			if _, ok := value.(bool); !ok {
				return nil, fmt.Errorf("bestiary.players_see_only_known is not true or false")
			}
		}
	}
	if telemetry, present := file["telemetry"]; present {
		object, ok := telemetry.(map[string]any)
		mode, _ := object["mode"].(string)
		if !ok || !slices.Contains(telemetryModes, mode) {
			return nil, fmt.Errorf("telemetry.mode is not off, local, or export")
		}
		if endpoint, present := object["endpoint"]; present {
			if text, ok := endpoint.(string); !ok || !endpointPattern.MatchString(text) {
				return nil, fmt.Errorf("telemetry.endpoint is not an http or https URL")
			}
		}
	}
	return file, nil
}

// Encode writes the file the one way both apps write it: keys sorted, two-space indent, a
// final newline, so a file exported by either app is byte for byte the same.
func (file File) Encode() []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(map[string]any(file)) // a map from Parse or Default always encodes
	return buffer.Bytes()
}
