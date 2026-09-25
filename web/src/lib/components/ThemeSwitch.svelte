<!-- Picks the theme: the system's, light, or dark (FR-083). -->
<script lang="ts">
  import Monitor from '@lucide/svelte/icons/monitor';
  import Moon from '@lucide/svelte/icons/moon';
  import Sun from '@lucide/svelte/icons/sun';
  import { theme, type ThemeChoice } from '$lib/util/theme.svelte';

  const current = theme();
  const options: { value: ThemeChoice; label: string; icon: typeof Sun }[] = [
    { value: 'system', label: 'System', icon: Monitor },
    { value: 'light', label: 'Light', icon: Sun },
    { value: 'dark', label: 'Dark', icon: Moon }
  ];
</script>

<div
  role="radiogroup"
  aria-label="Theme"
  class="inline-flex rounded-md border border-border-strong p-0.5"
>
  {#each options as option (option.value)}
    {@const Icon = option.icon}
    <button
      type="button"
      role="radio"
      aria-checked={current.choice === option.value}
      title={option.label}
      class="inline-flex items-center gap-1.5 rounded px-2 py-1 text-sm {current.choice ===
      option.value
        ? 'bg-accent text-accent-fg'
        : 'text-fg hover:bg-surface-2'}"
      onclick={() => current.set(option.value)}
    >
      <Icon class="size-4" aria-hidden="true" />
      <span>{option.label}</span>
    </button>
  {/each}
</div>
