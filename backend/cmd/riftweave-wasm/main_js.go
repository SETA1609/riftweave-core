//go:build js && wasm

package main

import (
	"embed"
	"syscall/js"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// ruleset is a copy of ruleset/data, ruleset/modules, and ruleset/schemas that
// scripts/build-wasm.sh makes before the build, so the core, its data, and the schemas the author
// command checks against come from one commit.
//
//go:embed all:ruleset
var ruleset embed.FS

func open() (*core.Core, error) { return openRuleset(ruleset) }

// main opens the ruleset, sets globalThis.riftweave = { run }, tells the page through
// globalThis.riftweaveReady(error or null), and then waits: the Go program must keep running for
// JavaScript to call into it.
func main() {
	opened, err := open()
	var ready js.Value = js.Global().Get("riftweaveReady")
	if err != nil {
		if ready.Type() == js.TypeFunction {
			ready.Invoke(err.Error())
		}
		return
	}
	var api js.Value = js.Global().Get("Object").New()
	api.Set("run", js.FuncOf(func(_ js.Value, arguments []js.Value) any {
		if len(arguments) != 2 {
			return map[string]any{"error": "run takes a command and its JSON input"}
		}
		result, refusal := call(opened, arguments[0].String(), arguments[1].String())
		if refusal != "" {
			return map[string]any{"error": refusal}
		}
		return map[string]any{"result": result}
	}))
	js.Global().Set("riftweave", api)
	if ready.Type() == js.TypeFunction {
		ready.Invoke(js.Null())
	}
	select {}
}
