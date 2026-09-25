// Whether the NAS can be reached (S02.2-T02). Any API call that gets no
// answer marks the connection lost; the shell then shows a banner and
// checks the health endpoint every few seconds until the server answers.
import { api, onApiError, unwrap } from '$lib/api/client';

export type Check = () => Promise<unknown>;

export class Connection {
  online = $state(true);
  checking = $state(false);
  private timer: ReturnType<typeof setTimeout> | undefined;

  constructor(
    private readonly check: Check,
    private readonly retryMs = 5000
  ) {}

  /** Records that a request got no answer, and starts retrying. */
  lost(): void {
    if (this.online) {
      this.online = false;
      this.schedule();
    }
  }

  /** Checks the server now; resolves to whether it answered. */
  async retry(): Promise<boolean> {
    clearTimeout(this.timer);
    this.checking = true;
    try {
      await this.check();
      this.online = true;
    } catch {
      this.online = false;
      this.schedule();
    } finally {
      this.checking = false;
    }
    return this.online;
  }

  private schedule(): void {
    clearTimeout(this.timer);
    this.timer = setTimeout(() => void this.retry(), this.retryMs);
  }
}

let shared: Connection | undefined;

/** The connection of this tab, fed by every API call. */
export function connection(): Connection {
  if (!shared) {
    const c = new Connection(() => unwrap(api.GET('/system/health')));
    onApiError((error) => {
      if (error.code === 'unreachable') {
        c.lost();
      }
    });
    shared = c;
  }
  return shared;
}
