// The light and dark mode toggle: one button that cycles light, dark, and the system's mode
// (docs/backend/frontend.md#the-app-context). It reads and sets the mode through AppContext.
import { Component, computed, inject } from '@angular/core';
import { AppContext } from '../context';
import { I18n } from '../i18n';
import { TranslatePipe } from '../translate-pipe';

@Component({
  imports: [TranslatePipe],
  selector: 'rw-theme-toggle',
  styleUrl: './theme-toggle.css',
  templateUrl: './theme-toggle.html',
})
export class ThemeToggle {
  readonly context = inject(AppContext);
  private readonly i18n = inject(I18n);
  /** What the button shows: the current mode. */
  readonly current = computed(() => this.context.theme());
  /** What the button says it does: the mode it switches to. */
  readonly action = computed(() => this.i18n.translate('theme.switch_to', { mode: this.i18n.translate(`theme.mode_${this.context.next()}`) }));
}
