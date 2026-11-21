// The interface language (docs/backend/frontend.md#languages): a select of the offered languages,
// each by its own name, bound to AppContext.locale. It sits in the header, so every screen has it.
import { Component, inject } from '@angular/core';
import { AppContext, LANGUAGES } from '../context';
import { I18n } from '../i18n';
import { TranslatePipe } from '../translate-pipe';
import { SelectField } from './select-field';

@Component({
  imports: [SelectField, TranslatePipe],
  selector: 'rw-language-select',
  styleUrl: './language-select.css',
  templateUrl: './language-select.html',
})
export class LanguageSelect {
  readonly context = inject(AppContext);
  readonly i18n = inject(I18n);
  readonly languages = LANGUAGES;
}
