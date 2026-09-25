<!-- One notification. The toaster (S02.2-T03) stacks them in a live region. -->
<script lang="ts" module>
  export type ToastKind = 'info' | 'success' | 'error';
</script>

<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import Info from '@lucide/svelte/icons/info';
  import X from '@lucide/svelte/icons/x';

  interface Props {
    kind?: ToastKind;
    message: string;
    /** More detail, such as the request ID of a failed request. */
    detail?: string;
    onclose?: () => void;
  }

  let { kind = 'info', message, detail, onclose }: Props = $props();

  const icons = { info: Info, success: CircleCheck, error: CircleAlert };
  const colors = { info: 'text-accent', success: 'text-success', error: 'text-danger' };
  const Icon = $derived(icons[kind]);
</script>

<div
  role={kind === 'error' ? 'alert' : 'status'}
  class="pointer-events-auto flex w-full max-w-sm items-start gap-3 rounded-lg border border-border bg-surface p-3 text-sm text-fg shadow-lg"
>
  <Icon class="mt-0.5 size-5 shrink-0 {colors[kind]}" aria-hidden="true" />
  <div class="flex-1">
    <p>{message}</p>
    {#if detail}<p class="mt-1 text-xs whitespace-pre-line text-fg-muted">{detail}</p>{/if}
  </div>
  {#if onclose}
    <button
      type="button"
      aria-label="Dismiss"
      class="-m-1 rounded p-1 text-fg-muted hover:bg-surface-2"
      onclick={onclose}
    >
      <X class="size-4" />
    </button>
  {/if}
</div>
