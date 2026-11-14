// The attack resolver: every combat input is a field, every roll an input, and the result is shown
// in pipeline order (docs/core/combat.md).
import { Component, inject, signal } from '@angular/core';
import { API } from './app';
import type { CombatInput, CombatResult } from './api';

const STEPS = ['outcome', 'target', 'margin', 'crit', 'evaded', 'blocked', 'reaction_tags', 'damage', 'hp', 'counter'] as const;

@Component({
  selector: 'rw-attack',
  template: `
    <h2>Attack</h2>
    <fieldset><legend>Attacker</legend>
      <label>skill <input type="number" [value]="form.attackerSkill" (input)="form.attackerSkill = +$any($event.target).value" /></label>
      <label>d100 <input type="number" min="1" max="100" [value]="form.attackerRoll" (input)="form.attackerRoll = +$any($event.target).value" /></label>
      <button (click)="form.attackerRoll = d100()">roll</button>
      <label>LCK <input type="number" [value]="form.luck" (input)="form.luck = +$any($event.target).value" /></label>
    </fieldset>
    <fieldset><legend>Defender</legend>
      <label>evasion <input type="number" [value]="form.evasion" (input)="form.evasion = +$any($event.target).value" /></label>
      <label>armor DR <input type="number" [value]="form.armor" (input)="form.armor = +$any($event.target).value" /></label>
      <label>HP <input type="number" [value]="form.hp" (input)="form.hp = +$any($event.target).value" /></label>
      <label>reaction
        <select (change)="form.reaction = $any($event.target).value">
          <option value="">none</option><option value="evade">evade</option><option value="block">block</option>
        </select>
      </label>
      <label>reaction d100 <input type="number" [value]="form.reactionRoll" (input)="form.reactionRoll = +$any($event.target).value" /></label>
    </fieldset>
    <fieldset><legend>Damage</legend>
      <label>roll <input type="number" [value]="form.damageRoll" (input)="form.damageRoll = +$any($event.target).value" /></label>
      <label>type <input [value]="form.damageType" (input)="form.damageType = $any($event.target).value" /></label>
      <label>delivery
        <select (change)="form.delivery = $any($event.target).value">
          <option value="single">single</option><option value="area">area</option><option value="instant">instant</option>
        </select>
      </label>
    </fieldset>
    <button (click)="resolve()">Resolve</button>
    @if (error()) { <p class="error">{{ error() }}</p> }
    @if (result(); as shown) {
      <ol>@for (step of steps; track step) { @if (has(shown, step)) { <li><b>{{ step }}</b> {{ value(shown, step) }}</li> } }</ol>
    }
  `,
})
export class Attack {
  private readonly api = inject(API);
  readonly steps = STEPS;
  readonly form = { attackerSkill: 72, attackerRoll: 47, luck: 0, evasion: 25, armor: 2, hp: 30, reaction: '', reactionRoll: 50, damageRoll: 8, damageType: 'slashing', delivery: 'single' };
  readonly result = signal<CombatResult | null>(null);
  readonly error = signal('');

  d100(): number {
    return 1 + Math.floor(Math.random() * 100);
  }

  /** The `combat` input the form describes; a field the form leaves out is left out. */
  input(): CombatInput {
    const f = this.form;
    const defender: CombatInput['defender'] = { evasion_skill: f.evasion, armor_dr: f.armor, hp: f.hp };
    if (f.reaction) {
      defender.reaction = { kind: f.reaction as 'evade' | 'block', roll: f.reactionRoll };
    }
    return {
      attacker: { skill: f.attackerSkill, roll: f.attackerRoll, attributes: { lck: f.luck } },
      defender,
      damage: { roll: f.damageRoll, type: f.damageType },
      delivery: f.delivery as CombatInput['delivery'],
    } as CombatInput;
  }

  has(shown: CombatResult, step: string): boolean {
    return step in (shown as object);
  }

  value(shown: CombatResult, step: string): string {
    return JSON.stringify((shown as Record<string, unknown>)[step]);
  }

  async resolve(): Promise<void> {
    this.error.set('');
    try { this.result.set(await this.api.combat(this.input())); } catch (failure) { this.error.set(String(failure)); }
  }
}
