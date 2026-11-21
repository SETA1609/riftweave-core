// The character list (FE_S07_04, docs/backend/files.md#the-character-list): the characters saved in
// this browser with who changed them last, and the hand-off between a player and the GM as files.
// An import follows the sync rule; a lower revision is refused and a diverged one asks, in the page.
import { Component, inject, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Characters, type ImportReport, type SavedCharacter } from './characters';
import { download } from './files';
import { Alert } from './ui/alert';
import { Card } from './ui/card';

@Component({
  imports: [DatePipe, Alert, Card],
  selector: 'rw-character-list',
  styleUrl: './character-list.css',
  templateUrl: './character-list.html',
})
export class CharacterList {
  readonly characters = inject(Characters);
  readonly error = signal('');
  readonly notice = signal('');
  /** A diverged import waiting for the person to keep one side, and the file it came from. */
  readonly diverged = signal<{ report: ImportReport; text: string } | null>(null);

  constructor() {
    void this.characters.load().catch((failure: unknown) => this.error.set(String(failure)));
  }

  level(saved: SavedCharacter): number {
    return saved.data.level;
  }

  async pick(event: Event): Promise<void> {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (file) {
      await this.read(file.name, await file.text());
    }
  }

  async exportFile(saved: SavedCharacter): Promise<void> {
    const file = await this.characters.export(saved.character.id);
    download(file.name, file.text);
  }

  async duplicate(saved: SavedCharacter): Promise<void> {
    await this.run(async () => {
      const copy = await this.characters.duplicate(saved.character.id);
      this.notice.set(`Duplicated as ${copy.name}.`);
    });
  }

  async remove(saved: SavedCharacter): Promise<void> {
    await this.run(async () => {
      await this.characters.remove(saved.character.id);
      this.notice.set(`Deleted ${saved.character.name}.`);
    });
  }

  /** Keep one side of a divergence; the other is dropped, never merged. */
  async keep(side: 'saved' | 'incoming'): Promise<void> {
    const pending = this.diverged();
    if (pending !== null) {
      await this.read('', pending.text, side);
    }
  }

  private async read(name: string, text: string, keep?: 'saved' | 'incoming'): Promise<void> {
    await this.run(async () => {
      const report = await this.characters.import(text, { keep });
      this.diverged.set(report.outcome === 'diverged' ? { report, text } : null);
      const said: Record<string, string> = {
        added: `Added ${report.character.name}, revision ${report.character.revision}.`,
        replaced: `Replaced ${report.character.name} with revision ${report.character.revision}.`,
        unchanged: `${report.character.name} is already saved at revision ${report.character.revision}; nothing changed.`,
        kept: `Kept the saved ${report.character.name}.`,
      };
      switch (report.outcome) {
        case 'refused':
          this.error.set(report.note ?? '');
          break;
        case 'diverged':
          this.notice.set(report.note ?? '');
          break;
        default:
          this.notice.set([said[report.outcome], ...report.warnings].join(' '));
      }
    }, name);
  }

  private async run(action: () => Promise<void>, name = ''): Promise<void> {
    this.error.set('');
    this.notice.set('');
    try {
      await action();
    } catch (failure) {
      const message = failure instanceof Error ? failure.message : String(failure);
      this.error.set(name ? `${name}: ${message}; nothing changed` : message);
    }
  }
}
