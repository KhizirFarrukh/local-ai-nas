import { describe, expect, it } from 'vitest';
import { carriesFiles, readDrop } from './drop';

// Stand-ins for the File System Entry API.
function fileEntry(name: string): FileSystemFileEntry {
  return {
    name,
    isFile: true,
    isDirectory: false,
    file: (ok: (f: File) => void) => ok(new File(['x'], name))
  } as unknown as FileSystemFileEntry;
}

/** A folder whose reader hands out children in batches, as browsers do. */
function dirEntry(name: string, children: FileSystemEntry[], batch = 2): FileSystemDirectoryEntry {
  return {
    name,
    isFile: false,
    isDirectory: true,
    createReader: () => {
      let at = 0;
      return {
        readEntries: (ok: (e: FileSystemEntry[]) => void) => {
          ok(children.slice(at, at + batch));
          at += batch;
        }
      };
    }
  } as unknown as FileSystemDirectoryEntry;
}

function transfer(items: { entry: FileSystemEntry | null; file: File | null }[]): DataTransfer {
  return {
    items: items.map(({ entry, file }) => ({
      kind: 'file',
      webkitGetAsEntry: () => entry,
      getAsFile: () => file
    }))
  } as unknown as DataTransfer;
}

describe('readDrop', () => {
  it('walks dropped folders, reading every batch, and keeps empty folders', async () => {
    const tree = dirEntry('trip', [
      fileEntry('a.jpg'),
      fileEntry('b.jpg'),
      fileEntry('c.jpg'), // a third batch entry: readEntries must be called again
      dirEntry('empty', []),
      dirEntry('day 2', [fileEntry('d.jpg')])
    ]);
    const got = await readDrop(
      transfer([
        { entry: tree, file: null },
        { entry: fileEntry('top.txt'), file: null }
      ])
    );
    expect(got.files.map((p) => p.relativePath)).toEqual([
      'trip/a.jpg',
      'trip/b.jpg',
      'trip/c.jpg',
      'trip/day 2/d.jpg',
      'top.txt'
    ]);
    expect(got.emptyFolders).toEqual(['trip/empty']);
  });

  it('falls back to the plain file when a browser gives no entry', async () => {
    const file = new File(['x'], 'plain.txt');
    const got = await readDrop(transfer([{ entry: null, file }]));
    expect(got).toEqual({ files: [{ file, relativePath: 'plain.txt' }], emptyFolders: [] });
  });

  it('skips items that are neither files nor folders', async () => {
    const odd = { name: 'odd', isFile: false, isDirectory: false } as FileSystemEntry;
    const got = await readDrop(transfer([{ entry: odd, file: null }]));
    expect(got).toEqual({ files: [], emptyFolders: [] });
  });
});

describe('carriesFiles', () => {
  it('is true only for drags with files', () => {
    const drag = (types: string[]) => ({ dataTransfer: { types } }) as unknown as DragEvent;
    expect(carriesFiles(drag(['Files']))).toBe(true);
    expect(carriesFiles(drag(['text/plain']))).toBe(false);
    expect(carriesFiles({ dataTransfer: null } as unknown as DragEvent)).toBe(false);
  });
});
