<script lang="ts" module>
  export interface Crumb {
    label: string;
    href: string;
  }
</script>

<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';

  interface Props {
    /** From the top level down; the last one is the current page. */
    crumbs: Crumb[];
    label?: string;
  }

  let { crumbs, label = 'Breadcrumbs' }: Props = $props();
</script>

<nav aria-label={label} class="min-w-0">
  <ol class="flex min-w-0 items-center gap-1 text-sm">
    {#each crumbs as crumb, i (crumb.href)}
      <li class="flex min-w-0 items-center gap-1">
        {#if i > 0}<ChevronRight class="size-4 shrink-0 text-fg-muted" aria-hidden="true" />{/if}
        {#if i === crumbs.length - 1}
          <span aria-current="page" class="truncate font-semibold">{crumb.label}</span>
        {:else}
          <a
            href={crumb.href}
            class="truncate rounded px-1 text-fg-muted hover:bg-surface-2 hover:text-fg"
            >{crumb.label}</a
          >
        {/if}
      </li>
    {/each}
  </ol>
</nav>
