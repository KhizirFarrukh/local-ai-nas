import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    // A single-page app (ADR-0009): every route falls back to index.html,
    // which the core serves for app paths (S02.1-T02).
    adapter: adapter({ fallback: 'index.html' })
  }
};

export default config;
