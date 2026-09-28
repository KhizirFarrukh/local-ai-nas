import { describe, expect, it } from 'vitest';
import { FolderListing, pageSize, type PageLoader } from './listing.svelte';
import type { FileItem } from './types';

/** A folder of n items served in pages, with the requests recorded. */
function folder(n: number, options: { fail?: (offset: number) => boolean } = {}) {
  const requests: number[] = [];
  const items = Array.from({ length: n }, (_, i): FileItem => ({
    path: `/f/${String(i).padStart(6, '0')}`,
    name: String(i).padStart(6, '0'),
    kind: 'file',
    size: i,
    mod_time: '2026-09-28T00:00:00Z'
  }));
  const load: PageLoader = async ({ offset, limit }) => {
    requests.push(offset);
    if (options.fail?.(offset)) {
      throw new Error(`page at ${offset} failed`);
    }
    return {
      item: { path: '/f', name: 'f', kind: 'dir', size: 0, mod_time: '2026-09-28T00:00:00Z' },
      items: items.slice(offset, offset + limit),
      total: n
    };
  };
  return { load, requests, items };
}

describe('FolderListing', () => {
  it('loads the first page, then pages on demand, each once', async () => {
    const f = folder(3 * pageSize + 10);
    const listing = new FolderListing(f.load, '/f');
    await listing.start();
    expect(listing.total).toBe(3 * pageSize + 10);
    expect(listing.folder?.name).toBe('f');
    expect(listing.at(0)?.name).toBe('000000');
    expect(listing.at(pageSize)).toBeUndefined();

    expect(await listing.loadRange(pageSize - 1, 2 * pageSize + 1)).toBe(true);
    expect(listing.at(2 * pageSize + 1)?.size).toBe(2 * pageSize + 1);
    // Asking again, or asking while a page is on its way, loads nothing new.
    listing.ensure(pageSize, pageSize + 5);
    await listing.loadRange(0, 2 * pageSize);
    expect(f.requests).toEqual([0, pageSize, 2 * pageSize]);
  });

  it('shares one request for a page asked for twice at once', async () => {
    const f = folder(2 * pageSize);
    const listing = new FolderListing(f.load, '/f');
    await listing.start();
    await Promise.all([
      listing.loadRange(pageSize, pageSize),
      listing.loadRange(pageSize + 1, pageSize + 2)
    ]);
    expect(f.requests).toEqual([0, pageSize]);
  });

  it('clamps a range past the end and loads nothing before the first page', async () => {
    const f = folder(10);
    const listing = new FolderListing(f.load, '/f');
    expect(await listing.loadRange(0, 5)).toBe(false); // no total yet
    await listing.start();
    expect(await listing.loadRange(-5, 1000)).toBe(true);
    expect(f.requests).toEqual([0]);
  });

  it('shows the error of the first page, and reports later failures', async () => {
    const broken = new FolderListing(folder(10, { fail: () => true }).load, '/f');
    await broken.start();
    expect(broken.error).toBeInstanceOf(Error);
    expect(broken.total).toBeUndefined();

    const f = folder(2 * pageSize, { fail: (offset) => offset > 0 });
    const listing = new FolderListing(f.load, '/f');
    await listing.start();
    expect(await listing.loadRange(pageSize, pageSize)).toBe(false);
    expect(listing.error).toBeUndefined(); // only the first page's error is shown
  });

  it('refreshes the pages on screen in one swap', async () => {
    const f = folder(2 * pageSize);
    const listing = new FolderListing(f.load, '/f');
    await listing.start();
    await listing.loadRange(pageSize, pageSize);
    const before = listing.version;
    await listing.refresh();
    expect(f.requests).toEqual([0, pageSize, 0, pageSize]);
    expect(listing.version).toBe(before + 1);
    expect(listing.at(pageSize)?.size).toBe(pageSize);
  });

  it('keeps what is shown when a refresh fails', async () => {
    let fail = false;
    const f = folder(5, { fail: () => fail });
    const listing = new FolderListing(f.load, '/f');
    await listing.start();
    fail = true;
    await listing.refresh();
    expect(listing.at(4)?.size).toBe(4);
    expect(listing.error).toBeUndefined();
  });

  it('drops a page that arrives after the folder was loaded again', async () => {
    let release: () => void = () => {};
    const gate = new Promise<void>((r) => (release = r));
    const f = folder(2 * pageSize);
    const slow: PageLoader = async (q) => {
      if (q.offset === pageSize) {
        await gate;
      }
      return f.load(q);
    };
    const listing = new FolderListing(slow, '/f');
    await listing.start();
    const late = listing.loadRange(pageSize, pageSize);
    await listing.start(); // a reload while the second page is on its way
    release();
    expect(await late).toBe(false);
    expect(listing.at(pageSize)).toBeUndefined();
  });

  it('finds an item by path among the loaded pages (bug S02-B11)', async () => {
    const f = folder(2 * pageSize);
    const listing = new FolderListing(f.load, '/f');
    await listing.start();
    expect(listing.indexOf('/f/000007')).toBe(7);
    expect(listing.indexOf(`/f/${String(pageSize + 3).padStart(6, '0')}`)).toBeUndefined(); // not loaded
    await listing.loadRange(pageSize, pageSize);
    expect(listing.indexOf(`/f/${String(pageSize + 3).padStart(6, '0')}`)).toBe(pageSize + 3);
    expect(listing.indexOf('/f/missing')).toBeUndefined();
  });

  it('passes the sort to every request', async () => {
    const seen: string[] = [];
    const listing = new FolderListing(
      async (q) => {
        seen.push(`${q.path} ${q.sort} ${q.order} ${q.offset} ${q.limit}`);
        return {
          item: { path: '/x', name: 'x', kind: 'dir', size: 0, mod_time: '' },
          items: [],
          total: 0
        };
      },
      '/x',
      'size',
      'desc'
    );
    await listing.start();
    expect(seen).toEqual([`/x size desc 0 ${pageSize}`]);
    expect(listing.total).toBe(0);
  });
});
