<script lang="ts">
  interface Props {
    /** Done so far, from 0 to max; undefined shows an ongoing bar. */
    value?: number;
    max?: number;
    /** What is in progress, for screen readers. */
    label: string;
    class?: string;
  }

  let { value, max = 1, label, class: extra = '' }: Props = $props();

  const percent = $derived(
    value === undefined ? undefined : Math.max(0, Math.min(100, (value / max) * 100))
  );
</script>

<div
  role="progressbar"
  aria-label={label}
  aria-valuemin={0}
  aria-valuemax={100}
  aria-valuenow={percent === undefined ? undefined : Math.round(percent)}
  class="h-2 w-full overflow-hidden rounded-full bg-surface-2 {extra}"
>
  {#if percent === undefined}
    <div class="h-full w-1/3 animate-pulse rounded-full bg-accent"></div>
  {:else}
    <div class="h-full rounded-full bg-accent transition-[width]" style:width="{percent}%"></div>
  {/if}
</div>
