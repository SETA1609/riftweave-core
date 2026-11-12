package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// The server is checked against the same files as the core: every vector, run as an HTTP call
// (docs/backend/api.md).

const repo = "../../.."

func server(t *testing.T) *httptest.Server {
	t.Helper()
	opened, err := core.Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := newHandler(opened, filepath.Join(repo, "ruleset", "schemas", "schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	served := httptest.NewServer(handler)
	t.Cleanup(served.Close)
	return served
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

// call sends a request and returns the status and the decoded JSON body.
func call(t *testing.T, served *httptest.Server, method, path string, body []byte) (int, map[string]any) {
	t.Helper()
	request, err := http.NewRequest(method, served.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	response, err := served.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	content, _ := io.ReadAll(response.Body)
	var decoded map[string]any
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("%s %s: body is not a JSON object: %s", method, path, content)
	}
	return response.StatusCode, decoded
}

func TestEveryVectorOverHTTP(t *testing.T) {
	served := server(t)
	files, err := filepath.Glob(filepath.Join(repo, "ruleset", "tests", "vectors", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vector files: %v", err)
	}
	var count int
	for _, file := range files {
		var stem string = strings.TrimSuffix(filepath.Base(file), ".json")
		vectors, ok := readJSON(t, file).([]any)
		if !ok {
			vectors = []any{readJSON(t, file)}
		}
		for _, raw := range vectors {
			var vector map[string]any = raw.(map[string]any)
			function, ok := vector["function"].(string)
			if !ok {
				function = stem
			}
			var input map[string]any = map[string]any{}
			for key, value := range vector {
				if key != "name" && key != "function" && key != "expect" {
					input[key] = value
				}
			}
			body, _ := json.Marshal(input)
			count++
			t.Run(stem+"/"+vector["name"].(string), func(t *testing.T) {
				_, result := call(t, served, http.MethodPost, "/v1/commands/"+function, body)
				for field, want := range vector["expect"].(map[string]any) {
					if got := result[field]; !reflect.DeepEqual(got, want) {
						t.Errorf("field %q: got %v, want %v", field, got, want)
					}
				}
			})
		}
	}
	t.Logf("%d vectors over HTTP", count)
}

func TestStatusCodes(t *testing.T) {
	served := server(t)
	cases := []struct {
		name, method, path, body string
		status                   int
	}{
		{"ran", http.MethodPost, "/v1/commands/formula", `{"formula":"ap","actor":{"attributes":{"agi":6}}}`, http.StatusOK},
		{"no such command", http.MethodPost, "/v1/commands/nope", `{}`, http.StatusNotFound},
		{"no such path", http.MethodGet, "/v2/anything", ``, http.StatusNotFound},
		{"bad input", http.MethodPost, "/v1/commands/resolution", `{"actor":{"attributes":{"str":1}},"roll":5,"extra":1}`, http.StatusBadRequest},
		{"not JSON", http.MethodPost, "/v1/commands/resolution", `{`, http.StatusBadRequest},
		{"an unknown key", http.MethodPost, "/v1/commands/loot", `{"table":"no_such_table","rolls":[1]}`, http.StatusNotFound},
		{"a missing input", http.MethodPost, "/v1/commands/economy", `{"operation":"convert"}`, http.StatusBadRequest},
		{"a rule refuses", http.MethodPost, "/v1/commands/loot", `{"table":"hoard_minor","rolls":[1]}`, http.StatusUnprocessableEntity},
		{"wrong method", http.MethodGet, "/v1/commands/formula", ``, http.StatusMethodNotAllowed},
		{"dev is not served", http.MethodPost, "/v1/commands/dev", `{"operation":"xp","add":1,"session":{}}`, http.StatusNotFound},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			status, body := call(t, served, testCase.method, testCase.path, []byte(testCase.body))
			if status != testCase.status {
				t.Errorf("status %d, want %d (%v)", status, testCase.status, body)
			}
			if status != http.StatusOK {
				if _, ok := body["error"].(string); !ok {
					t.Errorf("an error body has an error sentence: %v", body)
				}
			}
		})
	}
	big := `{"formula":"ap","pad":"` + strings.Repeat("x", 1<<20) + `"}`
	if status, _ := call(t, served, http.MethodPost, "/v1/commands/formula", []byte(big)); status != http.StatusRequestEntityTooLarge {
		t.Errorf("a body over 1 MiB: status %d, want 413", status)
	}
}

func TestTheCommandListIsTheCore(t *testing.T) {
	served := server(t)
	response, err := served.Client().Get(served.URL + "/v1/commands")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var entries []struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(response.Body).Decode(&entries); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name)
		if entry.Summary == "" {
			t.Errorf("%s has no summary", entry.Name)
		}
	}
	if !slices.Equal(names, servedCommands()) || slices.Contains(names, "dev") {
		t.Errorf("listed %v, served %v; dev is never served", names, servedCommands())
	}
	if len(names) != len(core.Commands())-1 {
		t.Errorf("listed %d commands, the core has %d and only dev is left out", len(names), len(core.Commands()))
	}
}

func TestTheSessionEndpointsAreTheSessionCommand(t *testing.T) {
	served := server(t)
	actor := readJSON(t, filepath.Join(repo, "ruleset", "tests", "vectors", "stored_actor.json")).([]any)[0].(map[string]any)["actor"]
	start, _ := json.Marshal(map[string]any{"start": actor})
	status, started := call(t, served, http.MethodPost, "/v1/sessions", start)
	if status != http.StatusOK {
		t.Fatalf("start: %d %v", status, started)
	}
	direct, _ := json.Marshal(map[string]any{"operation": "start", "start": actor})
	_, viaCommand := call(t, served, http.MethodPost, "/v1/commands/session", direct)
	if !reflect.DeepEqual(started, viaCommand) {
		t.Errorf("/v1/sessions %v, the session command %v", started, viaCommand)
	}
	step, _ := json.Marshal(map[string]any{"session": started["session"], "step": map[string]any{"command": "derived", "input": map[string]any{}}})
	if status, stepped := call(t, served, http.MethodPost, "/v1/sessions/step", step); status != http.StatusOK || stepped["result"] == nil {
		t.Fatalf("step: %d %v", status, stepped)
	}
	replay, _ := json.Marshal(map[string]any{"session": started["session"]})
	if status, replayed := call(t, served, http.MethodPost, "/v1/sessions/replay", replay); status != http.StatusOK || replayed["matches"] != true {
		t.Errorf("replay: %d %v", status, replayed)
	}
}

// resolve follows a local JSON pointer (#/a/b) in the document.
func resolve(document any, pointer string) (any, error) {
	if !strings.HasPrefix(pointer, "#/") {
		return nil, errors.New("not a local reference")
	}
	var node any = document
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "#/"), "/") {
		object, ok := node.(map[string]any)
		if !ok {
			return nil, errors.New("no object at " + part)
		}
		if node, ok = object[part]; !ok {
			return nil, errors.New("no " + part)
		}
	}
	return node, nil
}

func references(node any, found *[]string) {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if text, ok := child.(string); ok && key == "$ref" {
				*found = append(*found, text)
			}
			references(child, found)
		}
	case []any:
		for _, child := range value {
			references(child, found)
		}
	}
}

func TestTheOpenAPIDocumentListsEveryCommandAndStandsAlone(t *testing.T) {
	served := server(t)
	status, document := call(t, served, http.MethodGet, "/openapi.json", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if document["openapi"] != "3.1.0" {
		t.Errorf("openapi %v, want 3.1.0", document["openapi"])
	}
	paths := document["paths"].(map[string]any)
	if paths["/v1/commands/dev"] != nil {
		t.Error("the document lists dev, which is not served")
	}
	for _, name := range servedCommands() {
		operation, ok := paths["/v1/commands/"+name].(map[string]any)["post"].(map[string]any)
		if !ok {
			t.Errorf("no POST /v1/commands/%s", name)
			continue
		}
		if operation["requestBody"] == nil || operation["responses"].(map[string]any)["200"] == nil {
			t.Errorf("%s has no request body or 200 response", name)
		}
	}
	for _, path := range []string{"/v1/commands", "/v1/sessions", "/v1/sessions/step", "/v1/sessions/replay"} {
		if paths[path] == nil {
			t.Errorf("no %s", path)
		}
	}
	var found []string
	references(document, &found)
	if len(found) == 0 {
		t.Fatal("no references at all")
	}
	for _, reference := range found {
		if _, err := resolve(document, reference); err != nil {
			t.Errorf("$ref %s does not resolve: %v", reference, err)
		}
	}
}
