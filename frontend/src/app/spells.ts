// See spells.md.
import { Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { debounceTime, from, skip, switchMap } from 'rxjs';
import { API, type AuthorInput, type AuthorResult } from './api';
import { TranslatePipe } from './translate-pipe';
import { Alert } from './ui/alert';
import { Card } from './ui/card';
import { KeyValueList } from './ui/key-value-list';
import { NumberField } from './ui/number-field';
import { SelectField, type Option } from './ui/select-field';
import { TextField } from './ui/text-field';

export const AUTHOR = { Check: 'check', Load: 'load', Effects: 'effects' } as const;
export const SPELL_KIND = 'spell';
export const EXPORT_SHAPE = { Core: 'core' } as const;
export const EXPORT_FILE = 'spells.json';
export const CHECK_DELAY_MS = 300;
export const COLORS = ['red', 'orange', 'yellow', 'green', 'blue', 'indigo', 'violet', 'white', 'black'] as const;
export const TARGETS = ['self', 'touch', 'ranged', 'area', 'world'] as const;

export type AppliedEffect = { effect: string; magnitude?: number; duration?: number; target?: string; radius?: number; [field: string]: unknown };
export type SpellRow = { key: string; label: string; description: string; color: string; level: number; delivery?: string; effects: AppliedEffect[]; [field: string]: unknown };
export type CheckedSpell = NonNullable<AuthorResult['spells']>[number];

/** A request to the author command, typed as the contract's input. */
export function authorInput(fields: Record<string, unknown>): AuthorInput {
  return fields as unknown as AuthorInput;
}

const BLANK: SpellRow = { key: 'new_spell', label: 'New Spell', description: '', color: COLORS[0], level: 1, effects: [] };

@Component({
  imports: [Alert, Card, KeyValueList, NumberField, SelectField, TextField, TranslatePipe],
  selector: 'rw-spells',
  styleUrl: './spells.css',
  templateUrl: './spells.html',
})
export class Spells {
  private readonly api = inject(API);
  readonly toLoad = signal('');
  readonly draft = signal<SpellRow>(structuredClone(BLANK));
  readonly pool = signal<string[]>([]);
  readonly picked = signal('');
  readonly checked = signal<CheckedSpell | null>(null);
  readonly error = signal('');
  readonly colorOptions: readonly Option[] = COLORS.map((color) => ({ value: color, label: color }));
  readonly targetOptions: readonly Option[] = TARGETS.map((target) => ({ value: target, label: target }));
  readonly poolOptions = computed<readonly Option[]>(() => this.pool().map((ref) => ({ value: ref, label: ref })));
  readonly cost = computed(() => this.checked()?.cost ?? null);

  constructor() {
    toObservable(this.draft)
      .pipe(skip(1), debounceTime(CHECK_DELAY_MS), switchMap(() => from(this.check())), takeUntilDestroyed())
      .subscribe();
  }

  private async author(input: AuthorInput): Promise<AuthorResult | null> {
    this.error.set('');
    try {
      return await this.api.author(input);
    } catch (failure) {
      this.error.set(String(failure));
      return null;
    }
  }

  async load(key: string): Promise<void> {
    const result = await this.author(authorInput({ kind: SPELL_KIND, operation: AUTHOR.Load, spell: key.trim() }));
    if (!result?.row) return;
    this.draft.set(structuredClone(result.row) as unknown as SpellRow);
    await this.refreshPool();
  }

  async refreshPool(): Promise<void> {
    const result = await this.author(authorInput({ kind: SPELL_KIND, operation: AUTHOR.Effects, color: this.draft().color }));
    const refs = (result?.effects ?? []).map((entry) => entry.ref);
    this.pool.set(refs);
    this.picked.set(refs[0] ?? '');
  }

  async check(): Promise<void> {
    const result = await this.author(authorInput({ kind: SPELL_KIND, operation: AUTHOR.Check, spells: [this.draft()] }));
    this.checked.set(result?.spells?.[0] ?? null);
  }

  setField(field: keyof SpellRow, value: unknown): void {
    this.draft.update((row) => ({ ...row, [field]: value }));
    if (field === 'color') void this.refreshPool();
  }

  setEffect(index: number, field: keyof AppliedEffect, value: unknown): void {
    this.draft.update((row) => ({ ...row, effects: row.effects.map((effect, at) => (at === index ? { ...effect, [field]: value } : effect)) }));
  }

  addEffect(): void {
    const ref = this.picked();
    if (ref) this.draft.update((row) => ({ ...row, effects: [...row.effects, { effect: ref, magnitude: 1, target: TARGETS[2] }] }));
  }

  removeEffect(index: number): void {
    this.draft.update((row) => ({ ...row, effects: row.effects.filter((_, at) => at !== index) }));
  }

  exportInput(): AuthorInput {
    return authorInput({ kind: SPELL_KIND, operation: AUTHOR.Check, spells: [this.draft()], export: { shape: EXPORT_SHAPE.Core } });
  }

  async download(): Promise<void> {
    const result = await this.author(this.exportInput());
    const file = result?.export?.file;
    if (!file) return;
    const url = URL.createObjectURL(new Blob([JSON.stringify(file, null, 2) + '\n'], { type: 'application/json' }));
    const link = document.createElement('a');
    link.href = url;
    link.download = EXPORT_FILE;
    link.click();
    URL.revokeObjectURL(url);
  }
}
