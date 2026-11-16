// The shell: the three screens (docs/backend/frontend.md#screens) behind a tab bar, shown once the
// core can answer (at once over HTTP, after the WebAssembly core starts in a static build).
import { Component, inject, signal } from '@angular/core';
import { API } from './api';
import { Attack } from './attack';
import { Playthrough } from './playthrough';
import { Sheet } from './sheet';
import { Alert } from './ui/alert';

@Component({
  selector: 'rw-root',
  imports: [Sheet, Attack, Playthrough, Alert],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  readonly tabs = ['sheet', 'attack', 'playthrough'] as const;
  readonly current = signal<(typeof this.tabs)[number]>('sheet');
  readonly state = signal<'starting' | 'ready' | 'failed'>('starting');
  readonly failure = signal('');

  constructor() {
    const ready: Promise<void> = inject(API).ready ?? Promise.resolve();
    ready.then(
      () => this.state.set('ready'),
      (error: unknown) => {
        this.failure.set(`The rules engine did not start: ${error instanceof Error ? error.message : String(error)}`);
        this.state.set('failed');
      },
    );
  }
}
