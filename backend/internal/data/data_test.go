package data

import (
	"errors"
	"io/fs"
	"os"
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
