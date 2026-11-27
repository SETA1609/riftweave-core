// See screen-header.md.
import { DOCUMENT, Component, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterLink } from '@angular/router';
import { filter, map } from 'rxjs';
import { TranslatePipe } from '../translate-pipe';
import { MENU, pathOf } from './menu-grid';

export const MAIN_PATH = '/';
export const START_SEGMENT = 'play';
export const START_KEY = 'start';
export const MAIN_KEY = 'main';

/** The menu key a path stands for: `start` for `/play`, the last segment otherwise, `main` for `/`. */
export function keyOf(url: string): string {
  const segments = url.split(/[?#]/)[0].split('/').filter(Boolean);
  if (segments.length === 0) return MAIN_KEY;
  const last = segments[segments.length - 1];
  return segments.length === 1 && last === START_SEGMENT ? START_KEY : last;
}

/** Where Back goes from a path: the menu its entry sits in (backend/contract/menu.json), none from the main menu. A path with no entry goes to the menu its parent path names. */
export function parentPath(url: string): string | null {
  const key = keyOf(url);
  if (key === MAIN_KEY) return null;
  const entry = MENU.find((one) => one.key === key);
  if (entry) return pathOf(entry.parent);
  const segments = url.split(/[?#]/)[0].split('/').filter(Boolean);
  return segments.length > 1 ? `/${segments.slice(0, -1).join('/')}` : MAIN_PATH;
}

/** Whether Esc belongs to what has focus: a field, or an open dialog. */
export function escapeIsTaken(target: EventTarget | null, document: Document): boolean {
  const element = target as HTMLElement | null;
  const field = !!element?.closest?.('input, textarea, select, [contenteditable="true"]');
  return field || document.querySelector('dialog[open]') !== null;
}

@Component({
  imports: [RouterLink, TranslatePipe],
  selector: 'rw-screen-header',
  styleUrl: './screen-header.css',
  templateUrl: './screen-header.html',
  host: { '(document:keydown.escape)': 'escape($event)' },
})
export class ScreenHeader {
  private readonly router = inject(Router);
  private readonly document = inject(DOCUMENT);
  readonly url = toSignal(this.router.events.pipe(filter((event) => event instanceof NavigationEnd), map((event) => event.urlAfterRedirects)), { initialValue: this.router.url });
  readonly key = computed(() => keyOf(this.url()));
  readonly back = computed(() => parentPath(this.url()));
  readonly titleKey = computed(() => `menu.${this.key()}`);

  escape(event: Event): void {
    const back = this.back();
    if (back === null || escapeIsTaken(event.target, this.document)) return;
    event.preventDefault();
    void this.router.navigateByUrl(back);
  }
}
