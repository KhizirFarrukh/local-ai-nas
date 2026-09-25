<!--
  Confirms a delete (S02.5-T02, RK-19). It names what goes and how much,
  and says that deleting is permanent: there is no trash until S08. Cancel
  comes first and has the focus, so Enter never deletes by accident.
-->
<script lang="ts">
  import Button from '$lib/components/Button.svelte';
  import Dialog from '$lib/components/Dialog.svelte';
  import { formatCount } from '$lib/util/format';
  import { itemsText } from './operations';
  import type { FileItem } from './types';

  interface Props {
    open: boolean;
    items: readonly FileItem[];
    onconfirm: () => void;
  }

  let { open = $bindable(false), items, onconfirm }: Props = $props();

  const shown = $derived(items.slice(0, 5));
  const folders = $derived(items.some((item) => item.kind === 'dir'));

  function confirm() {
    open = false;
    onconfirm();
  }
</script>

<Dialog bind:open title="Delete {itemsText(items)}?" size="md">
  <div class="flex flex-col gap-3 text-sm">
    <p>
      <strong>This is permanent.</strong> There is no trash yet, so deleted items cannot be brought
      back.{folders ? ' Folders are deleted with everything in them.' : ''}
    </p>
    {#if items.length > 1}
      <ul class="list-inside list-disc text-fg-muted">
        {#each shown as item (item.path)}
          <li class="truncate">{item.name}</li>
        {/each}
        {#if items.length > shown.length}
          <li>and {formatCount(items.length - shown.length)} more</li>
        {/if}
      </ul>
    {/if}
  </div>
  {#snippet actions()}
    <Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
    <Button variant="danger" onclick={confirm}>Delete</Button>
  {/snippet}
</Dialog>
