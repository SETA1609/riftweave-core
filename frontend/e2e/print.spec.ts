// Printing a character (FE_S09_04, docs/backend/print.md#printing-a-character): the Print action of
// the character list and the sheet opens the `print` command's page, every number the rules
// computed for the session, and `fresh` the same character at full resources with no condition.
import { readFile } from 'node:fs/promises';
import { expect, test, type Page } from '@playwright/test';

const fixture = '../ruleset/tests/fixtures/files/character-synced.riftweave.json';

/** The page a Print button opens. */
async function printed(page: Page, button: () => Promise<void>): Promise<Page> {
  const opening = page.waitForEvent('popup');
  await button();
  const printout = await opening;
  await printout.waitForLoadState();
  return printout;
}

test('the character list prints a wounded character as the session holds it, and fresh (FE_S09_04)', async ({ page }) => {
  const saved = JSON.parse(await readFile(fixture, 'utf8')) as { data: { resources: { hp: number }; skills: Record<string, number> } };
  await page.goto('/#/characters');
  await page.locator('input[type=file]').setInputFiles(fixture);
  const row = page.getByRole('row', { name: /Mira Vell/ });

  const current = await printed(page, () => row.getByRole('button', { name: 'Print', exact: true }).click());
  await expect(current.locator('[data-current="hp"]')).toHaveText(String(saved.data.resources.hp));
  await expect(current.locator('[data-stat="hp"]')).toHaveText('71');
  await expect(current.locator('[data-target="blades"]')).toHaveText(String(saved.data.skills['blades']));
  await expect(current.locator('[data-condition="burning"]')).toHaveCount(1);

  const fresh = await printed(page, () => row.getByRole('button', { name: 'Print fresh' }).click());
  await expect(fresh.locator('[data-current="hp"]')).toHaveText('71');
  await expect(fresh.locator('[data-condition]')).toHaveCount(0);
});

test('the sheet prints the character in play (FE_S09_04)', async ({ page }) => {
  await page.goto('/#/play/character_creation');
  await page.getByRole('button', { name: 'Create character' }).click();
  await expect(page.getByRole('region', { name: 'Stored actor' })).toBeVisible();
  const sheet = await printed(page, () => page.getByRole('button', { name: 'Print', exact: true }).click());
  const most = await sheet.locator('[data-stat="hp"]').textContent();
  await expect(sheet.locator('[data-current="hp"]')).toHaveText(most!);
});
