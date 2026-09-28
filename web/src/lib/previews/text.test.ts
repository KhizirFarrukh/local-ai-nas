import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '$lib/api/errors';
import { readTextStart } from './text';

const encoder = new TextEncoder();

/** A fetch that answers with body, and records the request headers. */
function answering(body: Uint8Array | string, init: ResponseInit = {}) {
  const seen: Record<string, string>[] = [];
  const fetcher = vi.fn(async (_src: RequestInfo | URL, req?: RequestInit) => {
    seen.push((req?.headers ?? {}) as Record<string, string>);
    return new Response(body as BodyInit, init);
  });
  return { fetcher: fetcher as unknown as typeof fetch, seen };
}

describe('readTextStart', () => {
  it('reads a small file whole, without a Range', async () => {
    const { fetcher, seen } = answering('hello\n');
    const got = await readTextStart('/f', 6, undefined, 16, fetcher);
    expect(got).toEqual({ text: 'hello\n', partial: false });
    expect(seen[0]).toEqual({});
  });

  it('asks for the first bytes of a large file and marks it partial', async () => {
    const { fetcher, seen } = answering('0123456789abcdef', { status: 206 });
    const got = await readTextStart('/f', 1000, undefined, 16, fetcher);
    expect(seen[0]).toEqual({ Range: 'bytes=0-15' });
    expect(got).toEqual({ text: '0123456789abcdef', partial: true });
  });

  it('stops at the limit even when the server sends everything', async () => {
    const { fetcher } = answering('x'.repeat(100));
    const got = await readTextStart('/f', 100, undefined, 10, fetcher);
    expect(got.text).toBe('x'.repeat(10));
    expect(got.partial).toBe(true);
  });

  it('leaves out a character cut in two at the limit', async () => {
    // "aé" is 61 C3 A9; a limit of 2 cuts é in half.
    const { fetcher } = answering(encoder.encode('aé'), { status: 206 });
    const got = await readTextStart('/f', 3, undefined, 2, fetcher);
    expect(got.text).toBe('a');
  });

  it('shows invalid bytes as the replacement character', async () => {
    const { fetcher } = answering(new Uint8Array([0x61, 0xff, 0x62]));
    const got = await readTextStart('/f', 3, undefined, 16, fetcher);
    expect(got.text).toBe('a�b');
  });

  it('turns an error answer into an ApiError', async () => {
    const problem = JSON.stringify({ status: 404, code: 'not_found', title: 'Not found' });
    const { fetcher } = answering(problem, { status: 404 });
    await expect(readTextStart('/f', 3, undefined, 16, fetcher)).rejects.toMatchObject({
      code: 'not_found'
    });
    const { fetcher: html } = answering('<html>', { status: 500 });
    await expect(readTextStart('/f', 3, undefined, 16, html)).rejects.toMatchObject({
      code: 'unexpected_response'
    });
  });

  it('reports an unreachable server, but passes an abort on unchanged', async () => {
    const down = (async () => {
      throw new TypeError('Failed to fetch');
    }) as unknown as typeof fetch;
    await expect(readTextStart('/f', 3, undefined, 16, down)).rejects.toBeInstanceOf(ApiError);

    const controller = new AbortController();
    controller.abort();
    const aborted = new DOMException('aborted', 'AbortError');
    const abortingFetch = (async () => {
      throw aborted;
    }) as unknown as typeof fetch;
    await expect(readTextStart('/f', 3, controller.signal, 16, abortingFetch)).rejects.toBe(
      aborted
    );
  });
});
