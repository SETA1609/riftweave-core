// The shell: the three screens (docs/backend/frontend.md#screens) behind a tab bar, shown once the
// core can answer (at once over HTTP, after the WebAssembly core starts in a static build).
import { Component, inject, signal } from '@angular/core';
import { API } from './api';
import { Attack } from './attack';
import { Playthrough } from './playthrough';
import { SavedGames } from './saved-games';
import { Saves } from './saves';
import { Sheet } from './sheet';
import { Alert } from './ui/alert';
import { ThemeToggle } from './ui/theme-toggle';

/** The rules this build ships: the release tag in the Pages build (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;

@Component({
  selector: 'rw-root',
  imports: [Sheet, Attack, Playthrough, SavedGames, Alert, ThemeToggle],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  readonly rulesVersion = RIFTWEAVE_RULES_VERSION;
  readonly tabs = ['sheet', 'attack', 'playthrough', 'saves'] as const;
  readonly current = signal<(typeof this.tabs)[number]>('sheet');
  readonly state = signal<'starting' | 'ready' | 'failed'>('starting');
  readonly failure = signal('');

  constructor() {
    const ready: Promise<void> = inject(API).ready ?? Promise.resolve();
    const saves = inject(Saves);
    ready.then(
      // Resume the latest save; a browser without IndexedDB still plays, unsaved.
      () => saves.load().catch(() => undefined).finally(() => this.state.set('ready')),
      (error: unknown) => {
        this.failure.set(`The rules engine did not start: ${error instanceof Error ? error.message : String(error)}`);
        this.state.set('failed');
      },
    );
  }
}
