// The selection of the file browser (S02.5-T01, FR-021), as in common
// file managers: a click selects one item, Ctrl/Cmd-click toggles one,
// Shift-click selects the range from the last clicked item (with Ctrl/Cmd,
// adding to what is selected), and "select all" takes the whole folder.
//
// Items are kept by path, so a selection survives pages loading, a
// refresh, and a new sort. "Select all" is kept as "everything except", so
// it is instant in a folder of 50,000 items and covers pages loaded later.
// A range over pages not loaded yet loads them first; a newer selection
// cancels a range still waiting for its pages.
import { SvelteMap } from 'svelte/reactivity';
import type { FileItem } from './types';

/** The items a selection is made from: the folder listing, or a fake in tests. */
export interface ItemSource {
  readonly total: number | undefined;
  at(index: number): FileItem | undefined;
  /** Loads the items at first..last; false when that failed. */
  loadRange(first: number, last: number): Promise<boolean>;
}

/** How a pick changes the selection. */
export type PickMode =
  /** Only this item (a click, or an arrow key). */
  | 'only'
  /** This item in or out, the rest kept (Ctrl/Cmd-click, Space). */
  | 'toggle'
  /** The range from the anchor, instead of the selection (Shift). */
  | 'range'
  /** The range from the anchor, added to the selection (Ctrl/Cmd+Shift). */
  | 'add-range';

export class Selection {
  /** True when everything is selected except `items`. */
  all = $state(false);
  /** The selected items, or with `all` the items left out, by path. */
  readonly items = new SvelteMap<string, FileItem>();
  /** Where ranges start: the last item picked without Shift. */
  anchor = $state<number | undefined>(undefined);

  /** Counts changes, so a range that waited for its pages knows it is stale. */
  private change = 0;

  /** Reports whether the item at path is selected. */
  has(path: string): boolean {
    return this.all !== this.items.has(path);
  }

  /** The number of selected items in a folder of `total` items. */
  count(total: number | undefined): number {
    return this.all ? Math.max(0, (total ?? 0) - this.items.size) : this.items.size;
  }

  /** Reports whether every item is selected. */
  get everything(): boolean {
    return this.all && this.items.size === 0;
  }

  clear(): void {
    this.change++;
    this.all = false;
    this.items.clear();
    this.anchor = undefined;
  }

  selectAll(): void {
    this.change++;
    this.all = true;
    this.items.clear();
  }

  /**
   * Changes the selection for the item at `index`. The pages needed are
   * loaded first; the promise resolves once the selection is updated, or
   * at once when a newer change made this one stale.
   */
  async pick(index: number, mode: PickMode, source: ItemSource): Promise<void> {
    const change = ++this.change;
    const ranged = mode === 'range' || mode === 'add-range';
    const from = ranged ? (this.anchor ?? index) : index;
    const [first, last] = from <= index ? [from, index] : [index, from];
    if (!(await this.loaded(first, last, source)) || change !== this.change) {
      return;
    }
    switch (mode) {
      case 'only':
        this.all = false;
        this.items.clear();
        this.add(source.at(index));
        this.anchor = index;
        break;
      case 'toggle': {
        const item = source.at(index);
        if (item) {
          this.set(item, !this.has(item.path));
        }
        this.anchor = index;
        break;
      }
      case 'range':
        this.all = false;
        this.items.clear();
      // falls through: a range replaces the selection, then adds like add-range
      case 'add-range':
        for (let i = first; i <= last; i++) {
          this.add(source.at(i));
        }
        this.anchor ??= index;
        break;
    }
  }

  /**
   * Leaves out items that are gone, such as after a delete or a move. With
   * "select all" the folder's total drops too, so the count stays right.
   */
  forget(paths: Iterable<string>): void {
    for (const path of paths) {
      this.items.delete(path);
    }
  }

  /**
   * The selected items, in the folder's order for "select all" (its pages
   * are loaded for that); undefined when a page could not be loaded.
   */
  async resolve(source: ItemSource): Promise<FileItem[] | undefined> {
    if (!this.all) {
      return [...this.items.values()];
    }
    const total = source.total ?? 0;
    if (total > 0 && !(await source.loadRange(0, total - 1))) {
      return undefined;
    }
    const out: FileItem[] = [];
    for (let i = 0; i < total; i++) {
      const item = source.at(i);
      if (item && !this.items.has(item.path)) {
        out.push(item);
      }
    }
    return out;
  }

  private async loaded(first: number, last: number, source: ItemSource): Promise<boolean> {
    for (let i = first; i <= last; i++) {
      if (!source.at(i)) {
        return source.loadRange(first, last);
      }
    }
    return true;
  }

  private add(item: FileItem | undefined): void {
    if (item) {
      this.set(item, true);
    }
  }

  /** Selects or unselects one item, in either kind of selection. */
  private set(item: FileItem, selected: boolean): void {
    if (selected === this.all) {
      this.items.delete(item.path);
    } else {
      this.items.set(item.path, item);
    }
  }
}
