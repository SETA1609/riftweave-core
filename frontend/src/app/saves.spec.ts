// The saves (docs/backend/frontend.md#saves) over a store in memory and a fake API: a step is in the
// store before the next runs, a reload resumes the latest, a backup is the session itself, and
// opening a file reports what its replay found. IndexedDB itself is covered by e2e/app.spec.ts.
import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it } from 'vitest';
import { API, type Session } from './api';
import { Playthrough } from './playthrough';
import { SAVE_STORE, Saves, backupName, memoryStore, needsBackup, rollD100, type SaveRecord } from './saves';

const actor = { xp: 0 } as unknown as Session['actor'];
function session(steps = 0, version = '0.1.0'): Session {
  return { rules_version: version, ruleset: { modules: [] }, start: actor, log: Array.from({ length: steps }, () => ({ command: 'rest', input: {} })), actor } as unknown as Session;
}

const events: string[] = [];
let replayMatches = true;
const fakeApi = {
  newCharacter: async () => actor,
  startSession: async () => session(),
  step: async (current: Session) => {
    events.push(`step ${current.log.length + 1}`);
    return { session: session(current.log.length + 1), result: {} };
  },
  replay: async (saved: Session) => ({ session: { ...saved, rules_version: '0.1.0' }, matches: replayMatches }),
};

function setup() {
  const store = memoryStore();
  const put = store.put;
  store.put = async (record: SaveRecord) => {
    events.push(`saved ${record.session.log.length}`);
    await put(record);
  };
  TestBed.configureTestingModule({ providers: [{ provide: API, useValue: fakeApi }, { provide: SAVE_STORE, useValue: store }] });
  return { store, saves: TestBed.inject(Saves) };
}

beforeEach(() => {
  events.length = 0;
  replayMatches = true;
});

describe('the saves', () => {
  it('write each playthrough step before the next one runs', async () => {
    setup();
    const fixture = TestBed.createComponent(Playthrough);
    await fixture.componentInstance.run();
    const steps = events.filter((event) => event.startsWith('step'));
    expect(steps.length).toBeGreaterThan(1);
    for (const [index, event] of events.entries()) {
      if (event.startsWith('step ') && index > 0) {
        const done = Number(event.split(' ')[1]) - 1;
        expect(events.slice(0, index)).toContain(`saved ${done}`);
      }
    }
  });

  it('resume the latest save after a reload', async () => {
    const { store, saves } = setup();
    await saves.begin('Older', session(1));
    await new Promise((resolve) => setTimeout(resolve, 5));
    await saves.begin('Newer', session(2));
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({ providers: [{ provide: API, useValue: fakeApi }, { provide: SAVE_STORE, useValue: store }] });
    const reloaded = TestBed.inject(Saves);
    await reloaded.load();
    expect(reloaded.list().map((save) => save.name)).toEqual(['Newer', 'Older']);
    expect(reloaded.current()?.name).toBe('Newer');
    expect(reloaded.current()?.rules_version).toBe('0.1.0');
  });

  it('back up a save as the session JSON itself and mark it backed up', async () => {
    const { saves } = setup();
    const record = await saves.begin('Mara the Bold', session(3));
    expect(needsBackup(saves.current()!)).toBe(true);
    const file = await saves.backup(record.id);
    expect(file.name).toBe('mara-the-bold.riftweave.json');
    expect(JSON.parse(file.text)).toEqual(record.session);
    expect(needsBackup(saves.list()[0])).toBe(false);
  });

  it('remind when the last backup is older than seven days', () => {
    const record = { last_backup: '2026-11-01T00:00:00Z' } as SaveRecord;
    expect(needsBackup(record, new Date('2026-11-07T00:00:00Z'))).toBe(false);
    expect(needsBackup(record, new Date('2026-11-09T00:00:00Z'))).toBe(true);
    expect(backupName({ name: 'A/B c' } as SaveRecord)).toBe('a-b-c.riftweave.json');
  });

  it('open a file that replays to its character', async () => {
    const { saves } = setup();
    const opened = await saves.open('mara.riftweave.json', JSON.stringify(session(2)));
    expect([opened.name, opened.matches, opened.savedVersion, opened.currentVersion]).toEqual(['mara', true, '0.1.0', '0.1.0']);
    const kept = await saves.keep(opened, 'replayed');
    expect(saves.current()?.id).toBe(kept.id);
  });

  it('flag an edited file', async () => {
    const { saves } = setup();
    replayMatches = false;
    const opened = await saves.open('edited.json', JSON.stringify(session(2)));
    expect(opened.matches).toBe(false);
  });

  it('replay a file from other rules under the current ones and say so', async () => {
    const { saves } = setup();
    const opened = await saves.open('old.riftweave.json', JSON.stringify(session(2, '0.0.1')));
    expect([opened.savedVersion, opened.currentVersion]).toEqual(['0.0.1', '0.1.0']);
  });

  it('refuse a file that is not a save', async () => {
    const { saves } = setup();
    await expect(saves.open('notes.json', '{"hello": 1}')).rejects.toThrow('is not a Riftweave save');
    await expect(saves.open('notes.txt', 'not json')).rejects.toThrow('is not JSON');
  });

  it('roll a d100 from the cryptographic generator', () => {
    for (let index = 0; index < 200; index++) {
      const roll = rollD100();
      expect(roll >= 1 && roll <= 100).toBe(true);
    }
  });
});
