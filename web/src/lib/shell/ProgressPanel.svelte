<!--
  The operations in progress (S02.2-T03), at the bottom left: uploads
  (S02.4-T01) and other long operations, each with progress and its
  actions. The panel can be folded to its title.
-->
<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import Button from '$lib/components/Button.svelte';
  import ProgressBar from '$lib/components/ProgressBar.svelte';
  import { uploads } from '$lib/uploads/state.svelte';
  import UploadList from '$lib/uploads/UploadList.svelte';
  import { tasks } from './tasks.svelte';

  let folded = $state(false);
  const manager = $derived(uploads.manager);
  const active = $derived(tasks.items.length + (manager?.active ?? 0));
  const shown = $derived(tasks.items.length > 0 || (manager?.entries.length ?? 0) > 0);
</script>

{#if shown}
  <section
    aria-label="In progress"
    class="fixed bottom-20 left-4 z-40 w-[calc(100%-2rem)] max-w-sm rounded-lg border border-border bg-surface shadow-lg md:bottom-4 md:left-60"
  >
    <div class="flex items-center gap-2 px-4 py-2">
      <button
        type="button"
        class="flex flex-1 items-center justify-between gap-2 text-sm font-semibold"
        aria-expanded={!folded}
        onclick={() => (folded = !folded)}
      >
        <span>{active > 0 ? `In progress (${active})` : 'Finished'}</span>
        {#if folded}<ChevronUp class="size-4" aria-hidden="true" />{:else}<ChevronDown
            class="size-4"
            aria-hidden="true"
          />{/if}
      </button>
      {#if manager && manager.entries.length > manager.active}
        <Button size="sm" variant="ghost" onclick={() => manager.clearFinished()}>Clear</Button>
      {/if}
    </div>
    {#if !folded}
      <div class="flex max-h-72 flex-col gap-3 overflow-auto border-t border-border px-4 py-3">
        {#if manager && manager.entries.length > 0}
          <UploadList {manager} />
        {/if}
        {#if tasks.items.length > 0}
          <ul class="flex flex-col gap-3">
            {#each tasks.items as task (task.id)}
              <li class="flex flex-col gap-1">
                <div class="flex items-center justify-between gap-2 text-sm">
                  <span class="truncate">{task.label}</span>
                  {#if task.cancel}
                    <Button
                      size="sm"
                      variant="ghost"
                      aria-label="Cancel {task.label}"
                      onclick={task.cancel}>Cancel</Button
                    >
                  {/if}
                </div>
                <ProgressBar
                  value={task.total ? task.done : undefined}
                  max={task.total}
                  label={task.label}
                />
                {#if task.note}<span class="text-xs text-fg-muted">{task.note}</span>{/if}
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    {/if}
  </section>
{/if}
