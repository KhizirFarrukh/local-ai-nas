// Long operations in progress (S02.2-T03, FR-079): uploads, bulk
// operations, archives. The progress panel lists them; when one ends, a
// notification says how it went, and it leaves the list.
import { describe } from '$lib/api/messages';
import { toasts, type ToastInput, type Toasts } from './toasts.svelte';

export type TaskState = 'running' | 'done' | 'failed' | 'cancelled';

export class Task {
  /** Work done so far, in the unit of `total` (bytes, items). */
  done = $state(0);
  /** Undefined while the size is unknown. */
  total = $state<number | undefined>(undefined);
  /** A second line, such as "3 of 10 files". */
  note = $state('');
  state = $state<TaskState>('running');

  constructor(
    readonly id: number,
    public label: string,
    private readonly owner: Tasks,
    readonly cancel?: () => void
  ) {}

  /** Records progress. */
  update(done: number, total?: number, note?: string): void {
    this.done = done;
    if (total !== undefined) {
      this.total = total;
    }
    if (note !== undefined) {
      this.note = note;
    }
  }

  /** Ends the task successfully, with a notification. */
  finish(message: string): void {
    this.end('done');
    this.owner.notes.push({ kind: 'success', message });
  }

  /** Ends the task with an error, with a notification that stays. */
  fail(error: unknown, what = this.label): void {
    this.end('failed');
    const m = describe(error);
    this.owner.notes.push({
      kind: 'error',
      message: `${what}: ${m.title}. ${m.message}`,
      detail: m.detail
    });
  }

  /** Ends the task because the user cancelled it. */
  cancelled(message = `${this.label}: cancelled`): void {
    this.end('cancelled');
    this.owner.notes.push({ kind: 'info', message });
  }

  /** Ends the task with a notification of its own, such as a bulk run's summary. */
  summary(state: 'done' | 'failed' | 'cancelled', note: ToastInput): void {
    this.end(state);
    this.owner.notes.push(note);
  }

  private end(state: TaskState): void {
    this.state = state;
    this.owner.remove(this.id);
  }
}

export class Tasks {
  items = $state<Task[]>([]);
  private next = 1;

  constructor(readonly notes: Toasts) {}

  /** Starts a task. `cancel`, when given, adds a Cancel button. */
  start(label: string, options: { total?: number; cancel?: () => void } = {}): Task {
    const task = new Task(this.next++, label, this, options.cancel);
    task.total = options.total;
    this.items.push(task);
    return task;
  }

  remove(id: number): void {
    this.items = this.items.filter((t) => t.id !== id);
  }
}

/** The long operations of this tab. */
export const tasks = new Tasks(toasts);
