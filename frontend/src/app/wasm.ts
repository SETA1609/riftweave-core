// The Go core compiled to WebAssembly (docs/backend/frontend.md#the-core-in-the-browser,
// docs/theory/webassembly.md). It runs `core.Run` in the page, so the app needs no server; it is
// a transport, and holds no rule.

/** What the Go entry point (backend/cmd/riftweave-wasm) sets on globalThis.riftweave. */
interface GoCore {
  run(command: string, input: string): { result?: string; error?: string };
}

/** The class wasm_exec.js defines: it supplies the module's imports and runs its Go program. */
interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

type Host = typeof globalThis & {
  Go?: new () => GoRuntime;
  riftweave?: GoCore;
  riftweaveReady?: (error: string | null) => void;
};

/** Where the two files come from: a page fetches them, a spec reads them from disk. */
export interface CoreSource {
  /** Defines globalThis.Go: wasm_exec.js, from the Go version that built the module. */
  glue(): Promise<void>;
  /** The module's bytes, or the response that streams them. */
  module(): Promise<Response | BufferSource>;
}

/** The core, started: one call per command, as the HTTP API's POST /v1/commands/{name}. */
export interface Core {
  run(command: string, input: unknown): Promise<unknown>;
}

/** A page's source: both files are served beside index.html (frontend/public). */
export const pageSource: CoreSource = {
  glue: () =>
    new Promise<void>((resolve, reject) => {
      if ((globalThis as Host).Go) {
        resolve();
        return;
      }
      const script = document.createElement('script');
      script.src = 'wasm_exec.js';
      script.onload = () => resolve();
      script.onerror = () => reject(new Error('wasm_exec.js did not load'));
      document.head.append(script);
    }),
  module: () => fetch('riftweave.wasm'),
};

/** Fetch, compile, and start the core; it resolves once the Go program has opened the ruleset. */
export async function startCore(source: CoreSource): Promise<Core> {
  await source.glue();
  const host = globalThis as Host;
  if (!host.Go) {
    throw new Error('wasm_exec.js did not define Go');
  }
  const go = new host.Go();
  const bytes = await source.module();
  const { instance } =
    bytes instanceof Response
      ? await WebAssembly.instantiateStreaming(bytes, go.importObject)
      : await WebAssembly.instantiate(bytes, go.importObject);
  const opened = new Promise<void>((resolve, reject) => {
    host.riftweaveReady = (error) => (error === null ? resolve() : reject(new Error(error)));
  });
  void go.run(instance); // resolves only when the Go program exits, which it does not
  await opened;
  const core = host.riftweave;
  if (!core) {
    throw new Error('the core did not start');
  }
  return {
    async run(command, input) {
      const { result, error } = core.run(command, JSON.stringify(input));
      if (error !== undefined) {
        throw new Error(error);
      }
      return JSON.parse(result ?? 'null');
    },
  };
}
