# WebAssembly

Modified: 2026-11-15

Background for [the core in the browser](../backend/frontend.md#the-core-in-the-browser)
(APP-14): what WebAssembly is, how a Go program runs as one, and how this repo uses it so the
frontend needs no server. This page explains; it sets no rule. The rules stay in the Go core,
which the page runs unchanged.

Every quote below was read from a copy that `devops-tools fetch-docs` saved under `cache/docs/`,
pinned to the commit named with it. [V 2026-11-15]

## What it is

WebAssembly is a binary instruction format that browsers (and other hosts) run alongside
JavaScript. The specification's own summary:

> WebAssembly (abbreviated Wasm) is a *safe, portable, low-level code format* designed for
> efficient execution and compact representation. Its main goal is to enable high performance
> applications on the Web, but it does not make any Web-specific assumptions or provide
> Web-specific features, so it can be employed in other environments as well.

[WebAssembly/spec `document/core/intro/introduction.rst`](https://github.com/WebAssembly/spec/blob/943c5df07e84f11d2b83078964df5812530afbd2/document/core/intro/introduction.rst),
commit `943c5df` (footnote markers left out). Among its design goals:

> **Safe**: code is validated and executes in a memory-safe, sandboxed environment preventing
> data corruption or security breaches.

> **Language-independent**: does not privilege any particular language, programming model, or
> object model.

That is what makes it fit here: a language other than JavaScript (Go) compiles to it, and the
browser runs the result in the same sandbox as the page's own scripts.

## How a page runs one

MDN describes the two objects a page deals with
([mdn/content `webassembly/guides/concepts`](https://github.com/mdn/content/blob/742f106cbedfb40e6ebec554709d1ff9dcd29d52/files/en-us/webassembly/guides/concepts/index.md),
commit `742f106`):

> **Module**: Represents a WebAssembly binary that has been compiled by the browser into
> executable machine code. A Module is stateless […] A Module declares imports and exports just
> like an ES module.

> **Instance**: A Module paired with all the state it uses at runtime including Memories,
> Tables, and imported values.

and how the two sides talk:

> Given a WebAssembly instance, JavaScript code can synchronously call its exports, which are
> exposed as normal JavaScript functions. Arbitrary JavaScript functions can also be
> synchronously called by WebAssembly code by passing in those JavaScript functions as the
> imports to a WebAssembly instance.

A page loads a module with `WebAssembly.instantiateStreaming`, which
([mdn/content `instantiatestreaming_static`](https://github.com/mdn/content/blob/742f106cbedfb40e6ebec554709d1ff9dcd29d52/files/en-us/webassembly/reference/javascript_interface/instantiatestreaming_static/index.md))

> compiles and instantiates a WebAssembly module directly from a streamed underlying source.
> This is the most efficient, optimized way to load Wasm code.

So loading is three steps: fetch the `.wasm` file, compile and instantiate it with the imports it
needs, then call what it exports.

## How Go builds one

Go has a WebAssembly target. The Go wiki
([go.dev/wiki/WebAssembly](https://go.dev/wiki/WebAssembly), fetched 2026-11-15) shows the build:

> `$ GOOS=js GOARCH=wasm go build -o main.wasm`
>
> That will build the package and produce an executable WebAssembly module file named
> main.wasm.

A Go program needs more from its host than plain WebAssembly gives (a clock, random numbers,
memory management for its garbage collector), so Go ships a JavaScript support file that
supplies those imports and starts the program: `wasm_exec.js`, copied from
`$(go env GOROOT)/lib/wasm/`. Its version must match the compiler's:

> Note: The same major Go version of the compiler and wasm_exec.js support file must be used
> together. That is, if main.wasm file is compiled using Go version 1.N, the corresponding
> wasm_exec.js file must also be copied from Go version 1.N. Other combinations are not
> supported.

The page then runs it as the wiki's example does: `const go = new Go();`, then
`WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)`, then
`go.run(result.instance)`.

### Calling Go from JavaScript

The program reaches the page through the `syscall/js` package
([golang/go `src/syscall/js/js.go`](https://github.com/golang/go/blob/go1.26.1/src/syscall/js/js.go),
tag `go1.26.1`, commit `e87b10e`):

> Package js gives access to the WebAssembly host environment when using the js/wasm
> architecture. Its API is based on JavaScript semantics.
>
> This package is EXPERIMENTAL. Its current scope is only to allow tests to run, but not yet to
> provide a comprehensive API for users. It is exempt from the Go compatibility promise.

A Go function becomes a JavaScript one through `js.FuncOf`
([`src/syscall/js/func.go`](https://github.com/golang/go/blob/go1.26.1/src/syscall/js/func.go)):

> Invoking the wrapped Go function from JavaScript will pause the event loop and spawn a new
> goroutine. […] As a consequence, if one wrapped function blocks, JavaScript's event loop is
> blocked until that function returns. Hence, calling any async JavaScript API, which requires
> the event loop, like fetch (http.Client), will cause an immediate deadlock.

Two consequences for any Go program in a page: `main` must not return, or the functions it
exported stop working, so it ends in `select {}`; and an exported function must not wait on
the network. The core does neither: a command is pure computation over data already in memory.

### Size

Go's WebAssembly files are large, and the wiki says so:

> At present, Go generates large Wasm files, with the smallest possible size being around ~2MB.
> If your Go code imports libraries, this file size can increase dramatically. 10MB+ is common.

It names two remedies: compress the file when serving it (gzip or Brotli), or build with TinyGo,
which "supports a subset of the Go language".

## How this repo uses it

```
ruleset/data, ruleset/modules ──copied by scripts/build-wasm.sh──▶ backend/cmd/riftweave-wasm/ruleset
backend/cmd/riftweave-wasm (go:embed the copy, core.New on it)
        │  GOOS=js GOARCH=wasm go build
        ▼
frontend/public/riftweave.wasm  +  wasm_exec.js (same Go)  ──served beside index.html
        │  fetch, instantiateStreaming, go.run
        ▼
globalThis.riftweave.run(command, inputJSON) → { result } or { error }
        │
frontend/src/app/api.ts: wasmTransport, the same paths as the HTTP API
```

- **One entry point.** `riftweave.run(command, input)` calls `core.Run`, the function behind
  the CLI, the HTTP API, and MCP. The page sends the same JSON the API takes and gets the same
  JSON back, so the OpenAPI types still describe both, and a rule never moves into TypeScript.
- **The data is inside the binary.** The build copies `ruleset/data` and `ruleset/modules` next
  to the entry point and embeds them with `go:embed`, and the core reads them through the same
  `fs.FS` loader it uses on disk. The core and its data then come from one commit; a page cannot
  load rules from one version and data from another.
- **Ready before use.** The Go program opens the ruleset, sets `globalThis.riftweave`, and calls
  `globalThis.riftweaveReady`; the app shows "Starting the rules engine…" until then, and an
  error if the ruleset fails to open.
- **Standard Go, not TinyGo.** The core uses `encoding/json` with reflection, `html/template`
  (the printouts), and `embed` across its packages; TinyGo supports only part of the standard
  library, and the issue allowed it only if it covered what the core uses. The standard build
  works today and is checked by the same vectors. [A]
- **The size,** measured on 2026-11-15 with Go 1.26.1 (`-trimpath -ldflags="-s -w"`):
  `riftweave.wasm` is 9,326,790 bytes, 2,321,453 bytes gzipped; about 1 MB of it is the
  embedded data. A host that compresses sends the smaller number (GitHub Pages serves gzip [A]).
  The core starts in about 0.1 s once the file is compiled. [V 2026-11-15,
  `./scripts/build-wasm.sh` and Node 24.22.1]
- **Checked by the same vectors.** `backend/cmd/riftweave-wasm/call_test.go` runs every golden
  vector through the entry point's `call`, and the frontend specs and the Playwright tests run
  through both transports. [V 2026-11-15]

## What it is for

The frontend can then be a static site: no server to run or pay for, and a person with no
technical background opens a URL and plays (APP-15 publishes it on GitHub Pages, APP-16 keeps
saves in the browser). The server stays for the CLI, MCP, and anyone who wants the API; the
browser build adds a transport, not a second implementation.

## What it is not for

- **Hiding rules.** The `.wasm` file and its data can be read by anyone who loads the page, as
  the JSON can today. The ruleset is open (CC BY 4.0); nothing here is a secret.
- **Speed.** One command is small; the server and the page both answer in well under a frame.
  The reason is having no server, not performance. [A]
