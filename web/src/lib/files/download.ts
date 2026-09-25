// Downloads (S02.4-T04, FR-005, FR-006). The browser's own download
// manager fetches everything, so large downloads never pass through the
// page's memory and the progress is the browser's:
// - a file is a plain link to GET /files/content;
// - a folder or several items become one ZIP archive: POST /files/archives
//   checks them and returns a ticket, whose URL is then downloaded. The
//   server names it after the item, or download-<date>.zip.
import { api as sharedApi, unwrap, type Api } from '$lib/api/client';
import { tasks as sharedTasks, type Tasks } from '$lib/shell/tasks.svelte';
import { formatCount, formatSize } from '$lib/util/format';
import type { ItemSource, Selection } from './selection.svelte';
import type { FileItem } from './types';

/** The download URL of a file. */
export function contentHref(path: string): string {
  return `/api/v1/files/content?path=${encodeURIComponent(path)}`;
}

/** Reports whether an item can be downloaded: links and special files cannot. */
export function downloadable(item: Pick<FileItem, 'kind'>): boolean {
  return item.kind === 'file' || item.kind === 'dir';
}

/**
 * Hands a URL to the browser's download manager. The server's
 * Content-Disposition names the file; the download attribute keeps an
 * error answer from replacing the page.
 */
export function saveUrl(url: string): void {
  const link = document.createElement('a');
  link.href = url;
  link.download = '';
  document.body.append(link);
  link.click();
  link.remove();
}

/** What downloads use; tests pass their own. */
export interface DownloadDeps {
  api: Api;
  tasks: Tasks;
  save: (url: string) => void;
}

const shared: DownloadDeps = { api: sharedApi, tasks: sharedTasks, save: saveUrl };

/** Downloads items: one file directly, anything else as one ZIP archive. */
export async function download(
  items: readonly Pick<FileItem, 'path' | 'kind'>[],
  deps: DownloadDeps = shared
): Promise<void> {
  if (items.length === 1 && items[0].kind === 'file') {
    deps.save(contentHref(items[0].path));
    return;
  }
  await downloadArchive(
    items.map((item) => item.path),
    deps
  );
}

/**
 * Downloads what is selected in `folder` (S02.5-T01): the folder itself
 * when everything in it is selected (so any number of items works), else
 * the selected items. Links and special files are left out.
 */
export async function downloadSelection(
  selection: Selection,
  source: ItemSource,
  folder: string,
  deps: DownloadDeps = shared
): Promise<void> {
  if (selection.everything) {
    await downloadArchive([folder], deps);
    return;
  }
  const items = await selection.resolve(source);
  if (!items) {
    deps.tasks.notes.push({
      kind: 'error',
      message: 'Download: the folder’s items could not be loaded. Try again.'
    });
    return;
  }
  const wanted = items.filter(downloadable);
  if (wanted.length === 0) {
    deps.tasks.notes.push({ message: 'Links and special files cannot be downloaded.' });
    return;
  }
  await download(wanted, deps);
}

/**
 * Downloads the items at paths as one ZIP archive. While the server checks
 * them, a task shows "Preparing the download"; the notification then says
 * what the browser is downloading.
 */
export async function downloadArchive(
  paths: readonly string[],
  deps: DownloadDeps = shared
): Promise<void> {
  if (paths.length === 0) {
    return;
  }
  const task = deps.tasks.start('Preparing the download');
  try {
    const ticket = await unwrap(deps.api.POST('/files/archives', { body: { paths: [...paths] } }));
    deps.save(ticket.url);
    const items = `${formatCount(ticket.entries)} ${ticket.entries === 1 ? 'item' : 'items'}`;
    task.finish(
      `Downloading “${ticket.name}” (${items}, ${formatSize(ticket.size)}). Your browser shows the progress.`
    );
  } catch (error) {
    task.fail(error, 'Download');
  }
}
