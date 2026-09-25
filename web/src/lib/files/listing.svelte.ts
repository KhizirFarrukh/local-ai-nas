// One folder's items for the file browser (S02.3-T03, FR-002). Pages load
// by `offset` when their part of the list becomes visible, so a folder of
// 50,000 items costs only what is on screen. The items are kept in a plain
// map, not in deep reactive state: a version counter tells the views that
// a page arrived.
import type { FileItem, ItemsResponse, SortKey, SortOrder } from './types';

/** Loads one page; the app passes the API, tests pass a fake. */
export type PageLoader = (query: {
  path: string;
  offset: number;
  limit: number;
  sort: SortKey;
  order: SortOrder;
}) => Promise<ItemsResponse>;

export const pageSize = 500;

export class FolderListing {
  /** The folder itself, once the first page arrived. */
  folder = $state.raw<FileItem | undefined>(undefined);
  /** How many items the folder has; undefined until the first page. */
  total = $state<number | undefined>(undefined);
  /** The error of the first page; later pages retry by themselves. */
  error = $state.raw<unknown>(undefined);
  /** Changes whenever a page arrives, so views read it to update. */
  version = $state(0);

  // Plain collections on purpose: deep reactivity over tens of thousands
  // of items would cost more than it gives; `version` tells the views.
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  private pages = new Map<number, FileItem[]>();
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  private loading = new Map<number, Promise<boolean>>();
  private generation = 0;

  constructor(
    private readonly load: PageLoader,
    readonly path: string,
    readonly sort: SortKey = 'name',
    readonly order: SortOrder = 'asc'
  ) {}

  /** Loads the first page. Resolves when it is there or has failed. */
  async start(): Promise<void> {
    this.generation++;
    this.pages.clear();
    this.loading.clear();
    this.error = undefined;
    await this.fetch(0, true);
  }

  /**
   * Loads the pages on screen again, after a change in the folder, and
   * swaps them in at once, so the view keeps its place and does not flicker.
   */
  async refresh(): Promise<void> {
    const wanted = this.pages.size > 0 ? [...this.pages.keys()] : [0];
    const generation = ++this.generation;
    this.loading.clear();
    try {
      const results = await Promise.all(
        wanted.map((page) =>
          this.load({
            path: this.path,
            offset: page * pageSize,
            limit: pageSize,
            sort: this.sort,
            order: this.order
          })
        )
      );
      if (generation !== this.generation) {
        return;
      }
      this.pages.clear();
      wanted.forEach((page, i) => this.pages.set(page, results[i].items ?? []));
      this.folder = results[0].item;
      this.total = results[0].total ?? results[0].items?.length ?? 0;
      this.error = undefined;
      this.version++;
    } catch {
      // Keep what is shown; the next change or a reload tries again.
    }
  }

  /** The item at a position, or undefined while its page is loading. */
  at(index: number): FileItem | undefined {
    void this.version; // views re-read when a page arrives
    return this.pages.get(Math.floor(index / pageSize))?.[index % pageSize];
  }

  /** Makes sure the pages that hold positions first..last are loading. */
  ensure(first: number, last: number): void {
    void this.loadRange(first, last);
  }

  /**
   * Loads the pages that hold positions first..last, such as for a
   * selection over pages not seen yet (S02.5-T01). Resolves to false when
   * one of them failed.
   */
  async loadRange(first: number, last: number): Promise<boolean> {
    if (this.total === undefined) {
      return false;
    }
    const end = Math.min(last, this.total - 1);
    const pending: Promise<boolean>[] = [];
    for (let page = Math.floor(Math.max(0, first) / pageSize); page * pageSize <= end; page++) {
      if (!this.pages.has(page)) {
        pending.push(this.loading.get(page) ?? this.fetch(page, false));
      }
    }
    return (await Promise.all(pending)).every(Boolean);
  }

  private fetch(page: number, first: boolean): Promise<boolean> {
    const pending = this.fetchPage(page, first);
    this.loading.set(page, pending);
    return pending;
  }

  private async fetchPage(page: number, first: boolean): Promise<boolean> {
    const generation = this.generation;
    try {
      const result = await this.load({
        path: this.path,
        offset: page * pageSize,
        limit: pageSize,
        sort: this.sort,
        order: this.order
      });
      if (generation !== this.generation) {
        return false; // the folder was reloaded meanwhile
      }
      this.pages.set(page, result.items ?? []);
      this.folder = result.item;
      this.total = result.total ?? result.items?.length ?? 0;
      this.version++;
      return true;
    } catch (error) {
      if (generation === this.generation && first) {
        this.error = error;
      }
      return false;
    } finally {
      if (generation === this.generation) {
        this.loading.delete(page);
      }
    }
  }
}
