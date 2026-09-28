<!--
  The app's two live regions (S02.7-T02), hidden on screen. They are part
  of the page from the start, so screen readers read what announcer.say()
  puts in them.
-->
<script lang="ts">
  import { announcer } from './announcer.svelte';
</script>

<div class="sr-only" role="status" aria-atomic="true" data-testid="live-polite">
  {announcer.polite}
</div>
<div class="sr-only" role="alert" aria-atomic="true" data-testid="live-urgent">
  {announcer.urgent}
</div>
