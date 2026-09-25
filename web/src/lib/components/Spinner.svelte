<script lang="ts">
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';

  interface Props {
    /** What is loading, for screen readers. */
    label?: string;
    class?: string;
  }

  let { label = 'Loading', class: extra = 'size-5' }: Props = $props();
</script>

<span role="status" class="inline-flex items-center text-fg-muted">
  <LoaderCircle class="animate-spin {extra}" aria-hidden="true" />
  <span class="sr-only">{label}</span>
</span>
