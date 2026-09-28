// Test helper (S02.8-T01): an API client whose requests go to a handler
// instead of a server. The handler sees the method, the API path, the
// query, and the JSON body, and answers with a status and a body; problem
// answers use the real Problem shape, so errors pass through ApiError as
// they do in the app.
import { createApi, type Api } from '$lib/api/client';

export interface FakeRequest {
  method: string;
  /** The path below /api/v1, such as "/files/items". */
  path: string;
  query: URLSearchParams;
  body: unknown;
}

export interface FakeAnswer {
  status?: number;
  body?: unknown;
  headers?: Record<string, string>;
}

export type FakeHandler = (req: FakeRequest) => FakeAnswer | Promise<FakeAnswer>;

export interface FakeApi {
  api: Api;
  /** Every request, in order. */
  calls: FakeRequest[];
}

export function fakeApi(handler: FakeHandler): FakeApi {
  const calls: FakeRequest[] = [];
  const api = createApi({
    baseUrl: 'http://nas.test/api/v1',
    fetch: async (request: Request) => {
      const url = new URL(request.url);
      const text = await request.text();
      const req: FakeRequest = {
        method: request.method,
        path: url.pathname.replace(/^\/api\/v1/, ''),
        query: url.searchParams,
        body: text ? JSON.parse(text) : undefined
      };
      calls.push(req);
      const answer = await handler(req);
      const status = answer.status ?? 200;
      const noBody = status === 204 || answer.body === undefined;
      return new Response(noBody ? null : JSON.stringify(answer.body), {
        status,
        headers: {
          'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json',
          ...answer.headers
        }
      });
    }
  });
  return { api, calls };
}

/** A problem answer (docs/api/errors.md). */
export function problem(status: number, code: string, detail?: string): FakeAnswer {
  return { status, body: { status, code, title: code, ...(detail ? { detail } : {}) } };
}
