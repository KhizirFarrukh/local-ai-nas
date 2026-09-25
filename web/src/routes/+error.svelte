<!-- Any page that cannot be shown: an unknown address, or a failed load. -->
<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import { page } from '$app/state';
  import Button from '$lib/components/Button.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  const notFound = $derived(page.status === 404);
</script>

<svelte:head>
  <title>{notFound ? 'Not found' : 'Error'} · local-ai-nas</title>
</svelte:head>

<EmptyState
  icon={CircleAlert}
  title={notFound ? 'There is no page at this address' : 'Something went wrong'}
  description={notFound
    ? 'Check the address, or go to your files.'
    : (page.error?.message ?? 'The page could not be shown.')}
>
  {#snippet actions()}
    <Button variant="primary" onclick={() => (location.href = '/files')}>Go to Files</Button>
  {/snippet}
</EmptyState>
