// Plans the upload of a folder tree (S02.4-T02, FR-003): which folders to
// create, parents first, and where each file goes. Pure, so it is easy to
// test; the page runs the plan.
import { join, segments } from '$lib/util/paths';

export interface PickedFile {
  file: File;
  /** The path inside the picked tree, such as "photos/2026/a.jpg"; a plain file has just its name. */
  relativePath: string;
}

export interface UploadPlan {
  /** Folders to create, every parent before its children. */
  folders: string[];
  /** Each file with its target path. */
  files: { file: File; target: string }[];
}

/**
 * Plans an upload into `base`. `emptyFolders` are relative paths of folders
 * with nothing inside, which the tree still needs.
 */
export function planUpload(
  picked: PickedFile[],
  base: string,
  emptyFolders: string[] = []
): UploadPlan {
  const folders = new Set<string>();
  const baseNames = segments(base);
  const addFolder = (names: string[]) => {
    for (let i = 1; i <= names.length; i++) {
      folders.add(join([...baseNames, ...names.slice(0, i)]));
    }
  };
  const files = picked.map(({ file, relativePath }) => {
    const names = segments(relativePath);
    addFolder(names.slice(0, -1));
    return { file, target: join([...baseNames, ...names]) };
  });
  for (const folder of emptyFolders) {
    addFolder(segments(folder));
  }
  const ordered = [...folders].sort((a, b) => depth(a) - depth(b) || a.localeCompare(b));
  return { folders: ordered, files };
}

function depth(path: string): number {
  return segments(path).length;
}

/** The files of an `<input webkitdirectory>` or a plain file input. */
export function pickedFromInput(files: File[]): PickedFile[] {
  return files.map((file) => ({ file, relativePath: file.webkitRelativePath || file.name }));
}
