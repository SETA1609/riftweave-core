// End-to-end specs (docs/backend/frontend.md#tests), twice: Chromium on `ng serve` and its proxy
// with a real riftweave-server behind it (chromium), and on a build with no API, where the core
// runs in the page as WebAssembly (chromium-wasm, after ./scripts/build-wasm.sh). The Pages build
// (APP-15) has its own project, chromium-pages: the static site under its base path, offline. Each server is
// started unless one already answers, so CI and ./scripts/app.sh reuse theirs.
import { defineConfig, devices } from '@playwright/test';

const ci = Boolean(process.env['CI']);

export default defineConfig({
  testDir: 'e2e',
  forbidOnly: ci,
  retries: ci ? 1 : 0,
  reporter: ci ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: { baseURL: 'http://localhost:4200', trace: 'retain-on-failure' },
  projects: [
    { name: 'chromium', testIgnore: /pages\.spec\.ts/, use: { ...devices['Desktop Chrome'] } },
    { name: 'chromium-wasm', testIgnore: /pages\.spec\.ts/, use: { ...devices['Desktop Chrome'], baseURL: 'http://localhost:4300' } },
    { name: 'chromium-pages', testMatch: /pages\.spec\.ts/, use: { ...devices['Desktop Chrome'], baseURL: 'http://localhost:4400/riftweave-core/' } },
  ],
  webServer: [
    {
      command: 'go run ./cmd/riftweave-server --data ../ruleset/data --addr 127.0.0.1:8080',
      cwd: '../backend',
      url: 'http://127.0.0.1:8080/v1/commands',
      reuseExistingServer: true,
      timeout: 180_000,
    },
    { command: 'npm start -- --port 4200', url: 'http://localhost:4200', reuseExistingServer: true, timeout: 180_000 },
    { command: 'npm run start:wasm -- --port 4300', url: 'http://localhost:4300', reuseExistingServer: true, timeout: 180_000 },
    { command: 'npm run build:pages && npm run preview:pages', url: 'http://localhost:4400/riftweave-core/', reuseExistingServer: true, timeout: 300_000 },
  ],
});
