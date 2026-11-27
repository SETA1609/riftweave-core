// See bestiary.md.
import { KeyValuePipe } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { API, type BestiaryInput, type CatalogInput, type KnowledgeMark, type MonsterSheetInput, type MonsterSheetResult, type Session, type SessionStep } from '../api';
import { AppContext } from '../context';
import { Saves, rollD100 } from '../saves';
import { TranslatePipe } from '../translate-pipe';
import { Alert } from '../ui/alert';
import { Card } from '../ui/card';
import { KeyValueList } from '../ui/key-value-list';
import { NumberField } from '../ui/number-field';
import { SelectField, type Option } from '../ui/select-field';
import { TextField } from '../ui/text-field';

export const BESTIARY = { Reveal: 'reveal', View: 'view' } as const;
export const CATALOG = { List: 'list', Get: 'get' } as const;
export const LEVEL = { Unknown: 'unknown', Seen: 'seen', Known: 'known' } as const;
export const SHOWS = { Seen: 'seen', Full: 'full' } as const;
export const BESTIARY_COMMAND = 'bestiary';
export const MONSTERS = 'monsters';
export const LEVELS = [LEVEL.Unknown, LEVEL.Seen, LEVEL.Known] as const;
/** The viewer the page starts with: the GM, who sees every entry and every mark. */
export const GM_VIEWER = '';
/** The players' view of what the whole table was shown: a player with no mark of their own. */
export const TABLE_VIEWER = 'table';
export const SHEET_FIELDS = ['attributes', 'skills', 'level', 'role', 'perks', 'maneuvers'] as const;

export type Level = (typeof LEVELS)[number];
export type Entry = { monster: string; label: string; description: string; size: string; knowledge: Level; shows?: string; players?: Record<string, Level> };
/** The filter form; a level bound of 0 is no bound. */
export type Filter = { text: string; type: string; phase: string; levelMin: number; levelMax: number };
export type MonsterRow = Record<string, unknown> & { label?: string; level?: number; size?: string; phase?: string; tags?: string[] };
export type Opened = { entry: Entry; row: MonsterRow | null; sheet: MonsterSheetResult | null };

export const NO_FILTER: Filter = { text: '', type: '', phase: '', levelMin: 0, levelMax: 0 };

/** The catalog filters for the form: a creature type (the row's tags), a phase, a level range, and words. */
export function catalogFilters(filter: Filter): NonNullable<CatalogInput['filters']> {
  const filters: NonNullable<CatalogInput['filters']> = [];
  if (filter.type.trim()) filters.push({ field: 'tags', equals: filter.type.trim() });
  if (filter.phase.trim()) filters.push({ field: 'phase', equals: filter.phase.trim() });
  if (filter.levelMin > 0 || filter.levelMax > 0) {
    filters.push({ field: 'level', ...(filter.levelMin > 0 ? { min: filter.levelMin } : {}), ...(filter.levelMax > 0 ? { max: filter.levelMax } : {}) });
  }
  if (filter.text.trim()) filters.push({ text: filter.text.trim() });
  return filters;
}

/** A reveal of one monster at a level, for the table or one player. */
export function revealInput(monster: string, level: Level, player: string): BestiaryInput {
  return { operation: BESTIARY.Reveal, monster, level, ...(player ? { player } : {}) };
}

/** The monster_sheet input for a resolved monster row: the fields the sheet reads, as the row has them. */
export function sheetInput(row: MonsterRow): MonsterSheetInput {
  const monster: Record<string, unknown> = {};
  for (const field of SHEET_FIELDS) {
    if (row[field] !== undefined) monster[field] = row[field];
  }
  return { monster } as MonsterSheetInput;
}

/** A bestiary input as a step of the session's log. */
export function bestiaryStep(input: BestiaryInput): SessionStep {
  return { command: BESTIARY_COMMAND, input: input as unknown as SessionStep['input'] };
}

/** The players named in the marks, for the viewer list. */
export function playersOf(knowledge: readonly KnowledgeMark[]): string[] {
  return [...new Set(knowledge.map((mark) => mark.player).filter((player): player is string => !!player && player !== TABLE_VIEWER))].sort();
}

@Component({
  imports: [Alert, Card, KeyValueList, KeyValuePipe, NumberField, SelectField, TextField, TranslatePipe],
  selector: 'rw-bestiary',
  styleUrl: './bestiary.css',
  templateUrl: './bestiary.html',
})
export class Bestiary {
  private readonly api = inject(API);
  private readonly saves = inject(Saves);
  readonly context = inject(AppContext);

  readonly filter = signal<Filter>({ ...NO_FILTER });
  readonly rows = signal<Map<string, MonsterRow>>(new Map());
  readonly entries = signal<Entry[]>([]);
  readonly viewer = signal(GM_VIEWER);
  readonly player = signal('');
  readonly opened = signal<Opened | null>(null);
  readonly lore = signal<{ monster: string; target: number; success: boolean } | null>(null);
  readonly error = signal('');
  /** The marks when no game is in play; with one, the marks live in its session. */
  private readonly unsaved = signal<KnowledgeMark[]>([]);

  readonly session = computed<Session | null>(() => this.saves.current()?.session ?? null);
  readonly knowledge = computed<KnowledgeMark[]>(() => this.session()?.bestiary?.knowledge ?? this.unsaved());
  readonly isGm = computed(() => this.viewer() === GM_VIEWER);
  readonly levelOptions: readonly Option[] = LEVELS.map((level) => ({ value: level, label: level }));
  readonly viewerOptions = computed<readonly Option[]>(() => [
    { value: GM_VIEWER, label: 'GM' },
    { value: TABLE_VIEWER, label: TABLE_VIEWER },
    ...playersOf(this.knowledge()).map((player) => ({ value: player, label: player })),
  ]);

  /** The first listing, run when the page opens. */
  readonly ready: Promise<void> = this.list();

  private async call<T>(run: () => Promise<T>): Promise<T | null> {
    this.error.set('');
    try {
      return await run();
    } catch (failure) {
      this.error.set(String(failure));
      return null;
    }
  }

  /** List the monsters that match the filter, then view them as the viewer sees them. */
  async list(): Promise<void> {
    const filters = catalogFilters(this.filter());
    const listed = await this.call(() => this.api.catalog({ operation: CATALOG.List, collection: MONSTERS, ...(filters.length ? { filters } : {}) }));
    if (!listed) return;
    this.rows.set(new Map((listed.rows ?? []).map((row) => [row.ref, row.row as MonsterRow])));
    await this.view();
  }

  /** The listed monsters as the viewer sees them, through the bestiary command's view. */
  async view(): Promise<void> {
    const viewer = this.viewer();
    const viewed = await this.call(() =>
      this.api.bestiary({
        operation: BESTIARY.View,
        knowledge: this.knowledge(),
        monsters: [...this.rows().keys()],
        players_see_only_known: this.context.playersSeeOnlyKnown(),
        ...(viewer ? { player: viewer } : {}),
      }),
    );
    if (viewed) this.entries.set((viewed.entries ?? []) as Entry[]);
  }

  setFilter(field: keyof Filter, value: string | number): void {
    this.filter.update((filter) => ({ ...filter, [field]: value }));
  }

  async setViewer(viewer: string): Promise<void> {
    this.viewer.set(viewer);
    this.opened.set(null);
    await this.view();
  }

  async setPlayersSeeOnlyKnown(known: boolean): Promise<void> {
    const problem = this.context.setPlayersSeeOnlyKnown(known);
    await this.view();
    if (problem) this.error.set(problem);
  }

  /** Mark a monster for the table, or for the player named in the form: a bestiary step in the game in play, so it replays. */
  async mark(monster: string, level: Level): Promise<void> {
    const input = revealInput(monster, level, this.player().trim());
    const session = this.session();
    if (session) {
      const answer = await this.call(() => this.api.step(session, bestiaryStep(input)));
      if (answer) await this.saves.record(answer.session);
    } else {
      const revealed = await this.call(() => this.api.bestiary({ ...input, knowledge: this.unsaved() }));
      if (revealed?.knowledge) this.unsaved.set(revealed.knowledge);
    }
    await this.view();
  }

  /** A Lore check by the player named in the form, with the game's character, that reveals the monster to them on a success. */
  async loreCheck(monster: string, roll: number = rollD100()): Promise<void> {
    const session = this.session();
    const player = this.player().trim();
    if (!session || !player) return;
    const { attributes, skills, level } = session.actor as { attributes?: unknown; skills?: unknown; level?: number };
    const input = { operation: BESTIARY.Reveal, monster, player, check: { actor: { attributes, skills, level }, roll } } as BestiaryInput;
    const answer = await this.call(() => this.api.step(session, bestiaryStep(input)));
    if (!answer) return;
    await this.saves.record(answer.session);
    const check = (answer.result as { check?: { target: number; success: boolean } } | undefined)?.check;
    if (check) this.lore.set({ monster, target: check.target, success: check.success });
    await this.view();
  }

  /** An entry: its row and its sheet when it shows in full, its name, description, and size when only seen. */
  async open(entry: Entry): Promise<void> {
    if (entry.shows !== SHOWS.Full) {
      this.opened.set({ entry, row: null, sheet: null });
      return;
    }
    const got = await this.call(() => this.api.catalog({ operation: CATALOG.Get, collection: MONSTERS, ref: entry.monster }));
    const row = (got?.row?.row ?? this.rows().get(entry.monster) ?? null) as MonsterRow | null;
    const sheet = row ? await this.call(() => this.api.monsterSheet(sheetInput(row))) : null;
    this.opened.set({ entry, row, sheet });
  }

  /** What a list row shows of a field: the row's value when the entry shows in full, else a question mark. */
  shown(entry: Entry, field: 'level' | 'phase' | 'type'): string {
    if (entry.shows !== SHOWS.Full) return '?';
    const row = this.rows().get(entry.monster);
    const value = field === 'type' ? row?.tags?.[0] : row?.[field];
    return value === undefined || value === null ? '' : String(value);
  }
}
