<!--
  A text or code preview (S02.6-T03): the first 256 KiB as plain text, with
  a notice when there is more. It is always text: HTML shows its source and
  is never rendered.
-->
<script lang="ts">
  import { describe } from '$lib/api/messages';
  import Spinner from '$lib/components/Spinner.svelte';
  import type { FileItem } from '$lib/files/types';
  import { formatSize } from '$lib/util/format';
  import { readTextStart, textLimit, type TextStart } from './text';

  let { item, src, onfail }: { item: FileItem; src: string; onfail: (why: string) => void } =
    $props();

  let loaded = $state<TextStart | undefined>(undefined);

  $effect(() => {
    const stop = new AbortController();
    readTextStart(src, item.size, stop.signal).then(
      (result) => (loaded = result),
      (error) => {
        if (!stop.signal.aborted) {
          const m = describe(error);
          onfail(`This file could not be read. ${m.title}. ${m.message}`);
        }
      }
    );
    return () => stop.abort();
  });
</script>

{#if !loaded}
  <Spinner label="Loading the text" />
{:else}
  <div class="flex h-full w-full max-w-5xl flex-col gap-2">
    {#if loaded.partial}
      <p class="rounded-md bg-white/10 px-3 py-2 text-sm" role="note" data-testid="preview-notice">
        Showing the first {formatSize(textLimit)} of {formatSize(item.size)}. Download the file to
        see all of it.
      </p>
    {/if}
    <!-- The text scrolls, so the keyboard must be able to reach it (WCAG 2.1.1). -->
    <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
    <pre
      class="min-h-0 flex-1 overflow-auto rounded-md bg-white p-4 font-mono text-sm break-words whitespace-pre-wrap text-neutral-900"
      tabindex="0"
      aria-label="Text of {item.name}"
      data-testid="preview-text">{loaded.text || '(This file is empty.)'}</pre>
  </div>
{/if}
