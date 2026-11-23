package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/SETA1609/riftweave-core/backend/core"
)

var repo string = filepath.Join("..", "..", "..")

// The browser entry point answers every golden vector as core.Run does, so a page on the
// WebAssembly core sees the results the HTTP API returns.
func TestEveryVectorThroughCall(t *testing.T) {
	opened, err := core.Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(repo, "ruleset", "tests", "vectors", "*.json"))
	if len(files) == 0 {
		t.Fatal("no vector files")
	}
	for _, file := range files {
		content, _ := os.ReadFile(file)
		var vectors []map[string]any
		if json.Unmarshal(content, &vectors) != nil {
			var one map[string]any
			_ = json.Unmarshal(content, &one)
			vectors = []map[string]any{one}
		}
		var stem string = strings.TrimSuffix(filepath.Base(file), ".json")
		for _, vector := range vectors {
			function, ok := vector["function"].(string)
			if !ok {
				function = stem
			}
			expect, _ := vector["expect"].(map[string]any)
			var input map[string]any = map[string]any{}
			for key, value := range vector {
				if key != "name" && key != "function" && key != "expect" {
					input[key] = value
				}
			}
			body, _ := json.Marshal(input)
			name, _ := vector["name"].(string)
			t.Run(stem+"/"+name, func(t *testing.T) {
				output, refusal := call(opened, function, string(body))
				if want, isError := expect["error"]; isError {
					if refusal != want {
						t.Errorf("refusal %q, want %q", refusal, want)
					}
					return
				}
				if refusal != "" {
					t.Fatalf("refused: %s", refusal)
				}
				var result map[string]any
				_ = json.Unmarshal([]byte(output), &result)
				for field, want := range expect {
					if !reflect.DeepEqual(result[field], want) {
						t.Errorf("%s: got %v, want %v", field, result[field], want)
					}
				}
			})
		}
	}
}

func TestDevIsNotServedInTheBrowser(t *testing.T) {
	opened, err := core.Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	if _, refusal := call(opened, "dev", `{}`); !strings.Contains(refusal, "not served in the browser") {
		t.Errorf("dev ran in the browser: %q", refusal)
	}
	if _, refusal := call(opened, "nope", `{}`); !strings.Contains(refusal, "nope") {
		t.Errorf("an unknown command: %q", refusal)
	}
}

// The browser core opens with ruleset/schemas embedded beside the data (BE_S09_01), so author
// checks a spell in the page as it does over HTTP.
func TestTheBrowserCoreAuthorsWithTheSchemas(t *testing.T) {
	opened, err := openRuleset(os.DirFS(repo))
	if err != nil {
		t.Fatal(err)
	}
	loaded, refusal := call(opened, "author", `{"kind":"spell","operation":"load","spell":"fire_bolt"}`)
	if refusal != "" {
		t.Fatal(refusal)
	}
	var row struct{ Row json.RawMessage }
	_ = json.Unmarshal([]byte(loaded), &row)
	checked, refusal := call(opened, "author", `{"kind":"spell","spells":[`+string(row.Row)+`]}`)
	var result struct {
		Spells []struct {
			Problems []any
			Cost     struct{ Mana int64 }
		}
	}
	if refusal != "" || json.Unmarshal([]byte(checked), &result) != nil || len(result.Spells) != 1 || len(result.Spells[0].Problems) != 0 || result.Spells[0].Cost.Mana == 0 {
		t.Errorf("author in the browser: %s %s", checked, refusal)
	}
}

// The catalog answers in the browser as over HTTP (BE_S08_10).
func TestTheCatalogThroughCall(t *testing.T) {
	opened, err := core.Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	output, refusal := call(opened, "catalog", `{"operation":"get","collection":"weapons","ref":"core:weapon/longsword"}`)
	var result struct{ Row struct{ Row map[string]any } }
	if refusal != "" || json.Unmarshal([]byte(output), &result) != nil || result.Row.Row["label"] != "Longsword" {
		t.Errorf("output %s, refusal %q", output, refusal)
	}
}
