import { describe, expect, it } from 'vitest';
import { isMac, keyAction, shortcutLabel, shortcuts, type KeyInput } from './keys';

function key(k: string, mods: Partial<Omit<KeyInput, 'key'>> = {}): KeyInput {
  return { key: k, ctrlKey: false, metaKey: false, shiftKey: false, altKey: false, ...mods };
}

describe('isMac', () => {
  it('recognizes Apple platforms', () => {
    expect(isMac('Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0)')).toBe(true);
    expect(isMac('Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X)')).toBe(true);
    expect(isMac('Mozilla/5.0 (Windows NT 10.0; Win64; x64)')).toBe(false);
  });
});

describe('keyAction on Windows and Linux', () => {
  const pc = false;
  it.each([
    [key('F2'), 'rename'],
    [key('Delete'), 'delete'],
    [key('Backspace'), 'parent'],
    [key('ArrowUp', { altKey: true }), 'parent'],
    [key('N', { shiftKey: true }), 'new-folder'],
    [key('F10', { shiftKey: true }), 'menu'],
    [key('ContextMenu'), 'menu'],
    [key('?', { shiftKey: true }), 'help'],
    [key('?'), 'help'],
    [key('c', { ctrlKey: true }), 'copy'],
    [key('X', { ctrlKey: true }), 'cut'],
    [key('v', { ctrlKey: true }), 'paste']
  ])('%o → %s', (input, action) => {
    expect(keyAction(input, pc)).toBe(action);
  });

  it.each([
    key('n'),
    key('c', { metaKey: true }), // Cmd is not the modifier on a PC
    key('N', { ctrlKey: true, shiftKey: true }), // Chromium keeps it for a private window
    key('c', { ctrlKey: true, altKey: true }),
    key('ArrowUp', { altKey: true, shiftKey: true }),
    key('Backspace', { ctrlKey: true })
  ])('%o has no action', (input) => {
    expect(keyAction(input, pc)).toBeUndefined();
  });
});

describe('keyAction on a Mac', () => {
  it('uses Cmd, and Cmd+Backspace deletes as in the Finder', () => {
    expect(keyAction(key('c', { metaKey: true }), true)).toBe('copy');
    expect(keyAction(key('Backspace', { metaKey: true }), true)).toBe('delete');
    expect(keyAction(key('c', { ctrlKey: true }), true)).toBeUndefined();
  });
});

describe('shortcut labels', () => {
  it('lists every shortcut for the help dialog', () => {
    const list = shortcuts(false);
    expect(list.map((s) => s.what)).toContain('Rename');
    expect(list.find((s) => s.what === 'Select all')?.keys).toBe('Ctrl+A');
    expect(shortcuts(true).find((s) => s.what === 'Select all')?.keys).toBe('⌘+A');
  });

  it('labels each action for the menus', () => {
    expect(shortcutLabel('rename', false)).toBe('F2');
    expect(shortcutLabel('copy', false)).toBe('Ctrl+C');
    expect(shortcutLabel('copy', true)).toBe('⌘C');
    expect(shortcutLabel('delete', true)).toBe('⌘⌫');
    expect(shortcutLabel('new-folder', false)).toBe('Shift+N');
  });
});
