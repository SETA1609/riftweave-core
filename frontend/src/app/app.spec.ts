// The reusable components and the screens built from them (docs/backend/frontend.md#conventions):
// each component binds its value as a signal, and each screen sends the API the same input as before.
import { Component, signal, type Type } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { API } from './api';
import { Alert } from './ui/alert';
import { AppContext } from './context';
import { Attack } from './attack';
import { Card } from './ui/card';
import { KeyValueList } from './ui/key-value-list';
import { NumberField } from './ui/number-field';
import { Playthrough } from './playthrough';
import { provideRouter } from '@angular/router';
import { switchableTransport, type Transport } from './api';
import { routes } from './app.routes';
import { Resolution } from './resolution';
import { MODULE_IDS, SettingsScreens } from './settings-screens';
import { MENU, MenuGrid, pathOf } from './ui/menu-grid';
import { SelectField } from './ui/select-field';
import { Sheet } from './sheet';
import { SAVE_STORE, Saves, memoryStore } from './saves';
import { Injector } from '@angular/core';
import { Router } from '@angular/router';
import { installDevConsole } from './dev-console';
import { TextField } from './ui/text-field';
import { ThemeToggle } from './ui/theme-toggle';

function render<T>(host: new () => T) {
  const fixture = TestBed.createComponent(host);
  fixture.detectChanges();
  return { fixture, element: fixture.nativeElement as HTMLElement, host: fixture.componentInstance };
}

function type(input: HTMLInputElement | HTMLSelectElement, value: string, event = 'input'): void {
  input.value = value;
  input.dispatchEvent(new Event(event));
}

@Component({ imports: [NumberField, TextField, SelectField], template: `
  <rw-number-field label="STR" [min]="1" [(value)]="strength" />
  <rw-text-field label="Type" [(value)]="kind" />
  <rw-select-field label="Reaction" [options]="reactions" [(value)]="reaction" />` })
class Fields {
  readonly strength = signal(6);
  readonly kind = signal('slashing');
  readonly reaction = signal('');
  readonly reactions = [{ value: '', label: 'none' }, { value: 'evade', label: 'evade' }];
}

@Component({ imports: [Card, Alert, KeyValueList], template: `
  <rw-card title="Stats"><p class="inside">projected</p></rw-card>
  <rw-alert [message]="message()" />
  <rw-key-value-list [entries]="entries" />` })
class Blocks {
  readonly message = signal('');
  readonly entries = { hp: 59, mp: 46 };
}

describe('the field components', () => {
  it('label their input and write what is typed back to the bound signal', async () => {
    const { fixture, element, host } = render(Fields);
    const [number, text] = Array.from(element.querySelectorAll('input'));
    expect(element.textContent).toContain('STR');
    expect(number.value).toBe('6');
    type(number, '8');
    type(text, 'fire');
    type(element.querySelector('select')!, 'evade', 'change');
    await fixture.whenStable();
    expect(host.strength()).toBe(8);
    expect(host.kind()).toBe('fire');
    expect(host.reaction()).toBe('evade');
  });

  it('use the shared form-control class', () => {
    const { element } = render(Fields);
    for (const control of Array.from(element.querySelectorAll('input, select'))) {
      expect(control.classList).toContain('form-control');
    }
  });
});

describe('the block components', () => {
  it('project a card body under its title', () => {
    const { element } = render(Blocks);
    const card = element.querySelector('.card')!;
    expect(card.textContent).toContain('Stats');
    expect(card.querySelector('.inside')?.textContent).toBe('projected');
  });

  it('show an alert only with a message', async () => {
    const { fixture, element, host } = render(Blocks);
    expect(element.querySelector('.alert')).toBeNull();
    host.message.set('the server is down');
    fixture.detectChanges();
    await fixture.whenStable();
    expect(element.querySelector('.alert')?.textContent).toContain('the server is down');
  });

  it('list each entry as a row', () => {
    const { element } = render(Blocks);
    const rows = Array.from(element.querySelectorAll('tr')).map((row) => Array.from(row.children).map((cell) => cell.textContent?.trim()).join(' '));
    expect(rows).toEqual(['hp 59', 'mp 46']);
  });
});

const sent: { command: string; input: unknown }[] = [];
const fakeApi = {
  derived: async (actor: unknown) => { sent.push({ command: 'derived', input: actor }); return { hp: 67, mp: 42, sp: 10, ap: 4 }; },
  combat: async (input: unknown) => { sent.push({ command: 'combat', input }); return { outcome: 'hit', target: 72, damage: 6, hp: 24 }; },
  resolution: async (input: unknown) => { sent.push({ command: 'resolution', input }); return { target: 50, success: true, margin: 17, tags: [] }; },
};

beforeEach(() => {
  sent.length = 0;
  TestBed.configureTestingModule({ providers: [{ provide: API, useValue: fakeApi }] });
});

describe('the screens', () => {
  it('build their forms from the shared field components', () => {
    const screens: [Type<unknown>, number][] = [[Sheet, 9], [Attack, 12]];
    for (const [screen, fields] of screens) {
      const fixture = TestBed.createComponent(screen);
      fixture.detectChanges();
      const element = fixture.nativeElement as HTMLElement;
      expect(element.querySelectorAll('rw-number-field, rw-text-field, rw-select-field').length).toBe(fields);
      expect(element.querySelector('rw-card')).not.toBeNull();
    }
  });

  it('send the same combat input the attack form sent before', async () => {
    const fixture = TestBed.createComponent(Attack);
    fixture.detectChanges();
    await fixture.componentInstance.resolve();
    expect(sent).toEqual([{ command: 'combat', input: {
      attacker: { skill: 72, roll: 47, attributes: { lck: 0 } },
      defender: { evasion_skill: 25, armor_dr: 2, hp: 30 },
      damage: { roll: 8, type: 'slashing' },
      delivery: 'single',
    } }]);
  });

  it('send the sheet attributes to derived', async () => {
    const fixture = TestBed.createComponent(Sheet);
    fixture.detectChanges();
    await fixture.componentInstance.compute();
    expect(sent).toEqual([{ command: 'derived', input: { attributes: { str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 }, level: 1 } }]);
  });

  it('render the playthrough runner', () => {
    const fixture = TestBed.createComponent(Playthrough);
    fixture.detectChanges();
    expect((fixture.nativeElement as HTMLElement).querySelector('rw-card button.btn')).not.toBeNull();
  });
});
describe('the app context', () => {
  const root = (): HTMLElement => document.documentElement;
  /** A fresh context, as after a reload: a new injector reads what was saved. */
  function fresh(): AppContext {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({ providers: [{ provide: API, useValue: fakeApi }] });
    const context = TestBed.inject(AppContext);
    TestBed.tick();
    return context;
  }

  beforeEach(() => {
    localStorage.clear();
    root().removeAttribute('data-theme');
  });
  afterEach(() => vi.restoreAllMocks());

  it("starts in the system's mode with nothing saved, and sets no data-theme", () => {
    const context = fresh();
    expect(context.theme()).toBe('system');
    expect(root().hasAttribute('data-theme')).toBe(false);
  });

  it('cycles light, dark, and system, and puts the mode on <html>', () => {
    const context = fresh();
    context.cycleTheme();
    TestBed.tick();
    expect([context.theme(), root().getAttribute('data-theme')]).toEqual(['light', 'light']);
    context.cycleTheme();
    TestBed.tick();
    expect([context.theme(), root().getAttribute('data-theme')]).toEqual(['dark', 'dark']);
    context.cycleTheme();
    TestBed.tick();
    expect([context.theme(), root().hasAttribute('data-theme')]).toEqual(['system', false]);
  });

  it('keeps the choice across a reload', () => {
    fresh().theme.set('dark');
    TestBed.tick();
    expect(fresh().theme()).toBe('dark');
    expect(root().getAttribute('data-theme')).toBe('dark');
  });

  it("renders in the system's mode when storage is blocked", () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('blocked'); });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('blocked'); });
    const context = fresh();
    expect(context.theme()).toBe('system');
    context.cycleTheme();
    expect(() => TestBed.tick()).not.toThrow();
    expect(context.theme()).toBe('light');
  });

  it('shows the system mode it follows', () => {
    const context = fresh();
    context.systemDark.set(true);
    expect(context.shown()).toBe('dark');
    context.theme.set('light');
    expect(context.shown()).toBe('light');
  });
});

describe('the theme toggle', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('names the mode it switches to and cycles on a click', async () => {
    const { fixture, element } = render(ThemeToggle);
    const button = element.querySelector('button')!;
    expect(button.getAttribute('aria-label')).toBe('Switch to light mode');
    button.click();
    fixture.detectChanges();
    await fixture.whenStable();
    expect(button.getAttribute('aria-label')).toBe('Switch to dark mode');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });
});

describe('the menus of backend/contract/menu.json', () => {
  function menu(parent: string): HTMLElement {
    TestBed.configureTestingModule({ providers: [provideRouter(routes)] });
    const fixture = TestBed.createComponent(MenuGrid);
    fixture.componentRef.setInput('parent', parent);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }
  const cards = (element: HTMLElement) => Array.from(element.querySelectorAll('.menu-card')).map((card) => card.textContent!.trim());

  it('show the home page: the main menu without the CLI-only Quit', () => {
    const element = menu('main');
    expect(cards(element)).toEqual(['Start', 'Characters', 'Settings', 'Modules', 'Network', 'Import']);
    expect(element.querySelector('[aria-disabled="true"]')).toBeNull();
  });

  it('show the tools page and GM tools, each with Back to the menu above', () => {
    expect(cards(menu('start'))).toEqual(['GM tools', 'Character creation', 'Spell creation (APP-22)', 'Skill resolution', 'Bestiary (APP-29)', 'Saved games', 'Back']);
    TestBed.resetTestingModule();
    const gm = menu('gm_tools');
    expect(cards(gm)).toEqual(['Resolve an attack', 'Encounter tracker (APP-34)', 'Playthrough', 'Back']);
    expect(gm.querySelector('a.menu-card:last-child')!.getAttribute('href')).toBe('/play');
  });

  it('have a route for every entry that is not pending', () => {
    const paths = routes.map((route) => `/${route.path}`.replace(/\/$/, '') || '/');
    for (const entry of MENU.filter((entry) => !entry.issue)) {
      expect(paths, entry.key).toContain(pathOf(entry.key) === '/' ? '/' : pathOf(entry.key));
    }
  });
});

describe('the network setting', () => {
  function recorder(name: string, calls: string[]): Transport {
    return { ready: Promise.resolve(), post: async (path) => { calls.push(`${name} ${path}`); return {}; } };
  }

  it('switches the transport between the core in the page and the server', async () => {
    const calls: string[] = [];
    let choice = { mode: 'local', url: '' };
    const transport = switchableTransport(recorder('local', calls), (url) => recorder(`server ${url}`, calls), () => choice);
    await transport.post('/v1/commands/derived', {});
    choice = { mode: 'server', url: 'http://127.0.0.1:9000' };
    await transport.post('/v1/commands/resolution', {});
    choice = { mode: 'local', url: 'http://127.0.0.1:9000' };
    await transport.post('/v1/commands/combat', {});
    expect(calls).toEqual(['local /v1/commands/derived', 'server http://127.0.0.1:9000 /v1/commands/resolution', 'local /v1/commands/combat']);
  });
});

describe('the settings screens', () => {
  beforeEach(() => localStorage.clear());

  function screen(section: string) {
    const fixture = TestBed.createComponent(SettingsScreens);
    fixture.componentRef.setInput('section', section);
    fixture.detectChanges();
    return { fixture, element: fixture.nativeElement as HTMLElement, context: TestBed.inject(AppContext) };
  }

  it('list core checked and locked, then every module, and save a checked one', async () => {
    const { fixture, element, context } = screen('modules');
    const boxes = Array.from(element.querySelectorAll<HTMLInputElement>('input[type="checkbox"]'));
    expect(boxes.map((box) => [box.parentElement!.textContent!.trim(), box.checked, box.disabled])).toEqual([
      ['core (always loaded)', true, true], ...MODULE_IDS.map((id) => [id, false, false]),
    ]);
    boxes[MODULE_IDS.indexOf('factions') + 1].click();
    await fixture.whenStable();
    expect(context.modules()).toEqual(['factions']);
    expect(JSON.parse(context.exportSettings()).modules).toEqual(['factions']);
  });

  it('refuse server mode without a URL, and switch with one', async () => {
    const { fixture, element, context } = screen('network');
    element.querySelectorAll<HTMLInputElement>('input[type="radio"]')[1].click();
    await fixture.whenStable();
    expect(element.textContent).toContain('set the server URL first');
    expect(context.networkMode()).toBe('local');
    fixture.componentInstance.url.set('http://127.0.0.1:8080');
    fixture.componentInstance.useServer();
    expect([context.networkMode(), context.serverUrl()]).toEqual(['server', 'http://127.0.0.1:8080']);
  });

  it('save the bestiary choice in the settings file', () => {
    const { element, context } = screen('settings');
    element.querySelector<HTMLInputElement>('input[type="checkbox"]')!.click();
    expect(context.playersSeeOnlyKnown()).toBe(true);
    expect(JSON.parse(context.exportSettings()).bestiary).toEqual({ players_see_only_known: true });
  });
});

describe('skill resolution', () => {
  it('sends the skill, its value, LCK, the modifier, and the roll to resolution', async () => {
    const fixture = TestBed.createComponent(Resolution);
    fixture.detectChanges();
    fixture.componentInstance.bonus.set(10);
    fixture.componentInstance.roll.set(33);
    await fixture.componentInstance.resolve();
    expect(sent).toEqual([{ command: 'resolution', input: {
      actor: { skills: { lockpick: 40 }, attributes: { lck: 5 } }, skill: 'lockpick',
      modifiers: [{ target: 'lockpick', operation: 'add', value: 10, source: 'situation:gm' }], roll: 33,
    } }]);
    fixture.detectChanges();
    expect((fixture.nativeElement as HTMLElement).textContent).toContain('margin');
  });
});

describe('the dev console (APP-48)', () => {
  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({ providers: [provideRouter(routes), { provide: SAVE_STORE, useValue: memoryStore() }] });
  });

  it('puts the context, the API, the saves, and the router on rw', () => {
    const host: { rw?: ReturnType<typeof installDevConsole> } = {};
    const rw = installDevConsole(TestBed.inject(Injector), host);
    expect(host.rw).toBe(rw);
    expect(rw.context).toBe(TestBed.inject(AppContext));
    expect(rw.api).toBe(TestBed.inject(API));
    expect(rw.saves).toBe(TestBed.inject(Saves));
    expect(rw.router).toBe(TestBed.inject(Router));
  });

  it('changes the app through the context, as a screen does', () => {
    const rw = installDevConsole(TestBed.inject(Injector), {});
    rw.context.theme.set('dark');
    expect(TestBed.inject(AppContext).shown()).toBe('dark');
  });

  it('documents itself in rw.help()', () => {
    vi.spyOn(console, 'info').mockImplementation(() => undefined);
    const text = installDevConsole(TestBed.inject(Injector), {}).help();
    for (const name of ['rw.context', 'rw.api', 'rw.saves', 'rw.router', 'rw.screen()', 'rw.help()']) {
      expect(text).toContain(name);
    }
  });

  it('is not on globalThis in a test build, where RIFTWEAVE_DEV is false', () => {
    expect('rw' in globalThis).toBe(false);
  });
});
