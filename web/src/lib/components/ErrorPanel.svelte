<!-- An error in place of content that could not be loaded, with a retry. -->
<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import { describe } from '$lib/api/messages';
  import Button from './Button.svelte';

  interface Props {
    error: unknown;
    onretry?: () => void;
  }

  let { error, onretry }: Props = $props();
  const message = $derived(describe(error));
</script>

<div role="alert" class="flex flex-col items-center gap-3 px-6 py-12 text-center">
  <CircleAlert class="size-10 text-danger" aria-hidden="true" />
  <h2 class="text-base font-semibold">{message.title}</h2>
  <p class="max-w-md text-sm text-fg-muted">{message.message}</p>
  {#if message.detail}<p class="text-xs text-fg-muted">{message.detail}</p>{/if}
  {#if onretry}<Button onclick={onretry}>Try again</Button>{/if}
</div>
