// The app as a player uses it (docs/backend/frontend.md#tests): a browser, the dev proxy, and a real
// riftweave-server. Elements are found by label and role, so a restyle does not break a spec.
import { expect, test, type Page } from '@playwright/test';
import combat from '../../ruleset/tests/vectors/combat.json' with { type: 'json' };

/** The value cell of a key-value row, as rw-key-value-list prints it. */
function cell(page: Page, card: string, key: string) {
  return page.getByRole('region', { name: card }).getByRole('row', { name: new RegExp(`^${key}\\b`) }).getByRole('cell');
}

test('the sheet computes stats and creates a character', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('str', { exact: true }).fill('7');
  await page.getByRole('button', { name: 'Compute stats' }).click();
  await expect(cell(page, 'Derived stats', 'hp')).toHaveText(/^\d+$/);
  await page.getByRole('button', { name: 'Create character' }).click();
  await expect(page.getByRole('region', { name: 'Stored actor' })).toContainText('"str": 7');
});

test('an attack shows the result its vector expects', async ({ page }) => {
  const vector = combat[0] as { attacker: { skill: number; roll: number; attributes: { lck: number } };
    defender: { evasion_skill: number; armor_dr: number; hp: number }; damage: { roll: number; type: string };
    expect: Record<string, unknown> };
  await page.goto('/');
  await page.getByRole('button', { name: 'attack', exact: true }).click();
  await page.getByLabel('skill', { exact: true }).fill(String(vector.attacker.skill));
  await page.getByLabel('d100', { exact: true }).fill(String(vector.attacker.roll));
  await page.getByLabel('LCK', { exact: true }).fill(String(vector.attacker.attributes.lck));
  await page.getByLabel('evasion', { exact: true }).fill(String(vector.defender.evasion_skill));
  await page.getByLabel('armor DR', { exact: true }).fill(String(vector.defender.armor_dr));
  await page.getByLabel('HP', { exact: true }).fill(String(vector.defender.hp));
  await page.getByLabel('roll', { exact: true }).fill(String(vector.damage.roll));
  await page.getByLabel('type', { exact: true }).fill(vector.damage.type);
  await page.getByRole('button', { name: 'Resolve' }).click();
  for (const field of ['outcome', 'target', 'damage', 'hp']) {
    const value = vector.expect[field];
    await expect(cell(page, 'Result', field)).toHaveText(typeof value === 'string' ? value : JSON.stringify(value));
  }
});

test('the playthrough runs to a matching replay', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'playthrough', exact: true }).click();
  await page.getByRole('button', { name: 'Run the playthrough' }).click();
  await expect(page.getByRole('alert')).toHaveText('The replay matches the session.', { timeout: 15000 });
  await expect(page.getByText('differs from its vector')).toHaveCount(0);
});
