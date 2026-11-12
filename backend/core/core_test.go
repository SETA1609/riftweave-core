package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// The Go core is checked against the same files as the Python reference: every vector in
// ruleset/tests/vectors/ (docs/core/vectors.md), and the command list in backend/contract/.

const repo = "../.."

func open(t *testing.T) *Core {
	t.Helper()
	opened, err := Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	return opened
}

func readJSON(t *testing.T, path string) any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(content, &value); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return value
}

// coreOnly opens the ruleset with no module.
func coreOnly(t *testing.T) *Core {
	t.Helper()
	opened, err := New(os.DirFS(filepath.Join(repo, "ruleset", "data")))
	if err != nil {
		t.Fatal(err)
	}
	return opened
}

func TestEveryVectorWithoutModules(t *testing.T) {
	runCoreVectors(t, coreOnly(t))
}

func TestAModuleCommandNeedsItsModule(t *testing.T) {
	if _, err := coreOnly(t).Run("wuxing", []byte(`{"actor_phase":"fire","recipient_phase":"wood"}`)); err == nil {
		t.Errorf("wuxing with no module: want an error, got %v", err)
	}
}

func TestWuxingVectorsFromTheModule(t *testing.T) {
	opened := open(t)
	vectors := readJSON(t, filepath.Join(repo, "ruleset", "modules", "wuxing", "tests", "vectors.json")).([]any)
	if len(vectors) == 0 {
		t.Fatal("no wuxing vectors")
	}
	for _, raw := range vectors {
		var vector map[string]any = raw.(map[string]any)
		expect := vector["expect"].(map[string]any)
		input := map[string]any{}
		for key, value := range vector {
			if key != "name" && key != "function" && key != "expect" {
				input[key] = value
			}
		}
		body, _ := json.Marshal(input)
		out, err := opened.Run("wuxing", body)
		if err != nil {
			t.Fatalf("%s: %v", vector["name"], err)
		}
		var result map[string]any
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatal(err)
		}
		for field, want := range expect {
			if !reflect.DeepEqual(result[field], want) {
				t.Errorf("%s: field %q: got %v, want %v", vector["name"], field, result[field], want)
			}
		}
	}
}

func TestEveryVector(t *testing.T) {
	runCoreVectors(t, open(t))
}

func runCoreVectors(t *testing.T, opened *Core) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(repo, "ruleset", "tests", "vectors", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vector files: %v", err)
	}
	var count int = 0
	for _, file := range files {
		var stem string = strings.TrimSuffix(filepath.Base(file), ".json")
		vectors, ok := readJSON(t, file).([]any)
		if !ok {
			vectors = []any{readJSON(t, file)}
		}
		for _, raw := range vectors {
			var vector map[string]any = raw.(map[string]any)
			name, _ := vector["name"].(string)
			function, ok := vector["function"].(string)
			if !ok {
				function = stem
			}
			expect, ok := vector["expect"].(map[string]any)
			if !ok {
				t.Fatalf("%s: vector %q has no expect", file, name)
			}
			var input map[string]any = map[string]any{}
			for key, value := range vector {
				if key != "name" && key != "function" && key != "expect" {
					input[key] = value
				}
			}
			count++
			t.Run(stem+"/"+name, func(t *testing.T) {
				body, _ := json.Marshal(input)
				out, err := opened.Run(function, body)
				if err != nil && !errors.Is(err, ErrRule) {
					t.Fatalf("%s: %v", function, err)
				}
				var result map[string]any
				if err := json.Unmarshal(out, &result); err != nil {
					t.Fatalf("result is not JSON: %s", out)
				}
				for field, want := range expect {
					got, ok := result[field]
					if !ok {
						t.Errorf("result has no field %q: %s", field, out)
					} else if !reflect.DeepEqual(got, want) {
						t.Errorf("field %q: got %v, want %v", field, got, want)
					}
				}
			})
		}
	}
	t.Logf("%d vectors", count)
}

func TestCommandsMatchTheContract(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(repo, "backend", "contract", "*.json"))
	var contract []string
	for _, file := range files {
		if command, ok := readJSON(t, file).(map[string]any); ok && command["input_schema"] != nil {
			contract = append(contract, command["name"].(string))
		}
	}
	slices.Sort(contract)
	if got := Commands(); !slices.Equal(got, contract) {
		t.Errorf("commands %v, contract %v", got, contract)
	}
}

func TestBadInput(t *testing.T) {
	opened := open(t)
	if _, err := opened.Run("nope", []byte(`{}`)); !errors.Is(err, ErrUnknownCommand) {
		t.Errorf("unknown command: %v", err)
	}
	for _, body := range []string{`{"actor":{"attributes":{"str":1}},"roll":5,"extra":1}`, `{"actor":{"attributes":{"xyz":1}},"roll":5}`, `{`} {
		if _, err := opened.Run("resolution", []byte(body)); !errors.Is(err, ErrInput) {
			t.Errorf("%s: want ErrInput, got %v", body, err)
		}
	}
	out, err := opened.Run("loot", []byte(`{"table":"hoard_minor","rolls":[1]}`))
	if !errors.Is(err, ErrRule) || string(out) != `{"error":"hoard_minor takes 2 rolls, got 1"}` {
		t.Errorf("rule error: %s, %v", out, err)
	}
}

func TestTypedRuleErrors(t *testing.T) {
	opened := open(t)
	_, err := opened.Run("loot", []byte(`{"table":"no_such_table","rolls":[1]}`))
	var notDefined *NotDefinedError
	if !errors.Is(err, ErrRule) || !errors.As(err, &notDefined) || notDefined.Kind != "table" || notDefined.Key != "no_such_table" {
		t.Errorf("unknown table: want a NotDefinedError for table no_such_table, got %v", err)
	}
	_, err = opened.Run("economy", []byte(`{"operation":"convert"}`))
	var missing *MissingInputError
	if !errors.Is(err, ErrRule) || !errors.As(err, &missing) || missing.Needs != "cost" {
		t.Errorf("convert without cost: want a MissingInputError for cost, got %v", err)
	}
	_, err = opened.Run("loot", []byte(`{"table":"hoard_minor","rolls":[1]}`))
	if errors.As(err, &notDefined) || errors.As(err, &missing) {
		t.Errorf("wrong number of rolls is neither typed error, got %T", errors.Unwrap(err))
	}
}
