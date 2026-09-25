import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    // A single-page app (ADR-0009): every route falls back to index.html,
    // which the core serves for app paths (S02.1-T02).
    adapter: adapter({ fallback: 'index.html' }),
    // The Content Security Policy (NFR-022). SvelteKit adds the hashes of
    // its inline bootstrap script to the page's <meta> policy, and the core
    // sends the same policy as a header, plus frame-ancestors, which a
    // <meta> policy cannot carry. Nothing is loaded from other hosts (I6).
    csp: {
      mode: 'hash',
      directives: {
        'default-src': ['self'],
        'script-src': ['self'],
        'style-src': ['self'],
        'img-src': ['self', 'data:', 'blob:'],
        'media-src': ['self', 'blob:'],
        'font-src': ['self'],
        'connect-src': ['self'],
        'worker-src': ['self'],
        'object-src': ['none'],
        'base-uri': ['self'],
        'form-action': ['self']
      }
    }
  }
};

export default config;
