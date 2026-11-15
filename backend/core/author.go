package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/spellcraft"
)

// The `author` command (docs/backend/authoring.md, APP-10): check a draft spell as validate.py
// would, give the spellcraft cost, load a spell, and export a data file. It adds no rule.

type authorExport struct {
	Shape  string `json:"shape"`
	Module string `json:"module"`
}

type authorInput struct {
	Kind      string            `json:"kind"`
	Operation string            `json:"operation"`
	Spells    []json.RawMessage `json:"spells"`
	Spell     string            `json:"spell"`
	Export    *authorExport     `json:"export"`
}

type authorFinding struct {
	Check   string `json:"check"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type authorCost struct {
	Mana           int64 `json:"mana"`
	CastThreshold  int64 `json:"cast_threshold"`
	CraftThreshold int64 `json:"craft_threshold"`
}

type authoredSpell struct {
	Key      string          `json:"key"`
	Row      json.RawMessage `json:"row"`
	Problems []authorFinding `json:"problems"`
	Warnings []authorFinding `json:"warnings"`
	Cost     *authorCost     `json:"cost"`
}

type authorResultExport struct {
	Shape    string         `json:"shape"`
	Path     string         `json:"path"`
	File     map[string]any `json:"file"`
	Manifest map[string]any `json:"manifest,omitempty"`
	Refs     []string       `json:"refs"`
}

type authorResult struct {
	Spells []authoredSpell     `json:"spells,omitempty"`
	Export *authorResultExport `json:"export,omitempty"`
	Row    json.RawMessage     `json:"row,omitempty"`
}

func init() { commands["author"] = bind(runAuthor) }

func runAuthor(ruleset *data.Ruleset, input authorInput) (authorResult, error) {
	if input.Kind != "spell" {
		return authorResult{}, &data.NotDefinedError{Kind: "author kind", Key: input.Kind}
	}
	switch input.Operation {
	case "load":
		row, err := loadSpellRow(ruleset, input.Spell)
		return authorResult{Row: row}, err
	case "", "check":
	default:
		return authorResult{}, &data.NotDefinedError{Kind: "author operation", Key: input.Operation}
	}
	if len(input.Spells) == 0 {
		return authorResult{}, &data.MissingInputError{Subject: "author check", Needs: "spells"}
	}
	schema, err := spellSchema(ruleset)
	if err != nil {
		return authorResult{}, err
	}
	var refs map[string]bool = refIndex(ruleset)
	var result authorResult
	for _, draft := range input.Spells {
		result.Spells = append(result.Spells, checkSpell(ruleset, schema, refs, draft))
	}
	if input.Export != nil {
		export, err := exportSpells(input.Export, input.Spells, result.Spells)
		if err != nil {
			return authorResult{}, err
		}
		result.Export = export
	}
	return result, nil
}

// loadSpellRow is the spell as spells/core.json writes it, base and overrides included.
func loadSpellRow(ruleset *data.Ruleset, key string) (json.RawMessage, error) {
	content, err := fs.ReadFile(ruleset.Files, "spells/core.json")
	if err != nil {
		return nil, err
	}
	var file struct{ Spells []json.RawMessage }
	if err := json.Unmarshal(content, &file); err != nil {
		return nil, err
	}
	for _, row := range file.Spells {
		var named struct{ Key string }
		if json.Unmarshal(row, &named) == nil && named.Key == key {
			return row, nil
		}
	}
	return nil, &data.NotDefinedError{Kind: "spell", Key: key}
}

// schemaBase is the $id of ruleset/schemas/schema.json, so relative $refs between the schema
// files resolve the way they do for validate.py.
const schemaBase = "https://riftweave.dev/ruleset/schemas/"

var compiledSpellSchemas sync.Map // *data.Ruleset → *jsonschema.Schema

func spellSchema(ruleset *data.Ruleset) (*jsonschema.Schema, error) {
	if cached, ok := compiledSpellSchemas.Load(ruleset); ok {
		return cached.(*jsonschema.Schema), nil
	}
	if ruleset.Schemas == nil {
		return nil, &data.MissingInputError{Subject: "author", Needs: "the ruleset schemas (core.Open reads ruleset/schemas)"}
	}
	var compiler *jsonschema.Compiler = jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7)
	names, err := fs.Glob(ruleset.Schemas, "*.json")
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		content, err := fs.ReadFile(ruleset.Schemas, name)
		if err != nil {
			return nil, err
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if err := compiler.AddResource(schemaBase+name, document); err != nil {
			return nil, err
		}
	}
	schema, err := compiler.Compile(schemaBase + "spell.schema.json")
	if err != nil {
		return nil, err
	}
	compiledSpellSchemas.Store(ruleset, schema)
	return schema, nil
}

// collectionCategories is validate.py's COLLECTION_CATEGORIES: the collections whose keys it
// indexes as core:<category>/<key>.
var collectionCategories = map[string]string{
	"effects": "effect", "spells": "spell", "ingredients": "ingredient", "skills": "skill",
	"features": "feature", "races": "race", "conditions": "condition", "materials": "material",
	"gems": "gem", "monsters": "monster", "summons": "summon", "backgrounds": "background",
	"recipes": "recipe", "crafted_items": "crafted_item", "origins": "origin", "hazards": "hazard",
	"checks": "check", "natural_weapons": "natural_weapon", "monster_templates": "monster_template",
	"skill_groups": "skill_group", "summon_bonuses": "summon_bonus",
}

var refIndexes sync.Map // *data.Ruleset → map[string]bool

// refIndex holds every core:<category>/<key> the data defines, as validate.py builds it: the
// collections above, in core and module data, outside the bases directories.
func refIndex(ruleset *data.Ruleset) map[string]bool {
	if cached, ok := refIndexes.Load(ruleset); ok {
		return cached.(map[string]bool)
	}
	var refs map[string]bool = map[string]bool{}
	for _, files := range append([]fs.FS{ruleset.Files}, ruleset.ModuleFiles...) {
		_ = fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".json") || strings.Contains("/"+path, "/bases/") {
				return nil
			}
			content, readError := fs.ReadFile(files, path)
			if readError != nil {
				return nil
			}
			var collections map[string]json.RawMessage
			if json.Unmarshal(content, &collections) != nil {
				return nil
			}
			for collection, rows := range collections {
				category, indexed := collectionCategories[collection]
				if !indexed {
					continue
				}
				var keyed []struct{ Key string }
				if json.Unmarshal(rows, &keyed) == nil {
					for _, row := range keyed {
						if row.Key != "" {
							refs["core:"+category+"/"+row.Key] = true
						}
					}
				}
			}
			return nil
		})
	}
	refIndexes.Store(ruleset, refs)
	return refs
}

// englishMessages words the schema errors, as validate.py reports them in English.
var englishMessages *message.Printer = message.NewPrinter(language.English)

var coreRef = regexp.MustCompile(`^core:([a-z_]+)/[a-z0-9_]+$`)

func checkSpell(ruleset *data.Ruleset, schema *jsonschema.Schema, refs map[string]bool, draft json.RawMessage) authoredSpell {
	var spell authoredSpell = authoredSpell{Row: draft, Problems: []authorFinding{}, Warnings: []authorFinding{}}
	var row map[string]any
	if err := json.Unmarshal(draft, &row); err != nil {
		spell.Problems = append(spell.Problems, authorFinding{Check: "schema", Path: "", Message: "the draft is not a JSON object"})
		return spell
	}
	spell.Key, _ = row["key"].(string)

	// Schema: the draft in the spells wrapper, as the file will hold it.
	wrapper, _ := jsonschema.UnmarshalJSON(strings.NewReader(`{"spells":[` + string(draft) + `]}`))
	if err := schema.Validate(wrapper); err != nil {
		var invalid *jsonschema.ValidationError
		if ok := asValidationError(err, &invalid); ok {
			// The spells items are oneOf spell or spellWithBase; report the branch the draft is
			// written as, not both.
			var other string = "#/definitions/spellWithBase"
			if _, based := row["base"]; based {
				other = "#/definitions/spell"
			}
			var seen []string
			for _, leaf := range leaves(invalid) {
				if strings.Contains(leaf.SchemaURL, "spell.schema.json"+other+"/") || strings.HasSuffix(leaf.SchemaURL, "spell.schema.json"+other) {
					continue
				}
				var path string = "/" + strings.Join(leaf.InstanceLocation, "/")
				path = strings.TrimPrefix(strings.TrimPrefix(path, "/spells/0"), "/spells")
				var message string = leaf.ErrorKind.LocalizedString(englishMessages)
				if !slices.Contains(seen, path+message) {
					seen = append(seen, path+message)
					spell.Problems = append(spell.Problems, authorFinding{Check: "schema", Path: path, Message: message})
				}
			}
		} else {
			spell.Problems = append(spell.Problems, authorFinding{Check: "schema", Path: "", Message: err.Error()})
		}
	}

	// Refs: every core:<category>/<key> of an indexed category must name a row.
	walkStrings(row, "", func(path, value string) {
		match := coreRef.FindStringSubmatch(value)
		if match == nil || !slices.Contains(slices.Collect(mapValues(collectionCategories)), match[1]) {
			return
		}
		if !refs[value] {
			spell.Problems = append(spell.Problems, authorFinding{Check: "ref", Path: path, Message: value + " names no row"})
		}
	})

	// Channels and parameters, per effect.
	effects, _ := row["effects"].([]any)
	for index, raw := range effects {
		applied, _ := raw.(map[string]any)
		ref, _ := applied["effect"].(string)
		effect, known := ruleset.Effects.Get(ref)
		if !known {
			continue
		}
		var path string = fmt.Sprintf("/effects/%d", index)
		if !slices.Contains(effect.Channels, "spell") {
			spell.Problems = append(spell.Problems, authorFinding{Check: "channel", Path: path + "/effect",
				Message: fmt.Sprintf("%s does not list the spell channel (it has %s)", ref, strings.Join(effect.Channels, ", "))})
		}
		if parameter, given := applied["parameter"]; given && effect.ParameterCategory != "" {
			text, _ := parameter.(string)
			var category string = effect.ParameterCategory
			if !strings.HasPrefix(text, "core:"+category+"/") && !regexp.MustCompile(`^module:[a-z0-9_]+/`+category+`/`).MatchString(text) {
				spell.Problems = append(spell.Problems, authorFinding{Check: "parameter", Path: path + "/parameter",
					Message: fmt.Sprintf("%s needs a %s reference, got %v", ref, category, parameter)})
			}
		}
	}

	// Cost: the spellcraft formula; an authored mana that differs is a warning, not a problem.
	var drafted struct {
		Effects []data.AppliedEffect `json:"effects"`
		Cost    *struct {
			Mana *float64 `json:"mana"`
		} `json:"cost"`
	}
	if json.Unmarshal(draft, &drafted) == nil && len(drafted.Effects) > 0 {
		if mana, cast, craftThreshold, err := spellcraft.Cost(ruleset, drafted.Effects); err == nil {
			spell.Cost = &authorCost{Mana: mana, CastThreshold: cast, CraftThreshold: craftThreshold}
			if drafted.Cost != nil && drafted.Cost.Mana != nil && int64(*drafted.Cost.Mana) != mana {
				spell.Warnings = append(spell.Warnings, authorFinding{Check: "cost", Path: "/cost/mana",
					Message: fmt.Sprintf("the authored mana is %d; the spellcraft formula gives %d", int64(*drafted.Cost.Mana), mana)})
			}
		}
	}
	return spell
}

func asValidationError(err error, target **jsonschema.ValidationError) bool {
	invalid, ok := err.(*jsonschema.ValidationError)
	if ok {
		*target = invalid
	}
	return ok
}

// leaves are the errors with no causes: the specific keyword that failed, not the oneOf above it.
func leaves(invalid *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if len(invalid.Causes) == 0 {
		return []*jsonschema.ValidationError{invalid}
	}
	var found []*jsonschema.ValidationError
	for _, cause := range invalid.Causes {
		found = append(found, leaves(cause)...)
	}
	return found
}

func walkStrings(node any, path string, visit func(path, value string)) {
	switch value := node.(type) {
	case string:
		visit(path, value)
	case map[string]any:
		for _, key := range sortedKeys(value) {
			walkStrings(value[key], path+"/"+key, visit)
		}
	case []any:
		for index, item := range value {
			walkStrings(item, fmt.Sprintf("%s/%d", path, index), visit)
		}
	}
}

func mapValues[K comparable, V any](values map[K]V) func(func(V) bool) {
	return func(yield func(V) bool) {
		for _, value := range values {
			if !yield(value) {
				return
			}
		}
	}
}

// exportSpells writes the drafts as one data file in the core or the module shape.
func exportSpells(export *authorExport, drafts []json.RawMessage, checked []authoredSpell) (*authorResultExport, error) {
	var rows []any = make([]any, 0, len(drafts))
	for _, draft := range drafts {
		var row any
		if err := json.Unmarshal(draft, &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	var result *authorResultExport = &authorResultExport{Shape: export.Shape, Refs: []string{}}
	switch export.Shape {
	case "core":
		result.Path = "ruleset/data/spells/authored.json"
		result.File = map[string]any{"$schema": "../../schemas/spell.schema.json", "spells": rows}
		for _, spell := range checked {
			result.Refs = append(result.Refs, "core:spell/"+spell.Key)
		}
	case "module":
		if export.Module == "" {
			return nil, &data.MissingInputError{Subject: "a module export", Needs: "module"}
		}
		result.Path = "ruleset/modules/" + export.Module + "/data/spells.json"
		result.File = map[string]any{"$schema": "../../../schemas/spell.schema.json", "spells": rows}
		result.Manifest = map[string]any{
			"$schema":     "../../schemas/module.manifest.schema.json",
			"id":          export.Module,
			"version":     "0.1.0",
			"description": "Spells written with the author command (docs/backend/authoring.md).",
			"requires":    []string{},
			"extends":     map[string]any{"collections": []string{"spells"}},
			"data":        []string{"data/spells.json"},
			"schemas":     []string{},
		}
		for _, spell := range checked {
			result.Refs = append(result.Refs, "module:"+export.Module+"/spell/"+spell.Key)
		}
	default:
		return nil, &data.NotDefinedError{Kind: "export shape", Key: export.Shape}
	}
	return result, nil
}
