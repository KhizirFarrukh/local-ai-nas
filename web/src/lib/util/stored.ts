// Small per-browser preferences (view mode, sort), kept in localStorage.
// Storage may be blocked (private windows, policies): then the default is
// used and nothing is remembered.

const prefix = 'local-ai-nas.';

export function readStored<T extends string>(key: string, allowed: readonly T[], fallback: T): T {
  try {
    const value = globalThis.localStorage?.getItem(prefix + key);
    return value !== null && value !== undefined && (allowed as readonly string[]).includes(value)
      ? (value as T)
      : fallback;
  } catch {
    return fallback;
  }
}

export function writeStored(key: string, value: string): void {
  try {
    globalThis.localStorage?.setItem(prefix + key, value);
  } catch {
    // not remembered
  }
}
