// See navbar.md.
import { Component, ElementRef, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FaIconComponent } from '@fortawesome/angular-fontawesome';
import { faGithub } from '@fortawesome/free-brands-svg-icons';
import { faBars, faCube } from '@fortawesome/free-solid-svg-icons';
import { Saves } from '../saves';
import { TranslatePipe } from '../translate-pipe';
import { LanguageSelect } from './language-select';
import { MENU, pathOf } from './menu-grid';
import { ThemeToggle } from './theme-toggle';

export const REPOSITORY_URL = 'https://github.com/SETA1609/riftweave-core';
export const NAV_MENU_ID = 'rw-navbar-menu';
export const MAIN_MENU = 'main';

/** The rules this build ships (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;

@Component({
  imports: [FaIconComponent, LanguageSelect, RouterLink, ThemeToggle, TranslatePipe],
  selector: 'rw-navbar',
  styleUrl: './navbar.css',
  templateUrl: './navbar.html',
  host: { '(document:click)': 'outside($event)', '(keydown.escape)': 'escape($event)' },
})
export class Navbar {
  private readonly element = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly saves = inject(Saves);
  readonly github = faGithub;
  readonly bars = faBars;
  readonly cube = faCube;
  readonly repository = REPOSITORY_URL;
  readonly menuId = NAV_MENU_ID;
  readonly rulesVersion = RIFTWEAVE_RULES_VERSION;
  /** The main menu of backend/contract/menu.json, the same entries the home page shows. */
  readonly links = MENU.filter((entry) => entry.parent === MAIN_MENU && !entry.issue).map((entry) => ({ key: entry.key, path: pathOf(entry.key) }));
  /** The game in play, by its save's name; none before a game starts. */
  readonly playing = computed(() => this.saves.current()?.name ?? null);
  /** Whether the burger's menu is open; from 1024px the links show inline and the burger is hidden. */
  readonly open = signal(false);

  toggle(): void {
    this.open.update((open) => !open);
  }

  close(): void {
    this.open.set(false);
  }

  outside(event: Event): void {
    if (!this.element.nativeElement.contains(event.target as Node)) this.open.set(false);
  }

  /** Escape closes the menu, and goes no further while it was open. */
  escape(event: Event): void {
    if (!this.open()) return;
    this.open.set(false);
    event.stopPropagation();
  }
}
