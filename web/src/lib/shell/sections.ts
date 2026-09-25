// The app's sections (S02.2-T01, FR-079): the navigation shows them in the
// side bar on wide screens and in the bottom bar on phones.
import Folder from '@lucide/svelte/icons/folder';
import Images from '@lucide/svelte/icons/images';
import Settings from '@lucide/svelte/icons/settings';
import type { IconComponent } from '$lib/components/icon';

export interface Section {
  label: string;
  href: string;
  icon: IconComponent;
}

export const sections: readonly Section[] = [
  { label: 'Files', href: '/files', icon: Folder },
  { label: 'Photos', href: '/photos', icon: Images },
  { label: 'Settings', href: '/settings', icon: Settings }
];

/** Reports whether a URL path belongs to a section. */
export function inSection(pathname: string, section: Section): boolean {
  return pathname === section.href || pathname.startsWith(section.href + '/');
}
