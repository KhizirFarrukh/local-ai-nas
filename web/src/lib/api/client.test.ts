import { describe, expect, it } from 'vitest';
import { fakeApi, problem } from '$lib/testing/fake-api';
import { createApi, onApiError, unwrap } from './client';
import type { ApiError } from './errors';

describe('unwrap', () => {
  it('returns the data of a success', async () => {
    const { api, calls } = fakeApi(() => ({
      body: { item: { path: '/', name: '', kind: 'dir' } }
    }));
    const got = await unwrap(api.GET('/files/items', { params: { query: { path: '/' } } }));
    expect(got.item.kind).toBe('dir');
    expect(calls[0].path).toBe('/files/items');
    expect(calls[0].query.get('path')).toBe('/');
  });

  it('throws an ApiError for a problem, and tells the listeners', async () => {
    const { api } = fakeApi(() => problem(404, 'not_found', 'no item at /x'));
    const seen: ApiError[] = [];
    const stop = onApiError((e) => seen.push(e));
    await expect(
      unwrap(api.GET('/files/items', { params: { query: { path: '/x' } } }))
    ).rejects.toMatchObject({ code: 'not_found', status: 404, message: 'no item at /x' });
    stop();
    await expect(
      unwrap(api.GET('/files/items', { params: { query: { path: '/x' } } }))
    ).rejects.toBeDefined();
    expect(seen).toHaveLength(1); // the second error came after stop()
  });

  it('throws unexpected_response for a failure that is not a problem', async () => {
    const api = createApi({
      baseUrl: 'http://nas.test/api/v1',
      fetch: async () =>
        new Response('<html>busy</html>', { status: 503, statusText: 'Service Unavailable' })
    });
    await expect(unwrap(api.GET('/system/health'))).rejects.toMatchObject({
      code: 'unexpected_response',
      status: 503
    });
  });

  it('throws unreachable when no answer came', async () => {
    const api = createApi({
      baseUrl: 'http://nas.test/api/v1',
      fetch: async () => {
        throw new TypeError('Failed to fetch');
      }
    });
    await expect(unwrap(api.GET('/system/health'))).rejects.toMatchObject({
      code: 'unreachable',
      status: 0
    });
  });
});
