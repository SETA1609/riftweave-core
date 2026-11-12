// Command riftweave-server is the HTTP API (docs/backend/api.md): one POST per command, the
// session endpoints, and an OpenAPI document built from the contract. It parses, calls
// core.Run, and formats; it holds no rule and no state between calls.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SETA1609/riftweave-core/backend/contract"
	"github.com/SETA1609/riftweave-core/backend/core"
)

// maxBody is the largest request body the server reads.
const maxBody = 1 << 20

func main() {
	var dataFlag, address string
	var devFlag bool
	flag.StringVar(&dataFlag, "data", "", "the ruleset data directory (default: RIFTWEAVE_DATA, then ruleset/data above the working directory)")
	flag.StringVar(&address, "addr", "127.0.0.1:8080", "the address to listen on")
	flag.BoolVar(&devFlag, "dev", false, "turn on dev mode (docs/backend/dev-mode.md)")
	flag.Parse()
	directory, err := rulesetDirectory(dataFlag, os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	opened, err := core.Open(directory)
	if err != nil {
		log.Fatalf("loading %s: %v", directory, err)
	}
	if devFlag {
		opened.EnableDev()
	}
	handler, err := newHandler(opened, filepath.Join(directory, "..", "schemas", "schema.json"))
	if err != nil {
		log.Fatal(err)
	}
	var httpServer *http.Server = &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("riftweave-server on http://%s (ruleset %s)", address, directory)
	log.Fatal(httpServer.ListenAndServe())
}

// rulesetDirectory finds the data as the agent CLI does (docs/backend/cli.md).
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

// api is the handler: the core, and the documents built once at start-up.
type api struct {
	core     *core.Core
	openAPI  []byte
	commands []byte
}

// newHandler builds the OpenAPI document and the command list, and returns the router.
func newHandler(opened *core.Core, schemaPath string) (http.Handler, error) {
	document, err := buildOpenAPI(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("OpenAPI document: %w", err)
	}
	type entry struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	var entries []entry = []entry{}
	for _, name := range core.Commands() {
		var file struct {
			Summary string `json:"summary"`
		}
		content, err := fs.ReadFile(contract.Files, name+".json")
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, &file); err != nil {
			return nil, err
		}
		entries = append(entries, entry{Name: name, Summary: file.Summary})
	}
	list, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}
	return &api{core: opened, openAPI: document, commands: list}, nil
}

// sessionPaths are the session endpoints and the operation each fills in.
var sessionPaths = map[string]string{"/v1/sessions": "start", "/v1/sessions/step": "step", "/v1/sessions/replay": "replay"}

func (server *api) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var path string = request.URL.Path
	switch {
	case path == "/openapi.json":
		if allowed(writer, request, http.MethodGet) {
			write(writer, http.StatusOK, server.openAPI)
		}
	case path == "/v1/commands":
		if allowed(writer, request, http.MethodGet) {
			write(writer, http.StatusOK, server.commands)
		}
	case strings.HasPrefix(path, "/v1/commands/"):
		if !allowed(writer, request, http.MethodPost) {
			return
		}
		if body, ok := readBody(writer, request); ok {
			server.run(writer, strings.TrimPrefix(path, "/v1/commands/"), body)
		}
	case sessionPaths[path] != "":
		if !allowed(writer, request, http.MethodPost) {
			return
		}
		body, ok := readBody(writer, request)
		if !ok {
			return
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
			fail(writer, http.StatusBadRequest, "the body is not a JSON object")
			return
		}
		operation, _ := json.Marshal(sessionPaths[path])
		fields["operation"] = operation
		input, _ := json.Marshal(fields)
		server.run(writer, "session", input)
	default:
		fail(writer, http.StatusNotFound, fmt.Sprintf("no endpoint %s", path))
	}
}

// run calls the core and maps its errors to a status (docs/backend/api.md#status-codes).
func (server *api) run(writer http.ResponseWriter, name string, body []byte) {
	output, err := server.core.Run(name, body)
	if err == nil {
		write(writer, http.StatusOK, output)
		return
	}
	var notDefined *core.NotDefinedError
	var missing *core.MissingInputError
	switch {
	case errors.Is(err, core.ErrUnknownCommand):
		fail(writer, http.StatusNotFound, err.Error())
	case errors.Is(err, core.ErrInput):
		fail(writer, http.StatusBadRequest, err.Error())
	case errors.As(err, &notDefined):
		write(writer, http.StatusNotFound, output)
	case errors.As(err, &missing):
		write(writer, http.StatusBadRequest, output)
	case errors.Is(err, core.ErrRule):
		write(writer, http.StatusUnprocessableEntity, output)
	default:
		fail(writer, http.StatusInternalServerError, err.Error())
	}
}

func allowed(writer http.ResponseWriter, request *http.Request, method string) bool {
	if request.Method == method {
		return true
	}
	writer.Header().Set("Allow", method)
	fail(writer, http.StatusMethodNotAllowed, fmt.Sprintf("%s takes %s", request.URL.Path, method))
	return false
}

func readBody(writer http.ResponseWriter, request *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, maxBody))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		fail(writer, http.StatusRequestEntityTooLarge, "the body is over 1 MiB")
		return nil, false
	}
	if err != nil {
		fail(writer, http.StatusBadRequest, err.Error())
		return nil, false
	}
	return body, true
}

func write(writer http.ResponseWriter, status int, body []byte) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = writer.Write(body)
}

func fail(writer http.ResponseWriter, status int, message string) {
	body, _ := json.Marshal(map[string]string{"error": message})
	write(writer, status, body)
}
