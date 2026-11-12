// Package mcp serves the commands as Model Context Protocol tools on stdio (docs/backend/mcp.md):
// JSON-RPC 2.0, one message per line. Like every adapter it parses, calls the core, and
// formats; it holds no rule. The tools are generated from backend/contract/ at start-up.
package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/contract"
	"github.com/SETA1609/riftweave-core/backend/core"
)

// Runner runs one command on JSON input, as core.Core.Run does.
type Runner func(name string, input []byte) ([]byte, error)

// protocolVersions are the MCP versions the server speaks, newest first.
var protocolVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// JSON-RPC error codes.
const (
	parseError     = -32700
	methodNotFound = -32601
	invalidParams  = -32602
)

const defsPrefix = "#/$defs/"

// Tool is one entry of tools/list.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Server is one MCP session over a core: the tools, and the game session it holds.
type Server struct {
	run     Runner
	dev     bool
	tools   []Tool
	session json.RawMessage
}

// New builds the tools from the contract and the ruleset's schema.json at schemaPath.
func New(run Runner, schemaPath string, dev bool) (*Server, error) {
	pool, err := definitionPool(schemaPath)
	if err != nil {
		return nil, err
	}
	var server *Server = &Server{run: run, dev: dev}
	for _, name := range core.Commands() {
		if name == "dev" && !dev {
			continue
		}
		content, err := fs.ReadFile(contract.Files, name+".json")
		if err != nil {
			return nil, fmt.Errorf("contract for %s: %w", name, err)
		}
		var file struct {
			Summary       string         `json:"summary"`
			Doc           string         `json:"doc"`
			Documentation string         `json:"documentation"`
			InputSchema   map[string]any `json:"input_schema"`
			Definitions   map[string]any `json:"definitions"`
		}
		if err := json.Unmarshal(content, &file); err != nil {
			return nil, fmt.Errorf("contract for %s: %w", name, err)
		}
		var own map[string]any = map[string]any{}
		for key, schema := range file.Definitions {
			own[name+"_"+key] = rewrite(schema, name+"_")
		}
		schema, err := standalone(rewrite(file.InputSchema, name+"_").(map[string]any), pool, own)
		if err != nil {
			return nil, fmt.Errorf("%s input: %w", name, err)
		}
		var document string = file.Documentation
		if document == "" {
			document = file.Doc
		}
		server.tools = append(server.tools, Tool{Name: name, Description: file.Summary + " The rules are in " + document + ".", InputSchema: schema})
	}
	for _, session := range sessionTools(dev) {
		schema, err := standalone(session.InputSchema, pool, nil)
		if err != nil {
			return nil, fmt.Errorf("%s input: %w", session.Name, err)
		}
		session.InputSchema = schema
		server.tools = append(server.tools, session)
	}
	return server, nil
}

func object(properties map[string]any, required ...string) map[string]any {
	var schema map[string]any = map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func reference(name string) map[string]any { return map[string]any{"$ref": defsPrefix + name} }

// sessionTools hold one game session between calls (docs/backend/mcp.md#tools).
func sessionTools(dev bool) []Tool {
	var tools = []Tool{
		{"session_start", "Start a session from a stored actor and hold it.", object(map[string]any{"start": reference("stored_actor")}, "start")},
		{"session_step", "Run one step of the held session; returns the command's result and the actor after it.", object(map[string]any{"step": reference("session_step")}, "step")},
		{"session_save", "The held session, as the session command returns it, to save as JSON.", object(map[string]any{})},
		{"session_load", "Hold a saved session.", object(map[string]any{"session": reference("session")}, "session")},
		{"session_replay", "Rebuild the held session from its start and log; matches says whether it equals the saved actor.", object(map[string]any{})},
	}
	if dev {
		tools = append(tools, Tool{"session_dev", "Run one dev operation on the held session (docs/backend/dev-mode.md); returns the actor after it.",
			map[string]any{"type": "object", "properties": map[string]any{"operation": map[string]any{"type": "string"}}, "required": []string{"operation"}}})
	}
	return tools
}

// definitionPool is every shared definition a tool may reach, under its $defs name: common.json's
// under their own name, schema.json's with a ruleset_ prefix.
func definitionPool(schemaPath string) (map[string]any, error) {
	var pool map[string]any = map[string]any{}
	common, err := fs.ReadFile(contract.Files, "common.json")
	if err != nil {
		return nil, err
	}
	shared, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	for _, source := range []struct {
		content []byte
		prefix  string
	}{{common, ""}, {shared, "ruleset_"}} {
		var file struct {
			Definitions map[string]any `json:"definitions"`
		}
		if err := json.Unmarshal(source.content, &file); err != nil {
			return nil, err
		}
		for name, schema := range file.Definitions {
			pool[source.prefix+name] = rewrite(schema, source.prefix)
		}
	}
	return pool, nil
}

// rewrite copies a schema with every $ref pointing into $defs, and drops $schema. `local` is the
// prefix a `#/definitions/` ref gets in the file it came from.
func rewrite(node any, local string) any {
	switch value := node.(type) {
	case map[string]any:
		var copied map[string]any = make(map[string]any, len(value))
		for key, child := range value {
			if key == "$schema" {
				continue
			}
			if target, isText := child.(string); isText && key == "$ref" {
				copied[key] = retarget(target, local)
				continue
			}
			copied[key] = rewrite(child, local)
		}
		return copied
	case []any:
		var copied []any = make([]any, len(value))
		for index, child := range value {
			copied[index] = rewrite(child, local)
		}
		return copied
	}
	return node
}

func retarget(target, local string) string {
	for _, rule := range []struct{ prefix, into string }{
		{"common.json#/definitions/", ""},
		{"../../ruleset/schemas/schema.json#/definitions/", "ruleset_"},
		{"schema.json#/definitions/", "ruleset_"},
		{"#/definitions/", local},
	} {
		if name, found := strings.CutPrefix(target, rule.prefix); found {
			return defsPrefix + rule.into + name
		}
	}
	return target
}

// standalone adds to a schema, under $defs, every definition it reaches.
func standalone(schema map[string]any, pool, own map[string]any) (map[string]any, error) {
	var defs map[string]any = map[string]any{}
	var pending []any = []any{schema}
	for len(pending) > 0 {
		var node any = pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		switch value := node.(type) {
		case map[string]any:
			if target, ok := value["$ref"].(string); ok {
				name, local := strings.CutPrefix(target, defsPrefix)
				if !local {
					return nil, fmt.Errorf("$ref %q points outside the contract and schema.json", target)
				}
				if _, seen := defs[name]; !seen {
					definition, found := own[name]
					if !found {
						definition, found = pool[name]
					}
					if !found {
						return nil, fmt.Errorf("$ref %q names no definition", target)
					}
					defs[name] = definition
					pending = append(pending, definition)
				}
			}
			for _, child := range value {
				pending = append(pending, child)
			}
		case []any:
			pending = append(pending, value...)
		}
	}
	var copied map[string]any = make(map[string]any, len(schema)+1)
	for key, value := range schema {
		copied[key] = value
	}
	if len(defs) > 0 {
		copied["$defs"] = defs
	}
	return copied, nil
}

// Serve answers each line of `in` on `out` until `in` ends.
func (server *Server) Serve(in io.Reader, out io.Writer) error {
	var reader *bufio.Reader = bufio.NewReader(in)
	for {
		line, err := reader.ReadBytes('\n')
		if trimmed := bytes.TrimSpace(line); len(trimmed) > 0 {
			if answer := server.Handle(trimmed); answer != nil {
				if _, writeError := fmt.Fprintf(out, "%s\n", answer); writeError != nil {
					return writeError
				}
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func answer(id json.RawMessage, result any) []byte {
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
	return body
}

func failure(id json.RawMessage, code int, message string) []byte {
	if id == nil {
		id = json.RawMessage("null")
	}
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}})
	return body
}

// Handle answers one JSON-RPC message; a notification gets nil.
func (server *Server) Handle(line []byte) []byte {
	var message request
	if err := json.Unmarshal(line, &message); err != nil {
		return failure(nil, parseError, "not JSON: "+err.Error())
	}
	if message.ID == nil {
		return nil // a notification, such as notifications/initialized
	}
	switch message.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(message.Params, &params)
		var version string = protocolVersions[0]
		if slices.Contains(protocolVersions, params.ProtocolVersion) {
			version = params.ProtocolVersion
		}
		return answer(message.ID, map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "riftweave", "version": "0"},
		})
	case "ping":
		return answer(message.ID, map[string]any{})
	case "tools/list":
		return answer(message.ID, map[string]any{"tools": server.tools})
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(message.Params, &params); err != nil {
			return failure(message.ID, invalidParams, err.Error())
		}
		if len(params.Arguments) == 0 || string(params.Arguments) == "null" {
			params.Arguments = json.RawMessage("{}")
		}
		var arguments map[string]json.RawMessage
		if err := json.Unmarshal(params.Arguments, &arguments); err != nil {
			return failure(message.ID, invalidParams, "arguments must be a JSON object")
		}
		if !slices.ContainsFunc(server.tools, func(tool Tool) bool { return tool.Name == params.Name }) {
			return failure(message.ID, invalidParams, fmt.Sprintf("unknown tool %q", params.Name))
		}
		output, err := server.call(params.Name, params.Arguments, arguments)
		return answer(message.ID, toolResult(output, err))
	}
	return failure(message.ID, methodNotFound, fmt.Sprintf("method %q is not served", message.Method))
}

// toolResult is the JSON as text and as structuredContent; any error is an error result.
func toolResult(output []byte, err error) map[string]any {
	if err != nil && !errors.Is(err, core.ErrRule) {
		output, _ = json.Marshal(map[string]string{"error": err.Error()})
	}
	var structured any
	_ = json.Unmarshal(output, &structured)
	return map[string]any{
		"content":           []any{map[string]any{"type": "text", "text": string(output)}},
		"structuredContent": structured,
		"isError":           err != nil,
	}
}

var errNoSession = errors.New("no session is held; call session_start or session_load first")

// call runs one tool: a command on its arguments, or a session tool on the held session.
func (server *Server) call(name string, raw json.RawMessage, arguments map[string]json.RawMessage) ([]byte, error) {
	switch name {
	case "session_start":
		output, err := server.run("session", jsonObject("operation", "start", "start", arguments["start"]))
		if err != nil {
			return output, err
		}
		return server.hold(output)
	case "session_load":
		server.session = append(json.RawMessage(nil), arguments["session"]...)
		return server.session, nil
	case "session_save":
		if server.session == nil {
			return nil, errNoSession
		}
		return server.session, nil
	case "session_step", "session_replay", "session_dev":
		if server.session == nil {
			return nil, errNoSession
		}
	default:
		return server.run(name, raw)
	}
	var output []byte
	var err error
	switch name {
	case "session_step":
		output, err = server.run("session", jsonObject("operation", "step", "session", server.session, "step", arguments["step"]))
	case "session_replay":
		output, err = server.run("session", jsonObject("operation", "replay", "session", server.session))
	case "session_dev":
		var fields map[string]json.RawMessage = map[string]json.RawMessage{}
		for key, value := range arguments {
			fields[key] = value
		}
		fields["session"] = server.session
		body, _ := json.Marshal(fields)
		output, err = server.run("dev", body)
	}
	if err != nil {
		return output, err
	}
	var result struct {
		Session json.RawMessage `json:"session"`
		Result  json.RawMessage `json:"result"`
		Matches *bool           `json:"matches"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}
	server.session = result.Session
	if name == "session_replay" {
		return output, nil
	}
	var actor struct {
		Actor json.RawMessage `json:"actor"`
	}
	_ = json.Unmarshal(result.Session, &actor)
	var reply map[string]json.RawMessage = map[string]json.RawMessage{"actor": actor.Actor}
	if name == "session_step" {
		reply["result"] = result.Result
		if result.Result == nil {
			reply["result"] = json.RawMessage("null")
		}
	}
	return json.Marshal(reply)
}

// hold keeps the session a session command returned, and returns it.
func (server *Server) hold(output []byte) ([]byte, error) {
	var result struct {
		Session json.RawMessage `json:"session"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}
	server.session = result.Session
	return server.session, nil
}

// jsonObject builds a JSON object from key, value pairs; a value may be raw JSON.
func jsonObject(pairs ...any) []byte {
	var fields map[string]any = make(map[string]any, len(pairs)/2)
	for index := 0; index+1 < len(pairs); index += 2 {
		fields[pairs[index].(string)] = pairs[index+1]
	}
	body, _ := json.Marshal(fields)
	return body
}
