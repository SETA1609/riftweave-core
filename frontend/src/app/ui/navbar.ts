// See navbar.md.
import { Component, ElementRef, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FaIconComponent } from '@fortawesome/angular-fontawesome';
import { faGithub } from '@fortawesome/free-brands-svg-icons';
import { faBars } from '@fortawesome/free-solid-svg-icons';
import { TranslatePipe } from '../translate-pipe';
import { LanguageSelect } from './language-select';
import { ThemeToggle } from './theme-toggle';

export const REPOSITORY_URL = 'https://github.com/SETA1609/riftweave-core';
export const NAV_ITEMS_ID = 'rw-navbar-items';

@Component({
  imports: [FaIconComponent, LanguageSelect, RouterLink, ThemeToggle, TranslatePipe],
  selector: 'rw-navbar',
  styleUrl: './navbar.css',
  templateUrl: './navbar.html',
  host: { '(document:click)': 'outside($event)', '(keydown.escape)': 'escape($event)' },
})
export class Navbar {
  private readonly element = inject<ElementRef<HTMLElement>>(ElementRef);
  readonly github = faGithub;
  readonly bars = faBars;
  readonly repository = REPOSITORY_URL;
  readonly itemsId = NAV_ITEMS_ID;
  /** Whether the burger's dropdown is open; on a wide screen the items show inline and this does nothing. */
  readonly open = signal(false);

  toggle(): void {
    this.open.update((open) => !open);
  }

  outside(event: Event): void {
    if (!this.element.nativeElement.contains(event.target as Node)) this.open.set(false);
  }

  /** Escape closes the dropdown, and goes no further while it was open. */
  escape(event: Event): void {
    if (!this.open()) return;
    this.open.set(false);
    event.stopPropagation();
  }
}
