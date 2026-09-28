// Folder sizes for the views (S02.3-T05, FR-214). The server adds a folder
// up only when asked, so the views ask for the folders on screen, a few at
// a time, the most recently shown first. After the folder on screen is
// loaded again, sizes are asked for again; the old ones stay shown until
// the new ones arrive, so nothing flickers.
import { SvelteMap } from 'svelte/reactivity';

/** Adds up one folder; the app passes the API, tests pass a fake. */
export type UsageLoader = (path: string, signal: AbortSignal) => Promise<number>;

export class FolderSizes {
  // A size, or null when it could not be added up. Read by the views.
  private readonly known = new SvelteMap<string, number | null>();
  // Plain bookkeeping, never read by the views, so asking from an effect
  // cannot make it run again.
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  private asked = new Set<string>();
  private waiting: string[] = [];
  private running = 0;
  private generation = 0;
  private stamp: number | undefined;
  private controller = new AbortController();

  constructor(
    private readonly load: UsageLoader,
    private readonly parallel = 3
  ) {}

  /**
   * The size of the folder at path: a number of bytes, null when it could
   * not be added up, or undefined while it is not known yet.
   */
  get(path: string): number | null | undefined {
    return this.known.get(path);
  }

  /**
   * Asks for the sizes of these folders, unless asked already since the
   * folder on screen was loaded (stamp: `FolderListing.loads`).
   */
  want(paths: readonly string[], stamp: number): void {
    if (stamp !== this.stamp) {
      this.stamp = stamp;
      this.restart();
    }
    for (const path of paths) {
      if (!this.asked.has(path)) {
        this.asked.add(path);
        this.waiting.push(path);
      }
    }
    this.pump();
  }

  /** Stops what is running, and forgets what was asked. */
  private restart(): void {
    this.controller.abort();
    this.controller = new AbortController();
    this.generation++;
    this.asked.clear();
    this.waiting = [];
    this.running = 0;
  }

  private pump(): void {
    while (this.running < this.parallel && this.waiting.length > 0) {
      const path = this.waiting.pop()!; // the most recently shown first
      const generation = this.generation;
      this.running++;
      this.load(path, this.controller.signal)
        .then(
          (size) => generation === this.generation && this.known.set(path, size),
          () => generation === this.generation && this.known.set(path, null)
        )
        .finally(() => {
          if (generation === this.generation) {
            this.running--;
            this.pump();
          }
        });
    }
  }
}
