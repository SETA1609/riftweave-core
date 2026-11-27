// See encounter.md.
import { Component, computed, inject, signal } from '@angular/core';
import { API, type EncounterInput, type EncounterResult, type MonsterSheetInput, type Session, type SessionStep, type StoredActor } from '../api';
import { Characters } from '../characters';
import { Saves, rollD100 } from '../saves';
import { TranslatePipe } from '../translate-pipe';
import { Alert } from '../ui/alert';
import { Card } from '../ui/card';
import { NumberField } from '../ui/number-field';
import { SelectField, type Option } from '../ui/select-field';
import { TextField } from '../ui/text-field';

export const ENCOUNTER = { Start: 'start', TurnEnd: 'turn_end' } as const;
export const ENCOUNTER_COMMAND = 'encounter';
export const TURN_START = { Combat: 'combat', Surprise: 'surprise' } as const;
export const SOURCE = { Character: 'character', Monster: 'monster', Table: 'table' } as const;
export const MONSTER_PREFIX = 'core:monster/';
export const SHEET_FIELDS = ['attributes', 'skills', 'level', 'role', 'perks', 'maneuvers'] as const;
/** How many times a tie the roll-off did not break is rolled again before the page gives up. */
export const ROLL_OFF_TRIES = 5;
/** The refusal of a tie the rolls do not break (encounter.Start in the core). */
export const TIE_REFUSAL = 'roll off again';

export type Source = (typeof SOURCE)[keyof typeof SOURCE];
export type Combatant = { id: string; source: Source; initiative: number; agi: number; lck: number; surprise: boolean; hp: number; sp: number; mp: number; rolls: number[] };
export type Fight = EncounterResult;
type Pools = { initiative?: number; hp?: number; sp?: number; mp?: number };

/** An id not yet in the fight: the name, or the name and the next free number. */
export function uniqueId(name: string, taken: readonly string[]): string {
  const base = name.trim() || 'combatant';
  if (!taken.includes(base)) return base;
  let number = 2;
  while (taken.includes(`${base} ${number}`)) number++;
  return `${base} ${number}`;
}

/** A monster ref from what the GM typed: a key, or a full ref as it is. */
export function monsterRef(typed: string): string {
  const trimmed = typed.trim();
  return trimmed.includes(':') ? trimmed : `${MONSTER_PREFIX}${trimmed}`;
}

/** The encounter's start: every combatant's id, initiative, AGI, LCK, roll-off, and whether it ambushes. */
export function startInput(combatants: readonly Combatant[]): EncounterInput {
  return {
    operation: ENCOUNTER.Start,
    combatants: combatants.map(({ id, initiative, agi, lck, rolls, surprise }) => ({ id, initiative, agi, lck, rolls, ...(surprise ? { surprise } : {}) })),
  };
}

/** An encounter input as a step of the session's log. */
export function encounterStep(input: EncounterInput): SessionStep {
  return { command: ENCOUNTER_COMMAND, input: input as unknown as SessionStep['input'] };
}

function attribute(attributes: unknown, key: string): number {
  return Number((attributes as Record<string, number> | undefined)?.[key] ?? 0);
}

@Component({
  imports: [Alert, Card, NumberField, SelectField, TextField, TranslatePipe],
  selector: 'rw-encounter',
  styleUrl: './encounter.css',
  templateUrl: './encounter.html',
})
export class Encounter {
  private readonly api = inject(API);
  private readonly saves = inject(Saves);
  readonly characters = inject(Characters);

  readonly combatants = signal<Combatant[]>([]);
  readonly fight = signal<Fight | null>(null);
  readonly error = signal('');
  readonly pickedCharacter = signal('');
  readonly monster = signal('');
  readonly name = signal('');
  readonly initiative = signal(0);
  readonly agi = signal(5);
  readonly lck = signal(5);

  readonly session = computed<Session | null>(() => this.saves.current()?.session ?? null);
  readonly started = computed(() => this.fight() !== null);
  readonly characterOptions = computed<readonly Option[]>(() => this.characters.list().map((saved) => ({ value: saved.character.id, label: saved.character.name })));
  /** The fight's order, each with its combatant. */
  readonly order = computed(() => {
    const fight = this.fight();
    if (!fight) return [];
    const byId = new Map(this.combatants().map((combatant) => [combatant.id, combatant]));
    return fight.encounter.order.map((id, index) => ({ id, index, combatant: byId.get(id) ?? null, acted: fight.encounter.acted.includes(id), current: index === fight.encounter.turn }));
  });
  /** The first turn's partial AP pool, for the actions command (docs/core/actions.md#pool). */
  readonly turnStart = computed(() => this.fight()?.current.start ?? null);

  /** The first listing of the saved characters, run when the page opens. */
  readonly ready: Promise<void> = this.characters.load().catch(() => undefined).then(() => this.pickedCharacter.set(this.characterOptions()[0]?.value ?? ''));

  private async call<T>(run: () => Promise<T>): Promise<T | null> {
    this.error.set('');
    try {
      return await run();
    } catch (failure) {
      this.error.set(String(failure));
      return null;
    }
  }

  private add(name: string, source: Source, values: Omit<Combatant, 'id' | 'source' | 'surprise' | 'rolls'>): void {
    this.combatants.update((all) => [...all, { id: uniqueId(name, all.map((one) => one.id)), source, surprise: false, rolls: [], ...values }]);
  }

  /** A saved character, with its initiative and pools from the derived command and its current resources. */
  async addCharacter(id: string = this.pickedCharacter()): Promise<void> {
    const saved = this.characters.list().find((one) => one.character.id === id);
    if (!saved) return;
    await this.addActor(saved.character.name, saved.data);
  }

  /** The game's character, as its session holds it. */
  async addGameCharacter(name: string): Promise<void> {
    const actor = this.session()?.actor as StoredActor | undefined;
    if (actor) await this.addActor(name, actor);
  }

  private async addActor(name: string, actor: StoredActor): Promise<void> {
    const derived = (await this.call(() => this.api.derived({ attributes: actor.attributes, level: actor.level } as never))) as Pools | null;
    if (!derived) return;
    const resources = (actor as { resources?: Pools }).resources ?? {};
    this.add(name, SOURCE.Character, {
      initiative: derived.initiative ?? 0,
      agi: attribute(actor.attributes, 'agi'),
      lck: attribute(actor.attributes, 'lck'),
      hp: resources.hp ?? derived.hp ?? 0,
      sp: resources.sp ?? derived.sp ?? 0,
      mp: resources.mp ?? derived.mp ?? 0,
    });
  }

  /** A monster by key or ref, with its initiative and pools from monster_sheet. */
  async addMonster(typed: string = this.monster()): Promise<void> {
    if (!typed.trim()) return;
    const got = await this.call(() => this.api.catalog({ operation: 'get', collection: 'monsters', ref: monsterRef(typed) }));
    const row = got?.row?.row as Record<string, unknown> | undefined;
    if (!row) return;
    const monster: Record<string, unknown> = {};
    for (const field of SHEET_FIELDS) if (row[field] !== undefined) monster[field] = row[field];
    const sheet = (await this.call(() => this.api.monsterSheet({ monster } as MonsterSheetInput))) as Pools | null;
    if (!sheet) return;
    this.add(String(row['label'] ?? typed), SOURCE.Monster, {
      initiative: sheet.initiative ?? 0,
      agi: attribute(row['attributes'], 'agi'),
      lck: attribute(row['attributes'], 'lck'),
      hp: sheet.hp ?? 0,
      sp: sheet.sp ?? 0,
      mp: sheet.mp ?? 0,
    });
    if (this.monster() === typed) this.monster.set('');
  }

  /** Anyone else at the table, by the initiative, AGI, and LCK the GM reads off their sheet. */
  addAtTable(): void {
    this.add(this.name(), SOURCE.Table, { initiative: this.initiative(), agi: this.agi(), lck: this.lck(), hp: 0, sp: 0, mp: 0 });
    this.name.set('');
  }

  remove(id: string): void {
    this.combatants.update((all) => all.filter((one) => one.id !== id));
  }

  update(id: string, field: 'surprise' | 'hp' | 'sp' | 'mp', value: boolean | number): void {
    this.combatants.update((all) => all.map((one) => (one.id === id ? { ...one, [field]: value } : one)));
  }

  /** Start the fight: one d100 each for the roll-off, rolled again while a tie stands, kept in the step so it replays. */
  async start(roll: () => number = rollD100): Promise<void> {
    let combatants = this.combatants().map((one) => ({ ...one, rolls: [roll()] }));
    for (let attempt = 0; attempt < ROLL_OFF_TRIES; attempt++) {
      this.combatants.set(combatants);
      const fight = await this.run(startInput(combatants));
      if (fight) {
        this.fight.set(fight);
        return;
      }
      if (!this.error().includes(TIE_REFUSAL)) return;
      combatants = combatants.map((one) => ({ ...one, rolls: [...one.rolls, roll()] }));
    }
  }

  /** End the current turn: the next actor, or the next round. */
  async endTurn(): Promise<void> {
    const fight = this.fight();
    if (!fight) return;
    const next = await this.run({ operation: ENCOUNTER.TurnEnd, encounter: fight.encounter });
    if (next) this.fight.set(next);
  }

  /** A new fight with the same combatants. */
  reset(): void {
    this.fight.set(null);
  }

  /** One encounter command: a step of the game in play, autosaved (the session fills the state of a turn_end), or the command alone. */
  private async run(input: EncounterInput): Promise<Fight | null> {
    const session = this.session();
    if (!session) return this.call(() => this.api.encounter(input));
    const sent = input.operation === ENCOUNTER.TurnEnd ? { operation: ENCOUNTER.TurnEnd } : input;
    const answer = await this.call(() => this.api.step(session, encounterStep(sent as EncounterInput)));
    if (!answer) return null;
    await this.saves.record(answer.session);
    return answer.result as unknown as Fight;
  }
}
