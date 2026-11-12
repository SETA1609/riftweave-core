package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// The CLI is checked as a process sees it: arguments, stdin, stdout, stderr, and the exit
// code (docs/backend/cli.md).

const repo = "../../.."

var dataDirectory string = filepath.Join(repo, "ruleset", "data")

func noEnvironment(string) string { return "" }

// call runs the CLI and returns its exit code, stdout, and stderr.
func call(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	var code int = run(args, strings.NewReader(stdin), &stdout, &stderr, noEnvironment)
	return code, stdout.String(), stderr.String()
}

func decode(t *testing.T, text string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("not JSON: %q: %v", text, err)
	}
	return value
}

func TestEveryVectorThroughTheCLI(t *testing.T) {
	opened, err := core.Open(dataDirectory)
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(repo, "ruleset", "tests", "vectors", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no vector files: %v", err)
	}
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
			name, _ := vector["name"].(string)
			t.Run(stem+"/"+name, func(t *testing.T) {
				body, _ := json.Marshal(input)
				want, coreErr := opened.Run(function, body)
				code, stdout, stderr := call(t, string(body), "--data", dataDirectory, function)
				switch {
				case coreErr == nil:
					if code != 0 || stderr != "" {
						t.Fatalf("exit %d, stderr %q", code, stderr)
					}
					if !reflect.DeepEqual(decode(t, stdout), decode(t, string(want))) {
						t.Fatalf("stdout %s, core %s", stdout, want)
					}
				case errors.Is(coreErr, core.ErrRule):
					if code != 1 || stdout != "" {
						t.Fatalf("exit %d, stdout %q", code, stdout)
					}
					if !reflect.DeepEqual(decode(t, stderr), decode(t, string(want))) {
						t.Fatalf("stderr %s, core %s", stderr, want)
					}
				default:
					t.Fatalf("core: %v", coreErr)
				}
			})
		}
	}
}

func TestInputFromAFile(t *testing.T) {
	var path string = filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, []byte(`{"actor":{"skills":{"blades":40},"attributes":{"lck":5}},"skill":"blades","roll":30}`), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := call(t, "", "--data", dataDirectory, "resolution", path)
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if result, ok := decode(t, stdout).(map[string]any); !ok || result["success"] != true {
		t.Fatalf("stdout %s", stdout)
	}
}

func TestListMatchesTheContract(t *testing.T) {
	code, stdout, _ := call(t, "", "list")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var listed []struct{ Name, Summary string }
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, command := range listed {
		if command.Summary == "" {
			t.Errorf("%s has no summary", command.Name)
		}
		names = append(names, command.Name)
	}
	if !reflect.DeepEqual(names, core.Commands()) {
		t.Fatalf("list %v, core %v", names, core.Commands())
	}
}

func TestDescribePrintsTheContractFile(t *testing.T) {
	code, stdout, _ := call(t, "", "describe", "loot")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	content, err := os.ReadFile(filepath.Join(repo, "backend", "contract", "loot.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decode(t, stdout), decode(t, string(content))) {
		t.Fatalf("describe loot is not backend/contract/loot.json")
	}
}

func TestBadInputExitsTwoWithAJSONError(t *testing.T) {
	cases := map[string]struct {
		stdin string
		args  []string
	}{
		"not JSON":         {"{", []string{"--data", dataDirectory, "resolution"}},
		"an unknown field": {`{"bogus": 1}`, []string{"--data", dataDirectory, "resolution"}},
		"two values":       {`{} {}`, []string{"--data", dataDirectory, "resolution"}},
		"unknown command":  {`{}`, []string{"--data", dataDirectory, "fireball"}},
		"describe nothing": {"", []string{"describe"}},
		"describe unknown": {"", []string{"describe", "fireball"}},
		"no command":       {"", nil},
		"missing file":     {"", []string{"--data", dataDirectory, "resolution", "/nonexistent/input.json"}},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			code, stdout, stderr := call(t, test.stdin, test.args...)
			if code != 2 || stdout != "" {
				t.Fatalf("exit %d, stdout %q", code, stdout)
			}
			if result, ok := decode(t, stderr).(map[string]any); !ok || result["error"] == "" {
				t.Fatalf("stderr %q", stderr)
			}
		})
	}
}

func TestAMissingRulesetExitsThree(t *testing.T) {
	code, stdout, stderr := call(t, `{}`, "--data", filepath.Join(t.TempDir(), "nowhere"), "resolution")
	if code != 3 || stdout != "" {
		t.Fatalf("exit %d, stdout %q", code, stdout)
	}
	decode(t, stderr)
}

func TestTheRulesetFromTheEnvironment(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var environment func(string) string = func(name string) string {
		if name == "RIFTWEAVE_DATA" {
			return dataDirectory
		}
		return ""
	}
	var input string = `{"actor":{"skills":{"blades":40},"attributes":{"lck":5}},"skill":"blades","roll":30}`
	if code := run([]string{"resolution"}, strings.NewReader(input), &stdout, &stderr, environment); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
}

// docs/backend/dev-mode.md: the agent CLI refuses dev without --dev and runs it with it.
func TestDevNeedsTheDevFlag(t *testing.T) {
	var session string = `{"session":{"ruleset":{"modules":[]},"log":[],` +
		`"start":{"attributes":{"end":5},"level":1,"xp":0,"skills":{},"tags":[],"growth":{"hp":0,"mp":0,"sp":0},"resources":{"hp":10,"mp":0,"sp":0,"ap":0},"conditions":[],"modifiers":[],"flags":[],"inventory":[],"equipped":{}},` +
		`"actor":{"attributes":{"end":5},"level":1,"xp":0,"skills":{},"tags":[],"growth":{"hp":0,"mp":0,"sp":0},"resources":{"hp":10,"mp":0,"sp":0,"ap":0},"conditions":[],"modifiers":[],"flags":[],"inventory":[],"equipped":{}}},` +
		`"pool":"hp","set":3}`
	code, _, stderr := call(t, session, "--data", dataDirectory, "dev", "resource")
	if code != exitRule || !strings.Contains(stderr, "dev mode is off") {
		t.Errorf("without --dev: exit %d, stderr %q", code, stderr)
	}
	code, stdout, stderr := call(t, session, "--dev", "--data", dataDirectory, "dev", "resource")
	if code != exitOK {
		t.Fatalf("with --dev: exit %d, stderr %q", code, stderr)
	}
	var result struct {
		Session struct {
			Dev   bool
			Actor struct{ Resources struct{ HP int64 } }
		}
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || !result.Session.Dev || result.Session.Actor.Resources.HP != 3 {
		t.Errorf("stdout %s", stdout)
	}
}

func TestMCPServesOnStdio(t *testing.T) {
	var lines string = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"derived","arguments":{"actor":{"attributes":{"end":5,"int":5,"agi":6,"str":5,"per":5},"level":1}}}}
`
	code, stdout, stderr := call(t, lines, "--data", dataDirectory, "mcp")
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	var answers []string = strings.Split(strings.TrimSpace(stdout), "\n")
	if len(answers) != 2 {
		t.Fatalf("want two answers (the notification gets none), got %q", stdout)
	}
	var result map[string]any = decode(t, answers[1]).(map[string]any)["result"].(map[string]any)
	if result["isError"] != false || result["structuredContent"].(map[string]any)["hp"] == nil {
		t.Fatalf("derived through mcp: %v", result)
	}
}
