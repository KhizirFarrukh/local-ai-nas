// Reading the start of a text file for its preview (S02.6-T03): at most
// `limit` bytes, asked for with Range, so a 1 GiB log costs 256 KiB. The
// reading stops at the limit even if a server sends the whole file. Bytes
// are decoded as UTF-8; invalid ones become U+FFFD, and a character cut in
// two at the limit is left out instead of shown broken.
import { ApiError } from '$lib/api/errors';

/** How much of a text file the preview reads (256 KiB). */
export const textLimit = 256 * 1024;

export interface TextStart {
  text: string;
  /** False when the text is the whole file. */
  partial: boolean;
}

/** Reads the start of the text file at `src` of `size` bytes. */
export async function readTextStart(
  src: string,
  size: number,
  signal?: AbortSignal,
  limit = textLimit,
  fetcher: typeof fetch = fetch
): Promise<TextStart> {
  let response: Response;
  try {
    response = await fetcher(src, {
      headers: size > limit ? { Range: `bytes=0-${limit - 1}` } : {},
      signal
    });
  } catch (cause) {
    if (signal?.aborted) {
      throw cause;
    }
    throw ApiError.unreachable(cause);
  }
  if (!response.ok) {
    let body: unknown;
    try {
      body = await response.json();
    } catch {
      body = undefined;
    }
    throw ApiError.fromResponse(response, body);
  }
  const bytes = await readAtMost(response, limit);
  const partial = bytes.length < size;
  // `stream: true` keeps back a last character that is cut in two.
  const text = new TextDecoder('utf-8').decode(bytes, { stream: partial });
  return { text, partial };
}

/** Reads a response body up to `limit` bytes, then stops the transfer. */
async function readAtMost(response: Response, limit: number): Promise<Uint8Array> {
  const reader = response.body?.getReader();
  if (!reader) {
    return new Uint8Array(await response.arrayBuffer()).subarray(0, limit);
  }
  const out = new Uint8Array(limit);
  let length = 0;
  while (length < limit) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    const take = Math.min(value.length, limit - length);
    out.set(value.subarray(0, take), length);
    length += take;
  }
  await reader.cancel().catch(() => {});
  return out.subarray(0, length);
}
