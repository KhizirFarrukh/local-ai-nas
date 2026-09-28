import { describe, expect, it } from 'vitest';
import { basename, child, crumbs, filesHref, join, parent, pathFromParam, segments } from './paths';

describe('area paths', () => {
  it('splits and joins names', () => {
    expect(segments('/a/b')).toEqual(['a', 'b']);
    expect(segments('//a//b/')).toEqual(['a', 'b']);
    expect(segments('/')).toEqual([]);
    expect(join(['a', 'b'])).toBe('/a/b');
    expect(join([])).toBe('/');
  });

  it('finds the parent and the last name', () => {
    expect(parent('/a/b')).toBe('/a');
    expect(parent('/a')).toBe('/');
    expect(parent('/')).toBe('/');
    expect(basename('/a/b c.txt')).toBe('b c.txt');
    expect(basename('/')).toBe('');
  });

  it('builds a child path', () => {
    expect(child('/a', 'b c')).toBe('/a/b c');
    expect(child('/', 'x')).toBe('/x');
  });
});

describe('app URLs', () => {
  it('percent-encodes every name, so # % ? and spaces survive', () => {
    expect(filesHref('/')).toBe('/files');
    expect(filesHref('/docs/a b#1%?.txt')).toBe('/files/docs/a%20b%231%25%3F.txt');
    expect(filesHref('/Fotos/été')).toBe('/files/Fotos/%C3%A9t%C3%A9');
  });

  it('turns the route parameter back into an area path', () => {
    expect(pathFromParam(undefined)).toBe('/');
    expect(pathFromParam('')).toBe('/');
    expect(pathFromParam('docs/a b#1')).toBe('/docs/a b#1');
  });

  it('makes breadcrumbs from the root down', () => {
    expect(crumbs('/a/b c')).toEqual([
      { label: 'Files', href: '/files' },
      { label: 'a', href: '/files/a' },
      { label: 'b c', href: '/files/a/b%20c' }
    ]);
    expect(crumbs('/', 'Root')).toEqual([{ label: 'Root', href: '/files' }]);
  });
});
