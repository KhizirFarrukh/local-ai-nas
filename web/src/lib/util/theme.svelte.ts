// The theme choice (S02.1-T04, FR-083): follow the operating system, or a
// theme the user picked, remembered per browser. The colors themselves are
// CSS tokens in app.css; this module only sets data-theme on <html>.

export type ThemeChoice = 'system' | 'light' | 'dark';

export const themeChoices: readonly ThemeChoice[] = ['system', 'light', 'dark'];

const storageKey = 'local-ai-nas.theme';

/** The storage the choice is kept in; tests pass a fake one. */
export type ThemeStorage = Pick<Storage, 'getItem' | 'setItem'>;

export function isThemeChoice(value: unknown): value is ThemeChoice {
  return typeof value === 'string' && (themeChoices as readonly string[]).includes(value);
}

/** Reads the remembered choice; anything unknown means the system theme. */
export function readTheme(storage: ThemeStorage | undefined): ThemeChoice {
  try {
    const value = storage?.getItem(storageKey);
    return isThemeChoice(value) ? value : 'system';
  } catch {
    return 'system'; // storage can be blocked (private windows, policies)
  }
}

/** Applies a choice to an element (normally <html>). */
export function applyTheme(root: HTMLElement, choice: ThemeChoice): void {
  if (choice === 'system') {
    delete root.dataset.theme;
  } else {
    root.dataset.theme = choice;
  }
}

/** The app's theme state: read at startup, changed by the theme switch. */
export class Theme {
  choice = $state<ThemeChoice>('system');

  constructor(
    private readonly storage: ThemeStorage | undefined,
    private readonly root: HTMLElement
  ) {
    this.choice = readTheme(storage);
    applyTheme(root, this.choice);
  }

  set(choice: ThemeChoice): void {
    this.choice = choice;
    applyTheme(this.root, choice);
    try {
      this.storage?.setItem(storageKey, choice);
    } catch {
      // Not remembered, but still applied for this page.
    }
  }
}

let shared: Theme | undefined;

/** The theme of this browser tab, created on first use. */
export function theme(): Theme {
  shared ??= new Theme(globalThis.localStorage, document.documentElement);
  return shared;
}

// Deliberate lint error (S02.1-T05 acceptance), reverted in the next commit.
const lintProbe = 1;
