import { describe, expect, it } from 'vitest';
import { Selection, type ItemSource } from './selection.svelte';
import type { FileItem } from './types';

const make = (i: number): FileItem => ({
  path: `/f/${i}`,
  name: String(i),
  kind: 'file',
  size: i,
  mod_time: '2026-09-28T00:00:00Z',
  added_time: '2026-09-28T00:00:00Z'
});

/** A source of n items whose loaded range grows by loadRange. */
function source(n: number, loadedUpTo = n - 1, fail = false) {
  let upTo = loadedUpTo;
  const loads: [number, number][] = [];
  const src: ItemSource = {
    total: n,
    at: (i) => (i >= 0 && i <= upTo && i < n ? make(i) : undefined),
    loadRange: async (first, last) => {
      loads.push([first, last]);
      if (fail) return false;
      upTo = Math.max(upTo, last);
      return true;
    }
  };
  return { src, loads };
}

const names = (s: Selection) =>
  [...s.items.keys()].map((p) => Number(p.slice(3))).sort((a, b) => a - b);

describe('Selection', () => {
  it('selects one item, toggles, and ranges from the anchor', async () => {
    const { src } = source(10);
    const s = new Selection();
    await s.pick(2, 'only', src);
    expect(names(s)).toEqual([2]);
    expect(s.anchor).toBe(2);
    await s.pick(5, 'range', src);
    expect(names(s)).toEqual([2, 3, 4, 5]);
    expect(s.anchor).toBe(2); // Shift keeps the anchor
    await s.pick(8, 'toggle', src);
    await s.pick(3, 'toggle', src);
    expect(names(s)).toEqual([2, 4, 5, 8]);
    expect(s.anchor).toBe(3);
    await s.pick(0, 'add-range', src);
    expect(names(s)).toEqual([0, 1, 2, 3, 4, 5, 8]);
    expect(s.count(10)).toBe(7);
  });

  it('keeps "select all" as everything except, instantly', async () => {
    const { src } = source(50_000, 499);
    const s = new Selection();
    s.selectAll();
    expect(s.everything).toBe(true);
    expect(s.count(50_000)).toBe(50_000);
    expect(s.has('/f/49999')).toBe(true); // not loaded, still selected
    await s.pick(3, 'toggle', src);
    expect(s.has('/f/3')).toBe(false);
    expect(s.everything).toBe(false);
    expect(s.count(50_000)).toBe(49_999);
    await s.pick(3, 'toggle', src);
    expect(s.everything).toBe(true);
  });

  it('loads the pages of a range first', async () => {
    const { src, loads } = source(2000, 499);
    const s = new Selection();
    await s.pick(10, 'only', src);
    await s.pick(1500, 'range', src);
    expect(loads).toEqual([[10, 1500]]);
    expect(s.count(2000)).toBe(1491);
  });

  it('drops a range whose pages failed or that a newer pick overtook', async () => {
    const failing = source(2000, 499, true);
    const s = new Selection();
    await s.pick(10, 'only', failing.src);
    await s.pick(1500, 'range', failing.src);
    expect(names(s)).toEqual([10]);

    let release: (ok: boolean) => void = () => {};
    const slow: ItemSource = {
      total: 2000,
      at: (i) => (i < 500 ? make(i) : undefined),
      loadRange: () => new Promise((r) => (release = r))
    };
    const t = new Selection();
    await t.pick(1, 'only', slow);
    const waiting = t.pick(1500, 'range', slow);
    await t.pick(7, 'only', slow); // a newer click while the range waits
    release(true);
    await waiting;
    expect(names(t)).toEqual([7]);
  });

  it('clears, forgets gone items, and resolves the selection', async () => {
    const { src } = source(5);
    const s = new Selection();
    await s.pick(1, 'only', src);
    await s.pick(3, 'toggle', src);
    s.forget(['/f/1']);
    expect(await s.resolve(src)).toEqual([make(3)]);
    s.selectAll();
    await s.pick(2, 'toggle', src); // leave one out
    expect((await s.resolve(src))?.map((i) => i.size)).toEqual([0, 1, 3, 4]);
    s.clear();
    expect(s.count(5)).toBe(0);
    expect(s.anchor).toBeUndefined();
  });

  it('cannot resolve "select all" when a page fails', async () => {
    const { src } = source(2000, 499, true);
    const s = new Selection();
    s.selectAll();
    expect(await s.resolve(src)).toBeUndefined();
  });

  it('never counts below zero', () => {
    const s = new Selection();
    s.selectAll();
    s.items.set('/gone', make(0));
    expect(s.count(0)).toBe(0);
    expect(s.count(undefined)).toBe(0);
  });
});
