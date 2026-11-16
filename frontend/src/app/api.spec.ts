// The app's API through both transports (docs/backend/frontend.md#tests): a real riftweave-server
// over HTTP, and the core compiled to WebAssembly (./scripts/build-wasm.sh writes it to public/).
// Each creates a character and resolves attacks, and each attack matches its vector, so the
// browser core is held to the same vectors as the server.
import { describe, expect, it } from 'vitest';
import combatVectors from '../../../ruleset/tests/vectors/combat.json';
import { createApi, httpTransport, wasmTransport, type StoredActor } from './api';
import { startCore, type CoreSource } from './wasm';

const base: string = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env?.['RIFTWEAVE_API'] ?? 'http://127.0.0.1:8080';

/** Node's file reader, imported by name so the browser build does not try to bundle it. */
const nodeFs = 'node:fs/promises';
async function readPublic(name: string): Promise<Uint8Array> {
  const { readFile } = (await import(/* @vite-ignore */ nodeFs)) as { readFile(path: string): Promise<Uint8Array> };
  const root: string = (globalThis as { process?: { cwd(): string } }).process?.cwd() ?? '.';
  return readFile(`${root}/public/${name}`);
}

/** The two files read from disk, as a spec has no page to fetch them from. */
const diskSource: CoreSource = {
  glue: async () => {
    new Function(new TextDecoder().decode(await readPublic('wasm_exec.js')))(); // a classic script: it defines globalThis.Go
  },
  module: async () => new Uint8Array(await readPublic('riftweave.wasm')),
};

const transports = [
  ['HTTP', () => httpTransport(base)],
  ['WebAssembly', () => wasmTransport(startCore(diskSource))],
] as const;

const attributes = { str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 };

type Vector = { name: string; function?: string; expect: Record<string, unknown> } & Record<string, unknown>;
const vectors = (combatVectors as Vector[]).filter((vector) => !vector.function && !('error' in vector.expect)).slice(0, 12);

describe.each(transports)('over %s', (_transport, open) => {
  const api = createApi(open());

  it('starts', async () => {
    await expect(api.ready).resolves.toBeUndefined();
  });

  it('reads every derived stat for its attributes and level', async () => {
    const stats = await api.derived({ attributes, level: 1 });
    for (const stat of ['hp', 'mp', 'sp', 'ap', 'carry', 'initiative']) {
      expect(typeof stats[stat as keyof typeof stats]).toBe('number');
    }
  });

  it('creates a character as a session whose actor is the one sent', async () => {
    const actor: StoredActor = await api.newCharacter(attributes);
    const session = await api.startSession(actor);
    expect(session.actor).toEqual(actor);
    expect(session.log).toEqual([]);
  });

  it('refuses with the rule message', async () => {
    const { name: _name, expect: _expected, ...input } = vectors[0];
    await expect(api.combat({ ...input, damage: { roll: 1, type: 'nope' } } as Parameters<typeof api.combat>[0])).rejects.toThrow(
      'damage type "nope" is not defined',
    );
  });

  it.each(vectors.map((vector) => [vector.name, vector] as const))('resolves an attack: %s', async (_name, vector) => {
    const { name, expect: expected, ...input } = vector;
    const result = await api.combat(input as Parameters<typeof api.combat>[0]);
    for (const [field, value] of Object.entries(expected)) {
      expect((result as Record<string, unknown>)[field], `${name}: ${field}`).toEqual(value);
    }
  });
});
