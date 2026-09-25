// The API client (S02.1-T03, ADR-0002): openapi-fetch typed by
// schema.d.ts, which `pnpm generate` writes from api/openapi.yaml. The GUI
// uses only this public API, so login (S03) needs no screen rewrites.
import createClient, { type Client } from 'openapi-fetch';
import { ApiError } from './errors';
import type { paths } from './schema';

export type Api = Client<paths>;

export interface ApiOptions {
  /** Where the API lives; the default is the same origin. */
  baseUrl?: string;
  /** The fetch to use; tests pass their own. */
  fetch?: (input: Request) => Promise<Response>;
}

/** Creates a client. Screens use the shared `api` below. */
export function createApi(options: ApiOptions = {}): Api {
  return createClient<paths>({
    baseUrl: options.baseUrl ?? '/api/v1',
    ...(options.fetch ? { fetch: options.fetch } : {})
  });
}

/** The client for this origin. */
export const api: Api = createApi();

/** What an openapi-fetch call resolves to. */
interface Result<D> {
  data?: D;
  error?: unknown;
  response: Response;
}

/**
 * Waits for a call and returns its data, or throws an ApiError: for a
 * problem answer, for any other non-success answer, and when the server
 * cannot be reached.
 */
export async function unwrap<D>(pending: Promise<Result<D>>): Promise<D> {
  let result: Result<D>;
  try {
    result = await pending;
  } catch (cause) {
    throw ApiError.unreachable(cause);
  }
  if (result.error !== undefined || !result.response.ok) {
    throw ApiError.fromResponse(result.response, result.error);
  }
  return result.data as D;
}
