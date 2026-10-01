import js from '@eslint/js';
import { defineConfig } from 'eslint/config';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import ts from 'typescript-eslint';
import svelteConfig from './svelte.config.js';

export default defineConfig(
  js.configs.recommended,
  ts.configs.recommended,
  svelte.configs.recommended,
  {
    languageOptions: {
      globals: { ...globals.browser, ...globals.node }
    },
    rules: {
      // The app is always served at the root of its own origin (the core
      // embeds it, S02.1-T02), with no base path, and its links are built
      // from file paths at run time. resolve() adds nothing here.
      'svelte/no-navigation-without-resolve': 'off',
      // Security rules (S03.5-T08; the secure coding standard, section 2).
      // No code from strings, and never user data as raw HTML. The Svelte
      // rules are also in the recommended set; they are set here so a
      // change of that set cannot drop them.
      'no-eval': 'error',
      'no-implied-eval': 'error',
      'no-new-func': 'error',
      'no-script-url': 'error',
      'svelte/no-at-html-tags': 'error',
      'svelte/no-target-blank': 'error'
    }
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        svelteConfig
      }
    }
  },
  {
    // schema.d.ts is generated from api/openapi.yaml (pnpm generate).
    ignores: [
      'build/',
      '.svelte-kit/',
      'node_modules/',
      'src/lib/api/schema.d.ts',
      'static/pdfjs/',
      // Test output (S02.8)
      'coverage/',
      '.vitest/',
      'test-results/',
      'playwright-report/'
    ]
  }
);
