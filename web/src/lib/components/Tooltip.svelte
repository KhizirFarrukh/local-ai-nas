<!--
  A short text shown above an element while the pointer rests on it or it
  has keyboard focus. The element gets the text as its description:
  render it with the attributes passed to `children`.
-->
<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    text: string;
    children: Snippet<[{ 'aria-describedby': string }]>;
  }

  let { text, children }: Props = $props();
  const uid = $props.id();
</script>

<span class="group relative inline-flex">
  {@render children({ 'aria-describedby': `tip-${uid}` })}
  <span
    id="tip-{uid}"
    role="tooltip"
    class="pointer-events-none absolute bottom-full left-1/2 z-40 mb-2 -translate-x-1/2 rounded bg-fg px-2 py-1 text-xs whitespace-nowrap text-bg opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100"
  >
    {text}
  </span>
</span>
