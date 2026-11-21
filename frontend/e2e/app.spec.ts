// The app as a player uses it (docs/backend/frontend.md#tests): a browser, the dev proxy, and a real
// riftweave-server. Elements are found by label and role, so a restyle does not break a spec.
import { readFile } from 'node:fs/promises';
import { expect, test, type Page } from '@playwright/test';
import combat from '../../ruleset/tests/vectors/combat.json' with { type: 'json' };

/** The value cell of a key-value row, as rw-key-value-list prints it. */
function cell(page: Page, card: string, key: string) {
  return page.getByRole('region', { name: card }).getByRole('row', { name: new RegExp(`^${key}\\b`) }).getByRole('cell');
}

test('the sheet computes stats and creates a character', async ({ page }) => {
  await page.goto('/#/play/character_creation');
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
  await page.goto('/#/play/attack');
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
  await page.goto('/#/play/playthrough');
  await page.getByRole('button', { name: 'Run the playthrough' }).click();
  await expect(page.getByRole('alert')).toHaveText('The replay matches the session.', { timeout: 15000 });
  await expect(page.getByText('differs from its vector')).toHaveCount(0);
});

test('the mode toggle switches light and dark, keeps the choice, and goes back to the system', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'light' });
  await page.goto('/');
  const html = page.locator('html');
  await expect(html).not.toHaveAttribute('data-theme');
  await page.getByRole('button', { name: 'Switch to light mode' }).click();
  await expect(html).toHaveAttribute('data-theme', 'light');
  await page.getByRole('button', { name: 'Switch to dark mode' }).click();
  await expect(html).toHaveAttribute('data-theme', 'dark');
  await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(33, 37, 41)');
  await page.reload();
  await expect(html).toHaveAttribute('data-theme', 'dark');
  await page.getByRole('button', { name: "Switch to the system's mode" }).click();
  await expect(html).not.toHaveAttribute('data-theme');
  await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(255, 255, 255)');
});

test('the language select switches every string without a reload and keeps the choice (FE_S11_01)', async ({ page }) => {
  await page.goto('/#/settings');
  await expect(page.locator('html')).toHaveAttribute('lang', 'en');
  await page.getByLabel('Language').selectOption('de');
  await expect(page.locator('html')).toHaveAttribute('lang', 'de');
  await expect(page.getByRole('region', { name: 'Einstellungen' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Einstellungen exportieren' })).toBeVisible();
  await expect(page).toHaveTitle('Einstellungen');
  await page.reload();
  await expect(page.getByLabel('Sprache')).toHaveValue('de');
  await page.getByLabel('Sprache').selectOption('es');
  await page.goto('/#/play');
  await expect(page.getByRole('link', { name: 'Creación de personaje' })).toBeVisible();
  await expect(page.getByText('Reglas', { exact: false })).toBeVisible();
});

test("with storage blocked the app still renders, in the system's mode", async ({ page }) => {
  await page.addInitScript(() => {
    Object.defineProperty(window, 'localStorage', { get() { throw new Error('storage is blocked'); } });
  });
  await page.emulateMedia({ colorScheme: 'dark' });
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Riftweave' })).toBeVisible();
  await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(33, 37, 41)');
  await page.getByRole('button', { name: 'Switch to light mode' }).click();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
});

test('a playthrough is saved as it runs, and a reload resumes it', async ({ page }) => {
  await page.goto('/#/play/playthrough');
  await page.getByRole('button', { name: 'Run the playthrough' }).click();
  await expect(page.getByText('The replay matches the session.')).toBeVisible();
  await page.reload();
  await page.goto('/#/play/saved_games');
  await expect(page.getByRole('row', { name: /^Playthrough \(\d+ steps\)/ })).toBeVisible();
});

/** Create a character, then download its backup; returns the file's JSON. */
async function backup(page: Page): Promise<Record<string, unknown>> {
  await page.goto('/#/play/character_creation');
  await page.getByRole('button', { name: 'Create character' }).click();
  await expect(page.getByRole('region', { name: 'Stored actor' })).toBeVisible();
  await page.goto('/#/play/saved_games');
  const download = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Download backup' }).click();
  const file = await download;
  expect(file.suggestedFilename()).toMatch(/\.riftweave\.json$/);
  const chunks: Buffer[] = [];
  for await (const chunk of await file.createReadStream()) {
    chunks.push(chunk as Buffer);
  }
  return JSON.parse(Buffer.concat(chunks).toString('utf8')) as Record<string, unknown>;
}

/** Open a file in a fresh browser (no saves) and return its saved-games page. */
async function openFresh(browser: import('@playwright/test').Browser, baseURL: string, name: string, content: unknown): Promise<Page> {
  const context = await browser.newContext({ baseURL });
  const page = await context.newPage();
  await page.goto('/#/play/saved_games');
  await page.locator('input[type=file]').setInputFiles({ name, mimeType: 'application/json', buffer: Buffer.from(JSON.stringify(content)) });
  return page;
}

test('a backup opens in a fresh browser to the same character', async ({ page, browser, baseURL }) => {
  const saved = await backup(page);
  expect(Object.keys(saved)).toEqual(expect.arrayContaining(['rules_version', 'ruleset', 'start', 'log', 'actor']));
  const fresh = await openFresh(browser, baseURL!, 'mine.riftweave.json', saved);
  await expect(fresh.getByText('The file replays to the same character.')).toBeVisible();
  await fresh.getByRole('button', { name: 'Keep the replayed game' }).click();
  await expect(fresh.getByRole('row', { name: /^mine \(0 steps\)/ })).toBeVisible();
});

test('an edited backup is flagged, and one from other rules says so', async ({ page, browser, baseURL }) => {
  const saved = await backup(page);
  const edited = { ...saved, actor: { ...(saved['actor'] as object), xp: 99_999 } };
  const flagged = await openFresh(browser, baseURL!, 'edited.riftweave.json', edited);
  await expect(flagged.getByText('This file does not replay to the character it holds')).toBeVisible();
  await expect(flagged.getByRole('button', { name: 'Keep the saved character' })).toBeVisible();
  const older = await openFresh(browser, baseURL!, 'older.riftweave.json', { ...saved, rules_version: '0.0.1' });
  await expect(older.getByText(/made with rules 0\.0\.1 and was replayed under rules \S+; the result is the same/)).toBeVisible();
});

test('a dev build puts rw in the console (APP-48)', async ({ page }) => {
  await page.goto('/#/play/character_creation');
  await expect(page.getByRole('button', { name: 'Create character' })).toBeVisible();
  // The console module loads after bootstrap, by a dynamic import.
  await page.waitForFunction(() => 'rw' in globalThis);
  const help = await page.evaluate(() => (globalThis as unknown as { rw: { help(): string } }).rw.help());
  expect(help).toContain('rw.context');
  const hp = await page.evaluate(async () => {
    const rw = (globalThis as unknown as { rw: { api: { derived(actor: object): Promise<{ hp: number }> } } }).rw;
    return (await rw.api.derived({ attributes: { str: 4, per: 4, end: 4, int: 5, wil: 2, agi: 6, cha: 3, lck: 3 }, level: 3 })).hp;
  });
  expect(hp).toBe(59);
  await page.evaluate(() => (globalThis as unknown as { rw: { context: { theme: { set(theme: string): void } } } }).rw.context.theme.set('dark'));
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
});

const fixtures = '../ruleset/tests/fixtures/files/';

test('Import opens a character file, and stops on a module this browser does not load (APP-30)', async ({ page }) => {
  await page.goto('/#/import');
  await page.locator('input[type=file]').setInputFiles(fixtures + 'character.riftweave.json');
  await expect(page.getByText('Imported a character file')).toBeVisible();

  await page.locator('input[type=file]').setInputFiles(fixtures + 'missing-module.riftweave.json');
  await expect(page.getByText('module factions is not enabled here (module:factions/accessory/warden_badge)')).toBeVisible();
  await page.getByRole('button', { name: 'Import anyway' }).click();
  await expect(page.getByText('Imported a character file')).toBeVisible();

  await page.locator('input[type=file]').setInputFiles(fixtures + 'invalid.riftweave.json');
  await expect(page.getByText('nothing changed')).toBeVisible();
});

test('character creation exports a file in the one layout both apps write (APP-30)', async ({ page }) => {
  await page.goto('/#/play/character_creation');
  await page.getByRole('button', { name: 'Create character' }).click();
  await expect(page.getByRole('region', { name: 'Stored actor' })).toBeVisible();
  const downloading = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Export character' }).click();
  const file = await downloading;
  expect(file.suggestedFilename()).toBe('character.riftweave.json');
  const text = await readFile(await file.path(), 'utf8');
  const sortKeys = (value: unknown): unknown => Array.isArray(value) ? value.map(sortKeys)
    : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map((key) => [key, sortKeys((value as Record<string, unknown>)[key])])) : value;
  expect(text).toBe(JSON.stringify(sortKeys(JSON.parse(text)), null, 2) + '\n');
  expect(JSON.parse(text)).toMatchObject({ riftweave: 'character', modules: [] });
});
