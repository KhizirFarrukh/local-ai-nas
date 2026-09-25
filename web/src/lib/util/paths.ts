// Paths of the files area and the app URLs that show them (S02.2-T01,
// S02.3-T04). An area path starts with "/" and uses "/" between names
// (docs/api/conventions.md). The app shows the folder /docs/reports at
// /files/docs/reports, with every name percent-encoded, so names with
// spaces, "#", "%", or "?" survive the URL.

/** The names of an area path, without empty parts: "/a/b" → ["a", "b"]. */
export function segments(path: string): string[] {
  return path.split('/').filter((s) => s !== '');
}

/** Joins names into an area path: ["a", "b"] → "/a/b"; [] → "/". */
export function join(names: readonly string[]): string {
  return '/' + names.join('/');
}

/** The folder that contains a path: "/a/b" → "/a"; "/a" → "/". */
export function parent(path: string): string {
  return join(segments(path).slice(0, -1));
}

/** The last name of a path, or "" for the root. */
export function basename(path: string): string {
  return segments(path).at(-1) ?? '';
}

/** A path inside a folder: child("/a", "b c") → "/a/b c". */
export function child(folder: string, name: string): string {
  return join([...segments(folder), name]);
}

/** The app URL of a folder in the files area. */
export function filesHref(path: string): string {
  const names = segments(path);
  return names.length === 0 ? '/files' : '/files/' + names.map(encodeURIComponent).join('/');
}

/**
 * The area path from the route parameter of /files/[...path]. SvelteKit
 * passes the parameter decoded, as names joined by "/".
 */
export function pathFromParam(param: string | undefined): string {
  return join(segments(param ?? ''));
}

/** Breadcrumbs for a folder: "Files" first, then one per name. */
export function crumbs(path: string, rootLabel = 'Files'): { label: string; href: string }[] {
  const names = segments(path);
  return [
    { label: rootLabel, href: filesHref('/') },
    ...names.map((name, i) => ({ label: name, href: filesHref(join(names.slice(0, i + 1))) }))
  ];
}
