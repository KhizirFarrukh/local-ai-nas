<!-- The photos area is reserved until the media stages (S04.7). -->
<script lang="ts">
  import Images from '@lucide/svelte/icons/images';
  import EmptyState from '$lib/components/EmptyState.svelte';
</script>

<svelte:head>
  <title>Photos · local-ai-nas</title>
</svelte:head>

<EmptyState
  icon={Images}
  title="Photos come with stage 4"
  description="The photo library, with its timeline and albums, is a later stage of local-ai-nas. Until then, keep your files in the Files area."
/>
