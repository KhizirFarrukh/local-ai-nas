import { afterEach, describe, expect, it, vi } from 'vitest';
import { readStored, writeStored } from './stored';

function memoryStorage(): Storage {
  const data = new Map<string, string>();
  return {
    get length() {
      return data.size;
    },
    clear: () => data.clear(),
    getItem: (k) => data.get(k) ?? null,
    key: (i) => [...data.keys()][i] ?? null,
    removeItem: (k) => void data.delete(k),
    setItem: (k, v) => void data.set(k, String(v))
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('stored preferences', () => {
  it('remembers an allowed value under the app prefix', () => {
    const storage = memoryStorage();
    vi.stubGlobal('localStorage', storage);
    writeStored('view', 'grid');
    expect(storage.getItem('local-ai-nas.view')).toBe('grid');
    expect(readStored('view', ['list', 'grid'], 'list')).toBe('grid');
  });

  it('ignores a value that is not allowed', () => {
    const storage = memoryStorage();
    storage.setItem('local-ai-nas.view', 'tiles');
    vi.stubGlobal('localStorage', storage);
    expect(readStored('view', ['list', 'grid'], 'list')).toBe('list');
  });

  it('uses the default when nothing is stored or storage is missing', () => {
    vi.stubGlobal('localStorage', memoryStorage());
    expect(readStored('sort', ['name', 'size'], 'name')).toBe('name');
    vi.stubGlobal('localStorage', undefined);
    expect(readStored('sort', ['name', 'size'], 'size')).toBe('size');
    expect(() => writeStored('sort', 'size')).not.toThrow();
  });

  it('survives blocked storage', () => {
    const blocked = memoryStorage();
    blocked.getItem = () => {
      throw new DOMException('denied', 'SecurityError');
    };
    blocked.setItem = () => {
      throw new DOMException('denied', 'SecurityError');
    };
    vi.stubGlobal('localStorage', blocked);
    expect(readStored('view', ['list', 'grid'], 'list')).toBe('list');
    expect(() => writeStored('view', 'grid')).not.toThrow();
  });
});
