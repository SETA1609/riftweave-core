// Import on the home page (APP-30, docs/backend/files.md#in-the-frontend): one file, any kind. The
// envelope is read, the data checked by the core, and a module this browser does not load stops the
// import, naming the rows at stake, until the person enables it or confirms the loss.
import { Component, computed, inject, signal } from '@angular/core';
import { API, type Session, type StoredActor } from './api';
import { AppContext } from './context';
import { type MissingModule, type RiftweaveFile, missingModules, parseFile, rulesWarning } from './files';
import { I18n, type Said } from './i18n';
import { Saves } from './saves';
import { MODULE_IDS } from './settings-screens';
import { TranslatePipe } from './translate-pipe';
import { Alert } from './ui/alert';
import { Card } from './ui/card';

/** The CLI command that exports a file this screen opens. */
export const EXPORT_COMMAND = 'riftweave export';

/** A refused import: the file's name and why, and that nothing changed. */
function failed(name: string, reason: string): Said {
  return { key: 'import.failed', values: { name, reason } };
}

/** The rules this build ships (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;

@Component({
  selector: 'rw-file-import',
  imports: [Alert, Card, TranslatePipe],
  templateUrl: './file-import.html',
})
export class FileImport {
  private readonly api = inject(API);
  private readonly context = inject(AppContext);
  private readonly saves = inject(Saves);
  private readonly i18n = inject(I18n);
  /** The CLI command that writes these files, shown as code. */
  readonly exportCommand = EXPORT_COMMAND;
  readonly error = signal<Said | null>(null);
  readonly notice = signal<Said | null>(null);
  readonly warnings = signal<Said[]>([]);
  /** A checked file held back by a missing module, and the modules at stake. */
  readonly pending = signal<{ name: string; file: RiftweaveFile; missing: MissingModule[] } | null>(null);
  /** The alert for the missing modules, in the language shown; in English word for word the CLI's. */
  readonly alert = computed(() => {
    const pending = this.pending();
    if (pending === null) return '';
    const list = pending.missing
      .map((entry) => this.i18n.translate('import.module_missing', { module: entry.module, rows: entry.rows.length ? entry.rows.join(', ') : this.i18n.translate('import.no_rows') }))
      .join('; ');
    return this.i18n.translate('import.modules_alert', { list });
  });

  async pick(event: Event): Promise<void> {
    const input = event.target as HTMLInputElement;
    const picked = input.files?.[0];
    input.value = '';
    if (picked) {
      await this.read(picked.name, await picked.text());
    }
  }

  /** Read and check a file; apply it unless a module is missing. */
  async read(name: string, text: string): Promise<void> {
    this.error.set(null);
    this.notice.set(null);
    this.warnings.set([]);
    this.pending.set(null);
    const parsed = parseFile(text);
    if ('error' in parsed) {
      this.error.set(failed(name, parsed.error));
      return;
    }
    if ('settings' in parsed) {
      const problem = this.context.importSettings(text);
      if (problem) this.error.set(failed(name, problem));
      else this.notice.set({ key: 'import.settings_imported' });
      return;
    }
    try {
      await this.check(parsed.file);
    } catch (failure) {
      this.error.set(failed(name, failure instanceof Error ? failure.message : String(failure)));
      return;
    }
    if (rulesWarning(parsed.file, this.rulesVersion())) this.warnings.set([{ key: 'import.rules_differ', values: { made: parsed.file.rules_version, runs: this.rulesVersion() } }]);
    const missing = missingModules(parsed.file, this.context.modules());
    if (missing.length > 0) {
      this.pending.set({ name, file: parsed.file, missing });
      return;
    }
    await this.apply(name, parsed.file);
  }

  /** Enable the missing modules in the settings, then import. */
  async enable(): Promise<void> {
    const pending = this.pending();
    if (pending === null) return;
    const unknown = pending.missing.find((entry) => !MODULE_IDS.includes(entry.module));
    if (unknown) {
      this.error.set({ key: 'import.module_unknown', values: { name: pending.name, module: unknown.module } });
      return;
    }
    for (const entry of pending.missing) {
      if (!this.context.modules().includes(entry.module)) this.context.toggleModule(entry.module);
    }
    await this.apply(pending.name, pending.file);
  }

  /** Import without the missing modules: their rows are kept in the data but nothing here reads them. */
  async acceptLoss(): Promise<void> {
    const pending = this.pending();
    if (pending === null) return;
    this.warnings.update((list) => [...list, { key: 'import.without_modules', values: { modules: pending.missing.map((entry) => entry.module).join(', ') } }]);
    await this.apply(pending.name, pending.file);
  }

  private rulesVersion(): string {
    return this.saves.current()?.session.rules_version ?? RIFTWEAVE_RULES_VERSION;
  }

  /** The data against the contract, through the core: the same check the CLI's import makes. */
  private async check(file: RiftweaveFile): Promise<void> {
    switch (file.riftweave) {
      case 'character':
        await this.api.checkActor(file.data as StoredActor);
        return;
      case 'session':
        await this.api.replay(file.data as Session);
        return;
      case 'spell': {
        const result = (await this.api.author({ kind: 'spell', spells: [file.data as Record<string, unknown>] } as never)) as unknown as { spells: { problems: { path: string; message: string }[] }[] };
        const problems = result.spells[0]?.problems ?? [];
        if (problems.length > 0) throw new Error(`the spell does not check: ${problems.map((problem) => `${problem.path} ${problem.message}`.trim()).join('; ')}`);
        return;
      }
    }
  }

  private async apply(name: string, file: RiftweaveFile): Promise<void> {
    this.pending.set(null);
    const label = name.replace(/\.riftweave\.json$|\.json$/, '');
    switch (file.riftweave) {
      case 'character':
        await this.saves.begin(label, await this.api.startSession(file.data as StoredActor));
        this.notice.set({ key: 'import.character_imported', values: { label } });
        return;
      case 'session':
        await this.saves.begin(label, file.data as Session);
        this.notice.set({ key: 'import.session_imported', values: { label } });
        return;
      case 'spell':
        this.notice.set({ key: 'import.spell_checks' });
        return;
    }
  }
}
