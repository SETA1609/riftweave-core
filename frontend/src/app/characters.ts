// The character list in the browser (FE_S07_04, docs/backend/files.md#the-character-list): every
// saved character with its sync fields, in IndexedDB. Importing follows the sync rule, and its
// outcomes and notes are the Go CLI's (backend/cmd/riftweave/characters), so a file passes between
// a player's browser and a GM's CLI and back without either side merging anything.
import { InjectionToken, Service, inject, signal } from '@angular/core';
import { API, type StoredActor } from './api';
import { AppContext } from './context';
import { type FileModule, type RiftweaveFile, encodeFile, makeFile, missingModules, moduleAlert, parseFile, rulesWarning } from './files';
import { MODULE_VERSIONS } from './settings-screens';

/** The rules this build ships (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;

/** A character file's sync fields (common.json#/definitions/character). */
export interface CharacterSync {
  id: string;
  name: string;
  revision: number;
  changed_by: 'gm' | 'player';
  changed_at: string;
}

/** One saved character: its sync fields, its stored actor, and the rules and modules it was saved with. */
export interface SavedCharacter {
  character: CharacterSync;
  data: StoredActor;
  rules_version: string;
  modules: FileModule[];
}

export type Outcome = 'added' | 'replaced' | 'unchanged' | 'kept' | 'refused' | 'diverged';

/** What an import did. A refused or diverged import changes nothing; a diverged one holds both sides. */
export interface ImportReport {
  outcome: Outcome;
  character: CharacterSync;
  note?: string;
  saved?: SavedCharacter;
  incoming?: SavedCharacter;
  warnings: string[];
}

/** The choices an import takes, as the CLI's flags: --name, --as, --keep, --accept-loss. */
export interface ImportOptions {
  name?: string;
  changedBy?: 'gm' | 'player';
  keep?: 'saved' | 'incoming';
  acceptLoss?: boolean;
}

/** Where characters live. IndexedDB in the browser; a map in the unit specs. */
export interface CharacterStore {
  all(): Promise<SavedCharacter[]>;
  put(saved: SavedCharacter): Promise<void>;
  remove(id: string): Promise<void>;
}

/** The name a character's file is saved and exported as: the CLI's FileName. */
export function characterFileName(character: CharacterSync): string {
  const slug = character.name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
  return `${slug || 'character'}-${character.id.slice(0, 8)}.riftweave.json`;
}

/** The states of an import with sync fields, as small integers so the decision is one `switch` that
 * V8 compiles to a jump table: the order of the revisions (older, same, newer) is
 * `Math.sign(incoming - saved) + 1`, and a same revision splits on the content. */
const SYNC = { Older: 0, SameRevision: 1, Newer: 2, New: 3, Unchanged: 4, Diverged: 5 } as const;
type SyncState = (typeof SYNC)[keyof typeof SYNC];

/** Whether two sides hold the same character: the same name and actor. Who and when are not content. */
function sameContent(left: SavedCharacter, right: SavedCharacter): boolean {
  return left.character.name === right.character.name && encodeFile(makeFile('character', left.data, '')) === encodeFile(makeFile('character', right.data, ''));
}

function request<T>(call: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    call.onsuccess = () => resolve(call.result);
    call.onerror = () => reject(call.error);
  });
}

/** The characters in IndexedDB, database `riftweave-characters`, store `characters`, keyed by id.
 * A database of its own, so the saves' database (`riftweave`) needs no upgrade. */
export function indexedDbCharacterStore(name = 'riftweave-characters'): CharacterStore {
  let opened: Promise<IDBDatabase> | null = null;
  const database = (): Promise<IDBDatabase> => {
    opened ??= new Promise((resolve, reject) => {
      const open = indexedDB.open(name, 1);
      open.onupgradeneeded = () => open.result.createObjectStore('characters', { keyPath: 'character.id' });
      open.onsuccess = () => resolve(open.result);
      open.onerror = () => reject(open.error);
    });
    return opened;
  };
  const store = async (mode: IDBTransactionMode): Promise<IDBObjectStore> =>
    (await database()).transaction('characters', mode).objectStore('characters');
  const write = async (change: (objects: IDBObjectStore) => void): Promise<void> => {
    const objects = await store('readwrite');
    change(objects);
    await new Promise<void>((resolve, reject) => {
      objects.transaction.oncomplete = () => resolve();
      objects.transaction.onerror = () => reject(objects.transaction.error);
    });
  };
  return {
    all: async () => request((await store('readonly')).getAll() as IDBRequest<SavedCharacter[]>),
    put: (saved) => write((objects) => objects.put(saved)),
    remove: (id) => write((objects) => objects.delete(id)),
  };
}

/** Characters in memory, for the unit specs. */
export function memoryCharacterStore(): CharacterStore {
  const records = new Map<string, SavedCharacter>();
  return {
    all: async () => [...records.values()].map((saved) => structuredClone(saved)),
    put: async (saved) => void records.set(saved.character.id, structuredClone(saved)),
    remove: async (id) => void records.delete(id),
  };
}

export const CHARACTER_STORE = new InjectionToken<CharacterStore>('riftweave characters', {
  providedIn: 'root',
  factory: () => indexedDbCharacterStore(),
});

@Service()
export class Characters {
  private readonly store = inject(CHARACTER_STORE);
  private readonly api = inject(API);
  private readonly context = inject(AppContext, { optional: true });

  /** Every saved character, by name. */
  readonly list = signal<SavedCharacter[]>([]);

  async load(): Promise<void> {
    this.list.set((await this.store.all()).sort((a, b) => a.character.name.localeCompare(b.character.name) || a.character.id.localeCompare(b.character.id)));
  }

  private enabled(): string[] {
    return this.context?.modules() ?? [];
  }

  private now(): string {
    return new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
  }

  private find(id: string): SavedCharacter {
    const saved = this.list().find((entry) => entry.character.id === id);
    if (saved === undefined) throw new Error(`${id}: no saved character`);
    return saved;
  }

  /** The saved character as a file, stamped with the rules and modules it was saved with. Exporting
   * changes nothing, so the revision stays. */
  async export(id: string): Promise<{ name: string; text: string }> {
    await this.load();
    const saved = this.find(id);
    const file: RiftweaveFile & { character: CharacterSync } = { ...makeFile('character', saved.data, saved.rules_version, saved.modules), character: saved.character };
    return { name: characterFileName(saved.character), text: encodeFile(file) };
  }

  /** Apply a character file by the sync rule (docs/backend/files.md#the-sync-rule). */
  async import(text: string, options: ImportOptions = {}): Promise<ImportReport> {
    const parsed = parseFile(text);
    if ('error' in parsed) throw new Error(parsed.error);
    if ('settings' in parsed) throw new Error('a settings file is not a character');
    const file = parsed.file as RiftweaveFile & { character?: CharacterSync };
    if (file.riftweave !== 'character') throw new Error(`a ${file.riftweave} file is not a character`);
    await this.api.checkActor(file.data as StoredActor);
    const missing = missingModules(file, this.enabled());
    if (missing.length > 0 && !options.acceptLoss) throw new Error(moduleAlert(missing));
    const warnings: string[] = [];
    const rules = rulesWarning(file, RIFTWEAVE_RULES_VERSION);
    if (rules) warnings.push(rules);
    if (missing.length > 0) warnings.push(`imported without its modules: ${moduleAlert(missing)}`);
    await this.load();

    if (file.character === undefined) {
      const character: CharacterSync = {
        id: crypto.randomUUID(),
        name: options.name?.trim() || 'Unnamed character',
        revision: 1,
        changed_by: options.changedBy ?? 'player',
        changed_at: this.now(),
      };
      await this.put({ character, data: file.data as StoredActor, rules_version: file.rules_version, modules: file.modules });
      return { outcome: 'added', character, warnings };
    }
    const incoming: SavedCharacter = { character: file.character, data: file.data as StoredActor, rules_version: file.rules_version, modules: file.modules };
    const found = this.list().find((entry) => entry.character.id === incoming.character.id);
    if (found === undefined) {
      return this.keepIncoming(incoming, 'added', warnings);
    }
    const saved: SavedCharacter = found;
    let state = (Math.sign(incoming.character.revision - saved.character.revision) + 1) as SyncState;
    if (state === SYNC.SameRevision) state = sameContent(saved, incoming) ? SYNC.Unchanged : SYNC.Diverged;
    switch (state) {
      case SYNC.Newer:
        return this.keepIncoming(incoming, 'replaced', warnings);
      case SYNC.Older:
        return { outcome: 'refused', character: saved.character, warnings,
          note: `revision ${incoming.character.revision} is older than the saved revision ${saved.character.revision}; nothing changed` };
      case SYNC.Unchanged:
        return { outcome: 'unchanged', character: saved.character, warnings };
      default:
        // A divergence: the person keeps one side, or is asked. Never merged.
        switch (options.keep) {
          case 'incoming':
            return this.keepIncoming(incoming, 'replaced', warnings);
          case 'saved':
            return { outcome: 'kept', character: saved.character, warnings };
          default:
            return { outcome: 'diverged', character: saved.character, warnings, saved, incoming,
              note: `both sides changed revision ${saved.character.revision}; keep the saved or the incoming one` };
        }
    }
  }

  /** A copy under a new id, as a new character at revision 1. */
  async duplicate(id: string, changedBy: 'gm' | 'player' = 'player'): Promise<CharacterSync> {
    await this.load();
    const saved = this.find(id);
    const character: CharacterSync = { id: crypto.randomUUID(), name: `${saved.character.name} (copy)`, revision: 1, changed_by: changedBy, changed_at: this.now() };
    await this.put({ ...saved, character });
    return character;
  }

  async remove(id: string): Promise<void> {
    await this.store.remove(id);
    await this.load();
  }

  /** The modules enabled here, with their versions, for a file this app writes. */
  modules(): FileModule[] {
    return MODULE_VERSIONS.filter((module) => this.enabled().includes(module.id));
  }

  /** Save the incoming side and report the outcome. */
  private async keepIncoming(incoming: SavedCharacter, outcome: Outcome, warnings: string[]): Promise<ImportReport> {
    await this.put(incoming);
    return { outcome, character: incoming.character, warnings };
  }

  private async put(saved: SavedCharacter): Promise<void> {
    await this.store.put(saved);
    await this.load();
  }
}
