// The bestiary (FE_S08_06, docs/backend/bestiary.md): the monsters of core listed and filtered over
// the `catalog` command, and what the players know of each over the `bestiary` command. With
// "Players see only what they know" off, every entry shows in full; on, a player sees only what
// the GM marked, and the GM's view always shows every entry with its mark.
import { expect, test, type Page } from '@playwright/test';

/** The list's row for a monster, by its label. */
function row(page: Page, label: string) {
  return page.locator('tbody tr[data-monster]').filter({ has: page.getByRole('button', { name: label, exact: true }) });
}

test('the bestiary lists every monster, filters by creature type, and opens a full entry with its sheet', async ({ page }) => {
  await page.goto('/#/play/bestiary');
  await expect(row(page, 'Skeleton')).toBeVisible();
  await expect(row(page, 'Bandit')).toBeVisible();

  await page.getByLabel('creature type').fill('undead');
  await page.getByRole('button', { name: 'List', exact: true }).click();
  await expect(row(page, 'Bandit')).toHaveCount(0);
  await expect(row(page, 'Zombie')).toBeVisible();

  await row(page, 'Skeleton').getByRole('button', { name: 'Skeleton', exact: true }).click();
  const entry = page.getByRole('region', { name: 'Skeleton' });
  await expect(entry.getByRole('heading', { name: 'Sheet' })).toBeVisible();
  await expect(entry.getByRole('rowheader', { name: 'hp', exact: true })).toBeVisible();
});

test('with the box on, a player sees only what the GM marked, and a seen monster hides its stats', async ({ page }) => {
  await page.goto('/#/play/bestiary');
  await expect(row(page, 'Skeleton')).toBeVisible();
  await row(page, 'Skeleton').getByRole('button', { name: 'Mark seen: Skeleton' }).click();
  await expect(row(page, 'Skeleton').locator('.mark')).toHaveText('seen');

  // Off (the default): the players see every entry in full.
  await page.getByLabel('view as').selectOption('table');
  await expect(row(page, 'Bandit')).toBeVisible();

  // On: only the marked monster, as a seen one, with no marking controls.
  await page.getByLabel('Players see only the monsters they know').check();
  await expect(row(page, 'Bandit')).toHaveCount(0);
  await expect(row(page, 'Skeleton')).toBeVisible();
  await expect(row(page, 'Skeleton').locator('td').nth(1)).toHaveText('?');
  await expect(page.getByRole('columnheader', { name: 'set mark' })).toHaveCount(0);

  // The GM's view still shows every entry, with its mark.
  await page.getByLabel('view as').selectOption('');
  await expect(row(page, 'Bandit')).toBeVisible();
  await expect(row(page, 'Bandit').locator('.mark')).toHaveText('unknown');
});
