<!--
  Files and folders dropped anywhere on the page upload into the folder on
  screen (S02.4-T02). While files are dragged over the page, an outline
  says where they will go.
-->
<script lang="ts">
  import Upload from '@lucide/svelte/icons/upload';
  import { carriesFiles, readDrop, type Dropped } from './drop';

  interface Props {
    /** The folder name shown while dragging. */
    target: string;
    ondrop: (dropped: Dropped) => void;
  }

  let { target, ondrop }: Props = $props();

  let depth = $state(0); // dragenter and dragleave come in pairs per element

  function enter(event: DragEvent) {
    if (carriesFiles(event)) {
      event.preventDefault();
      depth++;
    }
  }

  function over(event: DragEvent) {
    if (carriesFiles(event)) {
      event.preventDefault();
      if (event.dataTransfer) {
        event.dataTransfer.dropEffect = 'copy';
      }
    }
  }

  function leave(event: DragEvent) {
    if (carriesFiles(event)) {
      depth = Math.max(0, depth - 1);
    }
  }

  async function drop(event: DragEvent) {
    if (!carriesFiles(event) || !event.dataTransfer) {
      return;
    }
    event.preventDefault();
    depth = 0;
    ondrop(await readDrop(event.dataTransfer));
  }
</script>

<svelte:window ondragenter={enter} ondragover={over} ondragleave={leave} ondrop={drop} />

{#if depth > 0}
  <div
    class="pointer-events-none fixed inset-0 z-30 flex items-center justify-center bg-overlay"
    role="status"
  >
    <div
      class="flex flex-col items-center gap-3 rounded-xl border-2 border-dashed border-accent bg-surface px-10 py-8 text-center shadow-xl"
    >
      <Upload class="size-10 text-accent" aria-hidden="true" />
      <p class="text-lg font-semibold">Drop to upload</p>
      <p class="text-sm text-fg-muted">into {target}</p>
    </div>
  </div>
{/if}
