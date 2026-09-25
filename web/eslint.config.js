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
      'svelte/no-navigation-without-resolve': 'off'
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
    ignores: ['build/', '.svelte-kit/', 'node_modules/', 'src/lib/api/schema.d.ts']
  }
);
