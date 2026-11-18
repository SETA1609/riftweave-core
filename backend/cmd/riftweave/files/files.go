// Package files is the file both apps trade (APP-30, docs/backend/files.md): one envelope,
// {"riftweave": kind, "rules_version", "modules", "data"}, around exactly what the command
// contract already returns. It checks a file before it applies: the envelope, the data against
// the contract, and the modules the data needs. It holds no rule and stays out of the core.
package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/SETA1609/riftweave-core/backend/contract"
)

// Kinds are the kinds of file the envelope carries. The settings file keeps its own format,
// which already names its kind (APP-25, docs/backend/settings.md).
var Kinds = []string{"character", "session", "spell"}

// Module is a module a file was made with, or one this device loads: its manifest id and version.
type Module struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// Runner is core.Run: the spell check goes through the author command.
type Runner func(name string, input []byte) ([]byte, error)

// Checker is what a check needs: the core, for a spell, and ruleset/schemas, which the
// contract's definitions point into (common.json names schema.json).
type Checker struct {
	Run     Runner
	Schemas fs.FS
}

// Missing is a module the file names that this device does not load, and the rows of the data
// that need it, which would be lost.
type Missing struct {
	Module string   `json:"module"`
	Rows   []string `json:"rows"`
}

// Read is what a checked file holds, and what to tell the person before it applies.
type Read struct {
	Kind         string          `json:"kind"`
	RulesVersion string          `json:"rules_version"`
	Modules      []Module        `json:"modules"`
	Data         json.RawMessage `json:"data"`
	Warnings     []string        `json:"warnings"`
	Missing      []Missing       `json:"missing"`
}

// contractSchemas compiles the contract's definitions once; a kind names the one its data is.
var contractSchemas = map[string]string{
	"character": "common.json#/definitions/stored_actor",
	"session":   "common.json#/definitions/session",
}

// The contract's files sit at backend/contract/ and the ruleset's schemas at ruleset/schemas/,
// so common.json's ../../ruleset/schemas/schema.json resolves as it does in the repository.
const (
	contractBase = "https://riftweave.dev/backend/contract/"
	schemasBase  = "https://riftweave.dev/ruleset/schemas/"
)

func compile(kind string, schemas fs.FS) (*jsonschema.Schema, error) {
	var compiler *jsonschema.Compiler = jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7)
	for _, source := range []struct {
		files fs.FS
		base  string
	}{{contract.Files, contractBase}, {schemas, schemasBase}} {
		if source.files == nil {
			return nil, fmt.Errorf("no ruleset/schemas to check against")
		}
		names, err := fs.Glob(source.files, "*.json")
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			content, err := fs.ReadFile(source.files, name)
			if err != nil {
				return nil, err
			}
			document, err := jsonschema.UnmarshalJSON(bytes.NewReader(content))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			if err := compiler.AddResource(source.base+name, document); err != nil {
				return nil, err
			}
		}
	}
	return compiler.Compile(contractBase + contractSchemas[kind])
}

// Check is the data of one kind against the contract: a character is a stored actor and a
// session a session (common.json), and a spell is a row the author command accepts.
func Check(kind string, data json.RawMessage, checker Checker) error {
	if !slices.Contains(Kinds, kind) {
		return fmt.Errorf("%q is not a kind of file (%s)", kind, strings.Join(Kinds, ", "))
	}
	if kind == "spell" {
		return checkSpell(data, checker.Run)
	}
	schema, err := compile(kind, checker.Schemas)
	if err != nil {
		return fmt.Errorf("the contract does not compile: %v", err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("the %s is not JSON: %v", kind, err)
	}
	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("the %s does not match the contract: %v", kind, err)
	}
	return nil
}

// checkSpell runs the author command's check on the one row, and refuses it on a problem.
func checkSpell(data json.RawMessage, run Runner) error {
	if run == nil {
		return fmt.Errorf("a spell is checked by the author command, and no core is open")
	}
	input, _ := json.Marshal(map[string]any{"kind": "spell", "spells": []json.RawMessage{data}})
	output, err := run("author", input)
	if err != nil {
		return fmt.Errorf("the spell does not check: %s", output)
	}
	var result struct {
		Spells []struct {
			Problems []struct {
				Path    string `json:"path"`
				Message string `json:"message"`
			} `json:"problems"`
		} `json:"spells"`
	}
	if err := json.Unmarshal(output, &result); err != nil || len(result.Spells) != 1 {
		return fmt.Errorf("the author command answered %s", output)
	}
	var messages []string
	for _, problem := range result.Spells[0].Problems {
		messages = append(messages, strings.TrimSpace(problem.Path+" "+problem.Message))
	}
	if len(messages) > 0 {
		return fmt.Errorf("the spell does not check: %s", strings.Join(messages, "; "))
	}
	return nil
}

// Wrap checks the data and puts it in an envelope with the rules and modules it was made with.
func Wrap(kind string, data json.RawMessage, rulesVersion string, modules []Module, checker Checker) ([]byte, error) {
	if err := Check(kind, data, checker); err != nil {
		return nil, err
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	if modules == nil {
		modules = []Module{}
	}
	sorted := slices.Clone(modules)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	var list []any = []any{}
	for _, module := range sorted {
		list = append(list, map[string]any{"id": module.ID, "version": module.Version})
	}
	return Encode(map[string]any{"riftweave": kind, "rules_version": rulesVersion, "modules": list, "data": value}), nil
}

// Encode writes a file the one way both apps write it: keys sorted, two-space indent, a final
// newline, so a file exported by either app is byte for byte the same (as the settings file).
func Encode(value any) []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
	return buffer.Bytes()
}

// Open reads a file and checks it before anything applies: the envelope, the data against the
// contract, the rules version (a difference warns), and the modules (a module this device does
// not load is reported with the rows that need it).
func Open(content []byte, rulesVersion string, here []Module, checker Checker) (Read, error) {
	var envelope struct {
		Riftweave    string          `json:"riftweave"`
		RulesVersion *string         `json:"rules_version"`
		Modules      *[]Module       `json:"modules"`
		Data         json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(content, &envelope); err != nil {
		return Read{}, fmt.Errorf("not a Riftweave file: it is not a JSON object")
	}
	if !slices.Contains(Kinds, envelope.Riftweave) {
		return Read{}, fmt.Errorf("not a Riftweave file: \"riftweave\" is %q, not one of %s", envelope.Riftweave, strings.Join(Kinds, ", "))
	}
	if envelope.RulesVersion == nil || envelope.Modules == nil || len(envelope.Data) == 0 {
		return Read{}, fmt.Errorf("not a Riftweave file: it needs rules_version, modules, and data")
	}
	if err := Check(envelope.Riftweave, envelope.Data, checker); err != nil {
		return Read{}, err
	}
	var read Read = Read{Kind: envelope.Riftweave, RulesVersion: *envelope.RulesVersion, Modules: *envelope.Modules, Data: envelope.Data, Warnings: []string{}, Missing: []Missing{}}
	if read.RulesVersion != rulesVersion {
		read.Warnings = append(read.Warnings, fmt.Sprintf("made with rules %s; this app runs %s", read.RulesVersion, rulesVersion))
	}
	var loaded = map[string]string{}
	for _, module := range here {
		loaded[module.ID] = module.Version
	}
	var rows map[string][]string = moduleRows(envelope.Data)
	for _, module := range read.Modules {
		version, ok := loaded[module.ID]
		switch {
		case !ok:
			found := rows[module.ID]
			if found == nil {
				found = []string{}
			}
			read.Missing = append(read.Missing, Missing{Module: module.ID, Rows: found})
		case version != module.Version:
			read.Warnings = append(read.Warnings, fmt.Sprintf("module %s is %s here; the file was made with %s", module.ID, version, module.Version))
		}
	}
	return read, nil
}

// moduleRows are the module refs in the data, by module id, sorted: the rows a device without
// that module would lose.
func moduleRows(data json.RawMessage) map[string][]string {
	var value any
	_ = json.Unmarshal(data, &value)
	var found = map[string][]string{}
	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case string:
			if rest, ok := strings.CutPrefix(typed, "module:"); ok {
				id, _, _ := strings.Cut(rest, "/")
				if !slices.Contains(found[id], typed) {
					found[id] = append(found[id], typed)
				}
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		case map[string]any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	for id := range found {
		slices.Sort(found[id])
	}
	return found
}

// Alert is the message for the missing modules: each module and the rows at stake.
func Alert(missing []Missing) string {
	var parts []string
	for _, module := range missing {
		var rows string = "no rows named"
		if len(module.Rows) > 0 {
			rows = strings.Join(module.Rows, ", ")
		}
		parts = append(parts, fmt.Sprintf("module %s is not enabled here (%s)", module.Module, rows))
	}
	return strings.Join(parts, "; ") + ": enable it, or confirm the loss"
}
