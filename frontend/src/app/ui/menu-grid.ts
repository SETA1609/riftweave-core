// One menu of backend/contract/menu.json as cards (docs/backend/app-layout.md): the home page, the
// tools page, and GM tools are this one component with another parent. A pending entry is a
// disabled card with its issue id, as the CLI dims it.
import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import tree from '../../../../backend/contract/menu.json';

export type MenuEntry = { readonly key: string; readonly parent: string; readonly label: string; readonly issue?: string; readonly apps?: readonly string[] };

/** The entries this app shows: the CLI-only ones (Quit) are left out. */
export const MENU: readonly MenuEntry[] = (tree.entries as MenuEntry[]).filter((entry) => !entry.apps || entry.apps.includes('web'));

/** The route of an entry: `/` for the main menu, `/play` for Start, `/<key>` for the rest of the
 * main menu, and `/play/<key>` for a tool. */
export function pathOf(key: string): string {
  if (key === 'main') return '/';
  if (key === 'start') return '/play';
  return MENU.find((entry) => entry.key === key)?.parent === 'main' ? `/${key}` : `/play/${key}`;
}

/** The label of a menu, "Riftweave" for the main one. */
export function titleOf(key: string): string {
  return MENU.find((entry) => entry.key === key)?.label ?? 'Riftweave';
}

@Component({
  selector: 'rw-menu-grid',
  imports: [RouterLink],
  templateUrl: './menu-grid.html',
  styleUrl: './menu-grid.css',
})
export class MenuGrid {
  /** The menu to show, a parent key of menu.json; a route sets it from its data. */
  readonly parent = input('main');
  readonly entries = computed(() => MENU.filter((entry) => entry.parent === this.parent()).map((entry) => ({ ...entry, path: pathOf(entry.key) })));
  readonly title = computed(() => titleOf(this.parent()));
  readonly back = computed(() => (this.parent() === 'main' ? null : pathOf(MENU.find((entry) => entry.key === this.parent())?.parent ?? 'main')));
}
