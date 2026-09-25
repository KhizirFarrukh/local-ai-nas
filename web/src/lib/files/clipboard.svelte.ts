// The app's own clipboard for files (S02.5-T04): Ctrl/Cmd+C or X keeps
// the selected items, and Ctrl/Cmd+V copies or moves them into the folder
// on screen. It lives for the tab, across folders, and holds items, not
// file contents; the system clipboard is not involved.
import type { FileItem } from './types';

export class FileClipboard {
  mode = $state<'copy' | 'move' | undefined>(undefined);
  items = $state.raw<FileItem[]>([]);

  set(mode: 'copy' | 'move', items: FileItem[]): void {
    this.mode = items.length > 0 ? mode : undefined;
    this.items = items;
  }

  clear(): void {
    this.mode = undefined;
    this.items = [];
  }

  // The cut paths as a lookup, rebuilt whole when the items change.
  private readonly cut = $derived(
    Object.fromEntries((this.mode === 'move' ? this.items : []).map((i) => [i.path, true]))
  );

  /** Reports whether the item at path waits to be moved (shown dimmed). */
  isCut(path: string): boolean {
    return Object.hasOwn(this.cut, path);
  }
}

export const clipboard = new FileClipboard();
