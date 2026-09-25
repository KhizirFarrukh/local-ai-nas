<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';
  import { buttonBase, buttonVariants, type ButtonVariant } from './Button.svelte';

  interface Props extends HTMLButtonAttributes {
    /** What the button does, for screen readers and as a tooltip. */
    label: string;
    variant?: ButtonVariant;
    size?: 'sm' | 'md';
    /** The icon. */
    children: Snippet;
  }

  let {
    label,
    variant = 'ghost',
    size = 'md',
    type = 'button',
    class: extra = '',
    children,
    ...rest
  }: Props = $props();
</script>

<button
  {type}
  aria-label={label}
  title={label}
  class="{buttonBase} {buttonVariants[variant]} {size === 'sm' ? 'size-8' : 'size-10'} {extra}"
  {...rest}
>
  {@render children()}
</button>
