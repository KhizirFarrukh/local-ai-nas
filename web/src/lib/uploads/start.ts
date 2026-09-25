// Runs an upload plan (S02.4-T02): creates the folders, parents first,
// then queues every file. Folders that already exist are kept, so a
// folder can be uploaded into one that has it already (on_conflict
// "overwrite" returns an existing folder; files still follow their own
// conflict policy).
import { api, unwrap } from '$lib/api/client';
import { ConflictBatch, conflicts } from '$lib/files/conflicts.svelte';
import { tasks } from '$lib/shell/tasks.svelte';
import { planUpload, type PickedFile } from './plan';
import { getUploader } from './state.svelte';
import type { ConflictPolicy } from './uploader.svelte';

export async function startUpload(
  picked: PickedFile[],
  base: string,
  emptyFolders: string[] = [],
  onConflict: ConflictPolicy = 'fail'
): Promise<void> {
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
      return; // the files would have nowhere to go
    }
  }
  // Taken names ask the conflict dialog, with "apply to all" for this batch.
  const manager = await getUploader();
  const batch = new ConflictBatch(conflicts, plan.files.length);
  for (const { file, target } of plan.files) {
    manager.add(file, target, onConflict, batch);
  }
}
