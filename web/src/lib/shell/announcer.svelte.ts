// Screen-reader announcements (S02.7-T02). Two live regions exist from the
// first render (LiveRegions.svelte), because a region added together with
// its text is often not read. Notifications, selection changes, and upload
// starts and failures are spoken through them; errors interrupt, the rest
// waits for a pause.

type Timer = (run: () => void, ms: number) => unknown;

export class Announcer {
  /** The polite region's text. */
  polite = $state('');
  /** The assertive region's text, for errors. */
  urgent = $state('');

  /** `timer` is setTimeout; tests pass their own. */
  constructor(private readonly timer: Timer = (run, ms) => setTimeout(run, ms)) {}

  /**
   * Speaks a message. The region is emptied first and filled a moment
   * later, so the same message twice in a row is read twice.
   */
  say(message: string, urgent = false): void {
    if (urgent) {
      this.urgent = '';
    } else {
      this.polite = '';
    }
    this.timer(() => {
      if (urgent) {
        this.urgent = message;
      } else {
        this.polite = message;
      }
    }, 50);
  }
}

/** The announcements of this tab. */
export const announcer = new Announcer();
