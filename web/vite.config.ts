import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { playwright } from '@vitest/browser-playwright';
import { defineConfig } from 'vitest/config';
import { tusFake } from './src/lib/testing/tus-fake';

// The libraries the app imports, for the dependency optimizer (S02.8-T01).
// Browser tests meeting one it has not bundled yet make Vite reload the
// page mid-run, which breaks the test; icons are imported one by one, so
// they are listed from the sources.
function prebundled(): string[] {
  const icons = new Set<string>();
  for (const file of readdirSync('src', { recursive: true, encoding: 'utf8' })) {
    if (/\.(svelte|ts)$/.test(file)) {
      const source = readFileSync(join('src', file), 'utf8');
      for (const m of source.matchAll(/'(@lucide\/svelte\/icons\/[a-z0-9-]+)'/g)) {
        icons.add(m[1]);
      }
    }
  }
  return [
    ...icons,
    '@tanstack/svelte-virtual',
    '@uppy/core',
    '@uppy/tus',
    'openapi-fetch',
    'pdfjs-dist',
    'vitest-browser-svelte'
  ];
}

export default defineConfig({
  // Under Vitest, a fake tus server answers at the upload path (S02.8-T01).
  plugins: [tailwindcss(), sveltekit(), ...(process.env.VITEST ? [tusFake()] : [])],
  optimizeDeps: { include: prebundled() },
  server: {
    // The development server sends API calls to a core running on this
    // computer (README, Development). Tests have no core behind them.
    proxy: process.env.VITEST ? undefined : { '/api': 'http://127.0.0.1:8080' }
  },
  // Tests (S02.8-T01): plain modules run in Node; components and the
  // modules that use runes (*.svelte.ts) run in a real browser, where
  // Svelte's reactivity works as in the app.
  test: {
    expect: { requireAssertions: true },
    coverage: {
      provider: 'v8',
      include: ['src/lib/**/*.{ts,svelte}'],
      exclude: [
        'src/lib/**/*.test.ts',
        'src/lib/testing/**',
        'src/lib/api/schema.d.ts',
        'src/lib/generated/**'
      ],
      reporter: ['text-summary', 'text', 'json-summary'],
      // An exit criterion of each stage's testing substage (RULES R6).
      thresholds: { statements: 80, branches: 80, functions: 80, lines: 80 }
    },
    projects: [
      {
        extends: true,
        test: {
          name: 'unit',
          environment: 'node',
          include: ['src/**/*.test.ts'],
          exclude: ['src/**/*.svelte.test.ts']
        }
      },
      {
        extends: true,
        test: {
          name: 'browser',
          include: ['src/**/*.svelte.test.ts'],
          browser: {
            enabled: true,
            headless: true,
            provider: playwright(),
            instances: [{ browser: 'chromium' }]
          }
        }
      }
    ]
  }
});
