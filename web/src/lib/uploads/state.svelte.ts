// The upload manager, once it exists, for the views (S02.4-T01). It is
// created on first use, after the server has named its chunk limit.
import { toasts } from '$lib/shell/toasts.svelte';
import { uploader, type Uploader } from './uploader.svelte';

class UploadsState {
  manager = $state.raw<Uploader | undefined>(undefined);
}

export const uploads = new UploadsState();

let wired = false;
let finishedInBatch = 0;

/** The upload manager, with the app's notification wired in. */
export async function getUploader(): Promise<Uploader> {
  const manager = await uploader();
  uploads.manager = manager;
  if (!wired) {
    wired = true;
    // One note per batch: when the last active upload is done.
    manager.onFinished(() => {
      finishedInBatch++;
      if (manager.active === 0) {
        toasts.push({
          kind: 'success',
          message: finishedInBatch === 1 ? 'Uploaded 1 file.' : `Uploaded ${finishedInBatch} files.`
        });
        finishedInBatch = 0;
      }
    });
  }
  return manager;
}
