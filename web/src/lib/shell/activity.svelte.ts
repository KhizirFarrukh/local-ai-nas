// Work in progress that the shell shows as a thin bar at the top
// (S02.2-T02): wrap a promise with track() while it runs.
import { untrack } from 'svelte';

class Activity {
  pending = $state(0);

  /**
   * Counts the work while it runs. Safe to call from an effect: the counter
   * is updated untracked, so the effect does not come to depend on it and
   * run again whenever it changes (bug S02-B04).
   */
  async track<T>(work: Promise<T>): Promise<T> {
    untrack(() => this.pending++);
    try {
      return await work;
    } finally {
      untrack(() => this.pending--);
    }
  }
}

/** The activity of this tab. */
export const activity = new Activity();
