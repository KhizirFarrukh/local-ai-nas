// The file browser's keyboard shortcuts (S02.5-T04, stage document 4.2).
// `keyAction` maps a key press to an action; the file view handles moving
// and selecting itself (arrows, Enter, Space, Ctrl/Cmd+A, Escape). The same
// table feeds the "?" help list, so the two cannot drift apart.
//
// New folder is Shift+N, not the design's Ctrl/Cmd+Shift+N: Chromium
// browsers keep that for a private window, and pages never see it.

export type KeyAction =
  'rename' | 'delete' | 'new-folder' | 'parent' | 'copy' | 'cut' | 'paste' | 'help' | 'menu';

export interface KeyInput {
  key: string;
  ctrlKey: boolean;
  metaKey: boolean;
  shiftKey: boolean;
  altKey: boolean;
}

/** Reports whether the platform uses Cmd for shortcuts. */
export function isMac(platform: string = globalThis.navigator?.userAgent ?? ''): boolean {
  return /Mac|iPhone|iPad/.test(platform);
}

/** The action of a key press, or undefined when it has none. */
export function keyAction(e: KeyInput, mac: boolean): KeyAction | undefined {
  const mod = mac ? e.metaKey : e.ctrlKey;
  const other = mac ? e.ctrlKey : e.metaKey;
  const key = e.key.length === 1 ? e.key.toLowerCase() : e.key;
  if (other) {
    return undefined;
  }
  if (mod) {
    if (e.altKey) {
      return undefined;
    }
    if (e.shiftKey) {
      return undefined;
    }
    const withMod: Record<string, KeyAction> = { c: 'copy', x: 'cut', v: 'paste' };
    // Cmd+Backspace deletes in the Mac's Finder.
    return withMod[key] ?? (mac && key === 'Backspace' ? 'delete' : undefined);
  }
  if (e.altKey) {
    return key === 'ArrowUp' && !e.shiftKey ? 'parent' : undefined;
  }
  if (e.shiftKey) {
    const shifted: Record<string, KeyAction> = { F10: 'menu', '?': 'help', n: 'new-folder' };
    return shifted[key];
  }
  const plain: Record<string, KeyAction> = {
    F2: 'rename',
    Delete: 'delete',
    Backspace: 'parent',
    ContextMenu: 'menu',
    '?': 'help'
  };
  return plain[key];
}

/** The shortcuts, for the help list and the menus. */
export function shortcuts(mac: boolean): { keys: string; what: string }[] {
  const mod = mac ? '⌘' : 'Ctrl';
  return [
    { keys: '↑ ↓ ← →, Home, End, Page Up, Page Down', what: 'Move, and select the item there' },
    { keys: `Shift + a move`, what: 'Select a range' },
    { keys: `${mod} + a move`, what: 'Move without changing the selection' },
    { keys: 'Space', what: 'Select or unselect the item' },
    { keys: `${mod}+A`, what: 'Select all' },
    { keys: 'Escape', what: 'Clear the selection' },
    { keys: 'Enter', what: 'Open' },
    { keys: 'Alt+↑ or Backspace', what: 'Go to the parent folder' },
    { keys: 'Shift+N', what: 'New folder' },
    { keys: 'F2', what: 'Rename' },
    { keys: mac ? 'Delete or ⌘+Backspace' : 'Delete', what: 'Delete (asks first)' },
    { keys: `${mod}+C, ${mod}+X`, what: 'Copy or cut, to paste in another folder' },
    { keys: `${mod}+V`, what: 'Paste into this folder' },
    { keys: 'Shift+F10 or the menu key', what: 'Open the menu of the selection' },
    { keys: '?', what: 'Show this list' }
  ];
}

/** The label of one action's shortcut, for menus: "F2", "Ctrl+C". */
export function shortcutLabel(action: KeyAction, mac: boolean): string {
  const mod = mac ? '⌘' : 'Ctrl+';
  const labels: Record<KeyAction, string> = {
    rename: 'F2',
    delete: mac ? '⌘⌫' : 'Del',
    'new-folder': mac ? '⇧N' : 'Shift+N',
    parent: 'Alt+↑',
    copy: `${mod}C`,
    cut: `${mod}X`,
    paste: `${mod}V`,
    help: '?',
    menu: 'Shift+F10'
  };
  return labels[action];
}
