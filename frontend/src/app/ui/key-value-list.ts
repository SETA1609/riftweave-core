// A two-column table of named values, such as a sheet's derived stats.
import { Component, computed, input } from '@angular/core';

@Component({ selector: 'rw-key-value-list', templateUrl: './key-value-list.html', styleUrl: './key-value-list.css' })
export class KeyValueList {
  readonly entries = input.required<Readonly<Record<string, unknown>>>();
  readonly rows = computed(() => Object.entries(this.entries()));

  shown(value: unknown): string {
    return typeof value === 'string' ? value : JSON.stringify(value);
  }
}
