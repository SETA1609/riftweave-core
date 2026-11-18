// The file both apps trade (APP-30): the fixtures the Go tests read are read here, written back
// byte for byte, and a bad file or a missing module is reported with the Go CLI's words.
import { describe, expect, it } from 'vitest';
import character from '../../../ruleset/tests/fixtures/files/character.riftweave.json';
import invalid from '../../../ruleset/tests/fixtures/files/invalid.riftweave.json';
import missingModule from '../../../ruleset/tests/fixtures/files/missing-module.riftweave.json';
import session from '../../../ruleset/tests/fixtures/files/session.riftweave.json';
import spell from '../../../ruleset/tests/fixtures/files/spell.riftweave.json';
import { encodeFile, makeFile, missingModules, moduleAlert, parseFile, rulesWarning, type RiftweaveFile } from './files';

const asFile = (value: unknown) => value as RiftweaveFile;

describe('the file envelope', () => {
  it.each([['character', character], ['session', session], ['spell', spell], ['missing-module', missingModule], ['invalid', invalid]])(
    'reads the %s fixture and writes it back in the one layout',
    (_, fixture) => {
      const text = encodeFile(asFile(fixture));
      const parsed = parseFile(text);
      expect('file' in parsed && encodeFile(parsed.file)).toBe(text);
    },
  );

  it('makes the file the CLI makes for the same data', () => {
    const fixture = asFile(character);
    expect(encodeFile(makeFile('character', fixture.data, fixture.rules_version, []))).toBe(encodeFile(fixture));
  });

  it.each([
    ['[]', 'not a JSON object'],
    ['{"riftweave": "save", "rules_version": "x", "modules": [], "data": {}}', 'not one of character, session, spell'],
    ['{"riftweave": "character", "data": {}}', 'rules_version, modules, and data'],
  ])('refuses %s', (body, says) => {
    const parsed = parseFile(body);
    expect('error' in parsed && parsed.error).toContain(says);
  });

  it('hands a settings file to the settings import', () => {
    expect(parseFile('{"riftweave": "settings", "version": 1}')).toEqual({ settings: true });
  });

  it('names a module this browser does not load, with the rows at stake, in the CLI words', () => {
    const missing = missingModules(asFile(missingModule), []);
    expect(missing).toEqual([{ module: 'factions', rows: ['module:factions/accessory/warden_badge'] }]);
    expect(moduleAlert(missing)).toBe('module factions is not enabled here (module:factions/accessory/warden_badge): enable it, or confirm the loss');
    expect(missingModules(asFile(missingModule), ['factions'])).toEqual([]);
  });

  it('warns about a file from other rules', () => {
    const fixture = asFile(character);
    expect(rulesWarning(fixture, fixture.rules_version)).toBeNull();
    expect(rulesWarning(fixture, '9.9.9')).toBe(`made with rules ${fixture.rules_version}; this app runs 9.9.9`);
  });
});
