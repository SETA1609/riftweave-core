// The settings file (APP-25): the fixture the Go tests read is read here with the same values,
// written back byte for byte, unknown keys included, and a bad file is refused with the Go CLI's
// messages and changes nothing.
import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it } from 'vitest';
import fixture from '../../../ruleset/tests/fixtures/settings/settings.json';
import { AppContext } from './context';
import { defaultSettings, encodeSettings, parseSettings } from './settings-file';

const text = encodeSettings(fixture as never);

describe('the settings file', () => {
  it('reads the shared fixture with the values the Go test reads', () => {
    const parsed = parseSettings(text);
    expect('settings' in parsed && parsed.settings).toMatchObject({ locale: 'de', theme: 'dark', network: { mode: 'server' }, telemetry: { mode: 'export' } });
  });

  it('writes a file back unchanged, keys it does not use included', () => {
    const parsed = parseSettings(text);
    expect('settings' in parsed && encodeSettings(parsed.settings)).toBe(text);
    expect(text).toContain('"x_from_a_later_app"');
  });

  it.each([
    ['[]', 'not a JSON object'],
    ['{"riftweave": "save", "version": 1}', 'not a settings file'],
    ['{"riftweave": "settings", "version": 2}', 'version 2'],
    ['{"riftweave": "settings", "version": 1, "theme": "blue"}', 'theme blue'],
    ['{"riftweave": "settings", "version": 1, "modules": ["core"]}', 'lists core'],
    ['{"riftweave": "settings", "version": 1, "network": {"mode": "lan"}}', 'network.mode'],
    ['{"riftweave": "settings", "version": 1, "telemetry": {"mode": "on"}}', 'telemetry.mode'],
    ['{"riftweave": "settings", "version": 1, "telemetry": {"mode": "export", "endpoint": "collector:4318"}}', 'telemetry.endpoint'],
  ])('refuses %s', (body, says) => {
    const parsed = parseSettings(body);
    expect('error' in parsed && parsed.error).toContain(says);
  });
});

describe('telemetry', () => {
  it('is off in a fresh install', () => {
    expect(defaultSettings()['telemetry']).toEqual({ mode: 'off' });
  });
});

describe('AppContext settings', () => {
  beforeEach(() => {
    localStorage.clear();
    TestBed.resetTestingModule();
  });

  it('imports a file, applies its theme and language, and exports it back unchanged', () => {
    const context = TestBed.inject(AppContext);
    expect(context.importSettings(text)).toBeNull();
    expect(context.theme()).toBe('dark');
    expect(context.locale()).toBe('de');
    expect(context.exportSettings()).toBe(text);
  });

  it('refuses a bad file and changes nothing', () => {
    const context = TestBed.inject(AppContext);
    const before = context.exportSettings();
    expect(context.importSettings('{"riftweave": "settings", "version": 9}')).toContain('version 9');
    expect(context.exportSettings()).toBe(before);
  });
});
