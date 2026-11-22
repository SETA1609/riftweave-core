// The app's own text in English, German, and Spanish (FE_S11_01, docs/backend/frontend.md#languages):
// every key of en.json is in every locale, every menu entry has a key, no template keeps a literal
// string, and switching the language re-renders a screen without a reload.
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import tree from '../../../backend/contract/menu.json';
import de from '../locales/de.json';
import en from '../locales/en.json';
import es from '../locales/es.json';
import { routes } from './app.routes';
import { AppContext, LANGUAGES, startLanguage } from './context';
import { I18n } from './i18n';
import { SettingsScreens } from './settings-screens';
import { LanguageSelect } from './ui/language-select';
import { MenuGrid } from './ui/menu-grid';

type Tree = { readonly [key: string]: string | Tree };

/** Every key of a locale file, as a dotted path. */
function keys(tree: Tree, prefix = ''): string[] {
  return Object.entries(tree).flatMap(([key, value]) => (typeof value === 'string' ? [prefix + key] : keys(value, `${prefix}${key}.`)));
}

/** The screens FE_S11_04_frontend-screen-strings moves to keys; until then they keep their English. */
const PENDING = ['attack.html', 'character-list.html', 'file-import.html', 'playthrough.html', 'resolution.html', 'saved-games.html', 'sheet.html'];
/** Words a template may show as they are: the name of the game. */
const ALLOWED = ['Riftweave'];
/** The attributes a person reads, which must be bound to a key, not written out. */
const READ_ATTRIBUTES = /\s(label|title|message|placeholder|aria-label|alt)="([^"]*)"/g;

/** The few calls of node:fs this spec reads the templates with. The app has no Node types, so the
 * module is loaded at run time, where Vitest runs on Node. */
type FileSystem = {
  readFileSync(path: string, encoding: 'utf8'): string;
  readdirSync(path: string, options: { withFileTypes: true }): { name: string; isDirectory(): boolean }[];
};
const NODE_FS = 'node:fs';
const fs = (await import(/* @vite-ignore */ NODE_FS)) as FileSystem;
const APP = `${(globalThis as unknown as { process: { cwd(): string } }).process.cwd()}/src/app`;

function templates(directory = APP): string[] {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = `${directory}/${entry.name}`;
    if (entry.isDirectory()) return templates(path);
    return entry.name.endsWith('.html') && !PENDING.includes(entry.name) ? [path] : [];
  });
}

/** The literal words a template shows: its text outside tags, interpolations, and control flow,
 * and the read attributes written out. */
function literals(template: string): string[] {
  const attributes = [...template.matchAll(READ_ATTRIBUTES)].map((match) => match[2]);
  const text = template
    .replace(/<!--[\s\S]*?-->/g, ' ')
    .replace(/\{\{[\s\S]*?\}\}/g, ' ')
    .replace(/<[^>]*>/g, ' ')
    .replace(/@(else if|if|else|for|switch|case|default|empty)\b[^{]*\{/g, ' ')
    .replace(/[{}]/g, ' ');
  return [...attributes, ...text.split(/\s+/)].map((word) => word.trim()).filter((word) => /\p{L}/u.test(word) && !ALLOWED.includes(word));
}

describe('the locale files', () => {
  const source = keys(en);

  it.each([['de', de], ['es', es]])('give %s every key of en.json, and no other', (_, locale) => {
    expect(keys(locale).sort()).toEqual([...source].sort());
  });

  it('name every menu entry of backend/contract/menu.json', () => {
    const missing = tree.entries.map((entry) => `menu.${entry.key}`).filter((key) => !source.includes(key));
    expect(missing).toEqual([]);
  });

  it('offer English, German, and Spanish, each by its own name and flag', () => {
    expect(LANGUAGES).toEqual([
      { value: 'en', label: 'English', flag: 'us' },
      { value: 'de', label: 'Deutsch', flag: 'de' },
      { value: 'es', label: 'Español', flag: 'es' },
    ]);
  });
});

describe('the templates', () => {
  it.each(templates().map((path) => [path.slice(APP.length + 1), path]))('%s shows no literal string', (_, path) => {
    expect(literals(fs.readFileSync(path, 'utf8'))).toEqual([]);
  });

  it('are caught when they write a string out', () => {
    expect(literals('<rw-card title="Settings"><p>Hello {{ name }}</p></rw-card>')).toEqual(['Settings', 'Hello']);
    expect(literals('@if (ok) { <rw-alert [message]="text()" /> } <h1>Riftweave</h1>')).toEqual([]);
  });
});

describe('the language', () => {
  beforeEach(() => localStorage.clear());
  afterEach(() => vi.restoreAllMocks());

  it('starts from the saved one, else the browser one if offered, else English', () => {
    expect(startLanguage('es', 'de-DE')).toBe('es');
    expect(startLanguage(null, 'de-DE')).toBe('de');
    expect(startLanguage(null, 'fr-FR')).toBe('en');
    expect(startLanguage('xx', 'en-GB')).toBe('en');
  });

  it('translates a key, fills its values, and falls back to English', () => {
    const i18n = TestBed.inject(I18n);
    const context = TestBed.inject(AppContext);
    context.locale.set('de');
    expect(i18n.translate('menu.characters')).toBe('Charaktere');
    expect(i18n.translate('app.rules', { version: '1.2.3' })).toBe('Regeln 1.2.3');
    context.locale.set('fr');
    expect(i18n.translate('menu.characters')).toBe('Characters');
  });

  it('re-renders a screen and sets <html lang> when it changes, and keeps the choice', async () => {
    TestBed.configureTestingModule({ providers: [provideRouter(routes)] });
    const fixture = TestBed.createComponent(MenuGrid);
    fixture.componentRef.setInput('parent', 'start');
    fixture.detectChanges();
    const element = fixture.nativeElement as HTMLElement;
    expect(element.textContent).toContain('Character creation');
    TestBed.inject(AppContext).locale.set('es');
    await fixture.whenStable();
    expect(element.textContent).toContain('Creación de personaje');
    expect(element.textContent).toContain('Volver');
    expect(document.documentElement.lang).toBe('es');
    expect(localStorage.getItem('riftweave.locale')).toBe('es');
  });

  it('is picked in the language select, which lists each language by its flag and own name', async () => {
    const fixture = TestBed.createComponent(LanguageSelect);
    fixture.detectChanges();
    const element = fixture.nativeElement as HTMLElement;
    const toggle = element.querySelector<HTMLButtonElement>('button[aria-haspopup]')!;
    expect(toggle.getAttribute('aria-label')).toBe('Language');
    expect(toggle.querySelector('img')!.getAttribute('src')).toBe('flags/us.svg');
    toggle.click();
    await fixture.whenStable();
    const items = Array.from(element.querySelectorAll<HTMLButtonElement>('[role="menuitemradio"]'));
    expect(items.map((item) => item.textContent!.trim())).toEqual(['English', 'Deutsch', 'Español']);
    expect(items.map((item) => item.querySelector('img')!.getAttribute('src'))).toEqual(['flags/us.svg', 'flags/de.svg', 'flags/es.svg']);
    items[1].click();
    await fixture.whenStable();
    expect(TestBed.inject(AppContext).locale()).toBe('de');
    expect(element.querySelector('[role="menu"]')).toBeNull();
    expect(toggle.getAttribute('aria-label')).toBe('Sprache');
    expect(toggle.querySelector('img')!.getAttribute('src')).toBe('flags/de.svg');
  });

  it('changes the settings screens with it', async () => {
    const fixture = TestBed.createComponent(SettingsScreens);
    fixture.componentRef.setInput('section', 'network');
    fixture.detectChanges();
    TestBed.inject(AppContext).locale.set('de');
    await fixture.whenStable();
    const text = (fixture.nativeElement as HTMLElement).textContent!;
    expect(text).toContain('Netzwerk');
    expect(text).toContain('Diesen Server verwenden');
  });
});
