<!--
  The download action of one item (S02.4-T04). A file is a plain link, so
  the browser's download manager fetches it and "Save link as" works; a
  folder downloads as a ZIP archive. Links and special files have none.
  The action sits inside the file view's grid, which the keyboard moves
  through by itself, so it is not a Tab stop.
-->
<script lang="ts">
  import Download from '@lucide/svelte/icons/download';
  import { buttonBase, buttonVariants } from '$lib/components/Button.svelte';
  import { contentHref, downloadArchive } from './download';
  import type { FileItem } from './types';

  let { item }: { item: Pick<FileItem, 'path' | 'name' | 'kind'> } = $props();

  const classes = `${buttonBase} ${buttonVariants.ghost} size-8`;

  // A double click on the action must not also open the item.
  function keep(event: MouseEvent) {
    event.stopPropagation();
  }

  // Reports whether a click is the second of a double click (people used to
  // opening items that way), which must not start a second download.
  function repeated(event: MouseEvent): boolean {
    return event.detail > 1;
  }
</script>

{#if item.kind === 'file'}
  <a
    href={contentHref(item.path)}
    download
    class={classes}
    aria-label="Download {item.name}"
    title="Download"
    tabindex="-1"
    ondblclick={keep}
    onclick={(e) => {
      if (repeated(e)) e.preventDefault();
    }}
  >
    <Download class="size-4" aria-hidden="true" />
  </a>
{:else if item.kind === 'dir'}
  <button
    type="button"
    class={classes}
    aria-label="Download {item.name} as a ZIP file"
    title="Download as a ZIP file"
    tabindex="-1"
    ondblclick={keep}
    onclick={(e) => {
      if (!repeated(e)) void downloadArchive([item.path]);
    }}
  >
    <Download class="size-4" aria-hidden="true" />
  </button>
{/if}
