// A labelled number input bound two-way to a signal (`[(value)]`).
import { Component, input, model } from '@angular/core';

@Component({ selector: 'rw-number-field', templateUrl: './number-field.html', styleUrl: './number-field.css' })
export class NumberField {
  readonly label = input.required<string>();
  readonly min = input<number | null>(null);
  readonly max = input<number | null>(null);
  readonly value = model(0);

  read(event: Event): void {
    this.value.set(Number((event.target as HTMLInputElement).value));
  }
}
