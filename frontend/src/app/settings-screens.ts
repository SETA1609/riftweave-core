// Settings, Modules, and Network (docs/backend/app-layout.md#settings-modules-and-network): three
// routes over one component, because all three edit the one settings file of APP-25 through the
// app context. Every change is checked by the file's rules and saved at once.
import { Component, computed, inject, input, signal } from '@angular/core';
import exampleModule from '../../../ruleset/modules/example-module/manifest.json';
import factions from '../../../ruleset/modules/factions/manifest.json';
import wuxing from '../../../ruleset/modules/wuxing/manifest.json';
import { AppContext, THEMES } from './context';
import { I18n } from './i18n';
import { TranslatePipe } from './translate-pipe';
import { Alert } from './ui/alert';
import { Card } from './ui/card';
import { SelectField } from './ui/select-field';
import { TextField } from './ui/text-field';

/** The modules in ruleset/modules/, by manifest id, as the CLI lists them; a new module adds its
 * manifest here. */
export const MODULE_IDS: readonly string[] = [exampleModule.id, factions.id, wuxing.id].sort();

/** The same modules with the versions their manifests name, for a file's `modules` (APP-30). */
export const MODULE_VERSIONS: readonly { id: string; version: string }[] = [exampleModule, factions, wuxing]
  .map((manifest) => ({ id: manifest.id, version: manifest.version }))
  .sort((a, b) => a.id.localeCompare(b.id));

@Component({
  selector: 'rw-settings-screens',
  imports: [Alert, Card, SelectField, TextField, TranslatePipe],
  templateUrl: './settings-screens.html',
  styleUrl: './settings-screens.css',
})
export class SettingsScreens {
  readonly context = inject(AppContext);
  private readonly i18n = inject(I18n);
  /** Which screen: settings, modules, or network; a route sets it from its data. */
  readonly section = input<'settings' | 'modules' | 'network'>('settings');
  readonly themes = computed(() => THEMES.map((theme) => ({ value: theme, label: this.i18n.translate(`theme.${theme}`) })));
  /** The issue that builds rooms, shown beside the disabled choice. */
  readonly roomIssue = 'APP-17';
  readonly moduleIds = MODULE_IDS;
  readonly url = signal(this.context.serverUrl());
  readonly problem = signal('');
  /** The locale key of what a change did, shown in the current language. */
  readonly notice = signal('');

  /** Show the rule a refused change breaks, or clear it and show the notice key. */
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
    if (file) this.report(this.context.importSettings(await file.text()), 'settings.imported');
  }

  useServer(): void {
    const url = this.url().trim();
    if (!url) return this.report(this.i18n.translate('network.needs_url'));
    this.report(this.context.setNetwork('server', url), 'network.server_notice');
  }
}
