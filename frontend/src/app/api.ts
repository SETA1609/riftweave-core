// The app's one door to the core (docs/backend/frontend.md#the-client): over HTTP when an API base
// URL is built in, else through the core compiled to WebAssembly in the page. Both follow the
// contract, so every type still comes from the server's OpenAPI document (src/api/openapi.d.ts,
// `npm run generate`); nothing here is a rule.
import { InjectionToken, inject } from '@angular/core';
import createClient from 'openapi-fetch';
import type { components, paths } from '../api/openapi';
import { AppContext } from './context';
import { pageSource, startCore, type Core } from './wasm';

type Json<T> = T extends { content: { 'application/json': infer Body } } ? Body : never;
type Post<Path extends keyof paths> = paths[Path] extends { post: infer Operation } ? Operation : never;
type Input<Path extends keyof paths> = Post<Path> extends { requestBody?: infer Body } ? Json<NonNullable<Body>> : never;
type Output<Path extends keyof paths> = Post<Path> extends { responses: { 200: infer Ok } } ? Json<Ok> : never;

export type StoredActor = components['schemas']['stored_actor'];
export type Session = components['schemas']['session'];
export type Attributes = components['schemas']['attributes'];
export type PrintInput = Input<'/v1/commands/print'>;
export type DerivedInput = Input<'/v1/commands/derived'>['actor'];
export type CombatInput = Input<'/v1/commands/combat'>;
export type CombatResult = Output<'/v1/commands/combat'>;
export type SessionStep = components['schemas']['session_step'];
export type ResolutionInput = Input<'/v1/commands/resolution'>;
export type ResolutionResult = Output<'/v1/commands/resolution'>;
export type EncounterInput = Input<'/v1/commands/encounter'>;
export type EncounterResult = Output<'/v1/commands/encounter'>;
export type AuthorInput = Input<'/v1/commands/author'>;
export type AuthorResult = Output<'/v1/commands/author'>;
export type CatalogInput = Input<'/v1/commands/catalog'>;
export type CatalogResult = Output<'/v1/commands/catalog'>;
export type BestiaryInput = Input<'/v1/commands/bestiary'>;
export type BestiaryResult = Output<'/v1/commands/bestiary'>;
export type MonsterSheetInput = Input<'/v1/commands/monster_sheet'>;
export type MonsterSheetResult = Output<'/v1/commands/monster_sheet'>;
export type KnowledgeMark = components['schemas']['knowledge_mark'];

/** One POST of the API: its path and body in, its result out, or an Error with the refusal. */
export interface Transport {
  post(path: string, body: unknown): Promise<unknown>;
  /** Resolves once the transport can answer: at once over HTTP, after start-up in WebAssembly. */
  ready: Promise<void>;
}

/** The HTTP API at `baseUrl`: the server in a test, '' (the dev proxy) in the app. */
export function httpTransport(baseUrl: string): Transport {
  const client = createClient<paths>({ baseUrl });
  return {
    ready: Promise.resolve(),
    async post(path, body) {
      // openapi-fetch types each path; the generic wrapper only forwards the body and unwraps the result.
      const { data, error } = await (client.POST as (path: string, init: { body: unknown }) => Promise<{ data?: unknown; error?: unknown }>)(path, { body });
      if (error !== undefined) {
        throw new Error((error as { error?: string }).error ?? JSON.stringify(error));
      }
      return data;
    },
  };
}

/** The session endpoints, as the server maps them onto the session command's operation. */
const sessionOperations: Record<string, string> = { '/v1/sessions': 'start', '/v1/sessions/step': 'step', '/v1/sessions/replay': 'replay' };

/** The core in the page: each path becomes the command the server would run for it. */
export function wasmTransport(core: Promise<Core>): Transport {
  return {
    ready: core.then(() => undefined),
    async post(path, body) {
      const started = await core;
      const operation = sessionOperations[path];
      if (operation !== undefined) {
        return started.run('session', { ...(body as object), operation });
      }
      return started.run(path.replace(/^\/v1\/commands\//, ''), body);
    },
  };
}

/** The network setting (docs/backend/app-layout.md#settings-modules-and-network): `local` is the
 * build's own transport, the core in the page (or the API a dev build names); `server` sends every
 * command to the HTTP API at the URL. The choice is read on each call, so changing it switches at
 * once. */
export function switchableTransport(local: Transport, server: (url: string) => Transport, choose: () => { mode: string; url: string }): Transport {
  const servers = new Map<string, Transport>();
  return {
    ready: local.ready,
    post(path, body) {
      const { mode, url } = choose();
      if (mode !== 'server' || !url) return local.post(path, body);
      if (!servers.has(url)) servers.set(url, server(url));
      return servers.get(url)!.post(path, body);
    },
  };
}

/** The API over a transport, or over HTTP at a base URL. */
export function createApi(over: Transport | string) {
  const transport: Transport = typeof over === 'string' ? httpTransport(over) : over;

  async function post<Path extends keyof paths>(path: Path, body: Input<Path>): Promise<Output<Path>> {
    return (await transport.post(path, body)) as Output<Path>;
  }

  async function derived(actor: DerivedInput) {
    return post('/v1/commands/derived', { actor });
  }

  return {
    ready: transport.ready,
    derived,
    combat: (input: CombatInput) => post('/v1/commands/combat', input),
    resolution: (input: ResolutionInput) => post('/v1/commands/resolution', input),
    /** Start a fight's initiative order, or end the current turn (the encounter command, BE_S08_07). */
    encounter: (input: EncounterInput) => post('/v1/commands/encounter', input),
    /** A printable page from the data (APP-11): its title and self-contained HTML. */
    print: (input: PrintInput) => post('/v1/commands/print', input),

    /** A new stored actor at level 1 from the core (stored_actor's new, BE_S23_03): its pools start full. */
    async newCharacter(attributes: Attributes, tags: string[] = []): Promise<StoredActor> {
      return (await post('/v1/commands/stored_actor', { new: { attributes, tags } })).actor;
    },

    /** Check a stored actor against the contract (the stored_actor command); it comes back unchanged. */
    checkActor: (actor: StoredActor) => post('/v1/commands/stored_actor', { actor }),
    /** Check spell drafts as the author command does (APP-10). */
    author: (input: AuthorInput) => post('/v1/commands/author', input),
    /** List, filter, and read the rows of a collection (the catalog command, BE_S08_10). */
    catalog: (input: CatalogInput) => post('/v1/commands/catalog', input),
    /** Mark what the players know of a monster, or view the bestiary as the GM or a player (BE_S08_05). */
    bestiary: (input: BestiaryInput) => post('/v1/commands/bestiary', input),
    /** A monster's derived stats and skills from its row (the monster_sheet command). */
    monsterSheet: (input: MonsterSheetInput) => post('/v1/commands/monster_sheet', input),

    startSession: async (start: StoredActor): Promise<Session> => (await post('/v1/sessions', { start })).session,
    step: (session: Session, step: SessionStep) => post('/v1/sessions/step', { session, step }),
    replay: (session: Session) => post('/v1/sessions/replay', { session }),
  };
}

export type Api = ReturnType<typeof createApi>;

/** The API base URL built in (angular.json `define`): '' is the dev proxy, null the core in the page. */
declare const RIFTWEAVE_API: string | null;

/** The API the screens share: the build's transport (HTTP when the build names an API, else
 * WebAssembly), or the server the network setting names. */
export const API = new InjectionToken<Api>('riftweave API', {
  providedIn: 'root',
  factory: () => {
    const context = inject(AppContext);
    const local: Transport = typeof RIFTWEAVE_API === 'string' ? httpTransport(RIFTWEAVE_API) : wasmTransport(startCore(pageSource));
    return createApi(switchableTransport(local, httpTransport, () => ({ mode: context.networkMode(), url: context.serverUrl() })));
  },
});
