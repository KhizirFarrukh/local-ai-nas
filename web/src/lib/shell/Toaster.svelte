<!--
  The notifications, stacked at the bottom right (above the bottom bar on
  phones). Screen readers hear each one: Toast has the status or alert
  role.
-->
<script lang="ts">
  import Toast from '$lib/components/Toast.svelte';
  import { toasts } from './toasts.svelte';
</script>

<section
  aria-label="Notifications"
  class="pointer-events-none fixed right-4 bottom-20 z-40 flex w-[calc(100%-2rem)] max-w-sm flex-col gap-2 md:bottom-4"
>
  {#each toasts.items as toast (toast.id)}
    <Toast
      kind={toast.kind}
      message={toast.message}
      detail={toast.detail}
      onclose={() => toasts.dismiss(toast.id)}
    />
  {/each}
</section>
