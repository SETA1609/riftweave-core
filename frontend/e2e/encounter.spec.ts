// The encounter tracker (FE_S08_08, docs/backend/sessions.md#encounters): a GM runs a fight of four
// characters and two monsters offline, and the page's initiative order and turns are the ones the
// encounter command gives in Go for the same combatants.
import { expect, test, type Page } from '@playwright/test';

const API = 'http://127.0.0.1:8080/v1/commands/encounter';

/** Add someone at the table by the initiative, AGI, and LCK on their sheet. */
async function atTable(page: Page, name: string, initiative: number, agi: number) {
  await page.getByLabel('name', { exact: true }).fill(name);
  await page.getByLabel('initiative', { exact: true }).fill(String(initiative));
  await page.getByLabel('AGI', { exact: true }).fill(String(agi));
  await page.getByRole('button', { name: 'Add at the table' }).click();
  await expect(page.locator(`tbody tr[data-combatant="${name}"]`)).toBeVisible();
}

async function addMonster(page: Page, key: string) {
  const before = await page.locator('tbody tr[data-combatant]').count();
  await page.getByLabel('monster key').fill(key);
  await page.getByRole('button', { name: 'Add monster' }).click();
  await expect(page.locator('tbody tr[data-combatant]')).toHaveCount(before + 1);
}

/** The combatants as the page lists them: id, initiative, AGI, LCK. */
async function combatants(page: Page) {
  const rows = await page.locator('tbody tr[data-combatant]').all();
  return Promise.all(rows.map(async (row) => {
    const cells = await row.locator('td').allTextContents();
    return { id: cells[0].trim(), initiative: Number(cells[1]), agi: Number(cells[2]), lck: Number(cells[3]) };
  }));
}

const order = (page: Page) => page.locator('[data-order]').allTextContents().then((ids) => ids.map((id) => id.trim()));

test('a GM runs four characters and two monsters, in the order and turns the encounter command gives', async ({ page, request }) => {
  await page.goto('/#/play/encounter_tracker');
  await atTable(page, 'Ana', 14, 6);
  await atTable(page, 'Bo', 11, 5);
  await atTable(page, 'Cyr', 9, 7);
  await atTable(page, 'Dee', 4, 4);
  await addMonster(page, 'skeleton');
  await addMonster(page, 'wolf');
  await expect(page.locator('tbody tr[data-combatant]')).toHaveCount(6);

  const listed = await combatants(page);
  const go = await (await request.post(API, { data: { operation: 'start', combatants: listed } })).json();

  await page.getByRole('button', { name: 'Start the fight' }).click();
  await expect(page.locator('[data-round]')).toHaveText('Round 1');
  expect(await order(page)).toEqual(go.encounter.order);
  await expect(page.locator('[data-current]')).toContainText(`Turn: ${go.current.id}`);
  await expect(page.locator('[data-current]')).toContainText('first turn');

  // A whole round of turns, and the first of the next, as turn_end gives them.
  let state = go;
  for (let turn = 0; turn < 7; turn++) {
    await page.getByRole('button', { name: 'End turn' }).click();
    state = await (await request.post(API, { data: { operation: 'turn_end', encounter: state.encounter } })).json();
    await expect(page.locator('[data-current]')).toContainText(`Turn: ${state.current.id}`);
    await expect(page.locator('[data-next]')).toHaveText(`Next: ${state.next}`);
  }
  await expect(page.locator('[data-round]')).toHaveText(`Round ${state.encounter.round}`);
  await expect(page.locator('[data-current]')).not.toContainText('first turn');
});

test('an ambusher acts alone in the surprise round', async ({ page }) => {
  await page.goto('/#/play/encounter_tracker');
  await atTable(page, 'Ana', 14, 6);
  await addMonster(page, 'wolf');
  await page.getByLabel('ambusher Wolf').check();
  await page.getByRole('button', { name: 'Start the fight' }).click();
  await expect(page.locator('[data-round]')).toHaveText('Surprise round');
  await expect(page.locator('[data-current]')).toContainText('Turn: Wolf');
  await expect(page.locator('[data-current]')).toContainText('only ambushers act');
  await page.getByRole('button', { name: 'End turn' }).click();
  await expect(page.locator('[data-round]')).toHaveText('Round 1');
  await expect(page.locator('[data-current]')).toContainText('Turn: Ana');
});
