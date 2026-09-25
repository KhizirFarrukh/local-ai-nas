// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    interface PageState {
      /** The file shown in the preview (S02.6-T01): Back closes it. */
      preview?: { path: string; index: number };
    }
    // interface Platform {}
  }
}

export {};
