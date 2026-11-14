// The app's one door to the HTTP API (docs/backend/frontend.md#the-client). Every type comes from
// the server's OpenAPI document (src/api/openapi.d.ts, `npm run generate`); nothing here is a rule.
import createClient from 'openapi-fetch';
import type { components, paths } from '../api/openapi';

type Json<T> = T extends { content: { 'application/json': infer Body } } ? Body : never;
type Post<Path extends keyof paths> = paths[Path] extends { post: infer Operation } ? Operation : never;
type Input<Path extends keyof paths> = Post<Path> extends { requestBody?: infer Body } ? Json<NonNullable<Body>> : never;
type Output<Path extends keyof paths> = Post<Path> extends { responses: { 200: infer Ok } } ? Json<Ok> : never;

export type StoredActor = components['schemas']['stored_actor'];
export type Session = components['schemas']['session'];
export type Attributes = components['schemas']['attributes'];
export type DerivedInput = Input<'/v1/commands/derived'>['actor'];
export type CombatInput = Input<'/v1/commands/combat'>;
export type CombatResult = Output<'/v1/commands/combat'>;
export type SessionStep = components['schemas']['session_step'];

/** The API at `baseUrl`: the server in a test, '' (the dev proxy) in the app. */
export function createApi(baseUrl: string) {
  const client = createClient<paths>({ baseUrl });

  async function post<Path extends keyof paths>(path: Path, body: Input<Path>): Promise<Output<Path>> {
    // openapi-fetch types each path; the generic wrapper only forwards the body and unwraps the result.
    const { data, error } = await (client.POST as (path: string, init: { body: unknown }) => Promise<{ data?: unknown; error?: unknown }>)(path, { body });
    if (error !== undefined) {
      throw new Error((error as { error?: string }).error ?? JSON.stringify(error));
    }
    return data as Output<Path>;
  }

  async function derived(actor: DerivedInput) {
    return post('/v1/commands/derived', { actor });
  }

  return {
    derived,
    combat: (input: CombatInput) => post('/v1/commands/combat', input),

    /** A new stored actor at level 1: its pools start full, at the maxima `derived` returns. */
    async newCharacter(attributes: Attributes, tags: string[] = []): Promise<StoredActor> {
      const stats = await derived({ attributes, level: 1 });
      return {
        attributes, level: 1, xp: 0, skills: {}, tags,
        growth: { hp: 0, mp: 0, sp: 0 },
        resources: { hp: stats.hp, mp: stats.mp, sp: stats.sp, ap: stats.ap },
        conditions: [], modifiers: [], flags: [], inventory: [], equipped: {},
      };
    },

    startSession: async (start: StoredActor): Promise<Session> => (await post('/v1/sessions', { start })).session,
    step: (session: Session, step: SessionStep) => post('/v1/sessions/step', { session, step }),
    replay: (session: Session) => post('/v1/sessions/replay', { session }),
  };
}

export type Api = ReturnType<typeof createApi>;
