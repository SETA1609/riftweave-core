// Command riftweave-wasm is the core built to WebAssembly for the browser
// (docs/backend/frontend.md#the-core-in-the-browser, docs/theory/webassembly.md). It exposes one
// JavaScript function, riftweave.run(command, input), over core.Run: the same entry point as the
// CLI, the HTTP API, and MCP. It holds no rule.
package main

import (
	"errors"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// notServed are the commands the browser never runs, as over HTTP: dev operations are for the
// CLIs only (docs/backend/dev-mode.md).
var notServed = []string{"dev"}

// call runs one command on JSON input. It returns the JSON result, or the message of a
// refusal: a rule error's own message, as the HTTP API sends it in {"error": ...}.
func call(opened *core.Core, name, input string) (string, string) {
	if slices.Contains(notServed, name) {
		return "", name + " is not served in the browser; use the CLI with --dev"
	}
	output, err := opened.Run(name, []byte(input))
	if err == nil {
		return string(output), ""
	}
	var ruleError *core.RuleError
	if errors.As(err, &ruleError) {
		return "", ruleError.Message
	}
	return "", err.Error()
}
