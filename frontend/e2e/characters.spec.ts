// The character list in the browser (FE_S07_04): a player imports a character file, exports it for
// the GM, takes the GM's higher revision back, and sees a lower or a diverged one in the page.
import { readFile } from 'node:fs/promises';
import { expect, test, type Page } from '@playwright/test';

const fixture = '../ruleset/tests/fixtures/files/character-synced.riftweave.json';

async function importText(page: Page, name: string, text: string): Promise<void> {
  await page.locator('input[type=file]').setInputFiles({ name, mimeType: 'application/json', buffer: Buffer.from(text) });
}

/** The fixture with its sync fields changed. */
async function edited(change: Record<string, unknown>): Promise<string> {
  const file = JSON.parse(await readFile(fixture, 'utf8'));
  Object.assign(file.character, change);
  return JSON.stringify(file, null, 2) + '\n';
}

test('the character list imports, exports, and follows the sync rule (FE_S07_04)', async ({ page }) => {
  await page.goto('/#/');
  await page.getByRole('link', { name: 'Characters' }).click();
  await expect(page.getByRole('heading', { name: 'Characters' })).toBeVisible();

  await page.locator('input[type=file]').setInputFiles(fixture);
  const row = page.getByRole('row', { name: /Mira Vell/ });
  await expect(row).toContainText('revision 3');

  const downloading = page.waitForEvent('download');
  await row.getByRole('button', { name: 'Export' }).click();
  const file = await downloading;
  expect(file.suggestedFilename()).toBe('mira-vell-5f0c7a52.riftweave.json');
  expect(await readFile(await file.path(), 'utf8')).toBe(await readFile(fixture, 'utf8'));

  await importText(page, 'from-the-gm.riftweave.json', await edited({ revision: 4, changed_by: 'gm' }));
  await expect(page.getByRole('row', { name: /Mira Vell/ })).toContainText('revision 4');

  await importText(page, 'old.riftweave.json', await edited({ revision: 2 }));
  await expect(page.getByText('revision 2 is older than the saved revision 4; nothing changed')).toBeVisible();

  await importText(page, 'other.riftweave.json', await edited({ revision: 4, name: 'Mira the Bold' }));
  await expect(page.getByText('both sides changed revision 4; keep the saved or the incoming one')).toBeVisible();
  await page.getByRole('button', { name: 'Keep the incoming one' }).click();
  await expect(page.getByRole('row', { name: /Mira the Bold/ })).toBeVisible();
});
