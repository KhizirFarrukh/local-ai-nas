// The upload manager (S02.4-T01, FR-080): files go to the tus endpoint
// (S01.4) through Uppy, three at a time. The queue shows each upload's
// progress, speed, and state, and uploads can be paused, resumed,
// cancelled, and retried. After a network drop they continue by
// themselves; after a page reload, adding the same file to the same place
// again continues where the server stopped.
import Uppy from '@uppy/core';
import Tus from '@uppy/tus';
import { ApiError, isProblem } from '$lib/api/errors';
import type { ConflictBatch } from '$lib/files/conflicts.svelte';
import { UploadsPath } from './endpoint';

export type UploadStatus =
  | 'queued'
  | 'uploading'
  | 'paused'
  | 'done'
  | 'failed'
  | 'cancelled'
  /** The name was taken and the user chose to skip it (S02.5-T03). */
  | 'skipped';
export type ConflictPolicy = 'fail' | 'rename' | 'overwrite';

/** Statuses worth retrying: 0 means no answer. */
const retryable = new Set([0, 423, 500, 502, 503, 504]);

/** The default chunk limit, when the server does not name one. */
export const defaultChunkBytes = 64 << 20;

export class UploadEntry {
  uploaded = $state(0);
  /** Bytes per second, smoothed. */
  speed = $state(0);
  status = $state<UploadStatus>('queued');
  error = $state.raw<ApiError | undefined>(undefined);
  /** Where the file ended up (Item-Path; differs with "rename"). */
  itemPath = $state<string | undefined>(undefined);
  /** What the server does with a taken name; the conflict dialog changes it. */
  policy: ConflictPolicy = 'fail';
  /** The upload's operation, which asks about taken names (S02.5-T03). */
  batch: ConflictBatch | undefined;

  private lastBytes = 0;
  private lastTime = 0;

  constructor(
    readonly id: string,
    readonly name: string,
    readonly target: string,
    readonly size: number
  ) {}

  /** Records progress and updates the speed. */
  progress(bytes: number, now: number): void {
    if (this.lastTime > 0 && now > this.lastTime) {
      const rate = ((bytes - this.lastBytes) * 1000) / (now - this.lastTime);
      this.speed = this.speed === 0 ? rate : this.speed * 0.7 + rate * 0.3;
    }
    this.lastBytes = bytes;
    this.lastTime = now;
    this.uploaded = bytes;
  }
}

type Meta = { target_path: string; on_conflict: ConflictPolicy; relativePath: string };

export class Uploader {
  entries = $state<UploadEntry[]>([]);
  private readonly uppy: Uppy<Meta, Record<string, never>>;
  // eslint-disable-next-line svelte/prefer-svelte-reactivity -- lookup only
  private readonly byId = new Map<string, UploadEntry>();
  // eslint-disable-next-line svelte/prefer-svelte-reactivity -- listeners only
  private readonly finished = new Set<(entry: UploadEntry) => void>();

  constructor(
    chunkBytes: number,
    private readonly now: () => number = () => Date.now()
  ) {
    this.uppy = new Uppy<Meta, Record<string, never>>({
      autoProceed: true,
      allowMultipleUploadBatches: true
    });
    this.uppy.use(Tus, {
      endpoint: UploadsPath,
      chunkSize: chunkBytes,
      limit: 3,
      // No data in the creation request: when the server refuses an upload
      // (409, 413, 507) it answers before reading the body and closes the
      // connection, which a browser still sending reports as a network error
      // (bug S02-B08). An empty POST always gets the answer.
      uploadDataDuringCreation: false,
      removeFingerprintOnSuccess: true,
      // Only these become tus metadata (docs/api/conventions.md).
      allowedMetaFields: ['target_path', 'on_conflict'],
      // Network drops continue by themselves; a longer outage needs Retry.
      retryDelays: [0, 1000, 3000, 5000, 10000, 20000],
      // Retry what can succeed later: no answer, busy (423), and passing
      // server trouble. Not 507 (no space) or 501: they stay (bug S02-B08).
      onShouldRetry: (err) => retryable.has(err.originalResponse?.getStatus() ?? 0),
      // The same file resumes only toward the same place: the target is
      // part of the fingerprint, so re-adding it elsewhere starts anew.
      fingerprint: (file, options) =>
        Promise.resolve(
          [
            'local-ai-nas',
            options.metadata?.target_path ?? '',
            (file as File).name,
            (file as File).size,
            (file as File).lastModified
          ].join('|')
        )
    });

    this.uppy.on('upload-progress', (file, p) => {
      const entry = file && this.byId.get(file.id);
      if (entry) {
        entry.status = entry.status === 'paused' ? 'paused' : 'uploading';
        entry.progress(p.bytesUploaded, this.now());
      }
    });
    this.uppy.on('upload-pause', (file, paused) => {
      const entry = file && this.byId.get(file.id);
      if (entry) {
        entry.status = paused ? 'paused' : 'uploading';
        entry.speed = 0;
      }
    });
    this.uppy.on('upload-success', (file, response) => {
      const entry = file && this.byId.get(file.id);
      if (!entry || !file) {
        return;
      }
      entry.status = 'done';
      entry.uploaded = entry.size;
      entry.speed = 0;
      const xhr = (response.body as { xhr?: XMLHttpRequest } | undefined)?.xhr;
      entry.itemPath = xhr?.getResponseHeader('Item-Path') ?? entry.target;
      // Let go of the file once Uppy is done with the upload. Removing it
      // right here would make the tus plugin cancel the finished upload
      // (a DELETE, answered 404; bug S02-B07).
      setTimeout(() => {
        if (this.uppy.getFile(file.id)) {
          this.uppy.removeFile(file.id);
        }
      }, 0);
      for (const listener of this.finished) {
        listener(entry);
      }
    });
    this.uppy.on('upload-error', (file, error, response) => {
      const entry = file && this.byId.get(file.id);
      if (!entry) {
        return;
      }
      entry.status = 'failed';
      entry.speed = 0;
      entry.error = uploadError(error, response);
      if (entry.error.code === 'conflict' && entry.policy === 'fail' && entry.batch) {
        void this.askConflict(entry, entry.batch);
      }
    });
  }

  /** Calls `listener` whenever an upload's file is in place; returns a function that stops it. */
  onFinished(listener: (entry: UploadEntry) => void): () => void {
    this.finished.add(listener);
    return () => this.finished.delete(listener);
  }

  /**
   * Queues a file for the path `target`. With a `batch`, a taken name asks
   * the conflict dialog; without one, the upload fails with "Already there".
   */
  add(
    file: File,
    target: string,
    onConflict: ConflictPolicy = 'fail',
    batch?: ConflictBatch
  ): UploadEntry {
    const id = this.uppy.addFile({
      name: file.name,
      type: file.type,
      data: file,
      // relativePath makes the Uppy file ID differ per target, so the same
      // file can go to two places at once.
      meta: { target_path: target, on_conflict: onConflict, relativePath: target }
    });
    const entry = new UploadEntry(id, file.name, target, file.size);
    entry.policy = onConflict;
    entry.batch = batch;
    this.byId.set(id, entry);
    this.entries.push(entry);
    return entry;
  }

  pause(entry: UploadEntry): void {
    if (entry.status === 'uploading' || entry.status === 'queued') {
      this.uppy.pauseResume(entry.id);
    }
  }

  resume(entry: UploadEntry): void {
    if (entry.status === 'paused') {
      this.uppy.pauseResume(entry.id);
    }
  }

  /** Stops an upload and removes its data from the server. */
  cancel(entry: UploadEntry): void {
    if (entry.status === 'done' || entry.status === 'cancelled' || entry.status === 'skipped') {
      return;
    }
    if (this.uppy.getFile(entry.id)) {
      this.uppy.removeFile(entry.id);
    }
    entry.status = 'cancelled';
    entry.speed = 0;
  }

  /** Asks what to do about a taken name, then skips or tries again. */
  private async askConflict(entry: UploadEntry, batch: ConflictBatch): Promise<void> {
    const choice = await batch.choose(entry.name, false, entry.target);
    if (entry.status !== 'failed') {
      return; // retried or cleared meanwhile
    }
    if (choice === 'skip' || choice === 'cancel') {
      if (this.uppy.getFile(entry.id)) {
        this.uppy.removeFile(entry.id);
      }
      entry.status = 'skipped';
      entry.error = undefined;
      return;
    }
    entry.policy = choice;
    const file = this.uppy.getFile(entry.id);
    this.uppy.setFileMeta(entry.id, { ...file.meta, on_conflict: choice });
    this.retry(entry);
  }

  retry(entry: UploadEntry): void {
    if (entry.status === 'failed' && this.uppy.getFile(entry.id)) {
      entry.status = 'uploading';
      entry.error = undefined;
      void this.uppy.retryUpload(entry.id);
    }
  }

  /** Forgets the uploads that are over (done, failed, or cancelled). */
  clearFinished(): void {
    for (const entry of this.entries) {
      if (entry.status === 'failed' && this.uppy.getFile(entry.id)) {
        this.uppy.removeFile(entry.id);
      }
      if (
        entry.status === 'done' ||
        entry.status === 'failed' ||
        entry.status === 'cancelled' ||
        entry.status === 'skipped'
      ) {
        this.byId.delete(entry.id);
      }
    }
    this.entries = this.entries.filter((e) => this.byId.has(e.id));
  }

  get active(): number {
    return this.entries.filter(
      (e) => e.status === 'queued' || e.status === 'uploading' || e.status === 'paused'
    ).length;
  }
}

/** Turns an Uppy upload error into an ApiError, from the problem body when there is one. */
export function uploadError(
  error: { message: string },
  response?: { status?: number; body?: unknown }
): ApiError {
  const xhr = (response?.body as { xhr?: XMLHttpRequest } | undefined)?.xhr;
  let body: unknown;
  try {
    body = xhr?.responseText ? JSON.parse(xhr.responseText) : undefined;
  } catch {
    body = undefined;
  }
  if (isProblem(body)) {
    return new ApiError({
      status: body.status,
      code: body.code,
      message: body.detail ?? body.title,
      rule: body.rule,
      detail: body.detail,
      correlationId: body.correlation_id
    });
  }
  if (!response?.status) {
    return ApiError.unreachable(error);
  }
  return new ApiError({
    status: response.status,
    code: 'unexpected_response',
    message: error.message
  });
}

/** Reads the server's chunk limit from the tus OPTIONS answer. */
export async function serverChunkBytes(fetcher: typeof fetch = fetch): Promise<number> {
  try {
    const response = await fetcher(UploadsPath, {
      method: 'OPTIONS',
      headers: { 'Tus-Resumable': '1.0.0' }
    });
    const value = Number(response.headers.get('Upload-Max-Chunk-Size'));
    return Number.isFinite(value) && value > 0 ? value : defaultChunkBytes;
  } catch {
    return defaultChunkBytes;
  }
}

let shared: Promise<Uploader> | undefined;

/** The upload manager of this tab, created on first use. */
export function uploader(): Promise<Uploader> {
  shared ??= serverChunkBytes().then((bytes) => new Uploader(bytes));
  return shared;
}
