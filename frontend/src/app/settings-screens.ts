// Settings, Modules, and Network (docs/backend/app-layout.md#settings-modules-and-network): three
// routes over one component, because all three edit the one settings file of APP-25 through the
// app context. Every change is checked by the file's rules and saved at once.
import { Component, inject, input, signal } from '@angular/core';
import exampleModule from '../../../ruleset/modules/example-module/manifest.json';
import factions from '../../../ruleset/modules/factions/manifest.json';
import wuxing from '../../../ruleset/modules/wuxing/manifest.json';
import { AppContext, THEMES } from './context';
import { Alert } from './ui/alert';
import { Card } from './ui/card';
import { SelectField } from './ui/select-field';
import { TextField } from './ui/text-field';

/** The modules in ruleset/modules/, by manifest id, as the CLI lists them; a new module adds its
 * manifest here. */
export const MODULE_IDS: readonly string[] = [exampleModule.id, factions.id, wuxing.id].sort();

@Component({
  selector: 'rw-settings-screens',
  imports: [Alert, Card, SelectField, TextField],
  templateUrl: './settings-screens.html',
  styleUrl: './settings-screens.css',
})
export class SettingsScreens {
  readonly context = inject(AppContext);
  /** Which screen: settings, modules, or network; a route sets it from its data. */
  readonly section = input<'settings' | 'modules' | 'network'>('settings');
  readonly themes = THEMES.map((theme) => ({ value: theme, label: theme }));
  readonly moduleIds = MODULE_IDS;
  readonly url = signal(this.context.serverUrl());
  readonly problem = signal('');
  readonly notice = signal('');

  /** Show the rule a refused change breaks, or clear it. */
  report(problem: string | null, notice = ''): void {
    this.problem.set(problem ?? '');
    this.notice.set(problem ? '' : notice);
  }

  setTheme(theme: string): void {
    this.context.theme.set(THEMES.find((known) => known === theme) ?? 'system');
  }

  exportSettings(): void {
    const link = document.createElement('a');
    link.href = URL.createObjectURL(new Blob([this.context.exportSettings()], { type: 'application/json' }));
    link.download = 'riftweave-settings.json';
    link.click();
    URL.revokeObjectURL(link.href);
  }

  async importSettings(event: Event): Promise<void> {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) this.report(this.context.importSettings(await file.text()), 'Settings imported.');
  }

  useServer(): void {
    this.report(this.context.setNetwork('server', this.url().trim()), 'Commands now go to the server.');
  }
}
