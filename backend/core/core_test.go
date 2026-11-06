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
	c, err := Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func readJSON(t *testing.T, path string) any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return v
}

func TestEveryVector(t *testing.T) {
	c := open(t)
	files, err := filepath.Glob(filepath.Join(repo, "ruleset", "tests", "vectors", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vector files: %v", err)
	}
	count := 0
	for _, file := range files {
		stem := strings.TrimSuffix(filepath.Base(file), ".json")
		vectors, ok := readJSON(t, file).([]any)
		if !ok {
			vectors = []any{readJSON(t, file)}
		}
		for _, raw := range vectors {
			vector := raw.(map[string]any)
			name, _ := vector["name"].(string)
			function, ok := vector["function"].(string)
			if !ok {
				function = stem
			}
			expect, ok := vector["expect"].(map[string]any)
			if !ok {
				t.Fatalf("%s: vector %q has no expect", file, name)
			}
			input := map[string]any{}
			for k, v := range vector {
				if k != "name" && k != "function" && k != "expect" {
					input[k] = v
				}
			}
			count++
			t.Run(stem+"/"+name, func(t *testing.T) {
				body, _ := json.Marshal(input)
				out, err := c.Run(function, body)
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
	for _, f := range files {
		if cmd, ok := readJSON(t, f).(map[string]any); ok && cmd["input_schema"] != nil {
			contract = append(contract, cmd["name"].(string))
		}
	}
	slices.Sort(contract)
	if got := Commands(); !slices.Equal(got, contract) {
		t.Errorf("commands %v, contract %v", got, contract)
	}
}

func TestBadInput(t *testing.T) {
	c := open(t)
	if _, err := c.Run("nope", []byte(`{}`)); !errors.Is(err, ErrUnknownCommand) {
		t.Errorf("unknown command: %v", err)
	}
	for _, body := range []string{`{"actor":{"attributes":{"str":1}},"roll":5,"extra":1}`, `{"actor":{"attributes":{"xyz":1}},"roll":5}`, `{`} {
		if _, err := c.Run("resolution", []byte(body)); !errors.Is(err, ErrInput) {
			t.Errorf("%s: want ErrInput, got %v", body, err)
		}
	}
	out, err := c.Run("loot", []byte(`{"table":"hoard_minor","rolls":[1]}`))
	if !errors.Is(err, ErrRule) || string(out) != `{"error":"hoard_minor takes 2 rolls, got 1"}` {
		t.Errorf("rule error: %s, %v", out, err)
	}
}
