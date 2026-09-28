import { describe, expect, it } from 'vitest';
import { Tasks } from '$lib/shell/tasks.svelte';
import { Toasts } from '$lib/shell/toasts.svelte';
import { fakeApi, problem, type FakeHandler } from '$lib/testing/fake-api';
import {
  contentHref,
  download,
  downloadArchive,
  downloadSelection,
  downloadable,
  saveUrl
} from './download';
import { Selection, type ItemSource } from './selection.svelte';
import type { FileItem } from './types';

const item = (path: string, kind: FileItem['kind'] = 'file'): FileItem => ({
  path,
  name: path.split('/').at(-1) ?? '',
  kind,
  size: 1,
  mod_time: ''
});

function setup(handler: FakeHandler) {
  const { api, calls } = fakeApi(handler);
  const toasts = new Toasts(() => 0);
  const saved: string[] = [];
  return {
    deps: { api, tasks: new Tasks(toasts), save: (u: string) => void saved.push(u) },
    calls,
    toasts,
    saved
  };
}

const ticket = {
  body: { url: '/api/v1/files/archives/t1', name: 'docs.zip', entries: 3, size: 2048 }
};

describe('downloads', () => {
  it('builds a file URL and knows what can be downloaded', () => {
    expect(contentHref('/a b/c#.txt')).toBe('/api/v1/files/content?path=%2Fa%20b%2Fc%23.txt');
    expect([
      downloadable(item('/a')),
      downloadable(item('/d', 'dir')),
      downloadable(item('/l', 'symlink'))
    ]).toEqual([true, true, false]);
  });

  it('downloads one file by its link, anything else as a ZIP', async () => {
    const { deps, calls, saved, toasts } = setup(() => ticket);
    await download([item('/a.txt')], deps);
    expect(saved).toEqual(['/api/v1/files/content?path=%2Fa.txt']);
    await download([item('/docs', 'dir')], deps);
    expect(calls[0].body).toEqual({ paths: ['/docs'] });
    expect(saved[1]).toBe('/api/v1/files/archives/t1');
    expect(toasts.items.at(-1)?.message).toBe(
      'Downloading “docs.zip” (3 items, 2 KB). Your browser shows the progress.'
    );
  });

  it('reports a refused archive', async () => {
    const { deps, toasts, saved } = setup(() =>
      problem(422, 'too_large_for_sync', 'the archive has too many items')
    );
    await downloadArchive(['/huge'], deps);
    expect(saved).toEqual([]);
    expect(toasts.items.at(-1)).toMatchObject({ kind: 'error' });
    expect(toasts.items.at(-1)?.message).toMatch(/^Download: Too much at once/);
    await downloadArchive([], deps); // nothing to do
    expect(toasts.items).toHaveLength(1);
  });

  it('downloads a selection: the folder for "select all", else the downloadable items', async () => {
    const items = [item('/f/a'), item('/f/l', 'symlink'), item('/f/d', 'dir')];
    const source: ItemSource = { total: 3, at: (i) => items[i], loadRange: async () => true };
    const { deps, calls, toasts } = setup(() => ticket);
    const all = new Selection();
    all.selectAll();
    await downloadSelection(all, source, '/f', deps);
    expect(calls[0].body).toEqual({ paths: ['/f'] });

    const some = new Selection();
    await some.pick(0, 'only', source);
    await some.pick(2, 'add-range', source);
    await downloadSelection(some, source, '/f', deps);
    expect(calls[1].body).toEqual({ paths: ['/f/a', '/f/d'] }); // the link left out

    const onlyLink = new Selection();
    await onlyLink.pick(1, 'only', source);
    await downloadSelection(onlyLink, source, '/f', deps);
    expect(toasts.items.at(-1)?.message).toBe('Links and special files cannot be downloaded.');
  });

  it('says so when the selection cannot be loaded', async () => {
    const { deps, toasts } = setup(() => ticket);
    const all = new Selection();
    all.selectAll();
    all.items.set('/f/x', item('/f/x')); // not "everything": needs the pages
    await downloadSelection(
      all,
      { total: 600, at: () => undefined, loadRange: async () => false },
      '/f',
      deps
    );
    expect(toasts.items.at(-1)).toMatchObject({ kind: 'error' });
  });

  it('hands a URL to the browser through a temporary link', () => {
    let clicked = '';
    const orig = HTMLAnchorElement.prototype.click;
    HTMLAnchorElement.prototype.click = function (this: HTMLAnchorElement) {
      clicked = `${this.getAttribute('href')} download=${this.hasAttribute('download')}`;
    };
    try {
      saveUrl('/api/v1/files/archives/t9');
    } finally {
      HTMLAnchorElement.prototype.click = orig;
    }
    expect(clicked).toBe('/api/v1/files/archives/t9 download=true');
    expect(document.querySelector('a[href$="t9"]')).toBeNull();
  });
});
