// Command riftweave is the agent CLI (docs/backend/cli.md): JSON in, JSON out, one call per
// process, no prompts and no colour. It parses, calls core.Run, and formats; it holds no
// rule.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SETA1609/riftweave-core/backend/cmd/riftweave/play"
	"github.com/SETA1609/riftweave-core/backend/contract"
	"github.com/SETA1609/riftweave-core/backend/core"
)

// Exit codes (docs/backend/cli.md#errors-and-exit-codes).
const (
	exitOK      = 0
	exitRule    = 1
	exitInput   = 2
	exitRuleset = 3
)

const usage = `usage: riftweave [--data <dir>] <command> [file]
       riftweave list
       riftweave describe <command>
       riftweave play

A command reads one JSON object from the file, or from stdin, and writes its result to
stdout. Errors are JSON on stderr. Exit 0 ok, 1 rule error, 2 bad input, 3 no ruleset.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv))
}

// fail writes one JSON error line on stderr and returns the exit code.
func fail(stderr io.Writer, code int, format string, args ...any) int {
	body, _ := json.Marshal(map[string]string{"error": fmt.Sprintf(format, args...)})
	fmt.Fprintf(stderr, "%s\n", body)
	return code
}

// run is the whole CLI, with its process boundaries as arguments so a test can drive it.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string) int {
	var dataFlag string
	if len(args) > 0 && strings.HasPrefix(args[0], "--data=") {
		dataFlag, args = strings.TrimPrefix(args[0], "--data="), args[1:]
	} else if len(args) > 1 && args[0] == "--data" {
		dataFlag, args = args[1], args[2:]
	}
	if len(args) == 0 {
		return fail(stderr, exitInput, "no command; run `riftweave help`")
	}
	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return exitOK
	case "list":
		return list(stdout, stderr)
	case "describe":
		if len(args) != 2 {
			return fail(stderr, exitInput, "describe takes one command")
		}
		content, err := fs.ReadFile(contract.Files, args[1]+".json")
		if err != nil || !slices.Contains(core.Commands(), args[1]) {
			return fail(stderr, exitInput, "unknown command %q", args[1])
		}
		_, _ = stdout.Write(content)
		return exitOK
	}
	if args[0] == "play" {
		return interactive(dataFlag, stdin, stdout, stderr, getenv)
	}
	return command(args, dataFlag, stdin, stdout, stderr, getenv)
}

// isTerminal reports whether the file is a character device, so play never waits on a pipe.
func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// interactive runs the human CLI (docs/backend/cli.md#human-mode).
func interactive(dataFlag string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string) int {
	if file, isFile := stdin.(*os.File); !isFile || !isTerminal(file) {
		return fail(stderr, exitInput, "play needs a terminal; agents call a command directly")
	}
	directory, err := rulesetDirectory(dataFlag, getenv)
	if err != nil {
		return fail(stderr, exitRuleset, "%v", err)
	}
	opened, err := core.Open(directory)
	if err != nil {
		return fail(stderr, exitRuleset, "loading %s: %v", directory, err)
	}
	var program *tea.Program = tea.NewProgram(play.New(opened.Run), tea.WithInput(stdin), tea.WithOutput(stdout))
	if _, err := program.Run(); err != nil {
		return fail(stderr, exitInput, "play needs a terminal: %v", err)
	}
	return exitOK
}

// list prints every command with the summary from its contract file.
func list(stdout, stderr io.Writer) int {
	type entry struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	var entries []entry = []entry{}
	for _, name := range core.Commands() {
		var file struct{ Summary string }
		content, err := fs.ReadFile(contract.Files, name+".json")
		if err == nil {
			err = json.Unmarshal(content, &file)
		}
		if err != nil {
			return fail(stderr, exitRuleset, "contract for %s: %v", name, err)
		}
		entries = append(entries, entry{Name: name, Summary: file.Summary})
	}
	body, _ := json.Marshal(entries)
	fmt.Fprintf(stdout, "%s\n", body)
	return exitOK
}

// command runs one core command on the input from a file or stdin.
func command(args []string, dataFlag string, stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string) int {
	var name string = args[0]
	if !slices.Contains(core.Commands(), name) {
		return fail(stderr, exitInput, "unknown command %q", name)
	}
	if len(args) > 2 {
		return fail(stderr, exitInput, "%s takes at most one input file", name)
	}
	var input []byte
	var err error
	if len(args) == 2 && args[1] != "-" {
		input, err = os.ReadFile(args[1])
	} else {
		input, err = io.ReadAll(stdin)
	}
	if err != nil {
		return fail(stderr, exitInput, "reading the input: %v", err)
	}
	directory, err := rulesetDirectory(dataFlag, getenv)
	if err != nil {
		return fail(stderr, exitRuleset, "%v", err)
	}
	opened, err := core.Open(directory)
	if err != nil {
		return fail(stderr, exitRuleset, "loading %s: %v", directory, err)
	}
	output, err := opened.Run(name, input)
	switch {
	case err == nil:
		fmt.Fprintf(stdout, "%s\n", output)
		return exitOK
	case errors.Is(err, core.ErrRule):
		fmt.Fprintf(stderr, "%s\n", output)
		return exitRule
	default:
		return fail(stderr, exitInput, "%v", err)
	}
}

// rulesetDirectory is the --data flag, else RIFTWEAVE_DATA, else the first ruleset/data from
// the current directory upward.
func rulesetDirectory(dataFlag string, getenv func(string) string) (string, error) {
	if dataFlag != "" {
		return dataFlag, nil
	}
	if fromEnvironment := getenv("RIFTWEAVE_DATA"); fromEnvironment != "" {
		return fromEnvironment, nil
	}
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		var candidate string = filepath.Join(directory, "ruleset", "data")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		var parent string = filepath.Dir(directory)
		if parent == directory {
			return "", errors.New("no ruleset/data found; pass --data <dir> or set RIFTWEAVE_DATA")
		}
		directory = parent
	}
}
