// What preview a file gets (S02.6-T01): from the item's media type first
// (the listing names it by extension, and the server's table differs by
// platform), then from the extension. Only kinds the app renders safely
// are named: images through <img>, audio and video through their
// elements, text as text, and PDF. Everything else gets the fallback card
// with a download button (stage document 4.2). HTML is text: its source
// is shown, never rendered (S02.6-T03).
import type { FileItem } from '$lib/files/types';

export type PreviewKind = 'image' | 'video' | 'audio' | 'text' | 'pdf' | 'none';

const byType: Record<string, PreviewKind> = {
  'image/jpeg': 'image',
  'image/png': 'image',
  'image/gif': 'image',
  'image/webp': 'image',
  'image/avif': 'image',
  'image/svg+xml': 'image',
  'image/bmp': 'image',
  'image/x-icon': 'image',
  'image/vnd.microsoft.icon': 'image',
  'video/mp4': 'video',
  'video/webm': 'video',
  'video/ogg': 'video',
  'video/quicktime': 'video',
  'audio/mpeg': 'audio',
  'audio/mp3': 'audio',
  'audio/ogg': 'audio',
  'audio/opus': 'audio',
  'audio/wav': 'audio',
  'audio/wave': 'audio',
  'audio/x-wav': 'audio',
  'audio/webm': 'audio',
  'audio/aac': 'audio',
  'audio/flac': 'audio',
  'audio/x-flac': 'audio',
  'audio/mp4': 'audio',
  'audio/x-m4a': 'audio',
  'application/pdf': 'pdf',
  'application/json': 'text',
  'application/xml': 'text',
  'application/javascript': 'text',
  'application/x-javascript': 'text',
  'application/x-sh': 'text',
  'application/x-yaml': 'text',
  'application/yaml': 'text',
  'application/toml': 'text',
  'application/sql': 'text',
  'application/xhtml+xml': 'text'
};

const image = ['jpg', 'jpeg', 'jfif', 'png', 'gif', 'webp', 'avif', 'svg', 'bmp', 'ico'];
const video = ['mp4', 'm4v', 'webm', 'ogv', 'mov'];
const audio = ['mp3', 'm4a', 'aac', 'flac', 'wav', 'ogg', 'oga', 'opus', 'weba'];
const text = [
  // plain text and data
  'txt',
  'text',
  'md',
  'markdown',
  'log',
  'csv',
  'tsv',
  'json',
  'jsonl',
  'ndjson',
  'xml',
  'yaml',
  'yml',
  'toml',
  'ini',
  'conf',
  'cfg',
  'env',
  'properties',
  'srt',
  'vtt',
  'tex',
  'rst',
  'adoc',
  'diff',
  'patch',
  'sql',
  'graphql',
  'proto',
  // code
  'html',
  'htm',
  'xhtml',
  'css',
  'scss',
  'sass',
  'less',
  'js',
  'mjs',
  'cjs',
  'ts',
  'mts',
  'cts',
  'jsx',
  'tsx',
  'svelte',
  'vue',
  'sh',
  'bash',
  'zsh',
  'fish',
  'ps1',
  'psm1',
  'bat',
  'cmd',
  'py',
  'go',
  'rs',
  'c',
  'h',
  'cc',
  'cpp',
  'cxx',
  'hpp',
  'java',
  'kt',
  'kts',
  'swift',
  'rb',
  'php',
  'pl',
  'lua',
  'r',
  'dart',
  'scala',
  'cs',
  'fs',
  'vb',
  'gradle',
  'cmake',
  'mk',
  'nix',
  'zig',
  'hs',
  'ex',
  'exs',
  'erl',
  'clj',
  'el',
  'vim',
  'gitignore',
  'gitattributes',
  'editorconfig',
  'lock'
];

const byExtension: Record<string, PreviewKind> = Object.fromEntries([
  ...image.map((e) => [e, 'image'] as const),
  ...video.map((e) => [e, 'video'] as const),
  ...audio.map((e) => [e, 'audio'] as const),
  ...text.map((e) => [e, 'text'] as const),
  ['pdf', 'pdf'] as const
]);

/** Whole names that are text: build and project files without an extension. */
const byName = new Set([
  'makefile',
  'dockerfile',
  'containerfile',
  'license',
  'readme',
  'changelog'
]);

/** The extension of a name, lower case, without the dot: "b.tar.gz" → "gz". */
export function extension(name: string): string {
  const dot = name.lastIndexOf('.');
  return dot < 0 || dot === name.length - 1 ? '' : name.slice(dot + 1).toLowerCase();
}

/** The preview a file gets; folders and links get none. */
export function previewKind(item: Pick<FileItem, 'kind' | 'name' | 'mime'>): PreviewKind {
  if (item.kind !== 'file') {
    return 'none';
  }
  const type = (item.mime ?? '').split(';')[0].trim().toLowerCase();
  const known = byType[type];
  if (known) {
    return known;
  }
  if (type.startsWith('text/') || type.endsWith('+json') || type.endsWith('+xml')) {
    return 'text';
  }
  // ".gitignore" has the extension "gitignore"; "Makefile" has none.
  return (
    byExtension[extension(item.name)] ?? (byName.has(item.name.toLowerCase()) ? 'text' : 'none')
  );
}
