// The saved games (docs/backend/frontend.md#saves): the saves in this browser, a backup file for
// each, and opening a file, by picking it or dropping it here. Opening replays the file first.
import { Component, inject, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Saves, needsBackup, type Opened, type SaveRecord } from './saves';
import { Alert } from './ui/alert';
import { Card } from './ui/card';

@Component({
  selector: 'rw-saved-games',
  imports: [DatePipe, Alert, Card],
  templateUrl: './saved-games.html',
  styleUrl: './saved-games.css',
})
export class SavedGames {
  readonly saves = inject(Saves);
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
