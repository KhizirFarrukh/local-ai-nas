// The upload manager against the fake tus server (src/lib/testing/tus-fake.ts),
// served by Vite at the app's upload path while Vitest runs.
import { describe, expect, it, vi } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';
import { ConflictBatch, ConflictResolver } from '$lib/files/conflicts.svelte';
import { announcer } from '$lib/shell/announcer.svelte';
import { toasts } from '$lib/shell/toasts.svelte';
import { fakeChunkBytes } from '$lib/testing/tus-fake';
import { startUpload } from './start';
import { getUploader, uploads } from './state.svelte';
import UploadList from './UploadList.svelte';
import {
  defaultChunkBytes,
  serverChunkBytes,
  uploadError,
  Uploader,
  type UploadEntry
} from './uploader.svelte';

interface Arrived {
  target: string;
  length: number;
  offset: number;
  itemPath?: string;
}

async function arrived(): Promise<Arrived[]> {
  return (await fetch('/__tus/uploads')).json();
}

const bytes = (n: number, name: string) => new File([new Uint8Array(n).fill(7)], name);
let unique = 0;
const path = (name: string) => `/t${++unique}/${name}`;

async function settled(entry: UploadEntry, status: UploadEntry['status'] = 'done') {
  await vi.waitFor(() => expect(entry.status).toBe(status), { timeout: 10_000 });
}

describe('Uploader', () => {
  it('uploads in chunks of the server limit and records where the file went', async () => {
    const up = new Uploader(fakeChunkBytes);
    const finished = vi.fn();
    const stop = up.onFinished(finished);
    const target = path('big.bin');
    const entry = up.add(bytes(3 * fakeChunkBytes + 10, 'big.bin'), target);
    expect(up.active).toBe(1);
    await settled(entry);
    expect(entry.uploaded).toBe(entry.size);
    expect(entry.itemPath).toBe(target);
    expect(finished).toHaveBeenCalledWith(entry);
    const server = (await arrived()).find((u) => u.target === target);
    expect(server).toMatchObject({
      length: 3 * fakeChunkBytes + 10,
      offset: 3 * fakeChunkBytes + 10
    });
    stop();
    expect(up.active).toBe(0);
  });

  it('keeps both on a taken name when the conflict answer says so', async () => {
    const up = new Uploader(fakeChunkBytes);
    const resolver = new ConflictResolver();
    const batch = new ConflictBatch(resolver, 1);
    const entry = up.add(bytes(5, 'taken.txt'), path('taken.txt'), 'fail', batch);
    await vi.waitFor(() => expect(resolver.current?.question.name).toBe('taken.txt'), {
      timeout: 10_000
    });
    resolver.answer({ choice: 'rename', all: false });
    await settled(entry);
    expect(entry.itemPath).toMatch(/taken \(1\)\.txt$/);
  });

  it('skips a taken name when asked to', async () => {
    const up = new Uploader(fakeChunkBytes);
    const resolver = new ConflictResolver();
    const entry = up.add(
      bytes(5, 'taken.md'),
      path('taken.md'),
      'fail',
      new ConflictBatch(resolver, 1)
    );
    await vi.waitFor(() => expect(resolver.current).toBeDefined(), { timeout: 10_000 });
    resolver.answer({ choice: 'skip', all: false });
    await settled(entry, 'skipped');
    expect(entry.error).toBeUndefined();
  });

  it('fails with the server problem, tells the listeners, and retries on demand', async () => {
    const up = new Uploader(fakeChunkBytes);
    const failed = vi.fn();
    up.onFailed(failed);
    const bad = up.add(bytes(5, 'CON.txt'), path('CON.txt'));
    await settled(bad, 'failed');
    expect(bad.error).toMatchObject({ code: 'invalid_name', rule: 'reserved_name' });
    expect(failed).toHaveBeenCalledWith(bad);
    // A taken name without a batch fails too, as "Already there".
    const taken = up.add(bytes(5, 'taken.csv'), path('taken.csv'));
    await settled(taken, 'failed');
    expect(taken.error?.code).toBe('conflict');
    // Retrying a refusal fails again; clearing forgets the finished ones.
    up.retry(bad);
    await settled(bad, 'failed');
    up.clearFinished();
    expect(up.entries).toHaveLength(0);
  });

  it('retries a server that is busy by itself', async () => {
    const up = new Uploader(fakeChunkBytes);
    const entry = up.add(bytes(10, 'flaky.txt'), path('flaky.txt'));
    await settled(entry);
  });

  it('pauses, resumes, and cancels', async () => {
    const up = new Uploader(fakeChunkBytes);
    const entry = up.add(bytes(40 * fakeChunkBytes, 'slow.bin'), path('slow.bin'));
    up.pause(entry);
    await settled(entry, 'paused');
    expect(up.active).toBe(1);
    up.resume(entry);
    await vi.waitFor(() => expect(['uploading', 'done']).toContain(entry.status));
    up.cancel(entry);
    expect(entry.status).toBe('cancelled');
    up.cancel(entry); // twice is harmless
    up.pause(entry); // so is pausing a cancelled upload
    expect(up.active).toBe(0);
  });
});

describe('uploadError and the chunk limit', () => {
  const xhr = (text: string) => ({
    body: { xhr: { responseText: text } as XMLHttpRequest },
    status: 413
  });

  it('reads the problem body when there is one', () => {
    const e = uploadError(
      { message: 'x' },
      xhr(
        JSON.stringify({
          status: 413,
          code: 'too_large',
          title: 'Too large',
          detail: 'at most 1 GiB',
          correlation_id: 'r1'
        })
      )
    );
    expect(e).toMatchObject({ code: 'too_large', message: 'at most 1 GiB', correlationId: 'r1' });
  });

  it('falls back to the status, or to "unreachable" without one', () => {
    expect(uploadError({ message: 'bad gateway' }, xhr('<html>'))).toMatchObject({
      code: 'unexpected_response',
      status: 413
    });
    expect(uploadError({ message: 'offline' })).toMatchObject({ code: 'unreachable' });
  });

  it('asks the server for its chunk limit, with a default', async () => {
    expect(await serverChunkBytes()).toBe(fakeChunkBytes);
    const none = (async () => new Response(null, { status: 204 })) as unknown as typeof fetch;
    expect(await serverChunkBytes(none)).toBe(defaultChunkBytes);
    const down = (async () => {
      throw new TypeError('offline');
    }) as unknown as typeof fetch;
    expect(await serverChunkBytes(down)).toBe(defaultChunkBytes);
  });
});

describe('startUpload and the shared manager', () => {
  it('creates the folders first, then queues the files, and notes the batch', async () => {
    const say = vi.spyOn(announcer, 'say');
    const base = `/s${++unique}`;
    await startUpload(
      [
        { file: bytes(3, 'a.txt'), relativePath: 'trip/a.txt' },
        { file: bytes(4, 'b.txt'), relativePath: 'trip/day 2/b.txt' }
      ],
      base,
      ['trip/empty']
    );
    const folders: string[] = await (await fetch('/__tus/folders')).json();
    expect(folders).toEqual(
      expect.arrayContaining([`${base}/trip`, `${base}/trip/day 2`, `${base}/trip/empty`])
    );
    expect(say).toHaveBeenCalledWith(`Uploading 2 files to s${unique}.`);
    const manager = await getUploader();
    expect(uploads.manager).toBe(manager);
    await vi.waitFor(
      () => expect(toasts.items.some((t) => t.message === 'Uploaded 2 files.')).toBe(true),
      {
        timeout: 10_000
      }
    );
    say.mockRestore();
  });

  it('stops when a folder cannot be created, and speaks a failed upload at once', async () => {
    await startUpload([{ file: bytes(1, 'x'), relativePath: 'fail-here/x' }], `/f${++unique}`);
    expect(toasts.items.at(-1)?.message).toMatch(
      /^Creating folders: Something went wrong on the NAS/
    );
    const say = vi.spyOn(announcer, 'say');
    await startUpload([{ file: bytes(1, 'nospace.bin'), relativePath: 'nospace.bin' }], '/');
    await vi.waitFor(
      () =>
        expect(say).toHaveBeenCalledWith(
          expect.stringMatching(/^Upload of nospace\.bin failed: Not enough space/),
          true
        ),
      {
        timeout: 10_000
      }
    );
    say.mockRestore();
  });
});

describe('UploadList', () => {
  it('shows each upload with its state and the actions that fit it', async () => {
    const up = new Uploader(fakeChunkBytes);
    const done = up.add(bytes(8, 'ok.txt'), path('ok.txt'));
    const bad = up.add(bytes(8, 'CON.log'), path('CON.log'));
    await settled(done);
    await settled(bad, 'failed');
    await render(UploadList, { manager: up });
    await expect.element(page.getByText('Done')).toBeVisible();
    await expect.element(page.getByText(/^Name not allowed\./)).toBeVisible();
    await expect.element(page.getByRole('button', { name: 'Retry CON.log' })).toBeVisible();
    await expect.element(page.getByRole('button', { name: 'Cancel CON.log' })).toBeVisible();
    expect(page.getByRole('button', { name: 'Cancel ok.txt' }).elements()).toHaveLength(0);
    await page.getByRole('button', { name: 'Cancel CON.log' }).click();
    await expect.element(page.getByText('Cancelled')).toBeVisible();
  });
});
