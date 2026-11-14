// The reusable components and the screens built from them (docs/backend/frontend.md#conventions):
// each component binds its value as a signal, and each screen sends the API the same input as before.
import { Component, signal, type Type } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it } from 'vitest';
import { API } from './api';
import { Alert } from './ui/alert';
import { Attack } from './attack';
import { Card } from './ui/card';
import { KeyValueList } from './ui/key-value-list';
import { NumberField } from './ui/number-field';
import { Playthrough } from './playthrough';
import { SelectField } from './ui/select-field';
import { Sheet } from './sheet';
import { TextField } from './ui/text-field';

function render<T>(host: new () => T) {
  const fixture = TestBed.createComponent(host);
  fixture.detectChanges();
  return { fixture, element: fixture.nativeElement as HTMLElement, host: fixture.componentInstance };
}

function type(input: HTMLInputElement | HTMLSelectElement, value: string, event = 'input'): void {
  input.value = value;
  input.dispatchEvent(new Event(event));
}

@Component({ imports: [NumberField, TextField, SelectField], template: `
  <rw-number-field label="STR" [min]="1" [(value)]="strength" />
  <rw-text-field label="Type" [(value)]="kind" />
  <rw-select-field label="Reaction" [options]="reactions" [(value)]="reaction" />` })
class Fields {
  readonly strength = signal(6);
  readonly kind = signal('slashing');
  readonly reaction = signal('');
  readonly reactions = [{ value: '', label: 'none' }, { value: 'evade', label: 'evade' }];
}

@Component({ imports: [Card, Alert, KeyValueList], template: `
  <rw-card title="Stats"><p class="inside">projected</p></rw-card>
  <rw-alert [message]="message()" />
  <rw-key-value-list [entries]="entries" />` })
class Blocks {
  readonly message = signal('');
  readonly entries = { hp: 59, mp: 46 };
}

describe('the field components', () => {
  it('label their input and write what is typed back to the bound signal', async () => {
    const { fixture, element, host } = render(Fields);
    const [number, text] = Array.from(element.querySelectorAll('input'));
    expect(element.textContent).toContain('STR');
    expect(number.value).toBe('6');
    type(number, '8');
    type(text, 'fire');
    type(element.querySelector('select')!, 'evade', 'change');
    await fixture.whenStable();
    expect(host.strength()).toBe(8);
    expect(host.kind()).toBe('fire');
    expect(host.reaction()).toBe('evade');
  });

  it('use the shared form-control class', () => {
    const { element } = render(Fields);
    for (const control of Array.from(element.querySelectorAll('input, select'))) {
      expect(control.classList).toContain('form-control');
    }
  });
});

describe('the block components', () => {
  it('project a card body under its title', () => {
    const { element } = render(Blocks);
    const card = element.querySelector('.card')!;
    expect(card.textContent).toContain('Stats');
    expect(card.querySelector('.inside')?.textContent).toBe('projected');
  });

  it('show an alert only with a message', async () => {
    const { fixture, element, host } = render(Blocks);
    expect(element.querySelector('.alert')).toBeNull();
    host.message.set('the server is down');
    fixture.detectChanges();
    await fixture.whenStable();
    expect(element.querySelector('.alert')?.textContent).toContain('the server is down');
  });

  it('list each entry as a row', () => {
    const { element } = render(Blocks);
    const rows = Array.from(element.querySelectorAll('tr')).map((row) => Array.from(row.children).map((cell) => cell.textContent?.trim()).join(' '));
    expect(rows).toEqual(['hp 59', 'mp 46']);
  });
});

const sent: { command: string; input: unknown }[] = [];
const fakeApi = {
  derived: async (actor: unknown) => { sent.push({ command: 'derived', input: actor }); return { hp: 67, mp: 42, sp: 10, ap: 4 }; },
  combat: async (input: unknown) => { sent.push({ command: 'combat', input }); return { outcome: 'hit', target: 72, damage: 6, hp: 24 }; },
};

beforeEach(() => {
  sent.length = 0;
  TestBed.configureTestingModule({ providers: [{ provide: API, useValue: fakeApi }] });
});

describe('the screens', () => {
  it('build their forms from the shared field components', () => {
    const screens: [Type<unknown>, number][] = [[Sheet, 9], [Attack, 11]];
    for (const [screen, fields] of screens) {
      const fixture = TestBed.createComponent(screen);
      fixture.detectChanges();
      const element = fixture.nativeElement as HTMLElement;
      expect(element.querySelectorAll('rw-number-field, rw-text-field, rw-select-field').length).toBe(fields);
      expect(element.querySelector('rw-card')).not.toBeNull();
    }
  });

  it('send the same combat input the attack form sent before', async () => {
    const fixture = TestBed.createComponent(Attack);
    fixture.detectChanges();
    await fixture.componentInstance.resolve();
    expect(sent).toEqual([{ command: 'combat', input: {
      attacker: { skill: 72, roll: 47, attributes: { lck: 0 } },
      defender: { evasion_skill: 25, armor_dr: 2, hp: 30 },
      damage: { roll: 8, type: 'slashing' },
      delivery: 'single',
    } }]);
  });

  it('send the sheet attributes to derived', async () => {
    const fixture = TestBed.createComponent(Sheet);
    fixture.detectChanges();
    await fixture.componentInstance.compute();
    expect(sent).toEqual([{ command: 'derived', input: { attributes: { str: 6, per: 5, end: 6, int: 5, wil: 4, agi: 5, cha: 4, lck: 5 }, level: 1 } }]);
  });

  it('render the playthrough runner', () => {
    const fixture = TestBed.createComponent(Playthrough);
    fixture.detectChanges();
    expect((fixture.nativeElement as HTMLElement).querySelector('rw-card button.btn')).not.toBeNull();
  });
});