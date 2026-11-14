package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// copyOfTheRuleset reads ruleset/data/ into memory, so a test can break files in it.
func copyOfTheRuleset(t *testing.T) fstest.MapFS {
	t.Helper()
	var files fstest.MapFS = fstest.MapFS{}
	var root fs.FS = os.DirFS("../../../ruleset/data")
	err := fs.WalkDir(root, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := fs.ReadFile(root, name)
		files[name] = &fstest.MapFile{Data: content}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestTheRulesetLoads(t *testing.T) {
	if _, err := Load(copyOfTheRuleset(t)); err != nil {
		t.Fatal(err)
	}
}

func TestEveryBrokenFileIsReported(t *testing.T) {
	var files fstest.MapFS = copyOfTheRuleset(t)
	files["rest/core.json"] = &fstest.MapFile{Data: []byte(`{`)}
	files["origins/core.json"] = &fstest.MapFile{Data: []byte(`not json`)}
	files["resolution/core.json"] = &fstest.MapFile{Data: []byte(`{"resolution": {"crit_window_per_two_luck": 0}}`)}

	_, err := Load(files)
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("want one joined error, got %v", err)
	}
	var reported []error = joined.Unwrap()
	if len(reported) != 3 {
		t.Fatalf("want 3 errors, one per broken file, got %d:\n%v", len(reported), err)
	}
	for index, file := range []string{"resolution/core.json", "rest/core.json", "origins/core.json"} {
		if !strings.HasPrefix(reported[index].Error(), file+":") {
			t.Errorf("error %d: want it to name %s, got %v", index, file, reported[index])
		}
	}
}

func TestABrokenEconomySkipsItemsAndLootOnce(t *testing.T) {
	var files fstest.MapFS = copyOfTheRuleset(t)
	files["economy/core.json"] = &fstest.MapFile{Data: []byte(`[]`)}

	_, err := Load(files)
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) || len(joined.Unwrap()) != 2 {
		t.Fatalf("want the economy error and one note that items and loot were skipped, got:\n%v", err)
	}
}

// docs/core/composition.md#a-base-that-is-an-item
func TestAnItemBaseIsTheItemOfItsKey(t *testing.T) {
	ruleset, err := Load(copyOfTheRuleset(t))
	if err != nil {
		t.Fatal(err)
	}
	if dagger, ok := ruleset.Items["core:weapon/dagger"]; !ok || dagger.Dice != "1d4" {
		t.Errorf("core:weapon/dagger = %+v, %v; want the dagger base's 1d4", dagger, ok)
	}
	for _, ref := range []string{"core:weapon/bow", "core:accessory/ring_base", "core:armor/plate"} {
		if _, ok := ruleset.Items[ref]; ok {
			t.Errorf("%s loaded; its base is a shape only", ref)
		}
	}
}

func TestABaseWithoutIsItemIsReported(t *testing.T) {
	var files fstest.MapFS = copyOfTheRuleset(t)
	files["equipment/bases/weapons/dagger.json"] = &fstest.MapFile{Data: []byte(`{"key":"dagger"}`)}
	if _, err := Load(files); err == nil || !strings.Contains(err.Error(), "is_item") {
		t.Errorf("error = %v; want one naming is_item", err)
	}
}

// docs/core/composition.md#a-spell-built-on-a-spell (M3-26)
func TestASpellBuiltOnASpellLoadsAsItsFlatRow(t *testing.T) {
	var files fstest.MapFS = copyOfTheRuleset(t)
	var spells map[string]any
	if err := json.Unmarshal(files["spells/core.json"].Data, &spells); err != nil {
		t.Fatal(err)
	}
	var rows []any = spells["spells"].([]any)
	rows = append(rows, map[string]any{
		"key": "raise_lord", "label": "Raise Lord", "base": "core:spell/raise_thrall",
		"overrides": map[string]any{"level": 20, "cost": map[string]any{"mana": 60}},
	})
	spells["spells"] = rows
	encoded, _ := json.Marshal(spells)
	files["spells/core.json"] = &fstest.MapFile{Data: encoded}

	ruleset, err := Load(files)
	if err != nil {
		t.Fatal(err)
	}
	lord, exists := ruleset.Spells.Get("raise_lord")
	base, _ := ruleset.Spells.Get("raise_thrall")
	if !exists {
		t.Fatal("raise_lord is not loaded")
	}
	if lord.Level != 20 || lord.Cost.Mana != 60 || lord.Color != base.Color {
		t.Errorf("level 20, mana 60, and the base's color: %+v", lord)
	}
	if !reflect.DeepEqual(lord.Cost.Reagents, base.Cost.Reagents) || len(base.Cost.Reagents) == 0 {
		t.Errorf("an override of cost.mana keeps the base's reagents: %v", lord.Cost.Reagents)
	}
	if !reflect.DeepEqual(lord.Effects, base.Effects) {
		t.Errorf("the base's effects: %v", lord.Effects)
	}
	if base.Level != 8 || base.Cost.Mana != 36 {
		t.Errorf("the base is unchanged: %+v", base)
	}
}

func TestASpellBaseThatNamesABaseIsReported(t *testing.T) {
	var files fstest.MapFS = copyOfTheRuleset(t)
	var spells map[string]any
	_ = json.Unmarshal(files["spells/core.json"].Data, &spells)
	spells["spells"] = append(spells["spells"].([]any),
		map[string]any{"key": "first", "label": "First", "base": "core:spell/unlock", "overrides": map[string]any{}},
		map[string]any{"key": "second", "label": "Second", "base": "core:spell/first", "overrides": map[string]any{}},
		map[string]any{"key": "third", "label": "Third", "base": "core:spell/nope", "overrides": map[string]any{}})
	encoded, _ := json.Marshal(spells)
	files["spells/core.json"] = &fstest.MapFile{Data: encoded}
	_, err := Load(files)
	if err == nil || !strings.Contains(err.Error(), "second") || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("want errors for second (a base that names a base) and third (nope), got %v", err)
	}
}
