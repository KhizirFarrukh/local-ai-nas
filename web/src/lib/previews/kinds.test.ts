import { describe, expect, it } from 'vitest';
import { extension, previewKind } from './kinds';

const file = (name: string, mime?: string) => ({ kind: 'file' as const, name, mime });

describe('previewKind', () => {
  it('gives folders, links, and special files no preview', () => {
    expect(previewKind({ kind: 'dir', name: 'a.png' })).toBe('none');
    expect(previewKind({ kind: 'symlink', name: 'a.png' })).toBe('none');
    expect(previewKind({ kind: 'other', name: 'a.png' })).toBe('none');
  });

  it('trusts a known media type, parameters and case ignored', () => {
    expect(previewKind(file('x.bin', 'image/PNG'))).toBe('image');
    expect(previewKind(file('x', 'video/mp4'))).toBe('video');
    expect(previewKind(file('x', 'audio/mpeg'))).toBe('audio');
    expect(previewKind(file('x', 'application/pdf'))).toBe('pdf');
    expect(previewKind(file('x', 'application/json; charset=utf-8'))).toBe('text');
  });

  it('shows any text type and structured suffixes as text', () => {
    expect(previewKind(file('x', 'text/x-weird'))).toBe('text');
    expect(previewKind(file('x', 'application/ld+json'))).toBe('text');
    expect(previewKind(file('x', 'application/atom+xml'))).toBe('text');
  });

  it('shows HTML as its source, never rendered (S02.6-T03)', () => {
    expect(previewKind(file('page.html', 'text/html; charset=utf-8'))).toBe('text');
    expect(previewKind(file('page.xhtml'))).toBe('text');
  });

  it('falls back to the extension, then to known whole names', () => {
    expect(previewKind(file('photo.JPG'))).toBe('image');
    expect(previewKind(file('clip.mov', 'application/octet-stream'))).toBe('video');
    expect(previewKind(file('song.flac'))).toBe('audio');
    expect(previewKind(file('doc.pdf'))).toBe('pdf');
    expect(previewKind(file('main.go'))).toBe('text');
    expect(previewKind(file('.gitignore'))).toBe('text');
    expect(previewKind(file('Makefile'))).toBe('text');
    expect(previewKind(file('LICENSE'))).toBe('text');
  });

  it('gives unknown files the fallback card', () => {
    expect(previewKind(file('setup.exe', 'application/octet-stream'))).toBe('none');
    expect(previewKind(file('archive.zip'))).toBe('none');
    expect(previewKind(file('noext'))).toBe('none');
  });
});

describe('extension', () => {
  it('takes the part after the last dot', () => {
    expect(extension('b.tar.gz')).toBe('gz');
    expect(extension('.gitignore')).toBe('gitignore');
    expect(extension('trailing.')).toBe('');
    expect(extension('none')).toBe('');
  });
});
