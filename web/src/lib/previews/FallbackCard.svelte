<!--
  What a preview shows for a file it cannot show (S02.6-T01, 4.2): the
  file's icon, name, type, and size, and a download button.
-->
<script lang="ts">
  import Download from '@lucide/svelte/icons/download';
  import { buttonBase, buttonVariants } from '$lib/components/Button.svelte';
  import { contentHref } from '$lib/files/download';
  import { iconFor, kindLabel } from '$lib/files/icons';
  import type { FileItem } from '$lib/files/types';
  import { formatSize } from '$lib/util/format';

  interface Props {
    item: FileItem;
    /** Why there is no preview; the default says the type has none. */
    reason?: string;
  }

  let { item, reason = 'There is no preview for this type of file.' }: Props = $props();

  const Icon = $derived(iconFor(item));
</script>

<div
  class="flex max-w-sm flex-col items-center gap-3 rounded-xl bg-surface p-8 text-center text-fg shadow-xl"
  data-testid="preview-fallback"
>
  <Icon class="size-16 text-fg-muted" aria-hidden="true" />
  <p class="font-semibold break-all">{item.name}</p>
  <p class="text-sm text-fg-muted">{kindLabel(item)} · {formatSize(item.size)}</p>
  <p class="text-sm">{reason}</p>
  <a href={contentHref(item.path)} download class="{buttonBase} {buttonVariants.primary} h-10 px-4">
    <Download class="size-4" aria-hidden="true" /> Download
  </a>
</div>
