<!--
  The items of a folder as a list or a grid (S02.3-T03, FR-002). Only the
  rows on screen exist: @tanstack/svelte-virtual places them, and the
  listing loads their pages. The view is one ARIA grid; the focused item is
  its active descendant, moved with the arrow keys, Home, End, Page Up, and
  Page Down, and opened with Enter or a double click.

  Selection (S02.5-T01) works as in common file managers. A click, or
  moving with the keys, selects one item. Ctrl/Cmd-click and Space toggle
  one. Shift with a click or a key selects a range (with Ctrl/Cmd, adding
  it). Ctrl/Cmd+A selects all, and Escape or a click on empty space clears
  the selection. Ctrl/Cmd with a key moves the focus only.

  Menus (S02.5-T04): a right-click, a long press on a touch screen, the
  menu key, or Shift+F10 asks the owner for the menu of the selection, or
  of the folder on empty space.
  An item not yet selected becomes the selection first, as in Explorer.

  Screen readers (S02.7-T02): each item is one cell whose name says all of
  it ("notes.txt, TXT file, 2 KB, added …, modified …"), whatever columns
  the width shows; the list has one column, the grid as many as fit.

  Sizes and dates (S02.3-T05, FR-214, FR-215): the list shows Size,
  Added, Modified, and Type as its own width allows (the row height
  follows the same width); where Added does not fit, a second line under
  the name has the size and the added date, as the grid's tiles do.
  Folders show their size once the server has added it up.

  reveal() scrolls to an item and blinks it twice (S02.4-T05, FR-216),
  such as a finished upload; with reduced motion it is lit steadily.
-->
<script lang="ts">
  import ArrowDown from '@lucide/svelte/icons/arrow-down';
  import ArrowUp from '@lucide/svelte/icons/arrow-up';
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack, type Snippet } from 'svelte';
  import { formatDate, formatDay, formatSize } from '$lib/util/format';
  import { focusFor } from '$lib/util/modality';
  import { iconFor, kindLabel } from './icons';
  import type { FolderListing } from './listing.svelte';
  import type { PickMode, Selection } from './selection.svelte';
  import type { FolderSizes } from './sizes.svelte';
  import type { FileItem, SortKey, SortOrder } from './types';

  interface Props {
    listing: FolderListing;
    mode: 'list' | 'grid';
    sort: SortKey;
    order: SortOrder;
    onsort: (key: SortKey) => void;
    /** Opens the item at a position: a folder, or a file's preview. */
    onopen: (item: FileItem, index: number) => void;
    /** The selected items; the owner clears it for a new folder. */
    selection: Selection;
    /** The focused position; the owner can move it. */
    focused?: number;
    /** Buttons for one item, shown on hover and on selected items. */
    actions?: Snippet<[FileItem]>;
    /** Opens a menu at x, y: for the selection, or the folder (no item). */
    onmenu?: (item: FileItem | undefined, x: number, y: number) => void;
    /** Items shown dimmed, such as cut items waiting to be moved. */
    dimmed?: (path: string) => boolean;
    /** Take the focus when shown, as after opening a folder from the view. */
    autofocus?: boolean;
    /** Folder sizes, asked for the folders on screen (S02.3-T05). */
    sizes?: FolderSizes;
  }

  let {
    listing,
    mode,
    sort,
    order,
    onsort,
    onopen,
    selection,
    focused = $bindable(0),
    actions,
    onmenu,
    dimmed,
    autofocus = false,
    sizes
  }: Props = $props();

  const uid = $props.id();
  const tileWidth = 150;
  const tileHeight = 156;

  let scroller: HTMLDivElement | undefined = $state();
  let width = $state(800);

  // What fits in the list's own width (not the window's, which also
  // holds the navigation), so the columns and the row height agree.
  const show = $derived({ added: width >= 560, modified: width >= 760, kind: width >= 900 });
  const compact = $derived(!show.added); // size and added date under the name
  const rowHeight = $derived(compact ? 56 : 44);

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
    const size = mode === 'grid' ? tileHeight : rowHeight; // read here, so a change re-runs this
    const options = {
      count: rows,
      estimateSize: () => size,
      getScrollElement: () => scroller ?? null
    };
    if (scroller) {
      untrack(() => {
        $virtualizer.setOptions(options);
        $virtualizer.measure();
      });
    }
  });

  // A folder opened from the view, by keyboard or pointer, keeps the focus
  // in the view, so keyboard users do not start over from the top.
  $effect(() => {
    if (autofocus && scroller) {
      untrack(() => focusFor(scroller));
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
      selection.anchor = undefined; // positions change with the sort
      $virtualizer.scrollToOffset(0);
    });
  });

  const columnsDef = $derived(
    [
      { key: 'name', label: 'Name', class: 'flex-1 min-w-0', shown: true },
      { key: 'size', label: 'Size', class: 'w-24 shrink-0 text-right', shown: !compact },
      { key: 'added_time', label: 'Added', class: 'w-44 shrink-0', shown: show.added },
      { key: 'mod_time', label: 'Modified', class: 'w-44 shrink-0', shown: show.modified },
      { key: 'kind', label: 'Type', class: 'w-28 shrink-0', shown: show.kind }
    ].filter((col) => col.shown) as { key: SortKey; label: string; class: string }[]
  );

  // Ask for the sizes of the folders on screen; again after the folder
  // is loaded anew. The store's answers are never read here, so asking
  // cannot make this run again.
  $effect(() => {
    if (!sizes) {
      return;
    }
    const visible = $virtualizer.getVirtualItems();
    const stamp = listing.loads;
    const folders: string[] = [];
    for (const row of visible) {
      for (let c = 0; c < columns; c++) {
        const item = listing.at(row.index * columns + c);
        if (item?.kind === 'dir') {
          folders.push(item.path);
        }
      }
    }
    untrack(() => sizes.want(folders, stamp));
  });

  /** The size to show: a file's, a folder's once added up, or nothing. */
  function sizeText(item: FileItem): string {
    if (item.kind === 'file') {
      return formatSize(item.size);
    }
    if (item.kind === 'dir' && sizes) {
      const size = sizes.get(item.path);
      return size === undefined ? '…' : size === null ? '' : formatSize(size);
    }
    return '';
  }

  // reveal(): the item blinking now, and when it stops.
  let flashing: string | undefined = $state();
  let flashTimer: ReturnType<typeof setTimeout> | undefined;

  /**
   * Scrolls the item at index (with path) into view and blinks it twice,
   * once its page is loaded, so the user sees where it is (FR-216).
   */
  export async function reveal(index: number, path: string): Promise<void> {
    await listing.loadRange(index, index);
    if (listing.at(index)?.path !== path) {
      return; // the folder changed meanwhile
    }
    $virtualizer.scrollToIndex(Math.floor(index / columns), { align: 'center' });
    clearTimeout(flashTimer);
    flashing = undefined;
    await Promise.resolve(); // a new blink starts from the beginning
    flashing = path;
    flashTimer = setTimeout(() => (flashing = undefined), 1500);
  }

  /** How a click or a key with these modifiers changes the selection. */
  function pickMode(event: MouseEvent | KeyboardEvent, key: boolean): PickMode | undefined {
    const ctrl = event.ctrlKey || event.metaKey;
    if (event.shiftKey) {
      return ctrl ? 'add-range' : 'range';
    }
    if (ctrl) {
      return key ? undefined : 'toggle'; // Ctrl/Cmd with a key only moves the focus
    }
    return 'only';
  }

  function move(to: number, event: KeyboardEvent) {
    if (total === 0) {
      return;
    }
    focused = Math.max(0, Math.min(total - 1, to));
    $virtualizer.scrollToIndex(Math.floor(focused / columns), { align: 'auto' });
    const how = pickMode(event, true);
    if (how) {
      void selection.pick(focused, how, listing);
    }
  }

  function keydown(event: KeyboardEvent) {
    const ctrl = event.ctrlKey || event.metaKey;
    if (event.key === 'ContextMenu' || (event.key === 'F10' && event.shiftKey && !ctrl)) {
      event.preventDefault();
      keyMenuAt = performance.now();
      void menuFor(total > 0 ? focused : undefined);
      return;
    }
    if (ctrl && event.key.toLowerCase() === 'a' && !event.shiftKey && !event.altKey) {
      event.preventDefault();
      selection.selectAll();
      return;
    }
    if (event.key === ' ' && total > 0) {
      event.preventDefault(); // not a scroll
      void selection.pick(focused, 'toggle', listing);
      return;
    }
    if (event.key === 'Escape' && selection.count(total) > 0) {
      event.preventDefault();
      selection.clear();
      return;
    }
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
      move(moves[event.key], event);
    } else if (event.key === 'Enter') {
      const item = listing.at(focused);
      if (item) {
        event.preventDefault();
        onopen(item, focused);
      }
    }
  }

  // Clicks keep the focus on the grid itself: a focused cell would lose
  // focus when it scrolls out and the virtual list removes it.
  // On a touch screen (S02.7-T01), as on phones' file managers: a tap
  // opens the item, a long press selects it (with its menu), and while
  // something is selected, taps add or remove items.
  let lastPointer = 'mouse';

  function click(index: number, event: MouseEvent) {
    focused = index;
    scroller?.focus();
    const item = listing.at(index);
    if (lastPointer === 'touch' && !event.shiftKey && !event.ctrlKey && !event.metaKey && item) {
      if (selection.count(total) === 0) {
        onopen(item, index);
      } else {
        void selection.pick(index, 'toggle', listing);
      }
      return;
    }
    void selection.pick(index, pickMode(event, false) ?? 'only', listing);
  }

  // The menu of the item at index (selected first when it is not), or of
  // the folder; placed at the pointer, or under the item for the keyboard.
  let keyMenuAt = -Infinity; // when the menu key last opened a menu
  async function menuFor(index: number | undefined, x?: number, y?: number) {
    if (!onmenu || !scroller) {
      return;
    }
    const item = index === undefined ? undefined : listing.at(index);
    if (index !== undefined && item) {
      focused = index;
      if (!selection.has(item.path)) {
        await selection.pick(index, 'only', listing);
      }
    } else {
      selection.clear();
    }
    if (x === undefined || y === undefined) {
      const cell = index === undefined ? null : document.getElementById(cellId(index));
      const box = (cell ?? scroller).getBoundingClientRect();
      x = box.left + 24;
      y = cell ? box.bottom : box.top + 8;
    }
    onmenu(item, x, y);
  }

  function contextmenu(event: MouseEvent) {
    if (!onmenu) {
      return;
    }
    event.preventDefault();
    endPress(); // a browser that sends this for a long press opens it here
    // The menu key also sends this event; its keydown has opened the menu.
    if (performance.now() - keyMenuAt < 500) {
      return;
    }
    scroller?.focus();
    const cell = (event.target as Element).closest('[role="gridcell"]');
    const index = cell ? Number(cell.id.slice(cell.id.lastIndexOf('-') + 1)) : undefined;
    void menuFor(index, event.clientX, event.clientY);
  }

  // A long press on a touch screen (500 ms without moving) opens the menu
  // in every browser; some also send `contextmenu`, which ends the press.
  // The click a browser may send when the finger lifts is dropped, since
  // the menu opens under the finger and the click would pick an entry.
  let press: { timer: ReturnType<typeof setTimeout>; x: number; y: number } | undefined;

  function dropNextClick() {
    const drop = (event: Event) => {
      event.stopPropagation();
      event.preventDefault();
      done();
    };
    const done = () => {
      window.removeEventListener('click', drop, true);
      window.removeEventListener('pointerdown', done, true);
    };
    window.addEventListener('click', drop, true);
    window.addEventListener('pointerdown', done, true); // the next touch clicks again
  }

  function pressStart(event: PointerEvent) {
    lastPointer = event.pointerType;
    endPress();
    if (event.pointerType !== 'touch' || !onmenu) {
      return;
    }
    const { clientX: x, clientY: y } = event;
    const cell = (event.target as Element).closest('[role="gridcell"]');
    const index = cell ? Number(cell.id.slice(cell.id.lastIndexOf('-') + 1)) : undefined;
    press = {
      x,
      y,
      timer: setTimeout(() => {
        press = undefined;
        dropNextClick();
        keyMenuAt = performance.now(); // a `contextmenu` right after is the same press
        void menuFor(index, x, y);
      }, 500)
    };
  }

  function pressMove(event: PointerEvent) {
    if (press && Math.hypot(event.clientX - press.x, event.clientY - press.y) > 10) {
      endPress(); // scrolling, not pressing
    }
  }

  function endPress() {
    if (press) {
      clearTimeout(press.timer);
      press = undefined;
    }
  }

  // A click on empty space clears the selection; one on the scroll bar
  // (outside the client area) does not.
  function clickEmpty(event: MouseEvent) {
    if (!scroller || (event.target as Element).closest('[role="gridcell"]')) {
      return;
    }
    const box = scroller.getBoundingClientRect();
    if (
      event.clientX - box.left < scroller.clientWidth &&
      event.clientY - box.top < scroller.clientHeight
    ) {
      selection.clear();
    }
  }

  function cellId(index: number) {
    return `${uid}-item-${index}`;
  }

  // Touch screens have no hover, so they always show the actions.
  function actionsClass(selected: boolean) {
    return selected ? '' : 'opacity-0 group-hover:opacity-100 pointer-coarse:opacity-100';
  }

  // Selected items are tinted; the focused one gets a ring while the
  // keyboard is in use.
  function cellClass(index: number, selected: boolean) {
    const ring =
      index === focused
        ? 'group-focus-visible/grid:ring-2 group-focus-visible/grid:ring-accent group-focus-visible/grid:ring-inset'
        : '';
    const item = listing.at(index);
    const dim = item && dimmed?.(item.path) ? 'opacity-50' : '';
    const flash = item && item.path === flashing ? 'flash' : '';
    return `${selected ? 'bg-accent-soft' : 'hover:bg-surface-2'} ${ring} ${dim} ${flash}`;
  }

  /** What a screen reader says for the item at a position. */
  function cellLabel(item: FileItem | undefined): string {
    if (!item) {
      return 'Loading';
    }
    const parts = [item.name, kindLabel(item)];
    const size = sizeText(item);
    if (size && size !== '…') {
      parts.push(size);
    }
    parts.push(`added ${formatDate(item.added_time)}`, `modified ${formatDate(item.mod_time)}`);
    if (dimmed?.(item.path)) {
      parts.push('cut, waiting to be pasted');
    }
    return parts.join(', ');
  }
</script>

{#snippet icon(item: FileItem)}
  {@const Icon = iconFor(item)}
  <Icon
    class="size-5 shrink-0 {item.kind === 'dir' ? 'text-accent' : 'text-fg-muted'}"
    aria-hidden="true"
  />
{/snippet}

{#snippet name(item: FileItem)}
  {@render icon(item)}
  <span class="truncate">{item.name}</span>
{/snippet}

<div class="flex min-h-0 flex-1 flex-col">
  {#if mode === 'list'}
    <div
      role="presentation"
      class="flex h-10 shrink-0 items-center gap-4 border-b border-border px-4 text-sm text-fg-muted pointer-coarse:h-11"
    >
      {#each columnsDef as col (col.key)}
        <div class={col.class}>
          <button
            type="button"
            class="inline-flex items-center gap-1 rounded px-1 font-medium hover:text-fg pointer-coarse:min-h-11 pointer-coarse:min-w-11"
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
    aria-colcount={columns}
    aria-multiselectable="true"
    aria-activedescendant={total > 0 ? cellId(focused) : undefined}
    data-testid="file-view"
    class="group/grid min-h-0 flex-1 overflow-auto select-none focus-visible:outline-offset-[-2px] [-webkit-touch-callout:none]"
    onkeydown={keydown}
    onclick={clickEmpty}
    oncontextmenu={contextmenu}
    onpointerdown={pressStart}
    onpointermove={pressMove}
    onpointerup={endPress}
    onpointercancel={endPress}
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
              {@const selected = item ? selection.has(item.path) : false}
              {#if mode === 'list'}
                <div
                  id={cellId(index)}
                  role="gridcell"
                  aria-selected={selected}
                  aria-label={cellLabel(item)}
                  class="group flex w-full cursor-default items-center gap-4 border-b border-border px-4 text-sm {cellClass(
                    index,
                    selected
                  )}"
                  tabindex="-1"
                  onmousedown={(e) => e.preventDefault()}
                  onclick={(e) => click(index, e)}
                  ondblclick={() => item && onopen(item, index)}
                  onkeydown={undefined}
                >
                  {#if item}
                    {#if compact}
                      <span class="flex min-w-0 flex-1 items-center gap-3">
                        {@render icon(item)}
                        <span class="flex min-w-0 flex-col">
                          <span class="truncate">{item.name}</span>
                          <span class="truncate text-xs text-fg-muted" data-testid="item-details"
                            >{[sizeText(item), `Added ${formatDate(item.added_time)}`]
                              .filter(Boolean)
                              .join(' · ')}</span
                          >
                        </span>
                      </span>
                    {:else}
                      <span class="flex min-w-0 flex-1 items-center gap-3"
                        >{@render name(item)}</span
                      >
                      <span
                        class="w-24 shrink-0 text-right text-fg-muted tabular-nums"
                        data-testid="item-size">{sizeText(item)}</span
                      >
                      <span class="w-44 shrink-0 text-fg-muted" data-testid="item-added"
                        >{formatDate(item.added_time)}</span
                      >
                      {#if show.modified}
                        <span class="w-44 shrink-0 text-fg-muted">{formatDate(item.mod_time)}</span>
                      {/if}
                      {#if show.kind}
                        <span class="w-28 shrink-0 truncate text-fg-muted">{kindLabel(item)}</span>
                      {/if}
                    {/if}
                    {#if actions}
                      <span class="flex w-8 shrink-0 justify-end {actionsClass(selected)}"
                        >{@render actions(item)}</span
                      >
                    {/if}
                  {:else}
                    <span class="h-3 w-1/3 animate-pulse rounded bg-surface-2" aria-hidden="true"
                    ></span>
                  {/if}
                </div>
              {:else}
                <div
                  id={cellId(index)}
                  role="gridcell"
                  aria-selected={selected}
                  aria-colindex={index - row.index * columns + 1}
                  aria-label={cellLabel(item)}
                  class="group relative flex min-w-0 flex-1 cursor-default flex-col items-center justify-center gap-2 rounded-lg p-2 text-center text-sm {cellClass(
                    index,
                    selected
                  )}"
                  tabindex="-1"
                  onmousedown={(e) => e.preventDefault()}
                  onclick={(e) => click(index, e)}
                  ondblclick={() => item && onopen(item, index)}
                  onkeydown={undefined}
                >
                  {#if item}
                    {@const Icon = iconFor(item)}
                    <Icon
                      class="size-12 {item.kind === 'dir' ? 'text-accent' : 'text-fg-muted'}"
                      aria-hidden="true"
                    />
                    <span class="line-clamp-2 w-full break-words">{item.name}</span>
                    <span
                      class="w-full truncate text-xs text-fg-muted"
                      title="Added {formatDate(item.added_time)}"
                      data-testid="item-details"
                      >{[sizeText(item), formatDay(item.added_time)]
                        .filter(Boolean)
                        .join(' · ')}</span
                    >
                    {#if actions}
                      <span class="absolute top-1 right-1 {actionsClass(selected)}"
                        >{@render actions(item)}</span
                      >
                    {/if}
                  {:else}
                    <span class="size-12 animate-pulse rounded bg-surface-2" aria-hidden="true"
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
