<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  interface Props extends Omit<HTMLInputAttributes, 'value'> {
    label: string;
    value?: string;
    /** An error message; the field is marked invalid while it is set. */
    error?: string;
    /** A short help text under the field. */
    hint?: string;
    /** The element, for focusing and selecting from outside. */
    input?: HTMLInputElement;
  }

  let {
    label,
    value = $bindable(''),
    error,
    hint,
    input = $bindable(),
    id: givenId,
    class: extra = '',
    ...rest
  }: Props = $props();

  const uid = $props.id();
  const id = $derived(givenId ?? `field-${uid}`);
</script>

<div class="flex flex-col gap-1 {extra}">
  <label for={id} class="text-sm font-medium">{label}</label>
  <input
    {id}
    bind:this={input}
    bind:value
    aria-invalid={error ? 'true' : undefined}
    aria-describedby={error ? `${id}-error` : hint ? `${id}-hint` : undefined}
    class="h-10 rounded-md border bg-surface px-3 text-fg placeholder:text-fg-muted pointer-coarse:h-11 {error
      ? 'border-danger'
      : 'border-border-strong'}"
    {...rest}
  />
  {#if error}
    <p id="{id}-error" class="text-sm text-danger">{error}</p>
  {:else if hint}
    <p id="{id}-hint" class="text-sm text-fg-muted">{hint}</p>
  {/if}
</div>
