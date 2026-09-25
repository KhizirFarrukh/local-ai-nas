// The type of an icon component (Lucide, bundled; FR-083), for components
// that take an icon as a prop. Every Lucide icon has the same type, so it
// is taken from one icon module rather than Lucide's full index.
import type Folder from '@lucide/svelte/icons/folder';

export type IconComponent = typeof Folder;
