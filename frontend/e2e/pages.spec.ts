// The published site (APP-15, docs/backend/frontend.md#github-pages): the Pages build under its base
// path, with no server. It shows its rules version, plays with the core in the page, and after one
// visit works offline through the service worker.
import { expect, test } from '@playwright/test';

test('the Pages build plays with no server and shows its rules version', async ({ page }) => {
  await page.goto('./#/play/character_creation');
  await expect(page.locator('footer')).toHaveText(/^Rules \S+$/);
  await page.getByRole('button', { name: 'Create character' }).click();
  await expect(page.getByRole('region', { name: 'Stored actor' })).toBeVisible();
});

test('after one visit the Pages build works offline', async ({ page, context }) => {
  await page.goto('./#/play/character_creation');
  await expect(page.getByRole('button', { name: 'Create character' })).toBeVisible();
  // Wait until the service worker controls the page and has cached the app and the core.
  await page.evaluate(async () => {
    await navigator.serviceWorker.ready;
    for (let attempt = 0; attempt < 100; attempt++) {
      const names = await caches.keys();
      const cached = await Promise.all(names.map(async (name) => (await (await caches.open(name)).keys()).map((key) => key.url)));
      if (cached.flat().some((url) => url.endsWith('riftweave.wasm'))) {
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
    throw new Error('the core was never cached');
  });
  await context.setOffline(true);
  await page.reload();
  await page.getByRole('button', { name: 'Create character' }).click();
  await expect(page.getByRole('region', { name: 'Stored actor' })).toBeVisible();
  await context.setOffline(false);
});
