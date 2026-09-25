<!--
  Picks the folder to move or copy items into (S02.5-T02). It browses
  folders only, through the listing API: sorted by type, folders come
  first, so reading stops at the first file even in a large folder. It
  starts in the folder on screen, and refuses a folder that is one of the
  items or inside one (a folder cannot go into itself).
-->
<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Folder from '@lucide/svelte/icons/folder';
  import { api as sharedApi, unwrap, type Api } from '$lib/api/client';
  import { describe } from '$lib/api/messages';
  import Button from '$lib/components/Button.svelte';
  import Dialog from '$lib/components/Dialog.svelte';
  import Spinner from '$lib/components/Spinner.svelte';
  import { basename, parent, segments, join } from '$lib/util/paths';
  import type { FileItem } from './types';

  interface Props {
    open: boolean;
    title: string;
    /** The button that picks, such as "Move here". */
    submitLabel: string;
    /** Where browsing starts. */
    start: string;
    /** The items to move or copy. */
    items: readonly FileItem[];
    /** True for a move: the items' own folder is no target then. */
    move?: boolean;
    onpick: (folder: string) => void;
    api?: Api;
  }

  let {
    open = $bindable(false),
    title,
    submitLabel,
    start,
    items,
    move = false,
    onpick,
    api = sharedApi
  }: Props = $props();

  let at = $state('/');
  let folders = $state.raw<FileItem[]>([]);
  let loading = $state(false);
  let error = $state<string | undefined>(undefined);
  let generation = 0;

  $effect(() => {
    if (open) {
      at = start;
    }
  });

  $effect(() => {
    if (open) {
      void load(at);
    }
  });

  /** Reads the subfolders of `folder`, page by page, until the first file. */
  async function load(folder: string) {
    const mine = ++generation;
    loading = true;
    error = undefined;
    const found: FileItem[] = [];
    try {
      let cursor: string | undefined;
      do {
        const page = await unwrap(
          api.GET('/files/items', {
            params: { query: { path: folder, sort: 'kind', order: 'asc', limit: 1000, cursor } }
          })
        );
        const items = page.items ?? [];
        found.push(...items.filter((item) => item.kind === 'dir'));
        const lastIsDir = items.length > 0 && items[items.length - 1].kind === 'dir';
        cursor = lastIsDir ? page.next_cursor : undefined;
      } while (cursor);
      if (mine === generation) {
        folders = found;
      }
    } catch (err) {
      if (mine === generation) {
        const m = describe(err);
        error = `${m.title}. ${m.message}`;
        folders = [];
      }
    } finally {
      if (mine === generation) {
        loading = false;
      }
    }
  }

  /** Why `folder` cannot take the items, or undefined when it can. */
  function refusal(folder: string): string | undefined {
    const inside = items.find(
      (item) => item.kind === 'dir' && (folder === item.path || folder.startsWith(item.path + '/'))
    );
    if (inside) {
      return `“${inside.name}” cannot go into itself.`;
    }
    if (move && items.every((item) => parent(item.path) === folder)) {
      return items.length === 1
        ? 'It is already in this folder.'
        : 'They are already in this folder.';
    }
    return undefined;
  }

  const why = $derived(refusal(at));
  const trail = $derived(segments(at));

  function pick() {
    if (!why) {
      open = false;
      onpick(at);
    }
  }
</script>

<Dialog bind:open {title} size="md" persistent>
  <nav aria-label="Folder" class="flex flex-wrap items-center gap-1 text-sm">
    <button type="button" class="rounded px-1 hover:bg-surface-2" onclick={() => (at = '/')}
      >Files</button
    >
    {#each trail as name, i (i)}
      <ChevronRight class="size-3.5 text-fg-muted" aria-hidden="true" />
      <button
        type="button"
        class="rounded px-1 hover:bg-surface-2 {i === trail.length - 1 ? 'font-semibold' : ''}"
        aria-current={i === trail.length - 1 ? 'location' : undefined}
        onclick={() => (at = join(trail.slice(0, i + 1)))}>{name}</button
      >
    {/each}
  </nav>
  <div
    class="flex h-64 flex-col overflow-auto rounded-md border border-border"
    data-testid="folder-picker"
  >
    {#if loading}
      <div class="flex flex-1 items-center justify-center"><Spinner label="Loading folders" /></div>
    {:else if error}
      <p class="p-4 text-sm text-danger">{error}</p>
    {:else if folders.length === 0}
      <p class="p-4 text-sm text-fg-muted">No folders in {at === '/' ? 'Files' : basename(at)}.</p>
    {:else}
      <ul>
        {#each folders as folder (folder.path)}
          <li>
            <button
              type="button"
              class="flex w-full items-center gap-3 px-3 py-2 text-left text-sm hover:bg-surface-2"
              onclick={() => (at = folder.path)}
            >
              <Folder class="size-5 shrink-0 text-accent" aria-hidden="true" />
              <span class="truncate">{folder.name}</span>
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
  <p class="text-sm {why ? 'text-danger' : 'text-fg-muted'}" aria-live="polite">
    {why ?? `Into ${at === '/' ? 'Files' : `“${basename(at)}”`}.`}
  </p>
  {#snippet actions()}
    <Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
    <Button variant="primary" disabled={!!why || loading} onclick={pick}>{submitLabel}</Button>
  {/snippet}
</Dialog>
