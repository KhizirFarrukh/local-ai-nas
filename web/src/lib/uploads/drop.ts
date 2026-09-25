// Reads what was dropped on the page (S02.4-T02): files, and folders with
// everything inside them, through the File System Entry API. A browser
// that gives no entries (or a synthetic drop) falls back to the plain file
// list, which has files but no folders.
import type { PickedFile } from './plan';

export interface Dropped {
  files: PickedFile[];
  /** Folders with nothing inside, as relative paths. */
  emptyFolders: string[];
}

export async function readDrop(data: DataTransfer): Promise<Dropped> {
  const entries = [...data.items]
    .filter((item) => item.kind === 'file')
    .map((item) => ({ entry: item.webkitGetAsEntry?.() ?? null, file: item.getAsFile() }));
  const out: Dropped = { files: [], emptyFolders: [] };
  for (const { entry, file } of entries) {
    if (entry) {
      await walk(entry, '', out);
    } else if (file) {
      out.files.push({ file, relativePath: file.name });
    }
  }
  return out;
}

async function walk(entry: FileSystemEntry, prefix: string, out: Dropped): Promise<void> {
  const path = prefix ? `${prefix}/${entry.name}` : entry.name;
  if (entry.isFile) {
    const file = await new Promise<File>((resolve, reject) =>
      (entry as FileSystemFileEntry).file(resolve, reject)
    );
    out.files.push({ file, relativePath: path });
    return;
  }
  if (!entry.isDirectory) {
    return;
  }
  const children = await readAll((entry as FileSystemDirectoryEntry).createReader());
  if (children.length === 0) {
    out.emptyFolders.push(path);
  }
  for (const childEntry of children) {
    await walk(childEntry, path, out);
  }
}

/** readEntries returns at most about 100 entries per call; read until empty. */
async function readAll(reader: FileSystemDirectoryReader): Promise<FileSystemEntry[]> {
  const all: FileSystemEntry[] = [];
  for (;;) {
    const batch = await new Promise<FileSystemEntry[]>((resolve, reject) =>
      reader.readEntries(resolve, reject)
    );
    if (batch.length === 0) {
      return all;
    }
    all.push(...batch);
  }
}

/** Reports whether a drag carries files (not text or links). */
export function carriesFiles(event: DragEvent): boolean {
  return event.dataTransfer?.types.includes('Files') ?? false;
}
