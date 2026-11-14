// The app against a real riftweave-server (docs/backend/frontend.md#tests): it creates a
// character and resolves attacks through the API, and each attack matches its vector.
import { describe, expect, it } from 'vitest';
import combatVectors from '../../../ruleset/tests/vectors/combat.json';
import { createApi, type StoredActor } from './api';

const base: string = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env?.['RIFTWEAVE_API'] ?? 'http://127.0.0.1:8080';
const api = createApi(base);

const attributes = { str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 };

describe('a character', () => {
  it('reads every derived stat for its attributes and level', async () => {
    const stats = await api.derived({ attributes, level: 1 });
    for (const stat of ['hp', 'mp', 'sp', 'ap', 'carry', 'initiative']) {
      expect(typeof stats[stat as keyof typeof stats]).toBe('number');
    }
  });

  it('is created as a session whose actor is the one sent', async () => {
    const actor: StoredActor = await api.newCharacter(attributes);
    const session = await api.startSession(actor);
    expect(session.actor).toEqual(actor);
    expect(session.log).toEqual([]);
  });
});

describe('an attack', () => {
  type Vector = { name: string; function?: string; expect: Record<string, unknown> } & Record<string, unknown>;
  const vectors = (combatVectors as Vector[]).filter((vector) => !vector.function && !('error' in vector.expect)).slice(0, 12);

  it.each(vectors.map((vector) => [vector.name, vector] as const))('%s', async (_name, vector) => {
    const { name, expect: expected, ...input } = vector;
    const result = await api.combat(input as Parameters<typeof api.combat>[0]);
    for (const [field, value] of Object.entries(expected)) {
      expect((result as Record<string, unknown>)[field], `${name}: ${field}`).toEqual(value);
    }
  });
});
