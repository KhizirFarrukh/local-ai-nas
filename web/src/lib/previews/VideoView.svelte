<!--
  A video preview (S02.6-T02, FR-019): the browser's own player on the
  file's download address, so seeking asks for byte ranges (206) instead of
  the whole file. It takes a single source on purpose: S04.8 gives it an HLS
  source and a quality menu (ADR-0020). A format or codec the browser cannot
  play shows the fallback card.
-->
<script lang="ts">
  import type { FileItem } from '$lib/files/types';

  let { item, src, onfail }: { item: FileItem; src: string; onfail: (why: string) => void } =
    $props();
</script>

<!-- A user's own videos come without caption files. -->
<!-- svelte-ignore a11y_media_has_caption -->
<video
  {src}
  controls
  preload="metadata"
  playsinline
  aria-label={item.name}
  class="max-h-full max-w-full bg-black"
  data-testid="preview-video"
  onerror={() =>
    onfail('This video cannot be played here: this browser does not play its format or codec.')}
></video>
