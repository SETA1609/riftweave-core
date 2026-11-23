// The character list (FE_S07_04, docs/backend/files.md#the-character-list): the characters saved in
// this browser with who changed them last, and the hand-off between a player and the GM as files.
// An import follows the sync rule; a lower revision is refused and a diverged one asks, in the page.
import { Component, inject, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { AUTHOR, Characters, OUTCOME, SIDE, type ImportReport, type SavedCharacter, type Side } from './characters';
import { API } from './api';
import { PRINT_STATE, download, printPage, sheetPrint, type PrintState } from './files';
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
  private readonly api = inject(API);
  readonly printState = PRINT_STATE;
  /** For the template: the sides a divergence keeps, and who changed a character. */
  readonly side = SIDE;
  readonly author = AUTHOR;
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

  /** Prints the character as saved (BE_S09_03): its session state, or fresh. */
  async print(saved: SavedCharacter, state: PrintState): Promise<void> {
    await this.run(async () => printPage((await this.api.print(sheetPrint(saved.data, state))).html), saved.character.name);
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
  async keep(side: Side): Promise<void> {
    const pending = this.diverged();
    if (pending !== null) {
      await this.read('', pending.text, side);
    }
  }

  private async read(name: string, text: string, keep?: Side): Promise<void> {
    await this.run(async () => {
      const report = await this.characters.import(text, { keep });
      this.diverged.set(report.outcome === OUTCOME.Diverged ? { report, text } : null);
      const said: Record<string, string> = {
        [OUTCOME.Added]: `Added ${report.character.name}, revision ${report.character.revision}.`,
        [OUTCOME.Replaced]: `Replaced ${report.character.name} with revision ${report.character.revision}.`,
        [OUTCOME.Unchanged]: `${report.character.name} is already saved at revision ${report.character.revision}; nothing changed.`,
        [OUTCOME.Kept]: `Kept the saved ${report.character.name}.`,
      };
      switch (report.outcome) {
        case OUTCOME.Refused:
          this.error.set(report.note ?? '');
          break;
        case OUTCOME.Diverged:
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
