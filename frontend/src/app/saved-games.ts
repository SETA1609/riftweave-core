// The saved games (docs/backend/frontend.md#saves): the saves in this browser, a backup file for
// each, and opening a file, by picking it or dropping it here. Opening replays the file first.
import { Component, inject, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Saves, needsBackup, type Opened, type SaveRecord } from './saves';
import { AppContext } from './context';
import { download, encodeFile, makeFile } from './files';
import { MODULE_VERSIONS } from './settings-screens';

/** The rules this build ships (angular.json `define`). */
declare const RIFTWEAVE_RULES_VERSION: string;
import { TranslatePipe } from './translate-pipe';
import { Alert } from './ui/alert';
import { Card } from './ui/card';

/** The ending of a backup file this screen opens. */
export const FILE_SUFFIX = '.riftweave.json';

@Component({
  selector: 'rw-saved-games',
  imports: [DatePipe, Alert, Card, TranslatePipe],
  templateUrl: './saved-games.html',
  styleUrl: './saved-games.css',
})
export class SavedGames {
  readonly saves = inject(Saves);
  readonly fileSuffix = FILE_SUFFIX;
  private readonly context = inject(AppContext);
  readonly opened = signal<Opened | null>(null);
  readonly error = signal('');
  readonly dragging = signal(false);

  stale(record: SaveRecord): boolean {
    return needsBackup(record);
  }

  async download(record: SaveRecord): Promise<void> {
    const file = await this.saves.backup(record.id);
    const link = document.createElement('a');
    link.href = URL.createObjectURL(new Blob([file.text], { type: 'application/json' }));
    link.download = file.name;
    link.click();
    URL.revokeObjectURL(link.href);
  }

  /** The save as a session file both apps open (APP-30). The backup stays the session itself. */
  exportFile(record: SaveRecord): void {
    const modules = MODULE_VERSIONS.filter((module) => this.context.modules().includes(module.id));
    const file = makeFile('session', record.session, record.session.rules_version ?? RIFTWEAVE_RULES_VERSION, modules);
    download(`${record.name.replace(/[^\w.-]+/g, '-').toLowerCase()}.session.riftweave.json`, encodeFile(file));
  }

  async pick(event: Event): Promise<void> {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (file) {
      await this.read(file);
    }
  }

  async drop(event: DragEvent): Promise<void> {
    event.preventDefault();
    this.dragging.set(false);
    const file = event.dataTransfer?.files[0];
    if (file) {
      await this.read(file);
    }
  }

  async keep(which: 'replayed' | 'saved'): Promise<void> {
    const opened = this.opened();
    if (opened !== null) {
      await this.saves.keep(opened, which);
      this.opened.set(null);
    }
  }

  private async read(file: File): Promise<void> {
    this.error.set('');
    this.opened.set(null);
    try {
      this.opened.set(await this.saves.open(file.name, await file.text()));
    } catch (failure) {
      this.error.set(failure instanceof Error ? failure.message : String(failure));
    }
  }
}
