// The shell: the three screens (docs/backend/frontend.md#screens) behind a tab bar.
import { Component, signal } from '@angular/core';
import { Attack } from './attack';
import { Playthrough } from './playthrough';
import { Sheet } from './sheet';

@Component({
  selector: 'rw-root',
  imports: [Sheet, Attack, Playthrough],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  readonly tabs = ['sheet', 'attack', 'playthrough'] as const;
  readonly current = signal<(typeof this.tabs)[number]>('sheet');
}
