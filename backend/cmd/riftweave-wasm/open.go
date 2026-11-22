package main

import (
	"io/fs"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// openRuleset opens the core from a tree that holds ruleset/data, ruleset/modules, and
// ruleset/schemas: the files scripts/build-wasm.sh embeds, or the repository in a test. The
// schemas let the author command check a draft in the page (BE_S09_01).
func openRuleset(root fs.FS) (*core.Core, error) {
	data, err := fs.Sub(root, "ruleset/data")
	if err != nil {
		return nil, err
	}
	schemas, err := fs.Sub(root, "ruleset/schemas")
	if err != nil {
		return nil, err
	}
	var modules []fs.FS
	entries, _ := fs.ReadDir(root, "ruleset/modules")
	for _, entry := range entries {
		var directory string = "ruleset/modules/" + entry.Name()
		if _, statError := fs.Stat(root, directory+"/manifest.json"); entry.IsDir() && statError == nil {
			module, err := fs.Sub(root, directory)
			if err != nil {
				return nil, err
			}
			modules = append(modules, module)
		}
	}
	return core.NewWithSchemas(data, schemas, modules...)
}
