import File from '@lucide/svelte/icons/file';
import FileArchive from '@lucide/svelte/icons/file-archive';
import FileAudio from '@lucide/svelte/icons/file-audio';
import FileCode from '@lucide/svelte/icons/file-code';
import FileImage from '@lucide/svelte/icons/file-image';
import FileQuestion from '@lucide/svelte/icons/file-question';
import FileText from '@lucide/svelte/icons/file-text';
import FileVideo from '@lucide/svelte/icons/file-video';
import Folder from '@lucide/svelte/icons/folder';
import Link from '@lucide/svelte/icons/link';
import { describe, expect, it } from 'vitest';
import { extension, iconFor, kindLabel } from './icons';

describe('extension', () => {
  it('is the lower-case part after the last dot', () => {
    expect(extension('Report.PDF')).toBe('pdf');
    expect(extension('archive.tar.gz')).toBe('gz');
    expect(extension('.hidden')).toBe('');
    expect(extension('README')).toBe('');
  });
});

describe('iconFor', () => {
  it('follows the kind first', () => {
    expect(iconFor({ kind: 'dir', name: 'a.zip' })).toBe(Folder);
    expect(iconFor({ kind: 'symlink', name: 'l' })).toBe(Link);
    expect(iconFor({ kind: 'other', name: 'dev' })).toBe(FileQuestion);
  });

  it('then the media type, then the extension', () => {
    expect(iconFor({ kind: 'file', name: 'x.bin', mime: 'image/png' })).toBe(FileImage);
    expect(iconFor({ kind: 'file', name: 'x', mime: 'video/mp4' })).toBe(FileVideo);
    expect(iconFor({ kind: 'file', name: 'x', mime: 'audio/ogg' })).toBe(FileAudio);
    expect(iconFor({ kind: 'file', name: 'a.ZIP' })).toBe(FileArchive);
    expect(iconFor({ kind: 'file', name: 'main.go', mime: 'text/plain' })).toBe(FileCode);
    expect(iconFor({ kind: 'file', name: 'notes', mime: 'text/plain' })).toBe(FileText);
    expect(iconFor({ kind: 'file', name: 'blob' })).toBe(File);
  });
});

describe('kindLabel', () => {
  it('names the kind for the Type column', () => {
    expect(kindLabel({ kind: 'dir', name: 'x' })).toBe('Folder');
    expect(kindLabel({ kind: 'symlink', name: 'x' })).toBe('Link');
    expect(kindLabel({ kind: 'other', name: 'x' })).toBe('Special file');
    expect(kindLabel({ kind: 'file', name: 'a.pdf' })).toBe('PDF file');
    expect(kindLabel({ kind: 'file', name: 'Makefile' })).toBe('File');
  });
});
