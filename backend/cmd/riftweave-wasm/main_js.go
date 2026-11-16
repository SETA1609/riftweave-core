//go:build js && wasm

package main

import (
	"embed"
	"io/fs"
	"syscall/js"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// ruleset is a copy of ruleset/data and ruleset/modules that scripts/build-wasm.sh makes before
// the build, so the core and its data come from one commit.
//
//go:embed all:ruleset
var ruleset embed.FS

func open() (*core.Core, error) {
	data, err := fs.Sub(ruleset, "ruleset/data")
	if err != nil {
		return nil, err
	}
	var modules []fs.FS
	entries, _ := fs.ReadDir(ruleset, "ruleset/modules")
	for _, entry := range entries {
		var directory string = "ruleset/modules/" + entry.Name()
		if _, statError := fs.Stat(ruleset, directory+"/manifest.json"); entry.IsDir() && statError == nil {
			module, err := fs.Sub(ruleset, directory)
			if err != nil {
				return nil, err
			}
			modules = append(modules, module)
		}
	}
	return core.New(data, modules...)
}

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
