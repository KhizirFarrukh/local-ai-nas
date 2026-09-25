<!--
  A menu at a screen position: a context menu at the pointer, or a menu
  under the button that opened it. Arrow keys, Home, and End move between
  items; Enter or Space picks one; Escape, Tab, a click outside, or a
  resize closes it, and focus returns to where it was.
-->
<script lang="ts" module>
  import type { IconComponent } from './icon';

  export interface MenuItem {
    label: string;
    icon?: IconComponent;
    /** Shown on the right, for example "F2". */
    shortcut?: string;
    danger?: boolean;
    disabled?: boolean;
    onselect: () => void;
  }

  export type MenuEntry = MenuItem | 'separator';
</script>

<script lang="ts">
  import { tick } from 'svelte';

  interface Props {
    open: boolean;
    items: MenuEntry[];
    /** Where to show the menu, in viewport pixels. */
    x: number;
    y: number;
    /** The menu's name for screen readers. */
    label: string;
    onclose?: () => void;
  }

  let { open = $bindable(false), items, x, y, label, onclose }: Props = $props();

  let menu: HTMLElement | undefined = $state();
  let left = $state(0);
  let top = $state(0);
  let returnFocus: HTMLElement | null = null;

  $effect(() => {
    if (!open) {
      return;
    }
    returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    left = x;
    top = y;
    tick().then(() => {
      if (!menu) {
        return;
      }
      // Keep the whole menu inside the window.
      const r = menu.getBoundingClientRect();
      left = Math.max(4, Math.min(x, window.innerWidth - r.width - 4));
      top = Math.max(4, Math.min(y, window.innerHeight - r.height - 4));
      buttons()[0]?.focus();
    });
  });

  function buttons(): HTMLButtonElement[] {
    return menu ? [...menu.querySelectorAll<HTMLButtonElement>('button:not([disabled])')] : [];
  }

  function close() {
    if (!open) {
      return;
    }
    open = false;
    onclose?.();
    returnFocus?.focus();
  }

  function choose(item: MenuItem) {
    close();
    item.onselect();
  }

  function keydown(event: KeyboardEvent) {
    const all = buttons();
    const at = all.indexOf(document.activeElement as HTMLButtonElement);
    let next: number;
    switch (event.key) {
      case 'ArrowDown':
        next = (at + 1) % all.length;
        break;
      case 'ArrowUp':
        next = (at - 1 + all.length) % all.length;
        break;
      case 'Home':
        next = 0;
        break;
      case 'End':
        next = all.length - 1;
        break;
      case 'Escape':
      case 'Tab':
        event.preventDefault();
        close();
        return;
      default:
        return;
    }
    event.preventDefault();
    all[next]?.focus();
  }

  function outside(event: PointerEvent) {
    if (menu && !menu.contains(event.target as Node)) {
      close();
    }
  }
</script>

<svelte:window onpointerdown={open ? outside : undefined} onresize={open ? close : undefined} />

{#if open}
  <div
    bind:this={menu}
    role="menu"
    aria-label={label}
    tabindex="-1"
    class="fixed z-50 min-w-48 rounded-md border border-border bg-surface py-1 text-sm text-fg shadow-lg"
    style:left="{left}px"
    style:top="{top}px"
    onkeydown={keydown}
  >
    {#each items as entry, i (i)}
      {#if entry === 'separator'}
        <div role="separator" class="my-1 border-t border-border"></div>
      {:else}
        {@const Icon = entry.icon}
        <button
          type="button"
          role="menuitem"
          tabindex="-1"
          disabled={entry.disabled}
          class="flex w-full items-center gap-3 px-3 py-2 text-left pointer-coarse:py-3 hover:bg-surface-2 focus:bg-surface-2 focus:outline-none disabled:opacity-50 {entry.danger
            ? 'text-danger'
            : ''}"
          onclick={() => choose(entry)}
        >
          {#if Icon}<Icon class="size-4 shrink-0" />{:else}<span class="size-4 shrink-0"
            ></span>{/if}
          <span class="flex-1">{entry.label}</span>
          {#if entry.shortcut}<kbd class="text-xs text-fg-muted">{entry.shortcut}</kbd>{/if}
        </button>
      {/if}
    {/each}
  </div>
{/if}
