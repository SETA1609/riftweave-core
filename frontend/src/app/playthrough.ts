// The playthrough runner: ruleset/tests/vectors/playthrough.json as session steps, then a replay.
import { Component, inject, signal } from '@angular/core';
import { JsonPipe } from '@angular/common';
import playthrough from '../../../ruleset/tests/vectors/playthrough.json';
import { API, type Session } from './api';
import { Alert } from './ui/alert';
import { Card } from './ui/card';

type Vector = { name: string; function: string; expect: Record<string, unknown> } & Record<string, unknown>;
type Row = { name: string; input: Record<string, unknown>; result: Record<string, unknown>; matches: boolean };

@Component({
  selector: 'rw-playthrough',
  imports: [JsonPipe, Alert, Card],
  templateUrl: './playthrough.html',
  styleUrl: './playthrough.css',
})
export class Playthrough {
  private readonly api = inject(API);
  readonly rows = signal<Row[]>([]);
  readonly replayed = signal<boolean | null>(null);
  readonly error = signal('');

  async run(): Promise<void> {
    this.error.set('');
    this.rows.set([]);
    this.replayed.set(null);
    try {
      const actor = await this.api.newCharacter({ str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 });
      let session: Session = await this.api.startSession(actor);
      for (const vector of playthrough as Vector[]) {
        const { name, function: command, expect, ...input } = vector;
        const answer = await this.api.step(session, { command, input: input as Record<string, never> });
        session = answer.session;
        const result = (answer.result ?? {}) as Record<string, unknown>;
        const matches = Object.entries(expect).every(([field, value]) => JSON.stringify(result[field]) === JSON.stringify(value));
        this.rows.update((rows) => [...rows, { name, input, result, matches }]);
      }
      this.replayed.set((await this.api.replay(session)).matches ?? false);
    } catch (failure) {
      this.error.set(String(failure));
    }
  }
}
