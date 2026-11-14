// The shell: the API the screens share, and the three screens (docs/backend/frontend.md#screens).
import { Component, InjectionToken, signal } from '@angular/core';
import { createApi, type Api } from './api';
import { Attack } from './attack';
import { Playthrough } from './playthrough';
import { Sheet } from './sheet';

/** The API; the app talks to it through the dev proxy, so its base URL is the page's own. */
export const API = new InjectionToken<Api>('riftweave API', { providedIn: 'root', factory: () => createApi('') });

@Component({
  selector: 'rw-root',
  imports: [Sheet, Attack, Playthrough],
  template: `
    <header><h1>Riftweave</h1>
      <nav>@for (tab of tabs; track tab) { <button [class.active]="tab === current()" (click)="current.set(tab)">{{ tab }}</button> }</nav>
    </header>
    @switch (current()) {
      @case ('sheet') { <rw-sheet /> }
      @case ('attack') { <rw-attack /> }
      @case ('playthrough') { <rw-playthrough /> }
    }
  `,
})
export class App {
  readonly tabs = ['sheet', 'attack', 'playthrough'] as const;
  readonly current = signal<(typeof this.tabs)[number]>('sheet');
}
