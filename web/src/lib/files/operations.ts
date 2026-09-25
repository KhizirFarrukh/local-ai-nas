// File operations from the GUI (S02.5-T02, FR-007, FR-081): new folder,
// rename, move, copy, and delete, on one item or on a selection. The S01
// operations work per item, so a bulk run goes one item at a time, with a
// progress entry that can be cancelled between items, and ends with one
// notification that lists what failed. A taken name is asked about through
// `onConflict` (the conflict dialog, S02.5-T03); without it, the item fails.
import { api as sharedApi, unwrap, type Api } from '$lib/api/client';
import { ApiError } from '$lib/api/errors';
import { describe } from '$lib/api/messages';
import { tasks as sharedTasks, type Tasks } from '$lib/shell/tasks.svelte';
import { formatCount } from '$lib/util/format';
import { child, parent } from '$lib/util/paths';
import type { FileItem } from './types';

/** What the server does when the target name is taken (docs/api/conventions.md). */
export type ConflictPolicy = 'fail' | 'rename' | 'overwrite';

/** An answer to a conflict: a policy for this item, or skip it, or stop the run. */
export type ConflictChoice = ConflictPolicy | 'skip' | 'cancel';

/** What operations use; tests pass their own. */
export interface OperationDeps {
  api: Api;
  tasks: Tasks;
}

const shared: OperationDeps = { api: sharedApi, tasks: sharedTasks };

/** Creates the folder `name` in `folder`. */
export function createFolder(
  folder: string,
  name: string,
  onConflict: ConflictPolicy = 'fail',
  deps: OperationDeps = shared
): Promise<FileItem> {
  return unwrap(
    deps.api.POST('/files/folders', {
      body: { path: child(folder, name), parents: false, on_conflict: onConflict }
    })
  );
}

/** Renames an item in its folder. */
export function renameItem(
  item: Pick<FileItem, 'path'>,
  newName: string,
  onConflict: ConflictPolicy = 'fail',
  deps: OperationDeps = shared
): Promise<FileItem> {
  return unwrap(
    deps.api.POST('/files/operations/rename', {
      body: { path: item.path, new_name: newName, on_conflict: onConflict }
    })
  );
}

export type BulkKind = 'move' | 'copy' | 'delete';

const words: Record<BulkKind, { doing: string; done: string; verb: string }> = {
  move: { doing: 'Moving', done: 'Moved', verb: 'move' },
  copy: { doing: 'Copying', done: 'Copied', verb: 'copy' },
  delete: { doing: 'Deleting', done: 'Deleted', verb: 'delete' }
};

/** "“a.txt”" for one item, "3 items" for more. */
export function itemsText(items: readonly Pick<FileItem, 'name'>[]): string {
  return items.length === 1 ? `“${items[0].name}”` : `${formatCount(items.length)} items`;
}

export interface BulkFailure {
  item: FileItem;
  error: unknown;
}

export interface BulkResult {
  /** The items the operation was done for. */
  done: FileItem[];
  failed: BulkFailure[];
  /** Left alone: skipped at a conflict, already in place, or after a cancel. */
  skipped: FileItem[];
  cancelled: boolean;
}

export interface BulkOptions {
  /** Asked when an item's target name is taken (S02.5-T03). */
  onConflict?: (item: FileItem, error: ApiError) => Promise<ConflictChoice>;
  deps?: OperationDeps;
}

/** Runs one operation for one item. */
function runOne(
  kind: BulkKind,
  item: FileItem,
  folder: string,
  policy: ConflictPolicy,
  api: Api
): Promise<unknown> {
  switch (kind) {
    case 'move':
      return unwrap(
        api.POST('/files/operations/move', {
          body: { from: item.path, to: child(folder, item.name), on_conflict: policy }
        })
      );
    case 'copy':
      return unwrap(
        api.POST('/files/operations/copy', {
          body: { from: item.path, to: child(folder, item.name), on_conflict: policy }
        })
      );
    case 'delete':
      return unwrap(
        api.DELETE('/files/items', {
          params: { query: { path: item.path, recursive: item.kind === 'dir' } }
        })
      );
  }
}

/**
 * Moves or copies items into `folder`, or deletes them (`folder` unused),
 * one at a time. Moving an item to the folder it is in does nothing.
 */
export async function runBulk(
  kind: BulkKind,
  items: readonly FileItem[],
  folder: string,
  options: BulkOptions = {}
): Promise<BulkResult> {
  const deps = options.deps ?? shared;
  const w = words[kind];
  const result: BulkResult = { done: [], failed: [], skipped: [], cancelled: false };
  const task = deps.tasks.start(`${w.doing} ${itemsText(items)}`, {
    total: items.length,
    cancel: () => (result.cancelled = true)
  });
  for (const [i, item] of items.entries()) {
    if (result.cancelled) {
      result.skipped.push(item);
      continue;
    }
    task.update(
      i,
      items.length,
      `${formatCount(i + 1)} of ${formatCount(items.length)}: ${item.name}`
    );
    if (kind === 'move' && parent(item.path) === folder) {
      result.skipped.push(item); // already there
      continue;
    }
    let policy: ConflictPolicy = 'fail';
    for (;;) {
      try {
        await runOne(kind, item, folder, policy, deps.api);
        result.done.push(item);
      } catch (error) {
        const conflict = error instanceof ApiError && error.code === 'conflict';
        if (conflict && policy === 'fail' && options.onConflict) {
          const choice = await options.onConflict(item, error);
          if (choice === 'skip' || choice === 'cancel') {
            result.skipped.push(item);
            result.cancelled ||= choice === 'cancel';
          } else {
            policy = choice;
            continue; // again, with the chosen policy
          }
        } else {
          result.failed.push({ item, error });
        }
      }
      break;
    }
  }
  task.update(items.length, items.length);
  summarize(kind, items, result, task);
  return result;
}

/** Ends the task with one notification: what was done, and what failed and why. */
function summarize(
  kind: BulkKind,
  items: readonly FileItem[],
  result: BulkResult,
  task: ReturnType<Tasks['start']>
): void {
  const w = words[kind];
  const done =
    result.done.length === items.length
      ? `${w.done} ${itemsText(items)}.`
      : `${w.done} ${formatCount(result.done.length)} of ${formatCount(items.length)} items.`;
  // Items skipped by choice (or already in place); after a stop, the rest
  // count as stopped instead.
  const skipped =
    result.skipped.length > 0 && !result.cancelled
      ? ` ${formatCount(result.skipped.length)} skipped.`
      : '';
  if (result.failed.length === 0) {
    const stopped = result.cancelled ? ` Stopped before the rest.` : '';
    const nothing = result.done.length === 0 && result.skipped.length > 0 && !result.cancelled;
    task.summary(result.cancelled ? 'cancelled' : 'done', {
      kind: result.cancelled || nothing ? 'info' : 'success',
      message: nothing
        ? `Skipped ${itemsText(result.skipped)}: nothing was ${w.done.toLowerCase()}.`
        : done + skipped + stopped
    });
    return;
  }
  const shown = result.failed.slice(0, 8).map(({ item, error }) => {
    const m = describe(error);
    return `“${item.name}”: ${m.title}. ${m.message}`;
  });
  if (result.failed.length > shown.length) {
    shown.push(`and ${formatCount(result.failed.length - shown.length)} more`);
  }
  const tooLarge = result.failed.some(
    ({ error }) => error instanceof ApiError && error.code === 'too_large_for_sync'
  );
  if (kind === 'copy' && tooLarge) {
    shown.push('A move has no such limit: it only changes where the items are listed.');
  }
  const failed =
    result.failed.length === 1
      ? `1 item could not be ${w.done.toLowerCase()}.`
      : `${formatCount(result.failed.length)} items could not be ${w.done.toLowerCase()}.`;
  task.summary('failed', {
    kind: 'error',
    message:
      result.done.length > 0 || result.skipped.length > 0
        ? `${done}${skipped} ${failed}`
        : `Could not ${w.verb} ${itemsText(items)}.`,
    detail: shown.join('\n')
  });
}
