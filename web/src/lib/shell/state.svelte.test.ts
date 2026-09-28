// The shell's small state modules: the clipboard, activity, notifications,
// tasks, the connection, and the announcer.
import { flushSync } from 'svelte';
import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '$lib/api/errors';
import { FileClipboard } from '$lib/files/clipboard.svelte';
import type { FileItem } from '$lib/files/types';
import { activity } from './activity.svelte';
import { Announcer } from './announcer.svelte';
import { Connection } from './connection.svelte';
import { Tasks } from './tasks.svelte';
import { Toasts } from './toasts.svelte';

const item = (path: string): FileItem => ({
  path,
  name: path.slice(1),
  kind: 'file',
  size: 1,
  mod_time: '',
  added_time: ''
});

/** A timer that runs callbacks when the test says so. */
function manualTimer() {
  const queue: { run: () => void; ms: number }[] = [];
  return {
    timer: (run: () => void, ms: number) => void queue.push({ run, ms }),
    queue,
    flush() {
      for (const t of queue.splice(0)) t.run();
    }
  };
}

describe('FileClipboard', () => {
  it('keeps cut items as dimmed until cleared, and copied ones never', () => {
    const c = new FileClipboard();
    c.set('move', [item('/a'), item('/b')]);
    flushSync();
    expect(c.isCut('/a')).toBe(true);
    expect(c.isCut('/c')).toBe(false);
    c.set('copy', [item('/a')]);
    expect(c.isCut('/a')).toBe(false);
    c.set('move', []);
    expect(c.mode).toBeUndefined();
    c.set('move', [item('/a')]);
    c.clear();
    expect(c.mode).toBeUndefined();
    expect(c.items).toEqual([]);
    expect(c.isCut('/a')).toBe(false);
  });
});

describe('activity', () => {
  it('counts work while it runs, also when it fails', async () => {
    let finish: () => void = () => {};
    const work = activity.track(new Promise<void>((r) => (finish = r)));
    expect(activity.pending).toBe(1);
    finish();
    await work;
    expect(activity.pending).toBe(0);
    await expect(activity.track(Promise.reject(new Error('x')))).rejects.toThrow('x');
    expect(activity.pending).toBe(0);
  });
});

describe('Toasts', () => {
  it('speaks each note, errors at once; notes go away except errors', () => {
    const t = manualTimer();
    const spoken: [string, boolean][] = [];
    const toasts = new Toasts(t.timer, (m, urgent) => spoken.push([m, urgent]));
    const id = toasts.push({ message: 'Saved.' });
    toasts.push({ kind: 'error', message: 'Failed.', detail: 'Request 1' });
    toasts.push({ kind: 'success', message: 'Quick.', timeout: 10 });
    expect(spoken).toEqual([
      ['Saved.', false],
      ['Failed.', true],
      ['Quick.', false]
    ]);
    expect(t.queue.map((q) => q.ms)).toEqual([5000, 10]); // the error has no timer
    t.flush();
    expect(toasts.items.map((x) => x.message)).toEqual(['Failed.']);
    toasts.dismiss(id); // already gone: no harm
    toasts.dismiss(toasts.items[0].id);
    expect(toasts.items).toEqual([]);
  });
});

describe('Tasks', () => {
  it('tracks progress and ends each task with a note', () => {
    const toasts = new Toasts(() => 0);
    const tasks = new Tasks(toasts);
    const a = tasks.start('Copying 2 items', { total: 2 });
    a.update(1, undefined, '1 of 2');
    expect([a.done, a.total, a.note]).toEqual([1, 2, '1 of 2']);
    a.update(2, 3);
    expect(a.total).toBe(3);
    a.finish('Copied 2 items.');
    expect(a.state).toBe('done');
    expect(tasks.items).toHaveLength(0);

    const b = tasks.start('Deleting “x”', { cancel: () => {} });
    expect(b.cancel).toBeTypeOf('function');
    b.fail(new ApiError({ status: 404, code: 'not_found', message: 'gone' }));
    const c = tasks.start('Uploading');
    c.cancelled();
    const d = tasks.start('Moving');
    d.summary('failed', { kind: 'error', message: 'Some failed.' });
    expect(toasts.items.map((n) => [n.kind, n.message])).toEqual([
      ['success', 'Copied 2 items.'],
      ['error', 'Deleting “x”: Not found. It may have been moved, renamed, or deleted.'],
      ['info', 'Uploading: cancelled'],
      ['error', 'Some failed.']
    ]);
    expect([b.state, c.state, d.state]).toEqual(['failed', 'cancelled', 'failed']);
  });
});

describe('Connection', () => {
  it('retries a lost connection on a timer until the server answers', async () => {
    vi.useFakeTimers();
    try {
      let up = false;
      const back = vi.fn();
      const c = new Connection(
        async () => {
          if (!up) throw new Error('down');
        },
        1000,
        back
      );
      c.lost();
      c.lost(); // once is enough
      expect(c.online).toBe(false);
      await vi.advanceTimersByTimeAsync(1000); // a check fails, another is scheduled
      expect(c.online).toBe(false);
      up = true;
      await vi.advanceTimersByTimeAsync(1000);
      expect(c.online).toBe(true);
      expect(c.checking).toBe(false);
      expect(back).toHaveBeenCalledTimes(1);
      expect(await c.retry()).toBe(true); // already online: no "back" note
      expect(back).toHaveBeenCalledTimes(1);
    } finally {
      vi.useRealTimers();
    }
  });
});

describe('Announcer', () => {
  it('empties the region first, so the same message is read again', () => {
    const t = manualTimer();
    const a = new Announcer(t.timer);
    a.say('3 selected');
    t.flush();
    expect(a.polite).toBe('3 selected');
    a.say('3 selected');
    expect(a.polite).toBe('');
    t.flush();
    expect(a.polite).toBe('3 selected');
    a.say('Upload failed', true);
    t.flush();
    expect([a.polite, a.urgent]).toEqual(['3 selected', 'Upload failed']);
  });
});
