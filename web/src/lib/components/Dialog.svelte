<!--
  A modal dialog on the native <dialog> element: the browser keeps focus
  inside it, makes the rest of the page inert, closes it on Escape, and
  returns focus to where it was when it closes.
-->
<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    open: boolean;
    title: string;
    description?: string;
    size?: 'sm' | 'md' | 'lg';
    /** The body of the dialog. */
    children?: Snippet;
    /** The buttons at the bottom, right-aligned. */
    actions?: Snippet;
    /** Called whenever the dialog closes (Escape, backdrop, or code). */
    onclose?: () => void;
    /** Keep the dialog open on a backdrop click, for dialogs with input. */
    persistent?: boolean;
  }

  let {
    open = $bindable(false),
    title,
    description,
    size = 'md',
    children,
    actions,
    onclose,
    persistent = false
  }: Props = $props();

  const uid = $props.id();
  let dialog: HTMLDialogElement | undefined = $state();

  const widths = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-2xl' };

  $effect(() => {
    if (!dialog) {
      return;
    }
    if (open && !dialog.open) {
      dialog.showModal();
    } else if (!open && dialog.open) {
      dialog.close();
    }
  });

  function closed() {
    open = false;
    onclose?.();
  }

  function backdrop(event: MouseEvent) {
    // The dialog element itself has no padding, so a click whose target is
    // the dialog landed on the backdrop around the content.
    if (!persistent && event.target === dialog) {
      dialog?.close();
    }
  }
</script>

<dialog
  bind:this={dialog}
  aria-labelledby="{uid}-title"
  aria-describedby={description ? `${uid}-description` : undefined}
  class="m-auto w-[calc(100%-2rem)] {widths[
    size
  ]} rounded-lg border border-border bg-surface p-0 text-fg shadow-xl backdrop:bg-overlay"
  onclose={closed}
  onclick={backdrop}
>
  <div class="flex flex-col gap-4 p-6">
    <h2 id="{uid}-title" class="text-lg font-semibold">{title}</h2>
    {#if description}
      <p id="{uid}-description" class="text-sm text-fg-muted">{description}</p>
    {/if}
    {#if children}{@render children()}{/if}
    {#if actions}
      <div class="flex flex-wrap justify-end gap-2">{@render actions()}</div>
    {/if}
  </div>
</dialog>
