// A labelled text input bound two-way to a signal (`[(value)]`).
import { Component, input, model } from '@angular/core';

@Component({ selector: 'rw-text-field', templateUrl: './text-field.html', styleUrl: './text-field.css' })
export class TextField {
  readonly label = input.required<string>();
  readonly value = model('');

  read(event: Event): void {
    this.value.set((event.target as HTMLInputElement).value);
  }
}
