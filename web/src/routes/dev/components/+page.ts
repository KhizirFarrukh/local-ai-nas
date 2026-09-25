// The component gallery exists in development builds only (S02.1-T04).
import { dev } from '$app/environment';
import { error } from '@sveltejs/kit';

export function load() {
  if (!dev) {
    error(404, 'Not Found');
  }
}
