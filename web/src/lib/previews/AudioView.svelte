<!--
  An audio preview (S02.6-T02): the file's icon and name over the browser's
  own player, which seeks with byte ranges. A format the browser cannot
  play shows the fallback card.
-->
<script lang="ts">
  import { iconFor } from '$lib/files/icons';
  import type { FileItem } from '$lib/files/types';

  let { item, src, onfail }: { item: FileItem; src: string; onfail: (why: string) => void } =
    $props();

  const Icon = $derived(iconFor(item));
</script>

<div class="flex w-full max-w-md flex-col items-center gap-4 rounded-xl bg-white/5 p-8">
  <Icon class="size-16 text-white/70" aria-hidden="true" />
  <p class="max-w-full truncate font-semibold">{item.name}</p>
  <audio
    {src}
    controls
    preload="metadata"
    aria-label={item.name}
    class="w-full"
    data-testid="preview-audio"
    onerror={() =>
      onfail('This audio cannot be played here: this browser does not play its format.')}
  ></audio>
</div>
