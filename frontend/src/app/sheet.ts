// The character sheet: attributes and level in, every derived stat out, and a session to play.
import { Component, inject, signal } from '@angular/core';
import { JsonPipe } from '@angular/common';
import { API } from './app';
import type { Attributes, Session } from './api';

const KEYS = ['str', 'per', 'end', 'int', 'wil', 'agi', 'cha', 'lck'] as const;

@Component({
  selector: 'rw-sheet',
  imports: [JsonPipe],
  template: `
    <h2>Character sheet</h2>
    <div class="row">
      @for (key of keys; track key) {
        <label>{{ key }} <input type="number" min="1" [value]="attributes()[key]" (input)="set(key, $event)" /></label>
      }
      <label>level <input type="number" min="1" [value]="level()" (input)="level.set(+$any($event.target).value)" /></label>
    </div>
    <button (click)="compute()">Compute stats</button>
    <button (click)="create()">Create character</button>
    @if (error()) { <p class="error">{{ error() }}</p> }
    @if (stats(); as values) {
      <table>@for (stat of entries(values); track stat[0]) { <tr><th>{{ stat[0] }}</th><td>{{ stat[1] }}</td></tr> }</table>
    }
    @if (session(); as created) { <h3>Stored actor</h3><pre>{{ created.actor | json }}</pre> }
  `,
})
export class Sheet {
  private readonly api = inject(API);
  readonly keys = KEYS;
  readonly attributes = signal<Attributes>({ str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 });
  readonly level = signal(1);
  readonly stats = signal<Record<string, unknown> | null>(null);
  readonly session = signal<Session | null>(null);
  readonly error = signal('');

  set(key: (typeof KEYS)[number], event: Event): void {
    this.attributes.update((current) => ({ ...current, [key]: +(event.target as HTMLInputElement).value }));
  }

  entries(values: Record<string, unknown>): [string, unknown][] {
    return Object.entries(values);
  }

  async compute(): Promise<void> {
    await this.run(async () => this.stats.set(await this.api.derived({ attributes: this.attributes(), level: this.level() })));
  }

  async create(): Promise<void> {
    await this.run(async () => this.session.set(await this.api.startSession(await this.api.newCharacter(this.attributes()))));
  }

  private async run(call: () => Promise<void>): Promise<void> {
    this.error.set('');
    try { await call(); } catch (failure) { this.error.set(String(failure)); }
  }
}
