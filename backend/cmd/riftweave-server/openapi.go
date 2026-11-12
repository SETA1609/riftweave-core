package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/contract"
)

// The OpenAPI document is built from backend/contract/ at start-up, never written by hand
// (docs/backend/api.md#openapi). Every $ref is rewritten to point inside the document.

const (
	componentPrefix = "#/components/schemas/"
	rulesetPrefix   = "ruleset_"
)

// commandFile is the part of a contract file the document reads.
type commandFile struct {
	Name          string `json:"name"`
	Summary       string `json:"summary"`
	Documentation string `json:"documentation"`
	InputSchema   any    `json:"input_schema"`
	ResultSchema  any    `json:"result_schema"`
}

// commandFiles reads every served command's contract file, sorted by name.
func commandFiles() ([]commandFile, error) {
	var files []commandFile
	for _, name := range servedCommands() {
		content, err := fs.ReadFile(contract.Files, name+".json")
		if err != nil {
			return nil, fmt.Errorf("contract for %s: %w", name, err)
		}
		var file commandFile
		if err := json.Unmarshal(content, &file); err != nil {
			return nil, fmt.Errorf("contract for %s: %w", name, err)
		}
		files = append(files, file)
	}
	return files, nil
}

// rewrite copies a schema with every $ref pointing into components.schemas. `local` is the
// prefix a `#/definitions/` ref gets in the file it came from.
func rewrite(node any, local string) (any, error) {
	switch value := node.(type) {
	case map[string]any:
		var copied map[string]any = make(map[string]any, len(value))
		for key, child := range value {
			if key == "$schema" {
				continue
			}
			if reference, isText := child.(string); isText && key == "$ref" {
				target, err := retarget(reference, local)
				if err != nil {
					return nil, err
				}
				copied[key] = target
				continue
			}
			rewritten, err := rewrite(child, local)
			if err != nil {
				return nil, err
			}
			copied[key] = rewritten
		}
		return copied, nil
	case []any:
		var copied []any = make([]any, len(value))
		for index, child := range value {
			rewritten, err := rewrite(child, local)
			if err != nil {
				return nil, err
			}
			copied[index] = rewritten
		}
		return copied, nil
	}
	return node, nil
}

func retarget(reference, local string) (string, error) {
	for prefix, component := range map[string]string{
		"common.json#/definitions/":                       "",
		"../../ruleset/schemas/schema.json#/definitions/": rulesetPrefix,
		"schema.json#/definitions/":                       rulesetPrefix,
		"#/definitions/":                                  local,
	} {
		if name, found := strings.CutPrefix(reference, prefix); found {
			return componentPrefix + component + name, nil
		}
	}
	return "", fmt.Errorf("$ref %q points outside the contract and schema.json", reference)
}

// definitions reads a file's `definitions`, rewritten, into the components under a prefix.
func definitions(content []byte, prefix string, components map[string]any) error {
	var file struct {
		Definitions map[string]any `json:"definitions"`
	}
	if err := json.Unmarshal(content, &file); err != nil {
		return err
	}
	for name, schema := range file.Definitions {
		rewritten, err := rewrite(schema, prefix)
		if err != nil {
			return fmt.Errorf("%s%s: %w", prefix, name, err)
		}
		components[prefix+name] = rewritten
	}
	return nil
}

func jsonBody(schema any, required bool) map[string]any {
	var body map[string]any = map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": schema}}}
	if required {
		body["required"] = true
	}
	return body
}

func errorResponses() map[string]any {
	var responses map[string]any = map[string]any{}
	for status, description := range map[string]string{
		"400": "The body is not the command's input, or a rule needs an input it was not given.",
		"404": "No such command, or the input names a key that does not exist.",
		"413": "The body is over 1 MiB.",
		"422": "The command refused for another rule reason.",
	} {
		responses[status] = map[string]any{"description": description, "content": map[string]any{
			"application/json": map[string]any{"schema": map[string]any{"$ref": componentPrefix + "error"}}}}
	}
	return responses
}

func post(operationID, summary string, request, result any) map[string]any {
	var responses map[string]any = errorResponses()
	responses["200"] = map[string]any{"description": "The result.", "content": map[string]any{"application/json": map[string]any{"schema": result}}}
	return map[string]any{"post": map[string]any{
		"operationId": operationID, "summary": summary, "requestBody": jsonBody(request, true), "responses": responses,
	}}
}

// buildOpenAPI is the OpenAPI 3.1 document for every command and the session endpoints.
func buildOpenAPI(schemaPath string) ([]byte, error) {
	var components map[string]any = map[string]any{
		"error": map[string]any{"type": "object", "properties": map[string]any{"error": map[string]any{"type": "string"}},
			"required": []any{"error"}},
	}
	common, err := fs.ReadFile(contract.Files, "common.json")
	if err != nil {
		return nil, err
	}
	if err := definitions(common, "", components); err != nil {
		return nil, err
	}
	shared, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	if err := definitions(shared, rulesetPrefix, components); err != nil {
		return nil, err
	}
	files, err := commandFiles()
	if err != nil {
		return nil, err
	}
	var paths map[string]any = map[string]any{}
	var sessionResult any
	for _, file := range files {
		input, err := rewrite(file.InputSchema, "")
		if err != nil {
			return nil, fmt.Errorf("%s input: %w", file.Name, err)
		}
		result, err := rewrite(file.ResultSchema, "")
		if err != nil {
			return nil, fmt.Errorf("%s result: %w", file.Name, err)
		}
		var operation map[string]any = post(file.Name, file.Summary, input, result)
		operation["post"].(map[string]any)["description"] = "The rules are in " + file.Documentation + "."
		paths["/v1/commands/"+file.Name] = operation
		if file.Name == "session" {
			sessionResult = result
		}
	}
	var ref = func(name string) map[string]any { return map[string]any{"$ref": componentPrefix + name} }
	var object = func(properties map[string]any, required ...string) map[string]any {
		var names []any
		for _, name := range required {
			names = append(names, name)
		}
		return map[string]any{"type": "object", "properties": properties, "required": names, "additionalProperties": false}
	}
	paths["/v1/sessions"] = post("session_start", "Start a session from a stored actor.",
		object(map[string]any{"start": ref("stored_actor")}, "start"), sessionResult)
	paths["/v1/sessions/step"] = post("session_step", "Run one step of a session.",
		object(map[string]any{"session": ref("session"), "step": ref("session_step")}, "session", "step"), sessionResult)
	paths["/v1/sessions/replay"] = post("session_replay", "Rebuild a session from its start and log.",
		object(map[string]any{"session": ref("session")}, "session"), sessionResult)
	paths["/v1/commands"] = map[string]any{"get": map[string]any{
		"operationId": "list_commands", "summary": "Every command with its contract summary.",
		"responses": map[string]any{"200": map[string]any{"description": "The commands, sorted by name.", "content": map[string]any{
			"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": object(map[string]any{
				"name": map[string]any{"type": "string"}, "summary": map[string]any{"type": "string"}}, "name", "summary")}}}}},
	}}
	return json.MarshalIndent(map[string]any{
		"openapi":    "3.1.0",
		"info":       map[string]any{"title": "Riftweave", "version": "1", "description": "The Riftweave ruleset as HTTP (docs/backend/api.md)."},
		"paths":      paths,
		"components": map[string]any{"schemas": components},
	}, "", "  ")
}
