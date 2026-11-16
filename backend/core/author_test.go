package core

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the author export fixtures in ruleset/tests/fixtures/author")

func authorCore(t *testing.T) *Core {
	t.Helper()
	opened, err := Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	return opened
}

func author(t *testing.T, opened *Core, input map[string]any) map[string]any {
	t.Helper()
	return runJSON(t, opened, "author", input)
}

type finding struct{ Check, Path, Message string }

func findings(t *testing.T, result map[string]any, field string) []finding {
	t.Helper()
	spells, _ := result["spells"].([]any)
	if len(spells) != 1 {
		t.Fatalf("want one spell in %v", result)
	}
	encoded, _ := json.Marshal(spells[0].(map[string]any)[field])
	var listed []finding
	_ = json.Unmarshal(encoded, &listed)
	return listed
}

// Authoring every spell in spells/core.json from its own fields gives back the same row, with no
// problems (docs/backend/authoring.md).
func TestAuthoringEveryCoreSpellReturnsItUnchangedWithNoProblems(t *testing.T) {
	opened := authorCore(t)
	content, err := os.ReadFile(filepath.Join(repo, "ruleset", "data", "spells", "core.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct{ Spells []map[string]any }
	if err := json.Unmarshal(content, &file); err != nil {
		t.Fatal(err)
	}
	for _, row := range file.Spells {
		t.Run(row["key"].(string), func(t *testing.T) {
			result := author(t, opened, map[string]any{"kind": "spell", "spells": []any{row}})
			spell := result["spells"].([]any)[0].(map[string]any)
			if !reflect.DeepEqual(spell["row"], row) {
				t.Errorf("row changed: %v", spell["row"])
			}
			if problems := findings(t, result, "problems"); len(problems) != 0 {
				t.Errorf("problems: %v", problems)
			}
		})
	}
}

func draftSpell() map[string]any {
	return map[string]any{
		"key": "authored_ember", "label": "Authored Ember", "description": "A test spell: a small flame.",
		"color": "red", "level": 1, "delivery": "single",
		"effects": []any{map[string]any{"effect": "core:effect/damage_fire", "magnitude": 6, "target": "ranged"}},
		"cost":    map[string]any{"mana": 4},
	}
}

func TestABrokenRefAWrongChannelAndABadParameterAreEachAProblem(t *testing.T) {
	opened := authorCore(t)
	cases := map[string]struct {
		effect map[string]any
		check  string
	}{
		"broken ref":    {map[string]any{"effect": "core:effect/no_such_effect", "magnitude": 6, "target": "ranged"}, "ref"},
		"wrong channel": {map[string]any{"effect": "core:effect/essence_fire", "magnitude": 2, "target": "self"}, "channel"},
		"bad parameter": {map[string]any{"effect": "core:effect/summon", "parameter": "core:effect/damage_fire", "magnitude": 1, "duration": 60, "target": "ranged"}, "parameter"},
		"schema":        {map[string]any{"effect": "core:effect/damage_fire", "magnitude": "six", "target": "ranged"}, "schema"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			draft := draftSpell()
			draft["effects"] = []any{testCase.effect}
			result := author(t, opened, map[string]any{"kind": "spell", "spells": []any{draft}})
			var checks []string
			for _, problem := range findings(t, result, "problems") {
				checks = append(checks, problem.Check)
			}
			if !strings.Contains(strings.Join(checks, ","), testCase.check) {
				t.Errorf("want a %s problem, got %v", testCase.check, findings(t, result, "problems"))
			}
		})
	}
}

func TestTheCostIsTheFormulasAndAnAuthoredCostThatDiffersIsAWarning(t *testing.T) {
	opened := authorCore(t)
	draft := draftSpell()
	result := author(t, opened, map[string]any{"kind": "spell", "spells": []any{draft}})
	spell := result["spells"].([]any)[0].(map[string]any)
	cost, _ := spell["cost"].(map[string]any)
	if cost == nil || cost["mana"].(float64) <= 0 || cost["cast_threshold"] == nil || cost["craft_threshold"] == nil {
		t.Fatalf("cost: %v", spell["cost"])
	}
	if len(findings(t, result, "problems")) != 0 {
		t.Errorf("problems: %v", findings(t, result, "problems"))
	}
	draft["cost"] = map[string]any{"mana": cost["mana"].(float64) + 3}
	warned := author(t, opened, map[string]any{"kind": "spell", "spells": []any{draft}})
	warnings := findings(t, warned, "warnings")
	if len(warnings) != 1 || warnings[0].Check != "cost" {
		t.Errorf("want one cost warning, got %v", warnings)
	}
	if len(findings(t, warned, "problems")) != 0 {
		t.Errorf("a different cost is not a problem: %v", findings(t, warned, "problems"))
	}
}

func TestLoadReturnsASpellAsWritten(t *testing.T) {
	result := author(t, authorCore(t), map[string]any{"kind": "spell", "operation": "load", "spell": "fire_bolt"})
	row, _ := result["row"].(map[string]any)
	if row["key"] != "fire_bolt" || row["color"] != "red" {
		t.Errorf("load: %v", result)
	}
}

// The exports are written to ruleset/tests/fixtures/author, where test_author_export.py runs
// validate.py on them; this test keeps the fixtures equal to what the command writes.
func TestTheExportsMatchTheirFixtures(t *testing.T) {
	opened := authorCore(t)
	directory := filepath.Join(repo, "ruleset", "tests", "fixtures", "author")
	moduleDraft := draftSpell()
	moduleDraft["key"], moduleDraft["label"] = "authored_spark", "Authored Spark"
	for _, export := range []struct {
		name   string
		export map[string]any
		draft  map[string]any
	}{
		{"core", map[string]any{"shape": "core"}, draftSpell()},
		{"module", map[string]any{"shape": "module", "module": "authored"}, moduleDraft},
	} {
		t.Run(export.name, func(t *testing.T) {
			result := author(t, opened, map[string]any{"kind": "spell", "spells": []any{export.draft}, "export": export.export})
			written, _ := json.MarshalIndent(result["export"], "", "  ")
			written = append(written, '\n')
			var path string = filepath.Join(directory, export.name+".json")
			if *update {
				if err := os.MkdirAll(directory, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, written, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			fixture, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run go test ./core -run TestTheExportsMatchTheirFixtures -update)", err)
			}
			if string(fixture) != string(written) {
				t.Errorf("%s differs from the export; run with -update", path)
			}
		})
	}
}

func TestASummonBlockIsPricedAsTheSummonEffectItReplaces(t *testing.T) {
	opened := authorCore(t)
	for _, test := range []struct {
		summon map[string]any
		mana   float64
	}{
		{map[string]any{"group": "core:summon/familiar", "duration": 60}, 18},
		{map[string]any{"group": "core:summon/elemental", "duration": 120, "minion": true, "count": "1d3+1"}, 108},
	} {
		draft := map[string]any{"key": "authored_call", "label": "Authored Call", "description": "A test summon.",
			"color": "orange", "level": 2, "effects": []any{}, "summon": test.summon, "cost": map[string]any{"mana": test.mana}}
		result := author(t, opened, map[string]any{"kind": "spell", "spells": []any{draft}})
		spell := result["spells"].([]any)[0].(map[string]any)
		cost, _ := spell["cost"].(map[string]any)
		if cost == nil || cost["mana"].(float64) != test.mana {
			t.Errorf("%v: cost %v, want mana %v", test.summon, spell["cost"], test.mana)
		}
		if warnings := findings(t, result, "warnings"); len(warnings) != 0 {
			t.Errorf("%v: warnings %v", test.summon, warnings)
		}
	}
}
