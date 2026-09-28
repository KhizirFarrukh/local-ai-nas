import { describe, expect, it } from 'vitest';
import { Tasks } from '$lib/shell/tasks.svelte';
import { Toasts } from '$lib/shell/toasts.svelte';
import { fakeApi, problem, type FakeHandler } from '$lib/testing/fake-api';
import { createFolder, itemsText, renameItem, runBulk, type BulkKind } from './operations';
import type { FileItem } from './types';

function item(path: string, kind: FileItem['kind'] = 'file'): FileItem {
  const name = path.split('/').at(-1) ?? '';
  return { path, name, kind, size: 1, mod_time: '2026-09-28T00:00:00Z' };
}

function setup(handler: FakeHandler) {
  const { api, calls } = fakeApi(handler);
  const toasts = new Toasts(() => 0);
  const tasks = new Tasks(toasts);
  return { deps: { api, tasks }, calls, toasts, tasks };
}

describe('single operations', () => {
  it('creates a folder and renames an item with the given conflict policy', async () => {
    const { deps, calls } = setup(() => ({ status: 201, body: item('/docs/new', 'dir') }));
    await createFolder('/docs', 'new', 'rename', deps);
    await renameItem(item('/docs/a.txt'), 'b.txt', 'fail', deps);
    expect(calls.map((c) => [c.method, c.path, c.body])).toEqual([
      ['POST', '/files/folders', { path: '/docs/new', parents: false, on_conflict: 'rename' }],
      [
        'POST',
        '/files/operations/rename',
        { path: '/docs/a.txt', new_name: 'b.txt', on_conflict: 'fail' }
      ]
    ]);
  });
});

describe('itemsText', () => {
  it('names one item and counts more', () => {
    expect(itemsText([item('/a.txt')])).toBe('“a.txt”');
    expect(itemsText([item('/a'), item('/b'), item('/c')])).toBe('3 items');
  });
});

describe('runBulk', () => {
  const ok: FakeHandler = () => ({ status: 200, body: {} });

  it.each<[BulkKind, string]>([
    ['move', '/files/operations/move'],
    ['copy', '/files/operations/copy']
  ])('%s sends each item to the folder', async (kind, path) => {
    const { deps, calls, toasts } = setup(ok);
    const result = await runBulk(kind, [item('/a.txt'), item('/d', 'dir')], '/to', { deps });
    expect(result.done).toHaveLength(2);
    expect(calls.map((c) => [c.path, c.body])).toEqual([
      [path, { from: '/a.txt', to: '/to/a.txt', on_conflict: 'fail' }],
      [path, { from: '/d', to: '/to/d', on_conflict: 'fail' }]
    ]);
    expect(toasts.items.at(-1)).toMatchObject({
      kind: 'success',
      message: `${kind === 'move' ? 'Moved' : 'Copied'} 2 items.`
    });
  });

  it('deletes folders recursively and files plainly', async () => {
    const { deps, calls } = setup(() => ({ status: 204 }));
    await runBulk('delete', [item('/a.txt'), item('/d', 'dir')], '/', { deps });
    expect(calls.map((c) => [c.method, c.query.get('path'), c.query.get('recursive')])).toEqual([
      ['DELETE', '/a.txt', 'false'],
      ['DELETE', '/d', 'true']
    ]);
  });

  it('skips a move into the folder the item is already in', async () => {
    const { deps, calls, toasts } = setup(ok);
    const result = await runBulk('move', [item('/to/a.txt')], '/to', { deps });
    expect(calls).toHaveLength(0);
    expect(result.skipped).toHaveLength(1);
    expect(toasts.items.at(-1)?.message).toBe('Skipped “a.txt”: nothing was moved.');
  });

  it('asks about a taken name and retries with the chosen policy', async () => {
    const { deps, calls } = setup((req) =>
      (req.body as { on_conflict: string }).on_conflict === 'fail'
        ? problem(409, 'conflict', 'taken')
        : ok(req)
    );
    const asked: string[] = [];
    const result = await runBulk('copy', [item('/a.txt'), item('/b.txt')], '/to', {
      deps,
      onConflict: async (it) => {
        asked.push(it.name);
        return it.name === 'a.txt' ? 'rename' : 'skip';
      }
    });
    expect(asked).toEqual(['a.txt', 'b.txt']);
    expect(result.done.map((i) => i.name)).toEqual(['a.txt']);
    expect(result.skipped.map((i) => i.name)).toEqual(['b.txt']);
    expect(calls.map((c) => (c.body as { on_conflict: string }).on_conflict)).toEqual([
      'fail',
      'rename',
      'fail'
    ]);
  });

  it('stops at "cancel" and leaves the rest', async () => {
    const { deps, toasts } = setup(() => problem(409, 'conflict'));
    const result = await runBulk('move', [item('/a'), item('/b'), item('/c')], '/to', {
      deps,
      onConflict: async () => 'cancel'
    });
    expect(result.cancelled).toBe(true);
    expect(result.skipped).toHaveLength(3);
    expect(toasts.items.at(-1)).toMatchObject({
      kind: 'info',
      message: 'Moved 0 of 3 items. Stopped before the rest.'
    });
  });

  it('stops between items when the progress entry is cancelled', async () => {
    const ref: { tasks?: Tasks } = {};
    const { deps } = setup(() => {
      ref.tasks?.items[0]?.cancel?.(); // the user presses Cancel during the first item
      return { status: 200, body: {} };
    });
    ref.tasks = deps.tasks;
    const result = await runBulk('copy', [item('/a'), item('/b')], '/to', { deps });
    expect(result.done.map((i) => i.name)).toEqual(['a']);
    expect(result.skipped.map((i) => i.name)).toEqual(['b']);
    expect(deps.tasks.items).toHaveLength(0); // the task ended
  });

  it('reports failures with reasons, at most eight, and the rest counted', async () => {
    const items = Array.from({ length: 10 }, (_, i) => item(`/f${i}`));
    const { deps, toasts } = setup((req) =>
      req.query.get('path') === '/f0' ? { status: 204 } : problem(404, 'not_found')
    );
    await runBulk('delete', items, '/', { deps });
    const note = toasts.items.at(-1);
    expect(note?.kind).toBe('error');
    expect(note?.message).toBe('Deleted 1 of 10 items. 9 items could not be deleted.');
    const lines = note?.detail?.split('\n') ?? [];
    expect(lines).toHaveLength(9);
    expect(lines[0]).toBe('“f1”: Not found. It may have been moved, renamed, or deleted.');
    expect(lines[8]).toBe('and 1 more');
  });

  it('says a move has no size limit when a copy is too large', async () => {
    const { deps, toasts } = setup(() =>
      problem(422, 'too_large_for_sync', 'the copy has 20,001 items')
    );
    await runBulk('copy', [item('/big', 'dir')], '/to', { deps });
    const note = toasts.items.at(-1);
    expect(note?.message).toBe('Could not copy “big”.');
    expect(note?.detail).toMatch(/A move has no such limit/);
  });
});
