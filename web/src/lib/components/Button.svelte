<script lang="ts" module>
  export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost';
  export type ButtonSize = 'sm' | 'md';

  export const buttonVariants: Record<ButtonVariant, string> = {
    primary: 'border-transparent bg-accent text-accent-fg hover:bg-accent-hover',
    secondary: 'border-border-strong bg-surface text-fg hover:bg-surface-2',
    danger: 'border-transparent bg-danger text-danger-fg hover:bg-danger-hover',
    ghost: 'border-transparent bg-transparent text-fg hover:bg-surface-2'
  };

  export const buttonBase =
    'inline-flex shrink-0 items-center justify-center gap-2 rounded-md border font-medium whitespace-nowrap transition-colors select-none disabled:pointer-events-none disabled:opacity-50';
</script>

<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  interface Props extends HTMLButtonAttributes {
    variant?: ButtonVariant;
    size?: ButtonSize;
    children: Snippet;
  }

  let {
    variant = 'secondary',
    size = 'md',
    type = 'button',
    class: extra = '',
    children,
    ...rest
  }: Props = $props();

  const sizes: Record<ButtonSize, string> = {
    sm: 'h-8 px-3 text-sm',
    md: 'h-10 px-4 text-sm'
  };
</script>

<button {type} class="{buttonBase} {buttonVariants[variant]} {sizes[size]} {extra}" {...rest}>
  {@render children()}
</button>
