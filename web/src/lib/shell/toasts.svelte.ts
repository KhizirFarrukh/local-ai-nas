// Notifications (S02.2-T03, FR-079). Info and success notes go away by
// themselves; errors stay until the user dismisses them, so none is missed.
import type { ToastKind } from '$lib/components/Toast.svelte';

export interface ToastItem {
  id: number;
  kind: ToastKind;
  message: string;
  detail?: string;
}

export interface ToastInput {
  kind?: ToastKind;
  message: string;
  detail?: string;
  /** How long it stays, in ms; 0 keeps it. Defaults: 5 s, errors 0. */
  timeout?: number;
}

type Timer = (run: () => void, ms: number) => unknown;

export class Toasts {
  items = $state<ToastItem[]>([]);
  private next = 1;

  /** `timer` is setTimeout; tests pass their own. */
  constructor(private readonly timer: Timer = (run, ms) => setTimeout(run, ms)) {}

  push(input: ToastInput): number {
    const id = this.next++;
    const kind = input.kind ?? 'info';
    this.items.push({ id, kind, message: input.message, detail: input.detail });
    const timeout = input.timeout ?? (kind === 'error' ? 0 : 5000);
    if (timeout > 0) {
      this.timer(() => this.dismiss(id), timeout);
    }
    return id;
  }

  dismiss(id: number): void {
    this.items = this.items.filter((t) => t.id !== id);
  }
}

/** The notifications of this tab. */
export const toasts = new Toasts();
