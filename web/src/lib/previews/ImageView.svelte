<!--
  An image preview (S02.6-T02): only ever an <img>, so an SVG's scripts and
  event handlers never run (4.2). It fits the frame; a format the browser
  cannot read, or a damaged file, shows the fallback card.
-->
<script lang="ts">
  import Spinner from '$lib/components/Spinner.svelte';
  import type { FileItem } from '$lib/files/types';

  let { item, src, onfail }: { item: FileItem; src: string; onfail: (why: string) => void } =
    $props();

  let loaded = $state(false);
</script>

{#if !loaded}
  <div class="absolute"><Spinner label="Loading the image" /></div>
{/if}
<img
  {src}
  alt={item.name}
  class="max-h-full max-w-full object-contain {loaded ? '' : 'invisible'}"
  data-testid="preview-image"
  onload={() => (loaded = true)}
  onerror={() =>
    onfail(
      'This image cannot be shown: the browser does not read its format, or the file is damaged.'
    )}
/>
