<!--
  The preview (S02.6-T01, FR-082): the file at `index` of the folder, over
  the whole window, with previous and next over the folder's files (pages
  load as needed), a download button, and close. Arrow keys step and
  Escape closes; the native dialog keeps the focus inside and gives it back
  when it closes. Each kind has its own view; a kind without one, or a
  file that fails to show, gets the fallback card.
-->
<script lang="ts" module>
  import type { Component } from 'svelte';
  import type { FileItem } from '$lib/files/types';
  import type { PreviewKind } from './kinds';
  import AudioView from './AudioView.svelte';
  import ImageView from './ImageView.svelte';
  import PdfView from './PdfView.svelte';
  import TextView from './TextView.svelte';
  import VideoView from './VideoView.svelte';

  /** The view of one kind; it calls `onfail` when the file cannot be shown. */
  export type PreviewView = Component<{
    item: FileItem;
    src: string;
    onfail: (why: string) => void;
  }>;

  /** The views by kind; kinds without one get the fallback card. */
  const views: Partial<Record<PreviewKind, PreviewView>> = {
    image: ImageView,
    video: VideoView,
    audio: AudioView,
    text: TextView,
    pdf: PdfView
  };
</script>

<script lang="ts">
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Download from '@lucide/svelte/icons/download';
  import X from '@lucide/svelte/icons/x';
  import { contentHref } from '$lib/files/download';
  import { kindLabel } from '$lib/files/icons';
  import type { FolderListing } from '$lib/files/listing.svelte';
  import { formatSize } from '$lib/util/format';
  import FallbackCard from './FallbackCard.svelte';
  import { previewKind } from './kinds';

  interface Props {
    listing: FolderListing;
    /** The position of the file in the folder. */
    index: number;
    /** Asks to show the file at another position. */
    onstep: (index: number, item: FileItem) => void;
    onclose: () => void;
  }

  let { listing, index, onstep, onclose }: Props = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  let stepping = false;

  const item = $derived(listing.at(index));
  const total = $derived(listing.total ?? 0);
  const kind = $derived(item ? previewKind(item) : 'none');
  const View = $derived(views[kind]);

  // A view's failure belongs to its file: another file starts afresh.
  let failure = $state<{ path: string; why: string } | undefined>(undefined);
  const failed = $derived(item && failure?.path === item.path ? failure.why : undefined);

  // The dialog itself takes the focus, so a screen reader names the file
  // and Enter does not press the first button.
  $effect(() => {
    dialog?.showModal();
    dialog?.focus();
  });

  // The file's page may not be loaded yet (a reload keeps the position).
  $effect(() => {
    if (!item) {
      void listing.loadRange(index, index);
    }
  });

  /** Shows the next file before or after this one, skipping folders. */
  async function step(direction: 1 | -1) {
    if (stepping) {
      return;
    }
    stepping = true;
    try {
      for (let i = index + direction; i >= 0 && i < total; i += direction) {
        if (!listing.at(i) && !(await listing.loadRange(i, i))) {
          return;
        }
        const next = listing.at(i);
        if (next?.kind === 'file') {
          onstep(i, next);
          return;
        }
      }
    } finally {
      stepping = false;
    }
  }

  // The keys are read on the window: a button that turns disabled (the
  // last page, the last file) drops the focus to the page, outside the
  // dialog, and the keys must still work.
  function keydown(event: KeyboardEvent) {
    // Media controls and fields use the arrow keys themselves.
    if ((event.target as Element).closest('video, audio, input, textarea, select')) {
      return;
    }
    if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
      event.preventDefault();
      void step(event.key === 'ArrowLeft' ? -1 : 1);
    }
  }
</script>

<svelte:window onkeydown={keydown} />

<dialog
  bind:this={dialog}
  aria-label={item ? `Preview of ${item.name}` : 'Preview'}
  class="m-0 h-dvh max-h-none w-screen max-w-none flex-col bg-neutral-950 p-0 text-white outline-none open:flex"
  tabindex="-1"
  data-testid="preview"
  {onclose}
>
  <header class="flex h-14 shrink-0 items-center gap-3 px-4">
    <div class="min-w-0 flex-1">
      <h2 class="truncate font-semibold">{item?.name ?? 'Loading…'}</h2>
      {#if item}
        <p class="truncate text-xs text-white/70">{kindLabel(item)} · {formatSize(item.size)}</p>
      {/if}
    </div>
    {#if item}
      <a
        href={contentHref(item.path)}
        download
        class="inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-medium hover:bg-white/15 pointer-coarse:h-11"
      >
        <Download class="size-4" aria-hidden="true" /> Download
      </a>
    {/if}
    <button
      type="button"
      class="inline-flex size-9 items-center justify-center rounded-md hover:bg-white/15 pointer-coarse:size-11"
      aria-label="Close the preview"
      title="Close (Escape)"
      onclick={() => dialog?.close()}
    >
      <X class="size-5" aria-hidden="true" />
    </button>
  </header>

  <div class="relative flex min-h-0 flex-1 items-center justify-center px-2 pb-4 sm:px-14">
    {#if item}
      {#key item.path}
        {#if View && !failed}
          <View
            {item}
            src={contentHref(item.path)}
            onfail={(why) => (failure = { path: item.path, why })}
          />
        {:else}
          <FallbackCard {item} reason={failed} />
        {/if}
      {/key}
    {/if}
    <button
      type="button"
      class="absolute top-1/2 left-2 inline-flex size-10 -translate-y-1/2 items-center justify-center rounded-full bg-black/40 hover:bg-white/25 disabled:opacity-30 pointer-coarse:size-11 sm:bg-white/10"
      aria-label="Previous file"
      title="Previous (←)"
      disabled={index <= 0}
      onclick={() => void step(-1)}
    >
      <ChevronLeft class="size-6" aria-hidden="true" />
    </button>
    <button
      type="button"
      class="absolute top-1/2 right-2 inline-flex size-10 -translate-y-1/2 items-center justify-center rounded-full bg-black/40 hover:bg-white/25 disabled:opacity-30 pointer-coarse:size-11 sm:bg-white/10"
      aria-label="Next file"
      title="Next (→)"
      disabled={index >= total - 1}
      onclick={() => void step(1)}
    >
      <ChevronRight class="size-6" aria-hidden="true" />
    </button>
  </div>
</dialog>
