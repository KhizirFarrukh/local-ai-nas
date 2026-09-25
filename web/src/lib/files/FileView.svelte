<!--
  The items of a folder as a list or a grid (S02.3-T03, FR-002). Only the
  rows on screen exist: @tanstack/svelte-virtual places them, and the
  listing loads their pages. The view is one ARIA grid; the focused item is
  its active descendant, moved with the arrow keys, Home, End, Page Up, and
  Page Down, and opened with Enter or a double click.
-->
<script lang="ts">
  import ArrowDown from '@lucide/svelte/icons/arrow-down';
  import ArrowUp from '@lucide/svelte/icons/arrow-up';
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack, type Snippet } from 'svelte';
  import { formatDate, formatSize } from '$lib/util/format';
  import { iconFor, kindLabel } from './icons';
  import type { FolderListing } from './listing.svelte';
  import type { FileItem, SortKey, SortOrder } from './types';

  interface Props {
    listing: FolderListing;
    mode: 'list' | 'grid';
    sort: SortKey;
    order: SortOrder;
    onsort: (key: SortKey) => void;
    onopen: (item: FileItem) => void;
    /** The focused position; the owner can move it. */
    focused?: number;
    /** Buttons for one item, shown on hover and on the focused item. */
    actions?: Snippet<[FileItem]>;
  }

  let {
    listing,
    mode,
    sort,
    order,
    onsort,
    onopen,
    focused = $bindable(0),
    actions
  }: Props = $props();

  const uid = $props.id();
  const rowHeight = 44;
  const tileWidth = 150;
  const tileHeight = 136;

  let scroller: HTMLDivElement | undefined = $state();
  let width = $state(800);

  const total = $derived(listing.total ?? 0);
  const columns = $derived(mode === 'grid' ? Math.max(1, Math.floor(width / tileWidth)) : 1);
  const rows = $derived(Math.ceil(total / columns));

  const virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scroller ?? null,
    estimateSize: () => rowHeight,
    overscan: 6
  });

  // setOptions() writes the store, so it runs untracked (S02.3-T01).
  $effect(() => {
    const options = {
      count: rows,
      estimateSize: () => (mode === 'grid' ? tileHeight : rowHeight),
      getScrollElement: () => scroller ?? null
    };
    if (scroller) {
      untrack(() => {
        $virtualizer.setOptions(options);
        $virtualizer.measure();
      });
    }
  });

  // Load the pages of the rows on screen.
  $effect(() => {
    const visible = $virtualizer.getVirtualItems();
    if (visible.length > 0) {
      const first = visible[0].index * columns;
      const last = (visible[visible.length - 1].index + 1) * columns - 1;
      untrack(() => listing.ensure(first, last));
    }
  });

  // A new folder or sort starts at the top.
  $effect(() => {
    void listing;
    untrack(() => {
      focused = 0;
      $virtualizer.scrollToOffset(0);
    });
  });

  const columnsDef: { key: SortKey; label: string; class: string }[] = [
    { key: 'name', label: 'Name', class: 'flex-1 min-w-0' },
    { key: 'size', label: 'Size', class: 'w-28 text-right hidden sm:block' },
    { key: 'mod_time', label: 'Modified', class: 'w-48 hidden md:block' },
    { key: 'kind', label: 'Type', class: 'w-32 hidden lg:block' }
  ];

  function move(to: number) {
    if (total === 0) {
      return;
    }
    focused = Math.max(0, Math.min(total - 1, to));
    $virtualizer.scrollToIndex(Math.floor(focused / columns), { align: 'auto' });
  }

  function keydown(event: KeyboardEvent) {
    const page = Math.max(
      1,
      Math.floor((scroller?.clientHeight ?? 400) / (mode === 'grid' ? tileHeight : rowHeight))
    );
    const moves: Record<string, number> = {
      ArrowDown: focused + columns,
      ArrowUp: focused - columns,
      ArrowRight: mode === 'grid' ? focused + 1 : focused,
      ArrowLeft: mode === 'grid' ? focused - 1 : focused,
      Home: 0,
      End: total - 1,
      PageDown: focused + page * columns,
      PageUp: focused - page * columns
    };
    if (event.key in moves) {
      event.preventDefault();
      move(moves[event.key]);
    } else if (event.key === 'Enter') {
      const item = listing.at(focused);
      if (item) {
        event.preventDefault();
        onopen(item);
      }
    }
  }

  // Clicks keep the focus on the grid itself: a focused cell would lose
  // focus when it scrolls out and the virtual list removes it.
  function pick(index: number) {
    focused = index;
    scroller?.focus();
  }

  function cellId(index: number) {
    return `${uid}-item-${index}`;
  }

  // Touch screens have no hover, so they always show the actions.
  function actionsClass(index: number) {
    return index === focused ? '' : 'opacity-0 group-hover:opacity-100 pointer-coarse:opacity-100';
  }
</script>

{#snippet name(item: FileItem)}
  {@const Icon = iconFor(item)}
  <Icon
    class="size-5 shrink-0 {item.kind === 'dir' ? 'text-accent' : 'text-fg-muted'}"
    aria-hidden="true"
  />
  <span class="truncate">{item.name}</span>
{/snippet}

<div class="flex min-h-0 flex-1 flex-col">
  {#if mode === 'list'}
    <div
      role="presentation"
      class="flex h-10 shrink-0 items-center gap-4 border-b border-border px-4 text-sm text-fg-muted"
    >
      {#each columnsDef as col (col.key)}
        <div class={col.class}>
          <button
            type="button"
            class="inline-flex items-center gap-1 rounded px-1 font-medium hover:text-fg"
            aria-label="Sort by {col.label.toLowerCase()}{col.key === sort
              ? `, now ${order === 'asc' ? 'ascending' : 'descending'}`
              : ''}"
            onclick={() => onsort(col.key)}
          >
            {col.label}
            {#if col.key === sort}
              {#if order === 'asc'}<ArrowUp class="size-3.5" aria-hidden="true" />{:else}<ArrowDown
                  class="size-3.5"
                  aria-hidden="true"
                />{/if}
            {/if}
          </button>
        </div>
      {/each}
      {#if actions}
        <div class="w-8 shrink-0" aria-hidden="true"></div>
      {/if}
    </div>
  {/if}

  <div
    bind:this={scroller}
    bind:clientWidth={width}
    role="grid"
    tabindex="0"
    aria-label="Items in {listing.folder?.name || 'Files'}"
    aria-rowcount={rows}
    aria-colcount={mode === 'list' ? 4 : columns}
    aria-activedescendant={total > 0 ? cellId(focused) : undefined}
    data-testid="file-view"
    class="min-h-0 flex-1 overflow-auto focus-visible:outline-offset-[-2px]"
    onkeydown={keydown}
  >
    <div class="relative w-full" style:height="{$virtualizer.getTotalSize()}px">
      {#each $virtualizer.getVirtualItems() as row (row.key)}
        <div
          role="row"
          aria-rowindex={row.index + 1}
          class="absolute top-0 left-0 flex w-full {mode === 'grid' ? 'gap-2 px-2' : ''}"
          style:height="{row.size}px"
          style:transform="translateY({row.start}px)"
        >
          {#each Array.from({ length: columns }, (_, c) => row.index * columns + c) as index (index)}
            {#if index < total}
              {@const item = listing.at(index)}
              {#if mode === 'list'}
                <div
                  id={cellId(index)}
                  role="gridcell"
                  aria-selected={index === focused}
                  class="group flex w-full cursor-default items-center gap-4 border-b border-border px-4 text-sm {index ===
                  focused
                    ? 'bg-accent-soft'
                    : 'hover:bg-surface-2'}"
                  tabindex="-1"
                  onmousedown={(e) => e.preventDefault()}
                  onclick={() => pick(index)}
                  ondblclick={() => item && onopen(item)}
                  onkeydown={undefined}
                >
                  {#if item}
                    <span class="flex min-w-0 flex-1 items-center gap-3">{@render name(item)}</span>
                    <span class="hidden w-28 text-right text-fg-muted tabular-nums sm:block"
                      >{item.kind === 'file' ? formatSize(item.size) : ''}</span
                    >
                    <span class="hidden w-48 text-fg-muted md:block"
                      >{formatDate(item.mod_time)}</span
                    >
                    <span class="hidden w-32 truncate text-fg-muted lg:block"
                      >{kindLabel(item)}</span
                    >
                    {#if actions}
                      <span class="flex w-8 shrink-0 justify-end {actionsClass(index)}"
                        >{@render actions(item)}</span
                      >
                    {/if}
                  {:else}
                    <span class="h-3 w-1/3 animate-pulse rounded bg-surface-2" aria-label="Loading"
                    ></span>
                  {/if}
                </div>
              {:else}
                <div
                  id={cellId(index)}
                  role="gridcell"
                  aria-selected={index === focused}
                  class="group relative flex min-w-0 flex-1 cursor-default flex-col items-center justify-center gap-2 rounded-lg p-2 text-center text-sm {index ===
                  focused
                    ? 'bg-accent-soft'
                    : 'hover:bg-surface-2'}"
                  tabindex="-1"
                  onmousedown={(e) => e.preventDefault()}
                  onclick={() => pick(index)}
                  ondblclick={() => item && onopen(item)}
                  onkeydown={undefined}
                >
                  {#if item}
                    {@const Icon = iconFor(item)}
                    <Icon
                      class="size-12 {item.kind === 'dir' ? 'text-accent' : 'text-fg-muted'}"
                      aria-hidden="true"
                    />
                    <span class="line-clamp-2 w-full break-words">{item.name}</span>
                    {#if actions}
                      <span class="absolute top-1 right-1 {actionsClass(index)}"
                        >{@render actions(item)}</span
                      >
                    {/if}
                  {:else}
                    <span class="size-12 animate-pulse rounded bg-surface-2" aria-label="Loading"
                    ></span>
                  {/if}
                </div>
              {/if}
            {:else}
              <div class="flex-1" aria-hidden="true"></div>
            {/if}
          {/each}
        </div>
      {/each}
    </div>
  </div>
</div>
