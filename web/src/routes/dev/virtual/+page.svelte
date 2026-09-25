<!--
  The virtualization prototype (S02.3-T01, ADR-0009): @tanstack/svelte-virtual
  on Svelte 5 with 50,000 synthetic rows, as a list and as a grid, with
  keyboard focus (arrows, Home, End). It exists to decide whether the file
  browser can use the package or needs its own windowing component.
-->
<script lang="ts">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack } from 'svelte';

  let count = $state(50000);
  let mode = $state<'list' | 'grid'>('list');
  let focused = $state(0);
  let scroller: HTMLDivElement | undefined = $state();
  let width = $state(800);

  const tile = 160;
  const columns = $derived(mode === 'grid' ? Math.max(1, Math.floor(width / tile)) : 1);
  const rows = $derived(Math.ceil(count / columns));

  const virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scroller ?? null,
    estimateSize: () => 40,
    overscan: 8
  });

  // The store starts before the scroll element is bound; hand it the
  // element and the current row count and size whenever they change.
  // setOptions() writes the store, so it runs untracked: reading $virtualizer
  // here would make the effect run again after every write, without end.
  $effect(() => {
    const options = {
      count: rows,
      estimateSize: () => (mode === 'grid' ? tile : 40),
      getScrollElement: () => scroller ?? null
    };
    if (scroller) {
      untrack(() => $virtualizer.setOptions(options));
    }
  });

  function keydown(event: KeyboardEvent) {
    const step = { ArrowDown: columns, ArrowUp: -columns, ArrowRight: 1, ArrowLeft: -1 }[event.key];
    let next: number;
    if (step !== undefined) {
      next = focused + step;
    } else if (event.key === 'Home') {
      next = 0;
    } else if (event.key === 'End') {
      next = count - 1;
    } else {
      return;
    }
    event.preventDefault();
    focused = Math.max(0, Math.min(count - 1, next));
    $virtualizer.scrollToIndex(Math.floor(focused / columns), { align: 'auto' });
  }
</script>

<div class="flex h-full flex-col gap-3 p-4">
  <div class="flex flex-wrap items-center gap-3 text-sm">
    <label>Items <input type="number" bind:value={count} class="w-28 rounded border px-2" /></label>
    <label><input type="radio" bind:group={mode} value="list" /> List</label>
    <label><input type="radio" bind:group={mode} value="grid" /> Grid</label>
    <span data-testid="rendered">Rendered: {$virtualizer.getVirtualItems().length}</span>
    <span>Focused: {focused}</span>
  </div>
  <div
    bind:this={scroller}
    bind:clientWidth={width}
    data-testid="scroller"
    role="grid"
    aria-label="Prototype items"
    aria-rowcount={rows}
    tabindex="0"
    class="min-h-0 flex-1 overflow-auto rounded border border-border"
    onkeydown={keydown}
  >
    <div class="relative w-full" style:height="{$virtualizer.getTotalSize()}px">
      {#each $virtualizer.getVirtualItems() as row (row.key)}
        <div
          role="row"
          aria-rowindex={row.index + 1}
          class="absolute top-0 left-0 flex w-full"
          style:height="{row.size}px"
          style:transform="translateY({row.start}px)"
        >
          {#each Array.from({ length: columns }, (_, c) => c) as c (c)}
            {@const index = row.index * columns + c}
            {#if index < count}
              <div
                role="gridcell"
                class="flex flex-1 items-center border-b border-border px-3 text-sm {index ===
                focused
                  ? 'bg-accent-soft'
                  : ''}"
              >
                Item {index + 1}
              </div>
            {/if}
          {/each}
        </div>
      {/each}
    </div>
  </div>
</div>
