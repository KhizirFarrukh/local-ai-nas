<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  interface Props extends Omit<HTMLInputAttributes, 'type' | 'checked'> {
    checked?: boolean;
    /** Shown as "some selected"; set by the owner, cleared on click. */
    indeterminate?: boolean;
    /** The visible label; without it, pass aria-label. */
    label?: string;
  }

  let {
    checked = $bindable(false),
    indeterminate = $bindable(false),
    label,
    class: extra = '',
    ...rest
  }: Props = $props();
</script>

<label class="inline-flex items-center gap-2 text-sm {extra}">
  <input
    type="checkbox"
    bind:checked
    bind:indeterminate
    class="size-4 rounded border-border-strong accent-accent"
    {...rest}
  />
  {#if label}<span>{label}</span>{/if}
</label>
