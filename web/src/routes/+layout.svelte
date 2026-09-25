<!--
  The app shell (S02.2-T01, FR-079): the header, the navigation (a side bar
  on wide screens, a bottom bar on phones), and the page.
-->
<script lang="ts">
  import '../app.css';
  import HardDrive from '@lucide/svelte/icons/hard-drive';
  import Monitor from '@lucide/svelte/icons/monitor';
  import Moon from '@lucide/svelte/icons/moon';
  import Sun from '@lucide/svelte/icons/sun';
  import { navigating } from '$app/state';
  import IconButton from '$lib/components/IconButton.svelte';
  import { activity } from '$lib/shell/activity.svelte';
  import ConnectionBanner from '$lib/shell/ConnectionBanner.svelte';
  import { connection } from '$lib/shell/connection.svelte';
  import Nav from '$lib/shell/Nav.svelte';
  import { theme, type ThemeChoice } from '$lib/util/theme.svelte';

  let { children } = $props();

  // Apply the remembered theme before the first page renders.
  const current = theme();

  const nextTheme: Record<ThemeChoice, ThemeChoice> = {
    system: 'light',
    light: 'dark',
    dark: 'system'
  };
  const themeIcons = { system: Monitor, light: Sun, dark: Moon };
  const themeNames = { system: 'system theme', light: 'light theme', dark: 'dark theme' };
  const ThemeIcon = $derived(themeIcons[current.choice]);

  const conn = connection();
  const busy = $derived(navigating.to !== null || activity.pending > 0);
</script>

<svelte:window ononline={() => conn.retry()} />

<a
  href="#main"
  class="sr-only z-50 rounded-md bg-surface px-4 py-2 focus:not-sr-only focus:fixed focus:top-2 focus:left-2"
  >Skip to content</a
>

<div class="flex h-full flex-col">
  <header class="flex h-14 shrink-0 items-center gap-3 border-b border-border bg-surface px-4">
    <a href="/files" class="flex items-center gap-2 font-semibold">
      <HardDrive class="size-6 text-accent" aria-hidden="true" />
      <span>local-ai-nas</span>
    </a>
    <div class="ml-auto">
      <IconButton
        label="Theme: {themeNames[current.choice]}. Switch to the {themeNames[
          nextTheme[current.choice]
        ]}."
        onclick={() => current.set(nextTheme[current.choice])}
      >
        <ThemeIcon class="size-5" />
      </IconButton>
    </div>
  </header>
  <div class="relative h-0.5 shrink-0 overflow-hidden" aria-hidden="true">
    {#if busy}<div class="absolute inset-y-0 w-1/3 animate-pulse bg-accent"></div>{/if}
  </div>
  <ConnectionBanner />

  <div class="flex min-h-0 flex-1">
    <aside class="hidden w-56 shrink-0 border-r border-border bg-surface md:block">
      <Nav layout="side" />
    </aside>
    <main id="main" tabindex="-1" class="min-w-0 flex-1 overflow-auto focus:outline-none">
      {@render children()}
    </main>
  </div>

  <div class="shrink-0 border-t border-border bg-surface md:hidden">
    <Nav layout="bottom" />
  </div>
</div>
