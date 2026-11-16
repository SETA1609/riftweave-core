// The viewer's settings in one place (docs/backend/frontend.md#the-app-context): the mode here and the
// language for APP-24. A screen reads them from this service; no component reads localStorage or
// matchMedia itself. Each choice is saved in localStorage, a per-viewer convenience, and every read
// and write is guarded, so a blocked storage still renders in the system's mode.
import { DOCUMENT, Service, computed, effect, inject, signal } from '@angular/core';

export type Theme = 'light' | 'dark' | 'system';
export const THEMES: readonly Theme[] = ['light', 'dark', 'system'];

const THEME_KEY = 'riftweave.theme';
const LOCALE_KEY = 'riftweave.locale';

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
  /** The interface language, an i18next language code (APP-24). */
  readonly locale = signal(load(LOCALE_KEY) ?? 'en');
  /** Whether the system asks for dark, kept current as it changes. */
  readonly systemDark = signal(this.query?.matches ?? false);
  /** The mode the page shows: the chosen one, or the system's for `system`. */
  readonly shown = computed<'light' | 'dark'>(() => {
    const theme = this.theme();
    return theme === 'system' ? (this.systemDark() ? 'dark' : 'light') : theme;
  });
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

  /** Move to the next mode. */
  cycleTheme(): void {
    this.theme.set(this.next());
  }
}
