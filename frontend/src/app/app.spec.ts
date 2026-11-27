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
import { ScreenHeader, escapeIsTaken, keyOf, parentPath } from './ui/screen-header';
import { ask, unsavedWorkGuard } from './unsaved-work-guard';
import { SelectField } from './ui/select-field';
import { Sheet } from './sheet';
import { Spells } from './spells';
import { SAVE_STORE, Saves, memoryStore } from './saves';
import { sheetPrint } from './files';
import { Injector } from '@angular/core';
import { Router } from '@angular/router';
import { installDevConsole } from './dev-console';
import { TextField } from './ui/text-field';
import { ThemeToggle } from './ui/theme-toggle';
import { Navbar, REPOSITORY_URL } from './ui/navbar';

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
  author: async (input: { operation?: string; spells?: { key: string }[] }) => {
    sent.push({ command: 'author', input });
    if (input.operation === 'load') return { row: FIRE_BOLT };
    if (input.operation === 'effects') return { effects: [{ ref: 'core:effect/damage_fire', row: { key: 'damage_fire', label: 'Fire Damage' } }] };
    return { spells: (input.spells ?? []).map((row) => ({ key: row.key, row, problems: [], warnings: [], cost: { mana: 6, cast_threshold: 6, craft_threshold: 8 } })) };
  },
};

/** fire_bolt as spells/core.json writes it, which the fake author loads. */
const FIRE_BOLT = { key: 'fire_bolt', label: 'Fire Bolt', description: 'Hurl a mote of fire at a target.', color: 'red', level: 1, delivery: 'single',
  template: true, effects: [{ effect: 'core:effect/damage_fire', magnitude: 8, target: 'ranged' }] };

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

  it("load a spell into the spell editor, list its color's effect pool, and check the draft (FE_S09_02)", async () => {
    const fixture = TestBed.createComponent(Spells);
    fixture.detectChanges();
    const editor = fixture.componentInstance;
    await editor.load('fire_bolt');
    await editor.check();
    fixture.detectChanges();
    expect(sent.map((call) => call.input)).toEqual([
      { kind: 'spell', operation: 'load', spell: 'fire_bolt' },
      { kind: 'spell', operation: 'effects', color: 'red' },
      { kind: 'spell', operation: 'check', spells: [FIRE_BOLT] },
    ]);
    expect(editor.pool()).toEqual(['core:effect/damage_fire']);
    const element = fixture.nativeElement as HTMLElement;
    expect(element.querySelector('[aria-label="Cost"]')!.textContent).toContain('6');
    expect(element.querySelector('[aria-label="Problems"]')!.textContent).toContain('No problems');
  });

  it("edit one effect of the draft and keep the row's other fields (FE_S09_02)", async () => {
    const fixture = TestBed.createComponent(Spells);
    fixture.detectChanges();
    const editor = fixture.componentInstance;
    await editor.load('fire_bolt');
    editor.setEffect(0, 'effect', 'core:effect/no_such_effect');
    editor.setEffect(0, 'magnitude', 10);
    expect(editor.draft().effects).toEqual([{ effect: 'core:effect/no_such_effect', magnitude: 10, target: 'ranged' }]);
    expect(editor.draft().template).toBe(true);
    expect(editor.exportInput()).toEqual({ kind: 'spell', operation: 'check', spells: [editor.draft()], export: { shape: 'core' } });
  });

  it('render the playthrough runner', () => {
    const fixture = TestBed.createComponent(Playthrough);
    fixture.detectChanges();
    expect((fixture.nativeElement as HTMLElement).querySelector('rw-card button.btn')).not.toBeNull();
  });

  it('print a character as its session holds it, or fresh, through the print command (FE_S09_04)', () => {
    const actor = { level: 3, resources: { hp: 41, mp: 20, sp: 30, ap: 4 } };
    expect(sheetPrint(actor as never, 'current')).toEqual({ kind: 'sheet', actor, state: 'current' });
    expect(sheetPrint(actor as never, 'fresh')).toEqual({ kind: 'sheet', actor, state: 'fresh' });
  });

  it('offer Print on the sheet only once a character is in play (FE_S09_04)', () => {
    const fixture = TestBed.createComponent(Sheet);
    fixture.detectChanges();
    expect(fixture.componentInstance.printInput('current')).toBeNull();
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

  it('show the tools page and GM tools, with Back in the screen header (FE_S08_12)', () => {
    expect(cards(menu('start'))).toEqual(['GM tools', 'Character creation', 'Spell creation', 'Skill resolution', 'Bestiary (FE_S08_06_bestiary-page)', 'Compendium (FE_S08_11_compendium-tables-page)', 'Saved games']);
    TestBed.resetTestingModule();
    expect(cards(menu('gm_tools'))).toEqual(['Resolve an attack', 'Encounter tracker (APP-34)', 'Playthrough']);
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
    expect(element.textContent).toContain('Set the server URL first.');
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

describe('the navbar (FE_S08_17)', () => {
  function navbar() {
    TestBed.configureTestingModule({ providers: [provideRouter(routes), { provide: SAVE_STORE, useValue: memoryStore() }] });
    const fixture = TestBed.createComponent(Navbar);
    fixture.detectChanges();
    return { fixture, element: fixture.nativeElement as HTMLElement };
  }

  it('hold the brand home, the rules version, the main menu, the language, the theme, the burger, and GitHub', () => {
    const { element } = navbar();
    expect(element.querySelector('h1 a')!.getAttribute('href')).toBe('/');
    expect(element.querySelector('.brand .tile svg')).not.toBeNull();
    expect(element.querySelector('.version')!.textContent).toContain('Rules');
    expect(Array.from(element.querySelectorAll('.links a')).map((link) => [link.textContent!.trim(), link.getAttribute('href')])).toEqual([
      ['Start', '/play'], ['Characters', '/characters'], ['Settings', '/settings'], ['Modules', '/modules'], ['Network', '/network'], ['Import', '/import'],
    ]);
    const controls = element.querySelector('.controls')!;
    expect(controls.querySelector('rw-language-select')).not.toBeNull();
    expect(controls.querySelector('rw-theme-toggle')).not.toBeNull();
    expect(controls.querySelector('button.burger')).not.toBeNull();
    const github = controls.querySelector<HTMLAnchorElement>('a.github')!;
    expect([github.href, github.target, github.rel, github.getAttribute('aria-label')]).toEqual([REPOSITORY_URL, '_blank', 'noopener', 'Riftweave on GitHub (opens in a new tab)']);
  });

  it('show the game in play by its save, and nothing without one', async () => {
    const { fixture, element } = navbar();
    expect(element.querySelector('.character')).toBeNull();
    const saves = TestBed.inject(Saves);
    saves.current.set(await saves.begin('Mara', { rules_version: 'test', ruleset: {}, start: {}, log: [], actor: {} } as never));
    fixture.detectChanges();
    expect(element.querySelector('.character')!.textContent).toContain('Mara');
  });

  it('open the main menu and GitHub from the burger, and close it on Escape, which goes no further, and on a click outside', () => {
    const { fixture, element } = navbar();
    const burger = element.querySelector<HTMLButtonElement>('button.burger')!;
    expect([burger.getAttribute('aria-expanded'), burger.getAttribute('aria-controls')]).toEqual(['false', 'rw-navbar-menu']);
    expect(element.querySelector('#rw-navbar-menu')).toBeNull();
    burger.click();
    fixture.detectChanges();
    expect(burger.getAttribute('aria-expanded')).toBe('true');
    const menu = element.querySelector('#rw-navbar-menu')!;
    expect(Array.from(menu.querySelectorAll('a')).map((link) => link.textContent!.trim())).toEqual(['Start', 'Characters', 'Settings', 'Modules', 'Network', 'Import', 'GitHub']);

    const reachedDocument = vi.fn();
    document.addEventListener('keydown', reachedDocument);
    burger.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    fixture.detectChanges();
    expect(burger.getAttribute('aria-expanded')).toBe('false');
    expect(reachedDocument).not.toHaveBeenCalled();
    // Closed, Escape is free to go on (to the screen header's Back).
    burger.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    expect(reachedDocument).toHaveBeenCalledTimes(1);
    document.removeEventListener('keydown', reachedDocument);

    burger.click();
    fixture.detectChanges();
    document.body.click();
    fixture.detectChanges();
    expect(element.querySelector('#rw-navbar-menu')).toBeNull();
  });

  it('close the menu when one of its links is followed', () => {
    const { fixture, element } = navbar();
    element.querySelector<HTMLButtonElement>('button.burger')!.click();
    fixture.detectChanges();
    element.querySelector<HTMLAnchorElement>('#rw-navbar-menu a')!.click();
    fixture.detectChanges();
    expect(element.querySelector('#rw-navbar-menu')).toBeNull();
  });
});

describe('the screen header (FE_S08_12)', () => {
  it("send every route but the main menu back to its parent menu in menu.json, also by its URL alone", () => {
    const backs = Object.fromEntries(routes.filter((route) => route.path !== '**').map((route) => [`/${route.path}`, parentPath(`/${route.path}`)]));
    expect(backs['/']).toBeNull();
    expect(backs['/play']).toBe('/');
    expect(backs['/characters']).toBe('/');
    expect(backs['/settings']).toBe('/');
    expect(backs['/play/gm_tools']).toBe('/play');
    expect(backs['/play/character_creation']).toBe('/play');
    expect(backs['/play/attack']).toBe('/play/gm_tools');
    expect(backs['/play/playthrough']).toBe('/play/gm_tools');
    // A path with no menu entry goes to the menu its parent path names.
    expect(parentPath('/play/no_such_tool')).toBe('/play');
    expect(keyOf('/play/attack?x=1')).toBe('attack');
    for (const [path, back] of Object.entries(backs)) {
      if (path !== '/') expect(back, path).not.toBeNull();
    }
  });

  it('show Back and the title on a screen, and nothing on the main menu', async () => {
    TestBed.configureTestingModule({ providers: [provideRouter(routes), { provide: SAVE_STORE, useValue: memoryStore() }] });
    const router = TestBed.inject(Router);
    const fixture = TestBed.createComponent(ScreenHeader);
    await router.navigateByUrl('/play/attack');
    fixture.detectChanges();
    const element = fixture.nativeElement as HTMLElement;
    expect(element.querySelector('h2')!.textContent).toBe('Resolve an attack');
    expect(element.querySelector('a.back')!.getAttribute('href')).toBe('/play/gm_tools');
    await router.navigateByUrl('/');
    fixture.detectChanges();
    expect(element.querySelector('.screen-header')).toBeNull();
  });

  it('go back on Esc, unless a field or an open dialog has it', async () => {
    TestBed.configureTestingModule({ providers: [provideRouter(routes), { provide: SAVE_STORE, useValue: memoryStore() }] });
    const router = TestBed.inject(Router);
    const fixture = TestBed.createComponent(ScreenHeader);
    await router.navigateByUrl('/play/gm_tools');
    fixture.detectChanges();
    const field = document.createElement('input');
    document.body.append(field);
    expect(escapeIsTaken(field, document)).toBe(true);
    field.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await fixture.whenStable();
    expect(router.url).toBe('/play/gm_tools');
    document.body.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await fixture.whenStable();
    expect(router.url).toBe('/play');
    field.remove();
  });

  it('ask before leaving a screen with unsaved work, and not without', () => {
    TestBed.configureTestingModule({ providers: [{ provide: API, useValue: fakeApi }] });
    const confirm = vi.spyOn(ask, 'confirm').mockReturnValue(false);
    const guard = (unsaved: boolean) => TestBed.runInInjectionContext(() => unsavedWorkGuard({ hasUnsavedWork: () => unsaved }, null as never, null as never, null as never));
    expect(guard(false)).toBe(true);
    expect(confirm).not.toHaveBeenCalled();
    expect(guard(true)).toBe(false);
    expect(confirm).toHaveBeenCalledWith(expect.stringContaining('Leave anyway?'));
    confirm.mockRestore();
  });

  it('count edits to the character form as unsaved until a character is created from it', async () => {
    TestBed.configureTestingModule({ providers: [{ provide: API, useValue: { ...fakeApi, newCharacter: async () => ({}), startSession: async () => ({ log: [] }) } }, { provide: SAVE_STORE, useValue: memoryStore() }] });
    const sheet = TestBed.createComponent(Sheet).componentInstance;
    expect(sheet.hasUnsavedWork()).toBe(false);
    sheet.set('str', 8);
    expect(sheet.hasUnsavedWork()).toBe(true);
    await sheet.create();
    expect(sheet.hasUnsavedWork()).toBe(false);
  });
});
