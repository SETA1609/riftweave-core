// The viewer's settings in one place (docs/backend/frontend.md#the-app-context): the mode and the
// interface language (FE_S11_01). A screen reads them from this service; no component reads localStorage or
// matchMedia itself. Each choice is saved in localStorage, a per-viewer convenience, and every read
// and write is guarded, so a blocked storage still renders in the system's mode.
import { DOCUMENT, Service, computed, effect, inject, signal } from '@angular/core';
import { type SettingsFile, defaultSettings, encodeSettings, parseSettings } from './settings-file';

export type Theme = 'light' | 'dark' | 'system';
export const THEMES: readonly Theme[] = ['light', 'dark', 'system'];

// LANGUAGES, FALLBACK_LANGUAGE, offeredLanguage, and startLanguage: see i18n.md.
export const LANGUAGES: readonly { value: string; label: string }[] = [
  { value: 'en', label: 'English' },
  { value: 'de', label: 'Deutsch' },
  { value: 'es', label: 'Español' },
];
export const FALLBACK_LANGUAGE = 'en';

export function offeredLanguage(code: string | null | undefined): string | null {
  const base = code?.toLowerCase().split('-')[0];
  return LANGUAGES.find((language) => language.value === base)?.value ?? null;
}

export function startLanguage(saved: string | null, browser: string | undefined): string {
  return offeredLanguage(saved) ?? offeredLanguage(browser) ?? FALLBACK_LANGUAGE;
}

const THEME_KEY = 'riftweave.theme';
const LOCALE_KEY = 'riftweave.locale';
// The whole settings file as last imported (APP-25), so the keys this app does not use survive.
const SETTINGS_KEY = 'riftweave.settings';

/** A saved value, or null when there is none or storage is blocked. */
function load(key: string): string | null {
  try {
    return globalThis.localStorage?.getItem(key) ?? null;
  } catch {
    return null;
  }
}

function save(key: string, value: string): void {
  try {
    globalThis.localStorage?.setItem(key, value);
  } catch {
    // Storage is blocked (a private window, a sandbox): the choice lasts until the page closes.
  }
}

/** The system's dark-mode query, or null where there is none (a test runner). */
function darkQuery(): MediaQueryList | null {
  try {
    return globalThis.matchMedia?.('(prefers-color-scheme: dark)') ?? null;
  } catch {
    return null;
  }
}

@Service()
export class AppContext {
  private readonly document = inject(DOCUMENT);
  private readonly query = darkQuery();

  /** The chosen mode: light, dark, or the system's. */
  readonly theme = signal<Theme>(THEMES.find((theme) => theme === load(THEME_KEY)) ?? 'system');
  /** The interface language, an i18next language code; I18n follows it (FE_S11_01). */
  readonly locale = signal(startLanguage(load(LOCALE_KEY), globalThis.navigator?.language));
  /** Whether the system asks for dark, kept current as it changes. */
  readonly systemDark = signal(this.query?.matches ?? false);
  /** The mode the page shows: the chosen one, or the system's for `system`. */
  readonly shown = computed<'light' | 'dark'>(() => {
    const theme = this.theme();
    return theme === 'system' ? (this.systemDark() ? 'dark' : 'light') : theme;
  });
  /** The rest of the settings file (APP-25) as last saved: bestiary, modules, network, and any key
   * this app does not use. Theme and locale live in their own signals above. */
  readonly file = signal<SettingsFile>(savedFile());
  /** Whether players see only the monsters they know (APP-29). */
  readonly playersSeeOnlyKnown = computed(() => (this.file()['bestiary'] as { players_see_only_known?: boolean } | undefined)?.players_see_only_known === true);
  /** The enabled optional modules, by manifest id; core is always loaded and never listed. */
  readonly modules = computed(() => (this.file()['modules'] as string[] | undefined) ?? []);
  /** local (the core in the page, offline), server (the HTTP API at serverUrl), or room (APP-17). */
  readonly networkMode = computed(() => ((this.file()['network'] as { mode?: string } | undefined)?.mode ?? 'local') as 'local' | 'server' | 'room');
  readonly serverUrl = computed(() => (this.file()['network'] as { server_url?: string } | undefined)?.server_url ?? '');
  /** The mode the toggle moves to next: light, dark, system, then light again. */
  readonly next = computed<Theme>(() => THEMES[(THEMES.indexOf(this.theme()) + 1) % THEMES.length]);

  constructor() {
    this.query?.addEventListener?.('change', (change) => this.systemDark.set(change.matches));
    // `data-theme` on <html> picks the tokens in styles.css; `system` leaves it off, so the
    // prefers-color-scheme block decides.
    effect(() => {
      const theme = this.theme();
      const root = this.document.documentElement;
      if (theme === 'system') {
        root.removeAttribute('data-theme');
      } else {
        root.setAttribute('data-theme', theme);
      }
      save(THEME_KEY, theme);
    });
    effect(() => save(LOCALE_KEY, this.locale()));
  }

  /** Change the settings file: the edit is checked by the file's rules and saved, or refused
   * with the rule's message, and then nothing changes. Returns null on success. */
  update(edit: (file: SettingsFile) => SettingsFile): string | null {
    const parsed = parseSettings(encodeSettings(edit(structuredClone(this.file()))));
    if ('error' in parsed) return parsed.error;
    this.file.set(parsed.settings);
    save(SETTINGS_KEY, encodeSettings(parsed.settings));
    return null;
  }

  setPlayersSeeOnlyKnown(known: boolean): string | null {
    return this.update((file) => ({ ...file, bestiary: { ...(file['bestiary'] as object), players_see_only_known: known } }));
  }

  /** Check or uncheck a module. */
  toggleModule(id: string): string | null {
    const enabled = this.modules();
    return this.update((file) => ({ ...file, modules: enabled.includes(id) ? enabled.filter((module) => module !== id) : [...enabled, id] }));
  }

  setNetwork(mode: 'local' | 'server', serverUrl = this.serverUrl()): string | null {
    if (mode === 'server' && !serverUrl) return 'set the server URL first';
    return this.update((file) => ({ ...file, network: { ...(file['network'] as object), mode, server_url: serverUrl } }));
  }

  /** Move to the next mode. */
  cycleTheme(): void {
    this.theme.set(this.next());
  }

  /** The settings file to export (`riftweave-settings.json`, APP-25): the last imported file, or
   * the defaults, with this app's current theme and language. */
  exportSettings(): string {
    return encodeSettings({ ...this.file(), theme: this.theme(), locale: this.locale() });
  }

  /** Apply a settings file. Returns null on success, or the message of the rule it breaks, in
   * which case nothing changes. */
  importSettings(text: string): string | null {
    const parsed = parseSettings(text);
    if ('error' in parsed) return parsed.error;
    const theme = THEMES.find((known) => known === parsed.settings['theme']);
    if (theme) this.theme.set(theme);
    if (typeof parsed.settings['locale'] === 'string') this.locale.set(parsed.settings['locale']);
    save(SETTINGS_KEY, encodeSettings(parsed.settings));
    this.file.set(parsed.settings);
    return null;
  }
}

/** The settings file as last saved, or the defaults. */
function savedFile(): SettingsFile {
  const saved = parseSettings(load(SETTINGS_KEY) ?? '');
  return 'settings' in saved ? saved.settings : defaultSettings();
}
