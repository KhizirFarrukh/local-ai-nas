<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { IconComponent } from './icon';

  interface Props {
    icon?: IconComponent;
    title: string;
    description?: string;
    /** Buttons for the next step, such as "Upload" or "Retry". */
    actions?: Snippet;
  }

  let { icon: Icon, title, description, actions }: Props = $props();
</script>

<div class="flex flex-col items-center justify-center gap-3 px-6 py-12 text-center">
  {#if Icon}<Icon class="size-10 text-fg-muted" aria-hidden="true" />{/if}
  <h2 class="text-base font-semibold">{title}</h2>
  {#if description}<p class="max-w-md text-sm text-fg-muted">{description}</p>{/if}
  {#if actions}<div class="mt-2 flex flex-wrap justify-center gap-2">{@render actions()}</div>{/if}
</div>
