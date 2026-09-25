<!-- Shown while the NAS cannot be reached, with a way to check again. -->
<script lang="ts">
  import WifiOff from '@lucide/svelte/icons/wifi-off';
  import Button from '$lib/components/Button.svelte';
  import { connection } from './connection.svelte';

  const conn = connection();
</script>

{#if !conn.online}
  <div
    role="alert"
    class="flex flex-wrap items-center gap-3 border-b border-danger bg-surface px-4 py-2 text-sm"
  >
    <WifiOff class="size-5 shrink-0 text-danger" aria-hidden="true" />
    <p class="flex-1">
      <span class="font-semibold">Can’t reach the NAS.</span>
      Check that the local-ai-nas server is running on this computer. The app keeps trying.
    </p>
    <Button size="sm" disabled={conn.checking} onclick={() => conn.retry()}>
      {conn.checking ? 'Checking…' : 'Try now'}
    </Button>
  </div>
{/if}
