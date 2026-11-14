// The attack resolver: every combat input is a field, every roll an input, and the result is shown
// in pipeline order (docs/core/combat.md).
import { Component, computed, inject, signal } from '@angular/core';
import { API, type CombatInput, type CombatResult } from './api';
import { Alert } from './ui/alert';
import { Card } from './ui/card';
import { KeyValueList } from './ui/key-value-list';
import { NumberField } from './ui/number-field';
import type { Option } from './ui/select-field';
import { SelectField } from './ui/select-field';
import { TextField } from './ui/text-field';

const STEPS = ['outcome', 'target', 'margin', 'crit', 'evaded', 'blocked', 'reaction_tags', 'damage', 'hp', 'counter'] as const;

@Component({
  selector: 'rw-attack',
  imports: [Alert, Card, KeyValueList, NumberField, SelectField, TextField],
  templateUrl: './attack.html',
  styleUrl: './attack.css',
})
export class Attack {
  private readonly api = inject(API);
  readonly reactions: readonly Option[] = [{ value: '', label: 'none' }, { value: 'evade', label: 'evade' }, { value: 'block', label: 'block' }];
  readonly deliveries: readonly Option[] = [{ value: 'single', label: 'single' }, { value: 'area', label: 'area' }, { value: 'instant', label: 'instant' }];

  readonly attackerSkill = signal(72);
  readonly attackerRoll = signal(47);
  readonly luck = signal(0);
  readonly evasion = signal(25);
  readonly armor = signal(2);
  readonly hp = signal(30);
  readonly reaction = signal('');
  readonly reactionRoll = signal(50);
  readonly damageRoll = signal(8);
  readonly damageType = signal('slashing');
  readonly delivery = signal('single');

  readonly result = signal<CombatResult | null>(null);
  readonly error = signal('');
  /** The result's fields in pipeline order, the ones it has. */
  readonly steps = computed(() => {
    const shown = this.result() as Record<string, unknown> | null;
    return shown ? Object.fromEntries(STEPS.filter((step) => step in shown).map((step) => [step, shown[step]])) : null;
  });

  rollAttack(): void {
    this.attackerRoll.set(1 + Math.floor(Math.random() * 100));
  }

  /** The `combat` input the form describes; a field the form leaves out is left out. */
  input(): CombatInput {
    const defender: CombatInput['defender'] = { evasion_skill: this.evasion(), armor_dr: this.armor(), hp: this.hp() };
    if (this.reaction()) {
      defender.reaction = { kind: this.reaction() as 'evade' | 'block', roll: this.reactionRoll() };
    }
    return {
      attacker: { skill: this.attackerSkill(), roll: this.attackerRoll(), attributes: { lck: this.luck() } },
      defender,
      damage: { roll: this.damageRoll(), type: this.damageType() },
      delivery: this.delivery() as CombatInput['delivery'],
    } as CombatInput;
  }

  async resolve(): Promise<void> {
    this.error.set('');
    try {
      this.result.set(await this.api.combat(this.input()));
    } catch (failure) {
      this.error.set(String(failure));
    }
  }
}
