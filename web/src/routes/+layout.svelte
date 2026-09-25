<script lang="ts">
  import '../app.css';
  import { theme } from '$lib/util/theme.svelte';

  let { children } = $props();

  // Apply the remembered theme before the first page renders.
  theme();
</script>

{@render children()}
