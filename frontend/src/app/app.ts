// The shell: the routes of the app layout (docs/backend/app-layout.md) in the router outlet, shown
// once the core can answer (at once over HTTP, after the WebAssembly core starts in a static build).
import { Component, inject, signal } from '@angular/core';
import { API } from './api';
import { RouterLink, RouterOutlet } from '@angular/router';
import { Saves } from './saves';
import { Alert } from './ui/alert';
import { ThemeToggle } from './ui/theme-toggle';

/** The rules this build ships: the release tag in the Pages build (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;

@Component({
  selector: 'rw-root',
  imports: [RouterLink, RouterOutlet, Alert, ThemeToggle],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  readonly rulesVersion = RIFTWEAVE_RULES_VERSION;
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
