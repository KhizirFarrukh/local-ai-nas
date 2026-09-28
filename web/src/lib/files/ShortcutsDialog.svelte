<!--
  The keyboard shortcuts of the file browser (S02.5-T04), shown with "?"
  or from the folder's menu. The list comes from keys.ts, as the keys do.
-->
<script lang="ts">
  import Button from '$lib/components/Button.svelte';
  import Dialog from '$lib/components/Dialog.svelte';
  import { isMac, shortcuts } from './keys';

  let { open = $bindable(false) }: { open: boolean } = $props();

  const list = shortcuts(isMac());
</script>

<Dialog bind:open title="Keyboard shortcuts" size="lg" phone="full">
  <table class="w-full text-sm">
    <thead class="sr-only">
      <tr><th>Keys</th><th>What they do</th></tr>
    </thead>
    <tbody>
      {#each list as row (row.keys)}
        <tr class="border-b border-border last:border-0">
          <td class="py-1.5 pr-4 align-top whitespace-nowrap"
            ><kbd class="font-sans font-medium">{row.keys}</kbd></td
          >
          <td class="py-1.5 text-fg-muted">{row.what}</td>
        </tr>
      {/each}
    </tbody>
  </table>
  {#snippet actions()}
    <Button variant="primary" onclick={() => (open = false)}>Close</Button>
  {/snippet}
</Dialog>
