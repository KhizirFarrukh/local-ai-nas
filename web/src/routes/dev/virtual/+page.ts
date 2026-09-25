// The virtualization prototype (S02.3-T01) exists in development builds only.
import { dev } from '$app/environment';
import { error } from '@sveltejs/kit';

export function load() {
  if (!dev) {
    error(404, 'Not Found');
  }
}
