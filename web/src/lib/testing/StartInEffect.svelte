<!--
  Test helper (S02.3-T05): starts a folder listing in an effect, as the
  files page does. The effect must run once, not loop on the listing's load
  counter (a request storm found by the system tests).
-->
<script lang="ts">
  import type { FolderListing } from '$lib/files/listing.svelte';

  let { listing, runs }: { listing: FolderListing; runs: { count: number } } = $props();

  $effect(() => {
    runs.count++;
    void listing.start();
  });
</script>

<p>{listing.total ?? ''}</p>
