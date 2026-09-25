<!--
  The operations in progress (S02.2-T03), at the bottom left: one line
  each, with a progress bar and, where possible, Cancel. The panel can be
  folded to its title.
-->
<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import Button from '$lib/components/Button.svelte';
  import ProgressBar from '$lib/components/ProgressBar.svelte';
  import { tasks } from './tasks.svelte';

  let folded = $state(false);
  const count = $derived(tasks.items.length);
</script>

{#if count > 0}
  <section
    aria-label="In progress"
    class="fixed bottom-20 left-4 z-40 w-[calc(100%-2rem)] max-w-sm rounded-lg border border-border bg-surface shadow-lg md:bottom-4 md:left-60"
  >
    <button
      type="button"
      class="flex w-full items-center justify-between gap-2 px-4 py-2 text-sm font-semibold"
      aria-expanded={!folded}
      onclick={() => (folded = !folded)}
    >
      <span>In progress ({count})</span>
      {#if folded}<ChevronUp class="size-4" aria-hidden="true" />{:else}<ChevronDown
          class="size-4"
          aria-hidden="true"
        />{/if}
    </button>
    {#if !folded}
      <ul class="flex max-h-64 flex-col gap-3 overflow-auto border-t border-border px-4 py-3">
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
  </section>
{/if}
