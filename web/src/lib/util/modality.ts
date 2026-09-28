// Whether the keyboard is in use, for focus rings (S02.7-T02). Browsers
// show a ring after a script moves the focus by their own rules: Firefox
// only if the element that had the focus showed one. After a folder opens
// (the old view is gone) or a dialog closes, that element no longer
// counts, and keyboard users lost the ring. So the app moves the focus with
// focusFor(), which asks for the ring whenever the keyboard is in use.

let keyboard = false;

if (typeof window !== 'undefined') {
  const modifiers = new Set(['Control', 'Shift', 'Alt', 'Meta']);
  window.addEventListener(
    'keydown',
    (event) => {
      if (!modifiers.has(event.key)) {
        keyboard = true;
      }
    },
    true
  );
  window.addEventListener('pointerdown', () => (keyboard = false), true);
}

/** Reports whether the last input was a key rather than a pointer. */
export function usingKeyboard(): boolean {
  return keyboard;
}

/**
 * Focuses `el`, with a ring when the keyboard is in use. An element that
 * already has the focus without a ring (as after a dialog gave it back) is
 * focused again so the ring shows.
 */
export function focusFor(el: Element | null | undefined): void {
  if (!(el instanceof HTMLElement) || !el.isConnected) {
    return;
  }
  if (document.activeElement === el) {
    if (!keyboard || el.matches(':focus-visible')) {
      return;
    }
    el.blur();
  }
  el.focus({ focusVisible: keyboard });
}
