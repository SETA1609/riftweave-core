// Package catalog is the `catalog` command (docs/backend/api.md#catalog): every table a player or
// a GM studies to plan a build, from core and every loaded module, each row resolved through its
// base (docs/core/composition.md) as validate.py composes it. It reads rows and filters them; it
// holds no rule. The columns a table shows are display, kept in `collections` below.
package catalog

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/spellcraft"
)

// Operation is what the command does.
type Operation string

const (
	// OperationCollections lists the collections, with their columns and row counts.
	OperationCollections Operation = "collections"
	// OperationList returns one collection's rows, filtered and sorted.
	OperationList Operation = "list"
	// OperationGet returns one row by its ref.
	OperationGet Operation = "get"
)

// coreSource is the source of a row that ships with the ruleset; a module row names its module.
const coreSource = "core"

// Filter keeps the rows that match it. Field is a dotted path into the resolved row
// (`weapon.damage.type`). Equals matches the value, or an item of a list; Min and Max bound a
// number; Text is found in the label or the description, in any case. A row matches every filter.
type Filter struct {
	Field  string   `json:"field"`
	Equals any      `json:"equals"`
	Min    *float64 `json:"min"`
	Max    *float64 `json:"max"`
	Text   string   `json:"text"`
}

// Sort orders the rows by a field; a row without it sorts last. Without a sort, rows keep the
// data's order: core first, then each module in load order.
type Sort struct {
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}

// Input is one catalog call.
type Input struct {
	Operation  Operation `json:"operation"`
	Collection string    `json:"collection"`
	Ref        string    `json:"ref"`
	Filters    []Filter  `json:"filters"`
	Sort       *Sort     `json:"sort"`
}

// Column is one column a table shows: a dotted field path and its heading.
type Column struct {
	Field string `json:"field"`
	Label string `json:"label"`
}

// CollectionInfo describes one collection for a picker.
type CollectionInfo struct {
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Count   int      `json:"count"`
	Columns []Column `json:"columns"`
}

// Row is one resolved row with its namespaced ref and where it comes from.
type Row struct {
	Ref    string         `json:"ref"`
	Source string         `json:"source"`
	Row    map[string]any `json:"row"`
}

// Output is the result: the collections, a collection's rows, or one row.
type Output struct {
	Collections []CollectionInfo `json:"collections,omitempty"`
	Collection  string           `json:"collection,omitempty"`
	Columns     []Column         `json:"columns,omitempty"`
	Rows        []Row            `json:"rows,omitempty"`
	Row         *Row             `json:"row,omitempty"`
}

// collection is where one table's rows live and how a row names its base.
type collection struct {
	name, label string
	// category is the singular category of its refs (core:<category>/<key>).
	category string
	// file and wrapper are the core data file and the array in it.
	file, wrapper string
	// baseDirectory holds the base files a `core:<category>_base/` or `core:creature_type/` ref
	// names; "" when the collection has none.
	baseDirectory string
	// itemBases: a base with is_item true is also a row of its own key (composition.md).
	itemBases bool
	// identity are the fields a row built on another row of its collection does not inherit.
	identity []string
	columns  []Column
}

var equipmentIdentity = []string{"key", "source"}

// collections is every table the compendium shows, in the order a picker lists them.
var collections = []collection{
	{name: "weapons", label: "Weapons", category: "weapon", file: "equipment/weapons.json", wrapper: "equipment", baseDirectory: "equipment/bases/weapons", itemBases: true, identity: equipmentIdentity,
		columns: []Column{{"label", "Weapon"}, {"weapon.skill", "Skill"}, {"weapon.damage.dice", "Damage"}, {"weapon.damage.type", "Type"}, {"weapon.properties", "Properties"}, {"weapon.range", "Range"}, {"cost", "Cost"}, {"weight", "Weight"}}},
	{name: "armor", label: "Armor", category: "armor", file: "equipment/armor.json", wrapper: "equipment", baseDirectory: "equipment/bases/armor", itemBases: true, identity: equipmentIdentity,
		columns: []Column{{"label", "Armor"}, {"armor.slot", "Slot"}, {"armor.category", "Category"}, {"armor.dr_base", "DR"}, {"armor.evasion_penalty", "Evasion"}, {"cost", "Cost"}, {"weight", "Weight"}}},
	{name: "accessories", label: "Accessories", category: "accessory", file: "equipment/accessories.json", wrapper: "equipment", baseDirectory: "equipment/bases/accessories", itemBases: true, identity: equipmentIdentity,
		columns: []Column{{"label", "Accessory"}, {"accessory.accessory_type", "Kind"}, {"enchantment_slots", "Slots"}, {"phase", "Phase"}, {"cost", "Cost"}, {"weight", "Weight"}}},
	{name: "consumables", label: "Consumables", category: "consumable", file: "equipment/consumables.json", wrapper: "equipment", identity: equipmentIdentity,
		columns: []Column{{"label", "Consumable"}, {"type", "Kind"}, {"consumable.delivery", "Delivery"}, {"consumable.toxicity", "Toxicity"}, {"cost", "Cost"}, {"weight", "Weight"}}},
	{name: "gear", label: "Gear", category: "gear", file: "equipment/gear.json", wrapper: "equipment", identity: equipmentIdentity,
		columns: []Column{{"label", "Gear"}, {"type", "Kind"}, {"cost", "Cost"}, {"weight", "Weight"}}},
	{name: "materials", label: "Materials", category: "material", file: "materials/core.json", wrapper: "materials",
		columns: []Column{{"label", "Material"}, {"tier", "Tier"}, {"phase", "Phase"}, {"weight_factor", "Weight"}, {"value_factor", "Value"}, {"tags", "Tags"}}},
	{name: "gems", label: "Gems", category: "gem", file: "gems/core.json", wrapper: "gems",
		columns: []Column{{"label", "Gem"}, {"tier", "Tier"}, {"phase", "Phase"}}},
	{name: "monsters", label: "Monsters", category: "monster", file: "monsters/core.json", wrapper: "monsters", baseDirectory: "monsters/bases", identity: []string{"key", "source"},
		columns: []Column{{"label", "Monster"}, {"level", "Level"}, {"size", "Size"}, {"phase", "Phase"}, {"tags", "Tags"}}},
	{name: "spells", label: "Spells", category: "spell", file: "spells/core.json", wrapper: "spells", identity: []string{"key", "label", "description", "source"},
		columns: []Column{{"label", "Spell"}, {"color", "Color"}, {"level", "Level"}, {"cost.mana", "Mana"}, {"delivery", "Delivery"}}},
	{name: "features", label: "Perks and traits", category: "feature", file: "features/core.json", wrapper: "features",
		columns: []Column{{"label", "Feature"}, {"type", "Type"}, {"category", "Category"}, {"prerequisite", "Prerequisites"}, {"ranks", "Ranks"}}},
	{name: "races", label: "Races", category: "race", file: "races/core.json", wrapper: "races",
		columns: []Column{{"label", "Race"}, {"lineage.role", "Lineage"}, {"size", "Size"}, {"speed", "Speed"}, {"phase", "Phase"}, {"ability_modifiers", "Attributes"}}},
	{name: "backgrounds", label: "Backgrounds", category: "background", file: "backgrounds/core.json", wrapper: "backgrounds",
		columns: []Column{{"label", "Background"}, {"category", "Category"}, {"suggested_tag_skills", "Tag skills"}, {"wealth_bonus", "Wealth"}}},
	{name: "skills", label: "Skills", category: "skill", file: "skills/core.json", wrapper: "skills",
		columns: []Column{{"label", "Skill"}, {"associated_ability", "Attribute"}, {"category", "Category"}, {"taggable", "Taggable"}}},
	{name: "conditions", label: "Conditions", category: "condition", file: "conditions/core.json", wrapper: "conditions",
		columns: []Column{{"label", "Condition"}, {"stacking", "Stacks"}, {"max_stacks", "Max"}, {"applied_by", "Applied by"}}},
	{name: "effects", label: "Effects", category: "effect", file: "effects/core.json", wrapper: "effects",
		columns: []Column{{"label", "Effect"}, {"category", "Category"}, {"color", "Color"}, {"phase", "Phase"}, {"polarity", "Polarity"}, {"channels", "Channels"}, {"base_cost", "Cost"}}},
}

// equipmentCollection is the collection an equipment row of a module belongs to, by its `type`.
var equipmentCollection = map[string]string{"weapon": "weapons", "armor": "armor", "accessory": "accessories", "consumable": "consumables", "potion": "consumables", "gear": "gear"}

// tables is every collection's resolved rows, built once per ruleset.
type tables map[string][]Row

var built sync.Map // *data.Ruleset → tables

// Run runs one catalog call.
func Run(ruleset *data.Ruleset, input Input) (Output, error) {
	var rows tables = tablesOf(ruleset)
	switch input.Operation {
	case OperationCollections:
		var output Output
		for _, entry := range collections {
			output.Collections = append(output.Collections, CollectionInfo{Name: entry.name, Label: entry.label, Count: len(rows[entry.name]), Columns: entry.columns})
		}
		return output, nil
	case OperationList:
		entry, err := named(input.Collection, "catalog list")
		if err != nil {
			return Output{}, err
		}
		var kept []Row = []Row{}
		for _, row := range rows[entry.name] {
			if matchesAll(row.Row, input.Filters) {
				kept = append(kept, row)
			}
		}
		if input.Sort != nil {
			sortRows(kept, *input.Sort)
		}
		return Output{Collection: entry.name, Columns: entry.columns, Rows: kept}, nil
	case OperationGet:
		entry, err := named(input.Collection, "catalog get")
		if err != nil {
			return Output{}, err
		}
		if input.Ref == "" {
			return Output{}, &data.MissingInputError{Subject: "catalog get", Needs: "ref"}
		}
		for _, row := range rows[entry.name] {
			if row.Ref == input.Ref {
				return Output{Collection: entry.name, Columns: entry.columns, Row: &row}, nil
			}
		}
		return Output{}, &data.NotDefinedError{Kind: entry.category, Key: input.Ref}
	}
	return Output{}, &data.NotDefinedError{Kind: "catalog operation", Key: string(input.Operation)}
}

// named is the collection of the name, or why there is none.
func named(name, subject string) (*collection, error) {
	if name == "" {
		return nil, &data.MissingInputError{Subject: subject, Needs: "collection"}
	}
	for index := range collections {
		if collections[index].name == name {
			return &collections[index], nil
		}
	}
	return nil, &data.NotDefinedError{Kind: "collection", Key: name}
}

func tablesOf(ruleset *data.Ruleset) tables {
	if cached, ok := built.Load(ruleset); ok {
		return cached.(tables)
	}
	var result tables = build(ruleset)
	built.Store(ruleset, result)
	return result
}

// rawRow is a row as its file writes it, with where it comes from.
type rawRow struct {
	ref, source string
	row         map[string]any
}

// build reads every collection from core, then layers each module's rows on top: a row whose key
// the module's manifest lists in `overrides` replaces the core row of that key, keeping its ref;
// any other row is added under the module's namespace.
func build(ruleset *data.Ruleset) tables {
	var raw map[string][]rawRow = map[string][]rawRow{}
	for _, entry := range collections {
		for _, key := range itemBaseKeys(ruleset.Files, entry) {
			raw[entry.name] = append(raw[entry.name], rawRow{ref: ref(coreSource, entry.category, key), source: coreSource,
				row: map[string]any{"key": key, "base": "core:" + entry.category + "_base/" + key}})
		}
		for _, row := range readRows(ruleset.Files, entry.file, entry.wrapper) {
			key, _ := row["key"].(string)
			raw[entry.name] = append(raw[entry.name], rawRow{ref: ref(coreSource, entry.category, key), source: coreSource, row: row})
		}
	}
	for _, module := range ruleset.ModuleFiles {
		layerModule(ruleset.Files, module, raw)
	}
	var result tables = tables{}
	for index := range collections {
		var entry *collection = &collections[index]
		var resolver = resolver{files: ruleset.Files, entry: entry, raw: raw[entry.name], done: map[int]map[string]any{}}
		for position, row := range raw[entry.name] {
			var composed map[string]any = resolver.resolve(position, map[int]bool{})
			if entry.name == spellsCollection {
				composed = withMana(ruleset, composed)
			}
			result[entry.name] = append(result[entry.name], Row{Ref: row.ref, Source: row.source, Row: composed})
		}
	}
	return result
}

// spellsCollection is the table whose rows show a computed mana.
const spellsCollection = "spells"

// withMana is a composed spell row with cost.mana filled by the spellcraft formula, which no row
// stores (DE_S02_05). A row the formula cannot price is shown without it.
func withMana(ruleset *data.Ruleset, row map[string]any) map[string]any {
	encoded, err := json.Marshal(row)
	if err != nil {
		return row
	}
	var spell data.Spell
	if json.Unmarshal(encoded, &spell) != nil {
		return row
	}
	mana, err := spellcraft.SpellMana(ruleset, &spell)
	if err != nil {
		return row
	}
	var shown map[string]any = maps.Clone(row)
	var cost map[string]any = map[string]any{}
	if stored, ok := row["cost"].(map[string]any); ok {
		cost = maps.Clone(stored)
	}
	cost["mana"] = float64(mana)
	shown["cost"] = cost
	return shown
}

func ref(source, category, key string) string {
	if source == coreSource {
		return "core:" + category + "/" + key
	}
	return "module:" + source + "/" + category + "/" + key
}

// layerModule adds one module's rows, by the array each of its data files holds.
func layerModule(core, module fs.FS, raw map[string][]rawRow) {
	var manifest struct {
		ID        string   `json:"id"`
		Data      []string `json:"data"`
		Overrides []string `json:"overrides"`
	}
	content, err := fs.ReadFile(module, "manifest.json")
	if err != nil || json.Unmarshal(content, &manifest) != nil {
		return
	}
	for _, file := range manifest.Data {
		content, err := fs.ReadFile(module, file)
		if err != nil {
			continue
		}
		var arrays map[string]json.RawMessage
		if json.Unmarshal(content, &arrays) != nil {
			continue
		}
		for wrapper, encoded := range arrays {
			var rows []map[string]any
			if json.Unmarshal(encoded, &rows) != nil {
				continue
			}
			for _, row := range rows {
				var name string = collectionOf(core, wrapper, row)
				if name == "" {
					continue
				}
				entry, _ := named(name, "")
				key, _ := row["key"].(string)
				var replaced bool = false
				if slices.Contains(manifest.Overrides, key) || slices.Contains(manifest.Overrides, ref(coreSource, entry.category, key)) {
					for index, existing := range raw[name] {
						if existing.source == coreSource && existing.row["key"] == key {
							raw[name][index] = rawRow{ref: existing.ref, source: manifest.ID, row: row}
							replaced = true
						}
					}
				}
				if !replaced {
					raw[name] = append(raw[name], rawRow{ref: ref(manifest.ID, entry.category, key), source: manifest.ID, row: row})
				}
			}
		}
	}
}

// collectionOf is the collection a module row belongs to: by its array, or for equipment by its
// type, read from its base when the row names one.
func collectionOf(core fs.FS, wrapper string, row map[string]any) string {
	if wrapper != "equipment" {
		for _, entry := range collections {
			if entry.wrapper == wrapper {
				return entry.name
			}
		}
		return ""
	}
	kind, _ := row["type"].(string)
	if baseRef, _ := row["base"].(string); kind == "" && baseRef != "" {
		for _, entry := range collections {
			if strings.HasPrefix(baseRef, "core:"+entry.category+"_base/") {
				return entry.name
			}
		}
	}
	return equipmentCollection[kind]
}

// itemBaseKeys are the keys of the bases that are also items, in file-name order.
func itemBaseKeys(files fs.FS, entry collection) []string {
	if !entry.itemBases {
		return nil
	}
	entries, err := fs.ReadDir(files, entry.baseDirectory)
	if err != nil {
		return nil
	}
	var keys []string
	for _, file := range entries {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		var base struct {
			IsItem bool `json:"is_item"`
		}
		if content, err := fs.ReadFile(files, entry.baseDirectory+"/"+file.Name()); err == nil && json.Unmarshal(content, &base) == nil && base.IsItem {
			keys = append(keys, strings.TrimSuffix(file.Name(), ".json"))
		}
	}
	return keys
}

func readRows(files fs.FS, file, wrapper string) []map[string]any {
	content, err := fs.ReadFile(files, file)
	if err != nil {
		return nil
	}
	var arrays map[string]json.RawMessage // beside the array, a file holds its $schema string
	var rows []map[string]any
	if json.Unmarshal(content, &arrays) != nil || json.Unmarshal(arrays[wrapper], &rows) != nil {
		return nil
	}
	return rows
}

// resolver composes one collection's rows, each once.
type resolver struct {
	files fs.FS
	entry *collection
	raw   []rawRow
	done  map[int]map[string]any
}

// resolve is the row at position composed: its base (a base file, or another row of the
// collection, itself resolved), then its overrides merged field by field (data.MergeOverride),
// then its own fields on top. A base that cannot be read leaves the row as written.
func (resolver *resolver) resolve(position int, seen map[int]bool) map[string]any {
	if row, ok := resolver.done[position]; ok {
		return row
	}
	var row map[string]any = resolver.raw[position].row
	baseRef, _ := row["base"].(string)
	var base map[string]any
	if baseRef != "" && !seen[position] {
		seen[position] = true
		base = resolver.base(baseRef, seen)
	}
	var composed map[string]any = map[string]any{}
	for field, value := range base {
		composed[field] = value
	}
	if overrides, ok := row["overrides"].(map[string]any); ok && base != nil {
		composed, _ = data.MergeOverride(composed, overrides).(map[string]any)
	} else if ok {
		for field, value := range overrides {
			composed[field] = value
		}
	}
	for field, value := range row {
		if field != "base" && field != "overrides" {
			composed[field] = value
		}
	}
	resolver.done[position] = composed
	return composed
}

// base is the composed row a base ref names, nil when it names nothing readable.
func (resolver *resolver) base(baseRef string, seen map[int]bool) map[string]any {
	if key, sameCollection := strings.CutPrefix(baseRef, "core:"+resolver.entry.category+"/"); sameCollection {
		for position, row := range resolver.raw {
			if row.source == coreSource && row.row["key"] == key {
				var inherited map[string]any = map[string]any{}
				for field, value := range resolver.resolve(position, seen) {
					if !slices.Contains(resolver.entry.identity, field) {
						inherited[field] = value
					}
				}
				return inherited
			}
		}
		return nil
	}
	if resolver.entry.baseDirectory == "" {
		return nil
	}
	var file string = resolver.entry.baseDirectory + "/" + baseRef[strings.LastIndex(baseRef, "/")+1:] + ".json"
	content, err := fs.ReadFile(resolver.files, file)
	if err != nil {
		return nil
	}
	var base map[string]any
	if json.Unmarshal(content, &base) != nil {
		return nil
	}
	delete(base, "is_item")
	delete(base, "growth")
	return base
}

// field is the value at a dotted path, nil when the path is missing.
func field(row map[string]any, path string) any {
	var value any = row
	for _, part := range strings.Split(path, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = object[part]
	}
	return value
}

func matchesAll(row map[string]any, filters []Filter) bool {
	for _, filter := range filters {
		if !matches(row, filter) {
			return false
		}
	}
	return true
}

func matches(row map[string]any, filter Filter) bool {
	if filter.Text != "" {
		var text string = strings.ToLower(filter.Text)
		label, _ := row["label"].(string)
		description, _ := row["description"].(string)
		if !strings.Contains(strings.ToLower(label), text) && !strings.Contains(strings.ToLower(description), text) {
			return false
		}
	}
	if filter.Field == "" {
		return true
	}
	var value any = field(row, filter.Field)
	if filter.Equals != nil && !equal(value, filter.Equals) {
		if list, isList := value.([]any); !isList || !slices.ContainsFunc(list, func(item any) bool { return equal(item, filter.Equals) }) {
			return false
		}
	}
	if filter.Min != nil || filter.Max != nil {
		number, isNumber := value.(float64)
		if !isNumber || (filter.Min != nil && number < *filter.Min) || (filter.Max != nil && number > *filter.Max) {
			return false
		}
	}
	return true
}

// equal compares two JSON values by their encoding, so 1 and 1.0 are the same.
func equal(first, second any) bool {
	left, _ := json.Marshal(first)
	right, _ := json.Marshal(second)
	return string(left) == string(right)
}

func sortRows(rows []Row, by Sort) {
	slices.SortStableFunc(rows, func(first, second Row) int {
		var left, right any = field(first.Row, by.Field), field(second.Row, by.Field)
		switch {
		case left == nil && right == nil:
			return 0
		case left == nil:
			return 1
		case right == nil:
			return -1
		}
		var order int = compare(left, right)
		if by.Descending {
			return -order
		}
		return order
	})
}

func compare(left, right any) int {
	leftNumber, leftIsNumber := left.(float64)
	rightNumber, rightIsNumber := right.(float64)
	if leftIsNumber && rightIsNumber {
		return cmp.Compare(leftNumber, rightNumber)
	}
	return cmp.Compare(strings.ToLower(fmt.Sprint(left)), strings.ToLower(fmt.Sprint(right)))
}
