// The light and dark mode toggle: one button that cycles light, dark, and the system's mode
// (docs/backend/frontend.md#the-app-context). It reads and sets the mode through AppContext.
import { Component, computed, inject } from '@angular/core';
import { AppContext, type Theme } from '../context';

const LABELS: Record<Theme, string> = { light: 'light mode', dark: 'dark mode', system: "the system's mode" };

@Component({
  selector: 'rw-theme-toggle',
  styleUrl: './theme-toggle.css',
  templateUrl: './theme-toggle.html',
})
export class ThemeToggle {
  readonly context = inject(AppContext);
  /** What the button shows: the current mode. */
  readonly current = computed(() => this.context.theme());
  /** What the button says it does: the mode it switches to. */
  readonly action = computed(() => `Switch to ${LABELS[this.context.next()]}`);
}
