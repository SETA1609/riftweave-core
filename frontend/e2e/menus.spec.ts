// The app layout (docs/backend/app-layout.md): the home page, the tools page, and every route,
// with the menus built from backend/contract/menu.json and the pending entries disabled.
import { expect, test, type Page } from '@playwright/test';

const menu = (page: Page, name: string) => page.getByRole('navigation', { name });

test('the app opens on the home page, and Start opens the tools', async ({ page }) => {
  await page.goto('/');
  const home = menu(page, 'Riftweave');
  for (const label of ['Start', 'Characters', 'Settings', 'Modules', 'Network']) {
    await expect(home.getByRole('link', { name: label, exact: true })).toBeVisible();
  }
  await expect(home.getByText('Quit')).toHaveCount(0);

  await home.getByRole('link', { name: 'Start' }).click();
  const tools = menu(page, 'Start');
  await expect(tools.getByRole('link', { name: 'Spell creation', exact: true })).toBeVisible();
  await expect(tools.getByText('Bestiary (FE_S08_06_bestiary-page)')).toHaveAttribute('aria-disabled', 'true');
  await tools.getByRole('link', { name: 'GM tools' }).click();
  const gm = menu(page, 'GM tools');
  await expect(gm.getByText('Encounter tracker (APP-34)')).toHaveAttribute('aria-disabled', 'true');
  await gm.getByRole('link', { name: 'Resolve an attack' }).click();
  await expect(page.getByRole('button', { name: 'Resolve' })).toBeVisible();
  await page.goBack();
  await back(page).click();
  await expect(menu(page, 'Start')).toBeVisible();
  await back(page).click();
  await expect(menu(page, 'Riftweave')).toBeVisible();
  await expect(back(page)).toHaveCount(0);
});

/** The screen header's Back (FE_S08_12). */
const back = (page: Page) => page.getByRole('link', { name: 'Back', exact: true });

test('every screen shows its title and Back, which goes to its parent menu, also when opened by its URL (FE_S08_12)', async ({ page }) => {
  const screens: [string, string, string, string][] = [
    ['/play', 'Start', '/', 'Riftweave'],
    ['/characters', 'Characters', '/', 'Riftweave'],
    ['/settings', 'Settings', '/', 'Riftweave'],
    ['/modules', 'Modules', '/', 'Riftweave'],
    ['/network', 'Network', '/', 'Riftweave'],
    ['/import', 'Import', '/', 'Riftweave'],
    ['/play/gm_tools', 'GM tools', '/play', 'Start'],
    ['/play/character_creation', 'Character creation', '/play', 'Start'],
    ['/play/skill_resolution', 'Skill resolution', '/play', 'Start'],
    ['/play/spell_creation', 'Spell creation', '/play', 'Start'],
    ['/play/saved_games', 'Saved games', '/play', 'Start'],
    ['/play/attack', 'Resolve an attack', '/play/gm_tools', 'GM tools'],
    ['/play/playthrough', 'Playthrough', '/play/gm_tools', 'GM tools'],
  ];
  for (const [path, title, parent, parentMenu] of screens) {
    await page.goto(`/#${path}`);
    await expect(page.locator('rw-screen-header').getByRole('heading', { level: 2 }), path).toHaveText(title);
    await back(page).click();
    await expect(page, path).toHaveURL(new RegExp(`#${parent.replace(/\//g, '\\/')}$`));
    await expect(menu(page, parentMenu), path).toBeVisible();
  }
});

test('Esc goes back one menu, but not from a field (FE_S08_12)', async ({ page }) => {
  await page.goto('/#/play/skill_resolution');
  await page.getByLabel('value', { exact: true }).focus();
  await page.keyboard.press('Escape');
  await expect(page).toHaveURL(/#\/play\/skill_resolution$/);
  await page.getByLabel('value', { exact: true }).blur();
  await page.keyboard.press('Escape');
  await expect(menu(page, 'Start')).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(menu(page, 'Riftweave')).toBeVisible();
});

test('character creation asks before leaving with changes no character was made from (FE_S08_12)', async ({ page }) => {
  await page.goto('/#/play/character_creation');
  await page.getByLabel('str', { exact: true }).fill('8');
  page.once('dialog', (dialog) => dialog.dismiss());
  await back(page).click();
  await expect(page).toHaveURL(/#\/play\/character_creation$/);
  page.once('dialog', (dialog) => dialog.accept());
  await back(page).click();
  await expect(menu(page, 'Start')).toBeVisible();
});

test('every tool route renders', async ({ page }) => {
  const screens: [string, string][] = [
    ['character_creation', 'Compute stats'], ['skill_resolution', 'Resolve'], ['saved_games', 'Open file'],
    ['attack', 'Resolve'], ['playthrough', 'Run the playthrough'], ['spell_creation', 'Load'],
  ];
  for (const [route, button] of screens) {
    await page.goto(`/#/play/${route}`);
    await expect(page.getByText(button, { exact: true }).first(), route).toBeVisible();
  }
});

test('skill resolution resolves a check', async ({ page }) => {
  await page.goto('/#/play/skill_resolution');
  await page.getByLabel('value', { exact: true }).fill('50');
  await page.getByLabel('d100', { exact: true }).fill('33');
  await page.getByRole('button', { name: 'Resolve' }).click();
  const check = page.getByRole('region', { name: 'Check' });
  await expect(check.getByRole('row', { name: /^target\b/ })).toContainText('50');
  await expect(check.getByRole('row', { name: /^margin\b/ })).toContainText('17');
});

test('settings, modules, and network are kept in the settings file', async ({ page }) => {
  await page.goto('/#/settings');
  await page.getByLabel('Players see only the monsters they know').check();
  const download = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Export settings' }).click();
  expect((await download).suggestedFilename()).toBe('riftweave-settings.json');

  await page.goto('/#/modules');
  await expect(page.getByLabel('core (always loaded)')).toBeDisabled();
  await page.getByLabel('factions').check();
  await page.reload();
  await expect(page.getByLabel('factions')).toBeChecked();

  await page.goto('/#/network');
  await expect(page.getByLabel('Room (APP-17)')).toBeDisabled();
  await page.getByRole('textbox', { name: 'Server URL' }).fill('http://127.0.0.1:8080');
  await page.getByRole('button', { name: 'Use this server' }).click();
  await expect(page.getByText('Commands now go to the server.')).toBeVisible();
  await page.getByLabel('Offline: the core in the page').check();
  await expect(page.getByText('Offline: the core in the page.')).toBeVisible();
});
