// Import on the home page (APP-30, docs/backend/files.md#in-the-frontend): one file, any kind. The
// envelope is read, the data checked by the core, and a module this browser does not load stops the
// import, naming the rows at stake, until the person enables it or confirms the loss.
import { Component, inject, signal } from '@angular/core';
import { API, type Session, type StoredActor } from './api';
import { AppContext } from './context';
import { type MissingModule, type RiftweaveFile, missingModules, moduleAlert, parseFile, rulesWarning } from './files';
import { Saves } from './saves';
import { MODULE_IDS } from './settings-screens';
import { Alert } from './ui/alert';
import { Card } from './ui/card';

/** The rules this build ships (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;

@Component({
  selector: 'rw-file-import',
  imports: [Alert, Card],
  templateUrl: './file-import.html',
})
export class FileImport {
  private readonly api = inject(API);
  private readonly context = inject(AppContext);
  private readonly saves = inject(Saves);
  readonly error = signal('');
  readonly notice = signal('');
  readonly warnings = signal<string[]>([]);
  /** A checked file held back by a missing module, and the modules at stake. */
  readonly pending = signal<{ name: string; file: RiftweaveFile; missing: MissingModule[] } | null>(null);
  readonly alert = signal('');

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
    this.error.set('');
    this.notice.set('');
    this.warnings.set([]);
    this.pending.set(null);
    this.alert.set('');
    const parsed = parseFile(text);
    if ('error' in parsed) {
      this.error.set(`${name}: ${parsed.error}; nothing changed`);
      return;
    }
    if ('settings' in parsed) {
      const problem = this.context.importSettings(text);
      if (problem) this.error.set(`${name}: ${problem}; nothing changed`);
      else this.notice.set('Imported the settings file.');
      return;
    }
    try {
      await this.check(parsed.file);
    } catch (failure) {
      this.error.set(`${name}: ${failure instanceof Error ? failure.message : String(failure)}; nothing changed`);
      return;
    }
    const warning = rulesWarning(parsed.file, this.rulesVersion());
    if (warning) this.warnings.set([warning]);
    const missing = missingModules(parsed.file, this.context.modules());
    if (missing.length > 0) {
      this.pending.set({ name, file: parsed.file, missing });
      this.alert.set(moduleAlert(missing));
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
      this.error.set(`${pending.name}: module ${unknown.module} is not in this app; nothing changed`);
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
    this.warnings.update((list) => [...list, `imported without its modules: ${this.alert()}`]);
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
    this.alert.set('');
    const label = name.replace(/\.riftweave\.json$|\.json$/, '');
    switch (file.riftweave) {
      case 'character':
        await this.saves.begin(label, await this.api.startSession(file.data as StoredActor));
        this.notice.set(`Imported a character file as the save "${label}".`);
        return;
      case 'session':
        await this.saves.begin(label, file.data as Session);
        this.notice.set(`Imported a session file as the save "${label}".`);
        return;
      case 'spell':
        this.notice.set('Imported a spell file: it checks. Spell creation, where it opens, comes with APP-22.');
        return;
    }
  }
}
