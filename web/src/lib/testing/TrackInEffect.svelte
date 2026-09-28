<!--
  Test helper (S02.8-T01, bug S02-B04): an effect that tracks failing work
  at once, as the shell does when the server is down. It must run once, not
  loop on the activity counter it changes.
-->
<script lang="ts">
  import { activity } from '$lib/shell/activity.svelte';

  let { runs }: { runs: { count: number } } = $props();

  $effect(() => {
    runs.count++;
    for (let i = 0; i < 3; i++) {
      void activity.track(Promise.reject(new Error('down'))).catch(() => {});
    }
  });
</script>

<p>{activity.pending}</p>
