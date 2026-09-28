// Regression tests for the bugs recorded in S02 (stage document, section 12)
// that unit and component tests can reach. The others are system tests
// (tests/e2e): S02-B02, B03, B05, B06, B10.
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { activity } from '$lib/shell/activity.svelte';
import TrackInEffect from '$lib/testing/TrackInEffect.svelte';
import { fakeChunkBytes } from '$lib/testing/tus-fake';
import { Uploader } from '$lib/uploads/uploader.svelte';

interface Logged {
  method: string;
  id: string;
  bytes: number;
}

const log = async (): Promise<Logged[]> => (await fetch('/__tus/log')).json();
const file = (n: number, name: string) => new File([new Uint8Array(n)], name);
let unique = 0;

describe('S02-B04', () => {
  it('an effect that tracks failing work runs once instead of looping', async () => {
    const runs = { count: 0 };
    const errors: unknown[] = [];
    const onerror = (e: ErrorEvent) => errors.push(e.error);
    window.addEventListener('error', onerror);
    await render(TrackInEffect, { runs });
    await vi.waitFor(() => expect(activity.pending).toBe(0));
    await new Promise((r) => setTimeout(r, 100));
    window.removeEventListener('error', onerror);
    expect(runs.count).toBe(1);
    expect(errors).toEqual([]);
  });
});

describe('S02-B07', () => {
  it('a finished upload is never deleted from the server afterwards', async () => {
    const before = (await log()).length;
    const up = new Uploader(fakeChunkBytes);
    const entry = up.add(file(10, 'kept.txt'), `/b07-${++unique}/kept.txt`);
    await vi.waitFor(() => expect(entry.status).toBe('done'), { timeout: 10_000 });
    await new Promise((r) => setTimeout(r, 300)); // Uppy's clean-up runs after success
    const deletes = (await log()).slice(before).filter((l) => l.method === 'DELETE');
    expect(deletes).toEqual([]);
  });
});

describe('S02-B08', () => {
  it('creation requests carry no data, and a refusal shows at once instead of stalling', async () => {
    const before = (await log()).length;
    const up = new Uploader(fakeChunkBytes);
    const started = performance.now();
    const full = up.add(file(5000, 'nospace.bin'), `/b08-${++unique}/nospace.bin`);
    const taken = up.add(file(5000, 'taken.bin'), `/b08-${unique}/taken.bin`);
    await vi.waitFor(() => expect([full.status, taken.status]).toEqual(['failed', 'failed']), {
      timeout: 10_000
    });
    // 507 and 409 are not retried: both fail well before the first retry delay would pass.
    expect(performance.now() - started).toBeLessThan(900);
    expect(full.error?.code).toBe('insufficient_storage');
    expect(taken.error?.code).toBe('conflict');
    const posts = (await log()).slice(before).filter((l) => l.method === 'POST');
    expect(posts.length).toBeGreaterThanOrEqual(2);
    expect(posts.every((p) => p.bytes === 0)).toBe(true);
  });
});
