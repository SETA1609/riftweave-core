// Saves in the browser (docs/backend/frontend.md#saves): every session step is written to IndexedDB
// before the next one runs, and a backup file is the session JSON itself, so it replays in the
// agent CLI (`riftweave session`, operation replay) and through the API. localStorage is not used:
// it is small, and it blocks the page.
import { InjectionToken, Service, inject, signal } from '@angular/core';
import { API, type Session } from './api';

/** One saved game: the session and what the list shows of it. Dates are ISO strings. */
export interface SaveRecord {
  id: string;
  name: string;
  rules_version: string;
  last_played: string;
  last_backup: string | null;
  session: Session;
}

/** Where saves live. IndexedDB in the browser; a map in the unit specs. */
export interface SaveStore {
  all(): Promise<SaveRecord[]>;
  put(record: SaveRecord): Promise<void>;
  remove(id: string): Promise<void>;
}

/** What opening a backup file found, before the player keeps it. */
export interface Opened {
  name: string;
  saved: Session;
  replayed: Session;
  /** The replay rebuilt the saved actor from its start and log. */
  matches: boolean;
  /** The rules the file was made with, and the ones it was replayed under. */
  savedVersion: string;
  currentVersion: string;
}

/** A backup older than this many days, or none at all, is worth a reminder. */
export const BACKUP_DAYS = 7;

/** The backup file's name for a save. */
export function backupName(record: SaveRecord): string {
  return `${record.name.replace(/[^\w.-]+/g, '-').toLowerCase()}.riftweave.json`;
}

/** Whether a save's last backup is missing or older than BACKUP_DAYS at `now`. */
export function needsBackup(record: SaveRecord, now: Date = new Date()): boolean {
  if (record.last_backup === null) {
    return true;
  }
  return now.getTime() - new Date(record.last_backup).getTime() > BACKUP_DAYS * 24 * 60 * 60 * 1000;
}

/** A d100 roll from the browser's cryptographic generator; the caller writes it into the step. */
export function rollD100(): number {
  const [value] = crypto.getRandomValues(new Uint32Array(1));
  return 1 + (value % 100);
}

function request<T>(call: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    call.onsuccess = () => resolve(call.result);
    call.onerror = () => reject(call.error);
  });
}

/** The saves in IndexedDB, database `riftweave`, store `saves`, keyed by `id`. */
export function indexedDbStore(name = 'riftweave'): SaveStore {
  let opened: Promise<IDBDatabase> | null = null;
  const database = (): Promise<IDBDatabase> => {
    opened ??= new Promise((resolve, reject) => {
      const open = indexedDB.open(name, 1);
      open.onupgradeneeded = () => open.result.createObjectStore('saves', { keyPath: 'id' });
      open.onsuccess = () => resolve(open.result);
      open.onerror = () => reject(open.error);
    });
    return opened;
  };
  const store = async (mode: IDBTransactionMode): Promise<IDBObjectStore> => (await database()).transaction('saves', mode).objectStore('saves');
  const done = (objects: IDBObjectStore): Promise<void> =>
    new Promise((resolve, reject) => {
      objects.transaction.oncomplete = () => resolve();
      objects.transaction.onerror = () => reject(objects.transaction.error);
    });
  return {
    async all() {
      return request((await store('readonly')).getAll() as IDBRequest<SaveRecord[]>);
    },
    async put(record) {
      const objects = await store('readwrite');
      objects.put(record);
      await done(objects);
    },
    async remove(id) {
      const objects = await store('readwrite');
      objects.delete(id);
      await done(objects);
    },
  };
}

/** Saves in memory, for the unit specs. */
export function memoryStore(): SaveStore & { records: Map<string, SaveRecord> } {
  const records = new Map<string, SaveRecord>();
  return {
    records,
    all: async () => [...records.values()].map((record) => structuredClone(record)),
    put: async (record) => void records.set(record.id, structuredClone(record)),
    remove: async (id) => void records.delete(id),
  };
}

export const SAVE_STORE = new InjectionToken<SaveStore>('riftweave saves', { providedIn: 'root', factory: () => indexedDbStore() });

function isSession(value: unknown): value is Session {
  const session = value as Record<string, unknown> | null;
  return typeof session === 'object' && session !== null && 'start' in session && Array.isArray(session['log']) && 'actor' in session && 'ruleset' in session;
}

@Service()
export class Saves {
  private readonly store = inject(SAVE_STORE);
  private readonly api = inject(API);

  /** Every save, the latest played first. */
  readonly list = signal<SaveRecord[]>([]);
  /** The save in play: a reload resumes it. */
  readonly current = signal<SaveRecord | null>(null);
  /** Whether the browser agreed to keep this site's storage; null before it is asked. */
  readonly persisted = signal<boolean | null>(null);

  /** Read the saves and resume the latest. */
  async load(): Promise<void> {
    const all = (await this.store.all()).sort((a, b) => b.last_played.localeCompare(a.last_played));
    this.list.set(all);
    this.current.set(all[0] ?? null);
  }

  /** A new save for a session that just started; it is in the store when this returns. */
  async begin(name: string, session: Session): Promise<SaveRecord> {
    const record: SaveRecord = {
      id: crypto.randomUUID(),
      name,
      rules_version: session.rules_version ?? '',
      last_played: new Date().toISOString(),
      last_backup: null,
      session,
    };
    await this.write(record);
    void this.persist();
    return record;
  }

  /** Autosave: the session after a step, in the store before the next step runs. */
  async record(session: Session): Promise<void> {
    const current = this.current();
    if (current === null) {
      return;
    }
    await this.write({ ...current, session, last_played: new Date().toISOString() });
  }

  async resume(id: string): Promise<void> {
    const record = this.list().find((save) => save.id === id) ?? null;
    if (record !== null) {
      await this.write({ ...record, last_played: new Date().toISOString() });
    }
  }

  async remove(id: string): Promise<void> {
    await this.store.remove(id);
    await this.load();
  }

  /** The backup file of a save: the session JSON itself. Marks the save as backed up. */
  async backup(id: string): Promise<{ name: string; text: string }> {
    const record = this.list().find((save) => save.id === id);
    if (record === undefined) {
      throw new Error(`no save ${id}`);
    }
    await this.store.put({ ...record, last_backup: new Date().toISOString() });
    await this.load();
    this.current.set(this.list().find((save) => save.id === this.current()?.id) ?? this.current());
    return { name: backupName(record), text: `${JSON.stringify(record.session, null, 2)}\n` };
  }

  /** Read a backup file and replay it under the current rules. Nothing is saved until `keep`. */
  async open(name: string, text: string): Promise<Opened> {
    let parsed: unknown;
    try {
      parsed = JSON.parse(text);
    } catch {
      throw new Error(`${name} is not JSON`);
    }
    if (!isSession(parsed)) {
      throw new Error(`${name} is not a Riftweave save: it needs ruleset, start, log, and actor`);
    }
    const answer = await this.api.replay(parsed);
    const replayed = answer.session;
    return {
      name: name.replace(/\.riftweave\.json$|\.json$/, ''),
      saved: parsed,
      replayed,
      matches: answer.matches ?? false,
      savedVersion: parsed.rules_version ?? 'unknown',
      currentVersion: replayed.rules_version ?? '',
    };
  }

  /** Keep an opened file as a save: the replayed game, or the actor as the file saved it. */
  async keep(opened: Opened, which: 'replayed' | 'saved'): Promise<SaveRecord> {
    return this.begin(opened.name, which === 'replayed' ? opened.replayed : opened.saved);
  }

  /** Ask the browser to keep this site's storage (a browser may clear it otherwise). */
  async persist(): Promise<void> {
    if (this.persisted() !== null) {
      return;
    }
    try {
      this.persisted.set((await navigator.storage?.persist?.()) ?? false);
    } catch {
      this.persisted.set(false);
    }
  }

  private async write(record: SaveRecord): Promise<void> {
    await this.store.put(record);
    await this.load();
    this.current.set(this.list().find((save) => save.id === record.id) ?? record);
  }
}
