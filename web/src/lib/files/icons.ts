// The icon for an item, by kind, media type, and extension (S02.3-T03).
import File from '@lucide/svelte/icons/file';
import FileArchive from '@lucide/svelte/icons/file-archive';
import FileAudio from '@lucide/svelte/icons/file-audio';
import FileCode from '@lucide/svelte/icons/file-code';
import FileImage from '@lucide/svelte/icons/file-image';
import FileQuestion from '@lucide/svelte/icons/file-question';
import FileSpreadsheet from '@lucide/svelte/icons/file-spreadsheet';
import FileText from '@lucide/svelte/icons/file-text';
import FileVideo from '@lucide/svelte/icons/file-video';
import Folder from '@lucide/svelte/icons/folder';
import Link from '@lucide/svelte/icons/link';
import type { IconComponent } from '$lib/components/icon';
import type { FileItem } from './types';

const byExtension: Record<string, IconComponent> = {
  zip: FileArchive,
  '7z': FileArchive,
  rar: FileArchive,
  tar: FileArchive,
  gz: FileArchive,
  xz: FileArchive,
  csv: FileSpreadsheet,
  xls: FileSpreadsheet,
  xlsx: FileSpreadsheet,
  ods: FileSpreadsheet,
  js: FileCode,
  ts: FileCode,
  go: FileCode,
  py: FileCode,
  json: FileCode,
  html: FileCode,
  css: FileCode,
  sh: FileCode,
  md: FileText,
  txt: FileText,
  pdf: FileText,
  doc: FileText,
  docx: FileText,
  odt: FileText
};

/** The lower-case extension of a name, without the dot; "" when none. */
export function extension(name: string): string {
  const dot = name.lastIndexOf('.');
  return dot > 0 ? name.slice(dot + 1).toLowerCase() : '';
}

export function iconFor(item: Pick<FileItem, 'kind' | 'name' | 'mime'>): IconComponent {
  switch (item.kind) {
    case 'dir':
      return Folder;
    case 'symlink':
      return Link;
    case 'other':
      return FileQuestion;
  }
  const type = item.mime?.split('/')[0];
  if (type === 'image') return FileImage;
  if (type === 'video') return FileVideo;
  if (type === 'audio') return FileAudio;
  return byExtension[extension(item.name)] ?? (type === 'text' ? FileText : File);
}

/** A short name of the kind, for the Type column: "Folder", "PDF file". */
export function kindLabel(item: Pick<FileItem, 'kind' | 'name'>): string {
  switch (item.kind) {
    case 'dir':
      return 'Folder';
    case 'symlink':
      return 'Link';
    case 'other':
      return 'Special file';
  }
  const ext = extension(item.name);
  return ext ? `${ext.toUpperCase()} file` : 'File';
}
