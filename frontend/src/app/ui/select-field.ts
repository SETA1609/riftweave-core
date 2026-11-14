// A labelled select bound two-way to a signal (`[(value)]`).
import { Component, input, model } from '@angular/core';

export type Option = { readonly value: string; readonly label: string };

@Component({ selector: 'rw-select-field', templateUrl: './select-field.html', styleUrl: './select-field.css' })
export class SelectField {
  readonly label = input.required<string>();
  readonly options = input.required<readonly Option[]>();
  readonly value = model('');

  read(event: Event): void {
    this.value.set((event.target as HTMLSelectElement).value);
  }
}
