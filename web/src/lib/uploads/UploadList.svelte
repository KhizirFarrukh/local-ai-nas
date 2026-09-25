<!--
  The upload queue (S02.4-T01, FR-080): one line per file, with progress,
  speed, and the actions that fit its state. Errors are explained in plain
  words (describe()).
-->
<script lang="ts">
  import Pause from '@lucide/svelte/icons/pause';
  import Play from '@lucide/svelte/icons/play';
  import RotateCw from '@lucide/svelte/icons/rotate-cw';
  import X from '@lucide/svelte/icons/x';
  import { describe } from '$lib/api/messages';
  import IconButton from '$lib/components/IconButton.svelte';
  import ProgressBar from '$lib/components/ProgressBar.svelte';
  import { formatSize } from '$lib/util/format';
  import { parent } from '$lib/util/paths';
  import type { Uploader, UploadEntry } from './uploader.svelte';

  interface Props {
    manager: Uploader;
  }

  let { manager }: Props = $props();

  function status(e: UploadEntry): string {
    switch (e.status) {
      case 'queued':
        return 'Waiting';
      case 'uploading':
        return `${formatSize(e.uploaded)} of ${formatSize(e.size)}${e.speed > 0 ? ` · ${formatSize(e.speed)}/s` : ''}`;
      case 'paused':
        return `Paused at ${formatSize(e.uploaded)} of ${formatSize(e.size)}`;
      case 'done':
        return e.itemPath && e.itemPath !== e.target ? `Done, saved as ${e.itemPath}` : 'Done';
      case 'cancelled':
        return 'Cancelled';
      case 'failed': {
        const m = describe(e.error);
        return `${m.title}. ${m.message}`;
      }
    }
  }
</script>

<ul class="flex flex-col gap-3">
  {#each manager.entries as entry (entry.id)}
    <li class="flex flex-col gap-1" data-testid="upload">
      <div class="flex items-center gap-2 text-sm">
        <span class="min-w-0 flex-1 truncate" title="{entry.name} → {parent(entry.target)}"
          >{entry.name}</span
        >
        {#if entry.status === 'uploading' || entry.status === 'queued'}
          <IconButton label="Pause {entry.name}" size="sm" onclick={() => manager.pause(entry)}>
            <Pause class="size-4" />
          </IconButton>
        {:else if entry.status === 'paused'}
          <IconButton label="Resume {entry.name}" size="sm" onclick={() => manager.resume(entry)}>
            <Play class="size-4" />
          </IconButton>
        {:else if entry.status === 'failed'}
          <IconButton label="Retry {entry.name}" size="sm" onclick={() => manager.retry(entry)}>
            <RotateCw class="size-4" />
          </IconButton>
        {/if}
        {#if entry.status !== 'done' && entry.status !== 'cancelled'}
          <IconButton label="Cancel {entry.name}" size="sm" onclick={() => manager.cancel(entry)}>
            <X class="size-4" />
          </IconButton>
        {/if}
      </div>
      {#if entry.status !== 'failed' && entry.status !== 'cancelled'}
        <ProgressBar value={entry.uploaded} max={entry.size || 1} label="Uploading {entry.name}" />
      {/if}
      <span
        class="text-xs {entry.status === 'failed' ? 'text-danger' : 'text-fg-muted'}"
        data-testid="upload-status">{status(entry)}</span
      >
    </li>
  {/each}
</ul>
