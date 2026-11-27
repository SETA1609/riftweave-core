// Skill resolution: one d100 check through the `resolution` command (docs/core/resolution.md), as
// the CLI's form. A roll the viewer leaves to the app is rolled here and sent as an input, so the
// check can be replayed with the agent CLI.
import { Component, inject, signal } from '@angular/core';
import { API, type ResolutionInput, type ResolutionResult } from './api';
import { rollD100 } from './saves';
import { TranslatePipe } from './translate-pipe';
import { Alert } from './ui/alert';
import { Card } from './ui/card';
import { KeyValueList } from './ui/key-value-list';
import { NumberField } from './ui/number-field';
import { TextField } from './ui/text-field';

@Component({
  selector: 'rw-resolution',
  imports: [Alert, Card, KeyValueList, NumberField, TextField, TranslatePipe],
  templateUrl: './resolution.html',
  styleUrl: './resolution.css',
})
export class Resolution {
  private readonly api = inject(API);
  readonly skill = signal('lockpick');
  readonly value = signal(40);
  readonly luck = signal(5);
  readonly bonus = signal(0);
  readonly roll = signal(50);
  readonly result = signal<ResolutionResult | null>(null);
  readonly error = signal('');

  rollForMe(): void {
    this.roll.set(rollD100());
  }

  /** The `resolution` input: the skill at its value, LCK, one target modifier, and the roll. */
  input(): ResolutionInput {
    const skill = this.skill().trim();
    return {
      actor: { skills: { [skill]: this.value() }, attributes: { lck: this.luck() } },
      skill,
      modifiers: this.bonus() ? [{ target: skill, operation: 'add', value: this.bonus(), source: 'situation:gm' }] : [],
      roll: this.roll(),
    } as ResolutionInput;
  }

  async resolve(): Promise<void> {
    this.error.set('');
    try {
      this.result.set(await this.api.resolution(this.input()));
    } catch (failure) {
      this.error.set(String(failure));
    }
  }
}
