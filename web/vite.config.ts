import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()],
  server: {
    // The development server sends API calls to a core running on this
    // computer (README, Development).
    proxy: { '/api': 'http://127.0.0.1:8080' }
  }
});
