package mcp

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// The MCP server is checked against the same vectors as the core and the agent CLI
// (docs/backend/mcp.md#checks).

const repo = "../.."

var (
	dataDirectory = filepath.Join(repo, "ruleset", "data")
	schemaPath    = filepath.Join(repo, "ruleset", "schemas", "schema.json")
)

func server(t *testing.T, dev bool) (*Server, *core.Core) {
	t.Helper()
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if dev {
		opened.EnableDev()
	}
	served, err := New(opened.Run, schemaPath, dev)
	if err != nil {
		t.Fatal(err)
	}
	return served, opened
}

// call sends one request and returns the decoded response.
func call(t *testing.T, served *Server, method string, params any) map[string]any {
	t.Helper()
	request, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	var response map[string]any
	if err := json.Unmarshal(served.Handle(request), &response); err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	return response
}

// tool calls a tool and returns its structured result and whether it is an error.
func tool(t *testing.T, served *Server, name string, arguments any) (any, bool) {
	t.Helper()
	response := call(t, served, "tools/call", map[string]any{"name": name, "arguments": arguments})
	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatalf("%s: no result in %v", name, response)
	}
	content := result["content"].([]any)[0].(map[string]any)
	var text any
	if err := json.Unmarshal([]byte(content["text"].(string)), &text); err != nil {
		t.Fatalf("%s: text is not JSON: %v", name, err)
	}
	if !reflect.DeepEqual(text, result["structuredContent"]) {
		t.Fatalf("%s: text %v and structuredContent %v differ", name, text, result["structuredContent"])
	}
	isError, _ := result["isError"].(bool)
	return text, isError
}

func toolNames(t *testing.T, served *Server) []string {
	t.Helper()
	var names []string
	for _, entry := range call(t, served, "tools/list", map[string]any{})["result"].(map[string]any)["tools"].([]any) {
		names = append(names, entry.(map[string]any)["name"].(string))
	}
	return names
}

func TestInitializeAndPing(t *testing.T) {
	served, _ := server(t, false)
	result := call(t, served, "initialize", map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{},
		"clientInfo": map[string]any{"name": "test", "version": "0"}})["result"].(map[string]any)
	if result["protocolVersion"] != "2025-06-18" {
		t.Fatalf("protocolVersion %v", result["protocolVersion"])
	}
	if _, ok := result["capabilities"].(map[string]any)["tools"]; !ok {
		t.Fatal("no tools capability")
	}
	if served.Handle([]byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)) != nil {
		t.Fatal("a notification has no answer")
	}
	if pong := call(t, served, "ping", nil); !reflect.DeepEqual(pong["result"], map[string]any{}) {
		t.Fatalf("ping %v", pong)
	}
}

func TestJSONRPCErrors(t *testing.T) {
	served, _ := server(t, false)
	for _, check := range []struct {
		line string
		code float64
	}{
		{`not json`, -32700},
		{`{"jsonrpc":"2.0","id":1,"method":"no/such"}`, -32601},
		{`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"derived","arguments":[]}}`, -32602},
	} {
		var response map[string]any
		if err := json.Unmarshal(served.Handle([]byte(check.line)), &response); err != nil {
			t.Fatal(err)
		}
		if code := response["error"].(map[string]any)["code"]; code != check.code {
			t.Fatalf("%s: code %v, want %v", check.line, code, check.code)
		}
	}
}

func TestEveryContractCommandIsAToolWithAStandaloneSchema(t *testing.T) {
	served, _ := server(t, false)
	names := toolNames(t, served)
	for _, command := range core.Commands() {
		if command == "dev" {
			continue
		}
		if !slices.Contains(names, command) {
			t.Errorf("no tool for %s", command)
		}
	}
	for _, entry := range call(t, served, "tools/list", map[string]any{})["result"].(map[string]any)["tools"].([]any) {
		var schema map[string]any = entry.(map[string]any)["inputSchema"].(map[string]any)
		if schema["type"] != "object" {
			t.Errorf("%v: inputSchema type %v", entry.(map[string]any)["name"], schema["type"])
		}
		defs, _ := schema["$defs"].(map[string]any)
		var walk func(node any)
		walk = func(node any) {
			switch value := node.(type) {
			case map[string]any:
				if reference, ok := value["$ref"].(string); ok {
					name, local := strings.CutPrefix(reference, "#/$defs/")
					if _, found := defs[name]; !local || !found {
						t.Errorf("%v: $ref %q does not stand alone", entry.(map[string]any)["name"], reference)
					}
				}
				for _, child := range value {
					walk(child)
				}
			case []any:
				for _, child := range value {
					walk(child)
				}
			}
		}
		walk(schema)
	}
}

func TestEveryVectorThroughTools(t *testing.T) {
	served, opened := server(t, false)
	files, err := filepath.Glob(filepath.Join(repo, "ruleset", "tests", "vectors", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vector files: %v", err)
	}
	var count int
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var vectors []map[string]any
		if err := json.Unmarshal(content, &vectors); err != nil {
			var single map[string]any
			if json.Unmarshal(content, &single) != nil {
				t.Fatalf("%s: %v", file, err)
			}
			vectors = []map[string]any{single}
		}
		var stem string = strings.TrimSuffix(filepath.Base(file), ".json")
		for _, vector := range vectors {
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
			want, coreErr := opened.Run(function, body)
			if coreErr != nil && !errors.Is(coreErr, core.ErrRule) {
				t.Fatalf("%s %v: core: %v", stem, vector["name"], coreErr)
			}
			got, isError := tool(t, served, function, input)
			var wanted any
			_ = json.Unmarshal(want, &wanted)
			if !reflect.DeepEqual(got, wanted) || isError != (coreErr != nil) {
				t.Errorf("%s %v: tool %v (error %v), core %s", stem, vector["name"], got, isError, want)
			}
			count++
		}
	}
	t.Logf("%d vectors through tools/call", count)
}

func storedActor(t *testing.T) map[string]any {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repo, "ruleset", "tests", "vectors", "stored_actor.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vectors []map[string]any
	if err := json.Unmarshal(content, &vectors); err != nil {
		t.Fatal(err)
	}
	return vectors[0]["actor"].(map[string]any)
}

func TestASessionPersistsAcrossCallsAndSavesAsTheSessionCommandDoes(t *testing.T) {
	served, opened := server(t, false)
	if _, isError := tool(t, served, "session_save", map[string]any{}); !isError {
		t.Fatal("session_save with no session held must be an error")
	}
	actor := storedActor(t)
	steps := []map[string]any{
		{"command": "derived", "input": map[string]any{}},
		{"command": "rest", "input": map[string]any{"occasion": "long_rest"}},
	}
	if _, isError := tool(t, served, "session_start", map[string]any{"start": actor}); isError {
		t.Fatal("session_start failed")
	}
	for _, step := range steps {
		result, isError := tool(t, served, "session_step", map[string]any{"step": step})
		if isError {
			t.Fatalf("session_step %v: %v", step["command"], result)
		}
		if _, ok := result.(map[string]any)["actor"]; !ok {
			t.Fatalf("session_step returns no actor: %v", result)
		}
	}
	saved, _ := tool(t, served, "session_save", map[string]any{})

	// The same run through the session command, as the agent CLI makes it.
	body, _ := json.Marshal(map[string]any{"operation": "start", "start": actor})
	output, err := opened.Run("session", body)
	if err != nil {
		t.Fatal(err)
	}
	var held map[string]any
	_ = json.Unmarshal(output, &held)
	for _, step := range steps {
		body, _ := json.Marshal(map[string]any{"operation": "step", "session": held["session"], "step": step})
		output, err := opened.Run("session", body)
		if err != nil {
			t.Fatal(err)
		}
		_ = json.Unmarshal(output, &held)
	}
	if !reflect.DeepEqual(saved, held["session"]) {
		t.Fatalf("saved %v, session command %v", saved, held["session"])
	}

	replayed, _ := tool(t, served, "session_replay", map[string]any{})
	if replayed.(map[string]any)["matches"] != true {
		t.Fatalf("replay %v", replayed)
	}
	fresh, _ := server(t, false)
	if _, isError := tool(t, fresh, "session_load", map[string]any{"session": saved}); isError {
		t.Fatal("session_load failed")
	}
	if again, _ := tool(t, fresh, "session_save", map[string]any{}); !reflect.DeepEqual(again, saved) {
		t.Fatal("a loaded session saves to the JSON it was loaded from")
	}
}

func TestDevToolsOnlyWithDev(t *testing.T) {
	plain, _ := server(t, false)
	names := toolNames(t, plain)
	for _, name := range []string{"dev", "session_dev"} {
		if slices.Contains(names, name) {
			t.Errorf("%s is listed without --dev", name)
		}
	}
	if response := call(t, plain, "tools/call", map[string]any{"name": "session_dev", "arguments": map[string]any{}}); response["error"] == nil {
		t.Error("calling session_dev without --dev is an unknown tool")
	}
	dev, _ := server(t, true)
	names = toolNames(t, dev)
	for _, name := range []string{"dev", "session_dev"} {
		if !slices.Contains(names, name) {
			t.Errorf("%s is not listed with --dev", name)
		}
	}
	if _, isError := tool(t, dev, "session_start", map[string]any{"start": storedActor(t)}); isError {
		t.Fatal("session_start failed")
	}
	result, isError := tool(t, dev, "session_dev", map[string]any{"operation": "level", "set": 5})
	if isError {
		t.Fatalf("session_dev: %v", result)
	}
	if level := result.(map[string]any)["actor"].(map[string]any)["level"]; level != float64(5) {
		t.Fatalf("level %v after a dev level 5", level)
	}
}
