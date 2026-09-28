// System tests (S02.8-T02): the real binary with the embedded interface,
// driven in real browsers. Run `pnpm build` first (`pnpm test:e2e` does).
// Chromium and Firefox everywhere; Edge (the msedge channel) on Windows;
// Google Chrome (the chrome channel) when E2E_CHROME=1, as in CI, whose
// machines have it. Safari is not checked in S02 (the user's decision,
// S006 E009).
import { defineConfig, devices } from '@playwright/test';
import { prepare } from './tests/e2e/setup';

const port = Number(process.env.E2E_PORT ?? 18_300);
const server = prepare(port);
const base = `http://127.0.0.1:${port}`;

export default defineConfig({
  testDir: 'tests/e2e',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  // One server and one storage root: tests work in folders of their own,
  // but run one at a time so timings (the large folder) are not disturbed.
  workers: 1,
  retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { baseURL: base, trace: 'retain-on-failure', acceptDownloads: true },
  webServer: {
    command: `"${server.binary}" serve --storage-root "${server.root}" --server-bind 127.0.0.1:${port} --uploads-max-chunk-size 1MiB`,
    url: `${base}/api/v1/system/health`,
    timeout: 120_000,
    stdout: 'ignore',
    stderr: 'ignore'
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    ...(process.platform === 'win32'
      ? [{ name: 'edge', use: { ...devices['Desktop Edge'], channel: 'msedge' } }]
      : []),
    ...(process.env.E2E_CHROME === '1'
      ? [{ name: 'chrome', use: { ...devices['Desktop Chrome'], channel: 'chrome' } }]
      : [])
  ]
});
