// Runs an upload plan (S02.4-T02): creates the folders, parents first,
// then queues every file. Folders that already exist are kept, so a
// folder can be uploaded into one that has it already (on_conflict
// "overwrite" returns an existing folder; files still follow their own
// conflict policy).
//
// The batch it returns (S02.4-T05, FR-216) tells the page what arrived:
// the items the upload adds to the folder it went to, and when every file
// is in place, so the page can show the one item of a single upload.
import { api, unwrap } from '$lib/api/client';
import { ConflictBatch, conflicts } from '$lib/files/conflicts.svelte';
import { announcer } from '$lib/shell/announcer.svelte';
import { tasks } from '$lib/shell/tasks.svelte';
import { basename } from '$lib/util/paths';
import { planUpload, topItem, type PickedFile } from './plan';
import { getUploader } from './state.svelte';
import type { ConflictPolicy, UploadEntry, Uploader } from './uploader.svelte';

export interface UploadBatch {
  /** The folder the upload went to. */
  base: string;
  /** How many folders were created (or found) for it. */
  folders: number;
  /**
   * Resolves when every file is in place, with the items the upload added
   * to `base` (a file's final name, as "Keep both" may change it). It stays
   * pending when a file fails, is skipped, or is cancelled.
   */
  done: Promise<string[]>;
}

export async function startUpload(
  picked: PickedFile[],
  base: string,
  emptyFolders: string[] = [],
  onConflict: ConflictPolicy = 'fail'
): Promise<UploadBatch | undefined> {
  const plan = planUpload(picked, base, emptyFolders);
  if (plan.folders.length > 0) {
    const task = tasks.start('Creating folders', { total: plan.folders.length });
    try {
      for (const [i, folder] of plan.folders.entries()) {
        await unwrap(
          api.POST('/files/folders', {
            body: { path: folder, parents: true, on_conflict: 'overwrite' }
          })
        );
        task.update(i + 1, plan.folders.length, `${i + 1} of ${plan.folders.length} folders`);
      }
      task.finish(
        plan.folders.length === 1 ? 'Created 1 folder.' : `Created ${plan.folders.length} folders.`
      );
    } catch (error) {
      task.fail(error, 'Creating folders');
      return undefined; // the files would have nowhere to go
    }
  }
  // Taken names ask the conflict dialog, with "apply to all" for this batch.
  const manager = await getUploader();
  const batch = new ConflictBatch(conflicts, plan.files.length);
  const entries = plan.files.map(({ file, target }) =>
    manager.add(file, target, onConflict, batch)
  );
  if (plan.files.length > 0) {
    const n = plan.files.length;
    const into = base === '/' ? 'Files' : basename(base);
    announcer.say(`Uploading ${n === 1 ? '1 file' : `${n} files`} to ${into}.`);
  }
  return { base, folders: plan.folders.length, done: whenDone(manager, base, plan, entries) };
}

function whenDone(
  manager: Uploader,
  base: string,
  plan: ReturnType<typeof planUpload>,
  entries: UploadEntry[]
): Promise<string[]> {
  const tops = () => {
    const found = new Set(plan.folders.map((folder) => topItem(base, folder)));
    for (const entry of entries) {
      found.add(topItem(base, entry.itemPath ?? entry.target));
    }
    return [...found];
  };
  if (entries.length === 0) {
    return Promise.resolve(tops()); // only (empty) folders
  }
  return new Promise((resolve) => {
    const stop = manager.onFinished(() => {
      if (entries.every((e) => e.status === 'done')) {
        stop();
        resolve(tops());
      } else if (entries.some((e) => e.status === 'cancelled' || e.status === 'skipped')) {
        stop(); // not all of it arrives: nothing to show
      }
    });
  });
}
