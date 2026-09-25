// The app opens in the files area.
import { redirect } from '@sveltejs/kit';

export function load() {
  redirect(307, '/files');
}
