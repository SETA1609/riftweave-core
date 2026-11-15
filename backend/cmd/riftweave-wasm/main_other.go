//go:build !(js && wasm)

package main

import (
	"fmt"
	"os"
)

// main says how to build it: this command only runs as WebAssembly in a page.
func main() {
	fmt.Fprintln(os.Stderr, "riftweave-wasm runs in a browser: build it with ./scripts/build-wasm.sh (GOOS=js GOARCH=wasm)")
	os.Exit(2)
}
