package catalog

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/spellcraft"
)

var repo string = filepath.Join("..", "..", "..")

// testModule adds a feature, an equipment row on a core weapon base, and replaces one core
// feature through its manifest's overrides (docs/core/overrides.md).
var testModule fstest.MapFS = fstest.MapFS{
	"manifest.json": {Data: []byte(`{"id": "test_module", "version": "1.0.0", "data": ["data/rows.json", "data/equipment.json"], "overrides": ["hardened_hide"]}`)},
	"data/rows.json": {Data: []byte(`{"features": [
		{"key": "module_perk", "label": "Module Perk", "description": "Added by a module.", "type": "perk", "category": "combat", "ranks": 1},
		{"key": "hardened_hide", "label": "Hardened Hide, Module Edition", "description": "Replaced by a module.", "type": "creation", "category": "general", "ranks": 1}
	]}`)},
	"data/equipment.json": {Data: []byte(`{"equipment": [{"key": "frost_dagger", "base": "core:weapon_base/dagger", "overrides": {"label": "Frost Dagger", "phase": "water"}}]}`)},
}

func load(t *testing.T, modules ...fs.FS) *data.Ruleset {
	t.Helper()
	ruleset, err := data.Load(os.DirFS(filepath.Join(repo, "ruleset", "data")), modules...)
	if err != nil {
		t.Fatal(err)
	}
	return ruleset
}

func run(t *testing.T, ruleset *data.Ruleset, input Input) Output {
	t.Helper()
	output, err := Run(ruleset, input)
	if err != nil {
		t.Fatalf("%+v: %v", input, err)
	}
	return output
}

func find(rows []Row, ref string) *Row {
	for index := range rows {
		if rows[index].Ref == ref {
			return &rows[index]
		}
	}
	return nil
}

func refs(rows []Row) []string {
	var result []string
	for _, row := range rows {
		result = append(result, row.Ref)
	}
	return result
}

func TestTheCollectionsAreEveryTableToStudyABuild(t *testing.T) {
	var output Output = run(t, load(t), Input{Operation: OperationCollections})
	var names []string
	for _, collection := range output.Collections {
		names = append(names, collection.Name)
		if collection.Count == 0 || len(collection.Columns) == 0 || collection.Columns[0].Field != "label" {
			t.Errorf("%s: %d rows, columns %v", collection.Name, collection.Count, collection.Columns)
		}
	}
	var want = []string{"weapons", "armor", "accessories", "consumables", "gear", "materials", "gems", "monsters", "spells", "features", "races", "backgrounds", "skills", "conditions", "effects"}
	if !slices.Equal(names, want) {
		t.Errorf("collections %v, want %v", names, want)
	}
}

func TestAWeaponIsItsBaseWithItsOverrides(t *testing.T) {
	var output Output = run(t, load(t), Input{Operation: OperationList, Collection: "weapons"})
	var longsword *Row = find(output.Rows, "core:weapon/longsword")
	if longsword == nil {
		t.Fatalf("no longsword, a base that is an item, in %v", refs(output.Rows))
	}
	if longsword.Source != "core" || field(longsword.Row, "weapon.skill") != "core:skill/blades" {
		t.Errorf("longsword %+v", longsword)
	}
	var cursed *Row = find(output.Rows, "core:weapon/cursed_iron_longsword")
	if cursed == nil {
		t.Fatal("no cursed_iron_longsword")
	}
	if cursed.Row["key"] != "cursed_iron_longsword" || cursed.Row["phase"] != "earth" || cursed.Row["enchantment_slots"] != float64(-1) {
		t.Errorf("overrides not on top: %v", cursed.Row)
	}
	if field(cursed.Row, "cost.quantity") != float64(8) || field(cursed.Row, "cost.unit") != "gold" || field(cursed.Row, "weapon.skill") != "core:skill/blades" {
		t.Errorf("an object override merges field by field: %v", cursed.Row)
	}
	if _, kept := cursed.Row["base"]; kept {
		t.Errorf("a resolved row has no base: %v", cursed.Row)
	}
}

func TestAMonsterAndASpellResolveThroughTheirBase(t *testing.T) {
	var ruleset *data.Ruleset = load(t)
	var monsters Output = run(t, ruleset, Input{Operation: OperationList, Collection: "monsters"})
	var skeleton *Row = find(monsters.Rows, "core:monster/skeleton")
	if skeleton == nil || skeleton.Row["label"] != "Skeleton" || skeleton.Row["key"] != "skeleton" {
		t.Errorf("skeleton %+v", skeleton)
	}
	var direWolf *Row = find(monsters.Rows, "core:monster/dire_wolf")
	if direWolf == nil || direWolf.Row["key"] != "dire_wolf" || direWolf.Row["level"] == nil {
		t.Errorf("dire wolf %+v", direWolf)
	}
	var spell Output = run(t, ruleset, Input{Operation: OperationGet, Collection: "spells", Ref: "core:spell/greater_unlock"})
	if spell.Row == nil || spell.Row.Row["label"] != "Greater Unlock" || spell.Row.Row["color"] != "indigo" || spell.Row.Row["level"] != float64(15) {
		t.Errorf("greater unlock %+v", spell.Row)
	}
}

func TestFiltersAndSortWorkOnWeaponsAndSpells(t *testing.T) {
	var ruleset *data.Ruleset = load(t)
	var slashing Output = run(t, ruleset, Input{Operation: OperationList, Collection: "weapons",
		Filters: []Filter{{Field: "weapon.damage.type", Equals: "slashing"}}, Sort: &Sort{Field: "weight", Descending: true}})
	if len(slashing.Rows) < 2 {
		t.Fatalf("slashing weapons %v", refs(slashing.Rows))
	}
	for index, row := range slashing.Rows {
		if field(row.Row, "weapon.damage.type") != "slashing" {
			t.Errorf("%s is not slashing", row.Ref)
		}
		if index > 0 && number(field(slashing.Rows[index-1].Row, "weight")) < number(field(row.Row, "weight")) {
			t.Errorf("not heaviest first at %s", row.Ref)
		}
	}
	var low = 1.0
	var high = 2.0
	var red Output = run(t, ruleset, Input{Operation: OperationList, Collection: "spells",
		Filters: []Filter{{Field: "color", Equals: "red"}, {Field: "level", Min: &low, Max: &high}}, Sort: &Sort{Field: "level"}})
	if find(red.Rows, "core:spell/fire_bolt") == nil {
		t.Errorf("fire bolt is red and level 1: %v", refs(red.Rows))
	}
	for index, row := range red.Rows {
		var level float64 = number(row.Row["level"])
		if row.Row["color"] != "red" || level < 1 || level > 2 {
			t.Errorf("%s: color %v, level %v", row.Ref, row.Row["color"], level)
		}
		if index > 0 && number(red.Rows[index-1].Row["level"]) > level {
			t.Errorf("not lowest level first at %s", row.Ref)
		}
	}
	var text Output = run(t, ruleset, Input{Operation: OperationList, Collection: "spells", Filters: []Filter{{Text: "FIRE BOLT"}}})
	if !slices.Contains(refs(text.Rows), "core:spell/fire_bolt") {
		t.Errorf("text is found in the label, in any case: %v", refs(text.Rows))
	}
}

func TestAModuleAddsRowsAndOverridesACoreRow(t *testing.T) {
	var ruleset *data.Ruleset = load(t, testModule)
	var features Output = run(t, ruleset, Input{Operation: OperationList, Collection: "features"})
	var added *Row = find(features.Rows, "module:test_module/feature/module_perk")
	if added == nil || added.Source != "test_module" {
		t.Errorf("module perk %+v", added)
	}
	var replaced *Row = find(features.Rows, "core:feature/hardened_hide")
	if replaced == nil || replaced.Source != "test_module" || replaced.Row["label"] != "Hardened Hide, Module Edition" {
		t.Errorf("the override keeps the core ref and takes the module's row: %+v", replaced)
	}
	var weapons Output = run(t, ruleset, Input{Operation: OperationList, Collection: "weapons"})
	var dagger *Row = find(weapons.Rows, "module:test_module/weapon/frost_dagger")
	if dagger == nil || dagger.Row["phase"] != "water" || field(dagger.Row, "weapon.skill") == nil {
		t.Errorf("a module's weapon on a core base: %+v", dagger)
	}
}

func TestRefusals(t *testing.T) {
	var ruleset *data.Ruleset = load(t)
	var notDefined *data.NotDefinedError
	if _, err := Run(ruleset, Input{Operation: OperationList, Collection: "classes"}); !errors.As(err, &notDefined) {
		t.Errorf("an unknown collection: %v", err)
	}
	if _, err := Run(ruleset, Input{Operation: OperationGet, Collection: "spells", Ref: "core:spell/nope"}); !errors.As(err, &notDefined) {
		t.Errorf("an unknown ref: %v", err)
	}
	var missing *data.MissingInputError
	if _, err := Run(ruleset, Input{Operation: OperationList}); !errors.As(err, &missing) {
		t.Errorf("list without a collection: %v", err)
	}
	if _, err := Run(ruleset, Input{Operation: "drop"}); !errors.As(err, &notDefined) {
		t.Errorf("an unknown operation: %v", err)
	}
}

func number(value any) float64 {
	result, _ := value.(float64)
	return result
}

// DE_S02_05 (docs/modules/magic.md#cost-formula): no spell row stores its mana, so the spell table
// shows the mana the spellcraft formula gives, for every core spell, in the cost.mana column.
func TestTheSpellTableShowsTheManaTheFormulaGives(t *testing.T) {
	var ruleset *data.Ruleset = load(t)
	var spells Output = run(t, ruleset, Input{Operation: OperationList, Collection: "spells"})
	if len(spells.Rows) < 45 {
		t.Fatalf("%d spells", len(spells.Rows))
	}
	for _, row := range spells.Rows {
		spell, err := ruleset.Spells.Require("spell", row.Row["key"].(string))
		if err != nil {
			t.Fatal(err)
		}
		want, err := spellcraft.SpellMana(ruleset, spell)
		if err != nil {
			t.Fatal(err)
		}
		cost, _ := row.Row["cost"].(map[string]any)
		if cost == nil || cost["mana"] != float64(want) {
			t.Errorf("%s: cost %v, want mana %d", row.Ref, row.Row["cost"], want)
		}
	}
	// A plain spell, a spell on a base spell, and two summon spells, as casting_reference.spell_cost
	// prices them.
	for ref, want := range map[string]float64{"core:spell/fire_bolt": 6, "core:spell/greater_unlock": 30, "core:spell/summon_familiar": 18, "core:spell/raise_thrall": 36} {
		if row := find(spells.Rows, ref); row == nil || row.Row["cost"].(map[string]any)["mana"] != want {
			t.Errorf("%s: %+v, want mana %v", ref, row, want)
		}
	}
}
