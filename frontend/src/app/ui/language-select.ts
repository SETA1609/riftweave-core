// See language-select.md.
import { Component, ElementRef, computed, inject, input, signal } from '@angular/core';
import { AppContext, LANGUAGES } from '../context';
import { I18n } from '../i18n';
import { TranslatePipe } from '../translate-pipe';

@Component({
  imports: [TranslatePipe],
  selector: 'rw-language-select',
  styleUrl: './language-select.css',
  templateUrl: './language-select.html',
  host: { '(document:click)': 'outside($event)', '(keydown.escape)': 'open.set(false)', '[class.compact]': 'compact()' },
})
export class LanguageSelect {
  readonly context = inject(AppContext);
  readonly i18n = inject(I18n);
  private readonly element = inject<ElementRef<HTMLElement>>(ElementRef);
  readonly languages = LANGUAGES;
  readonly open = signal(false);
  /** Show the language's code (EN) instead of its name, for a narrow bar. */
  readonly compact = input(false);
  readonly current = computed(() => LANGUAGES.find((language) => language.value === this.i18n.language()) ?? LANGUAGES[0]);

  pick(code: string): void {
    this.context.locale.set(code);
    this.open.set(false);
  }

  outside(event: Event): void {
    if (!this.element.nativeElement.contains(event.target as Node)) this.open.set(false);
  }
}
