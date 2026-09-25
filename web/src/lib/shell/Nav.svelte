<!-- The navigation between sections: a side bar, or a bottom bar on phones. -->
<script lang="ts">
  import { page } from '$app/state';
  import { inSection, sections } from './sections';

  interface Props {
    layout: 'side' | 'bottom';
  }

  let { layout }: Props = $props();
</script>

<nav aria-label="Sections" class={layout === 'side' ? 'flex flex-col gap-1 p-2' : 'flex'}>
  {#each sections as section (section.href)}
    {@const Icon = section.icon}
    {@const active = inSection(page.url.pathname, section)}
    <a
      href={section.href}
      aria-current={active ? 'page' : undefined}
      class={layout === 'side'
        ? `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium pointer-coarse:py-3 ${active ? 'bg-accent-soft text-fg' : 'text-fg-muted hover:bg-surface-2 hover:text-fg'}`
        : `flex flex-1 flex-col items-center gap-0.5 py-2 text-xs font-medium ${active ? 'text-accent' : 'text-fg-muted'}`}
    >
      <Icon class={layout === 'side' ? 'size-5' : 'size-6'} aria-hidden="true" />
      <span>{section.label}</span>
    </a>
  {/each}
</nav>
