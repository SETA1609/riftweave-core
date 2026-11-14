// The character sheet: attributes and level in, every derived stat out, and a session to play.
import { Component, inject, signal } from '@angular/core';
import { JsonPipe } from '@angular/common';
import { API, type Attributes, type Session } from './api';
import { Alert } from './ui/alert';
import { Card } from './ui/card';
import { KeyValueList } from './ui/key-value-list';
import { NumberField } from './ui/number-field';

const KEYS = ['str', 'per', 'end', 'int', 'wil', 'agi', 'cha', 'lck'] as const;
type Key = (typeof KEYS)[number];

@Component({
  selector: 'rw-sheet',
  imports: [JsonPipe, Alert, Card, KeyValueList, NumberField],
  templateUrl: './sheet.html',
  styleUrl: './sheet.css',
})
export class Sheet {
  private readonly api = inject(API);
  readonly keys = KEYS;
  readonly attributes = signal<Attributes>({ str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 });
  readonly level = signal(1);
  readonly stats = signal<Record<string, unknown> | null>(null);
  readonly session = signal<Session | null>(null);
  readonly error = signal('');

  set(key: Key, value: number): void {
    this.attributes.update((current) => ({ ...current, [key]: value }));
  }

  async compute(): Promise<void> {
    await this.run(async () => this.stats.set(await this.api.derived({ attributes: this.attributes(), level: this.level() })));
  }

  async create(): Promise<void> {
    await this.run(async () => this.session.set(await this.api.startSession(await this.api.newCharacter(this.attributes()))));
  }

  private async run(call: () => Promise<void>): Promise<void> {
    this.error.set('');
    try {
      await call();
    } catch (failure) {
      this.error.set(String(failure));
    }
  }
}
