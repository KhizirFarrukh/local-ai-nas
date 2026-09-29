import { describe, expect, it } from 'vitest';
import { pickedFromInput, planUpload, topItem } from './plan';

const f = (name: string) => new File(['x'], name);

describe('planUpload', () => {
  it('puts plain files straight into the folder, with no folders to create', () => {
    const a = f('a.txt');
    expect(planUpload([{ file: a, relativePath: 'a.txt' }], '/docs')).toEqual({
      folders: [],
      files: [{ file: a, target: '/docs/a.txt' }]
    });
  });

  it('creates every folder of a tree once, parents first', () => {
    const picked = [
      { file: f('c.jpg'), relativePath: 'trip/2026/day 2/c.jpg' },
      { file: f('a.jpg'), relativePath: 'trip/2026/a.jpg' },
      { file: f('b.jpg'), relativePath: 'trip/b.jpg' }
    ];
    const plan = planUpload(picked, '/');
    expect(plan.folders).toEqual(['/trip', '/trip/2026', '/trip/2026/day 2']);
    expect(plan.files.map((x) => x.target)).toEqual([
      '/trip/2026/day 2/c.jpg',
      '/trip/2026/a.jpg',
      '/trip/b.jpg'
    ]);
  });

  it('keeps empty folders of the tree', () => {
    const plan = planUpload([{ file: f('a'), relativePath: 't/a' }], '/x', ['t/empty/deeper']);
    expect(plan.folders).toEqual(['/x/t', '/x/t/empty', '/x/t/empty/deeper']);
  });

  it('orders folders of the same depth by name', () => {
    const plan = planUpload(
      [
        { file: f('1'), relativePath: 'b/1' },
        { file: f('2'), relativePath: 'a/2' }
      ],
      '/'
    );
    expect(plan.folders).toEqual(['/a', '/b']);
  });
});

describe('pickedFromInput', () => {
  it('uses the folder path of a folder pick, else the name', () => {
    const inFolder = f('a.jpg');
    Object.defineProperty(inFolder, 'webkitRelativePath', { value: 'trip/a.jpg' });
    const plain = f('b.jpg');
    expect(pickedFromInput([inFolder, plain])).toEqual([
      { file: inFolder, relativePath: 'trip/a.jpg' },
      { file: plain, relativePath: 'b.jpg' }
    ]);
  });
});

describe('topItem', () => {
  it('names the item right below the folder an upload went to (S02.4-T05)', () => {
    expect(topItem('/', '/a.txt')).toBe('/a.txt');
    expect(topItem('/', '/trip/day 2/b.txt')).toBe('/trip');
    expect(topItem('/docs', '/docs/trip/a.txt')).toBe('/docs/trip');
    expect(topItem('/docs', '/docs/taken (1).txt')).toBe('/docs/taken (1).txt');
  });
});
