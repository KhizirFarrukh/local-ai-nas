// The folder-size store (S02.3-T05, FR-214).
import { describe, expect, it, vi } from 'vitest';
import { FolderSizes } from './sizes.svelte';

/** A loader whose answers the test gives, one call at a time. */
function controlled() {
  const calls: { path: string; signal: AbortSignal; answer: (size: number | Error) => void }[] = [];
  const load = vi.fn(
    (path: string, signal: AbortSignal) =>
      new Promise<number>((resolve, reject) => {
        calls.push({
          path,
          signal,
          answer: (size) => (size instanceof Error ? reject(size) : resolve(size))
        });
      })
  );
  return { load, calls };
}

describe('FolderSizes', () => {
  it('asks once per folder, a few at a time, the most recently shown first', async () => {
    const { load, calls } = controlled();
    const sizes = new FolderSizes(load, 2);
    sizes.want(['/a', '/b', '/c', '/d'], 1);
    sizes.want(['/a', '/b'], 1); // asked already
    expect(calls.map((c) => c.path)).toEqual(['/d', '/c']);
    expect(sizes.get('/d')).toBeUndefined();
    calls[0].answer(4096);
    await vi.waitFor(() => expect(sizes.get('/d')).toBe(4096));
    await vi.waitFor(() => expect(calls).toHaveLength(3)); // the next one starts
    calls[1].answer(new Error('gone'));
    await vi.waitFor(() => expect(sizes.get('/c')).toBeNull());
    expect(load).toHaveBeenCalledTimes(4);
  });

  it('asks again after the folder is loaded anew, keeping the old sizes meanwhile', async () => {
    const { calls } = controlled();
    const sizes = new FolderSizes((path, signal) => {
      const p = new Promise<number>((resolve, reject) =>
        calls.push({
          path,
          signal,
          answer: (size) => (size instanceof Error ? reject(size) : resolve(size))
        })
      );
      return p;
    });
    sizes.want(['/a', '/b'], 1);
    calls[1].answer(10); // '/a'
    await vi.waitFor(() => expect(sizes.get('/a')).toBe(10));
    sizes.want(['/a', '/b'], 2); // a refresh: the unanswered one is stopped
    expect(calls[0].signal.aborted).toBe(true);
    expect(sizes.get('/a')).toBe(10);
    calls[0].answer(99); // the stopped request answers late: ignored
    const fresh = calls.slice(2);
    expect(fresh.map((c) => c.path).sort()).toEqual(['/a', '/b']);
    fresh.find((c) => c.path === '/a')!.answer(20);
    await vi.waitFor(() => expect(sizes.get('/a')).toBe(20));
    expect(sizes.get('/b')).toBeUndefined();
  });
});
