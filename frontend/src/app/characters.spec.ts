// The character list in the browser (FE_S07_04, docs/backend/files.md#the-character-list): the sync
// rule and its words are the Go CLI's (backend/cmd/riftweave/characters), and a character file the
// page exports is the fixture the Go tests import, byte for byte.
import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it } from 'vitest';
import synced from '../../../ruleset/tests/fixtures/files/character-synced.riftweave.json';
import plain from '../../../ruleset/tests/fixtures/files/character.riftweave.json';
import session from '../../../ruleset/tests/fixtures/files/session.riftweave.json';
import { API } from './api';
import { CHARACTER_STORE, Characters, characterFileName, memoryCharacterStore } from './characters';
import { encodeFile, type RiftweaveFile } from './files';

const text = (value: unknown): string => encodeFile(value as RiftweaveFile);
/** The synced fixture with its sync fields changed. */
function edited(change: Record<string, unknown>, data?: Record<string, unknown>): string {
  const file = structuredClone(synced) as Record<string, unknown> & { character: Record<string, unknown>; data: Record<string, unknown> };
  Object.assign(file.character, change);
  if (data) Object.assign(file.data, data);
  return text(file);
}

describe('the character list', () => {
  let characters: Characters;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        { provide: CHARACTER_STORE, useValue: memoryCharacterStore() },
        { provide: API, useValue: { checkActor: async () => ({}) } },
      ],
    });
    characters = TestBed.inject(Characters);
  });

  it('adds a character file with sync fields and exports it back byte for byte', async () => {
    const report = await characters.import(text(synced));
    expect(report.outcome).toBe('added');
    expect(characters.list().map((entry) => [entry.character.name, entry.character.revision])).toEqual([['Mira Vell', 3]]);
    const exported = await characters.export(synced.character.id);
    expect(exported.name).toBe('mira-vell-5f0c7a52.riftweave.json');
    expect(exported.text).toBe(text(synced));
  });

  it('replaces the saved character with a higher revision', async () => {
    await characters.import(text(synced));
    const report = await characters.import(edited({ revision: 4, changed_by: 'gm' }, { xp: 999 }));
    expect(report.outcome).toBe('replaced');
    expect(characters.list()[0].character.revision).toBe(4);
  });

  it('refuses a lower revision, in the CLI words, and changes nothing', async () => {
    await characters.import(text(synced));
    const report = await characters.import(edited({ revision: 2 }, { xp: 999 }));
    expect(report).toMatchObject({ outcome: 'refused', note: 'revision 2 is older than the saved revision 3; nothing changed' });
    expect((await characters.export(synced.character.id)).text).toBe(text(synced));
  });

  it('leaves the same revision with the same content unchanged; who and when are not content', async () => {
    await characters.import(text(synced));
    expect((await characters.import(edited({ changed_by: 'gm', changed_at: '2026-11-21T09:00:00Z' }))).outcome).toBe('unchanged');
  });

  it('asks on a divergence, shows both sides, and keeps the one picked', async () => {
    await characters.import(text(synced));
    const other = edited({ name: 'Mira the Bold' });
    const report = await characters.import(other);
    expect(report).toMatchObject({ outcome: 'diverged', note: 'both sides changed revision 3; keep the saved or the incoming one' });
    expect(report.saved?.character.name).toBe('Mira Vell');
    expect(report.incoming?.character.name).toBe('Mira the Bold');
    expect(characters.list()[0].character.name).toBe('Mira Vell');
    expect((await characters.import(other, { keep: 'saved' })).outcome).toBe('kept');
    expect((await characters.import(other, { keep: 'incoming' })).outcome).toBe('replaced');
    expect(characters.list()[0].character.name).toBe('Mira the Bold');
  });

  it('imports a character file with no sync fields as a new character, revision 1', async () => {
    const report = await characters.import(text(plain));
    expect(report.outcome).toBe('added');
    expect(report.character).toMatchObject({ name: 'Unnamed character', revision: 1, changed_by: 'player' });
    const named = await characters.import(text(plain), { name: 'Tam', changedBy: 'gm' });
    expect(named.character).toMatchObject({ name: 'Tam', revision: 1, changed_by: 'gm' });
    expect(named.character.id).not.toBe(report.character.id);
  });

  it('refuses a file that is not a character', async () => {
    await expect(characters.import(text(session))).rejects.toThrow('a session file is not a character');
  });

  it('duplicates as a new character and deletes', async () => {
    await characters.import(text(synced));
    const copy = await characters.duplicate(synced.character.id);
    expect(copy).toMatchObject({ name: 'Mira Vell (copy)', revision: 1, changed_by: 'player' });
    expect(copy.id).not.toBe(synced.character.id);
    await characters.remove(synced.character.id);
    expect(characters.list().map((entry) => entry.character.name)).toEqual(['Mira Vell (copy)']);
  });

  it('names a file as the CLI does: the name in lowercase words and the start of the id', () => {
    expect(characterFileName({ ...synced.character, changed_by: 'player' })).toBe('mira-vell-5f0c7a52.riftweave.json');
    expect(characterFileName({ ...synced.character, changed_by: 'player', name: '???' })).toBe('character-5f0c7a52.riftweave.json');
  });
});
