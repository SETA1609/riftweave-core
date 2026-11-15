// End-to-end specs (docs/backend/frontend.md#tests): Chromium on `ng serve` and its proxy, with a
// real riftweave-server behind it. Each server is started unless one already answers, so CI and
// ./scripts/app.sh reuse theirs.
import { defineConfig, devices } from '@playwright/test';

const ci = Boolean(process.env['CI']);

export default defineConfig({
  testDir: 'e2e',
  forbidOnly: ci,
  retries: ci ? 1 : 0,
  reporter: ci ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: { baseURL: 'http://localhost:4200', trace: 'retain-on-failure' },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'go run ./cmd/riftweave-server --data ../ruleset/data --addr 127.0.0.1:8080',
      cwd: '../backend',
      url: 'http://127.0.0.1:8080/v1/commands',
      reuseExistingServer: true,
      timeout: 180_000,
    },
    { command: 'npm start -- --port 4200', url: 'http://localhost:4200', reuseExistingServer: true, timeout: 180_000 },
  ],
});
