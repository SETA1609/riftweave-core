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
  await expect(tools.getByText('Spell creation (APP-22)')).toHaveAttribute('aria-disabled', 'true');
  await expect(tools.getByText('Bestiary (APP-29)')).toHaveAttribute('aria-disabled', 'true');
  await tools.getByRole('link', { name: 'GM tools' }).click();
  const gm = menu(page, 'GM tools');
  await expect(gm.getByText('Encounter tracker (APP-34)')).toHaveAttribute('aria-disabled', 'true');
  await gm.getByRole('link', { name: 'Resolve an attack' }).click();
  await expect(page.getByRole('button', { name: 'Resolve' })).toBeVisible();
  await page.goBack();
  await menu(page, 'GM tools').getByRole('link', { name: 'Back' }).click();
  await menu(page, 'Start').getByRole('link', { name: 'Back' }).click();
  await expect(menu(page, 'Riftweave')).toBeVisible();
});

test('every tool route renders', async ({ page }) => {
  const screens: [string, string][] = [
    ['character_creation', 'Compute stats'], ['skill_resolution', 'Resolve'], ['saved_games', 'Open file'],
    ['attack', 'Resolve'], ['playthrough', 'Run the playthrough'],
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
