// See theme-toggle.md.
import { Component, computed, inject } from '@angular/core';
import { FaIconComponent } from '@fortawesome/angular-fontawesome';
import { type IconDefinition, faCircleHalfStroke, faMoon, faSun } from '@fortawesome/free-solid-svg-icons';
import { AppContext, type Theme } from '../context';
import { I18n } from '../i18n';

const ICONS: Readonly<Record<Theme, IconDefinition>> = { light: faSun, dark: faMoon, system: faCircleHalfStroke };

@Component({
  imports: [FaIconComponent],
  selector: 'rw-theme-toggle',
  styleUrl: './theme-toggle.css',
  templateUrl: './theme-toggle.html',
})
export class ThemeToggle {
  readonly context = inject(AppContext);
  private readonly i18n = inject(I18n);
  readonly icon = computed(() => ICONS[this.context.theme()]);
  readonly action = computed(() => this.i18n.translate('theme.switch_to', { mode: this.i18n.translate(`theme.mode_${this.context.next()}`) }));
}
