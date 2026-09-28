<!--
  Asks what to do about a taken name (S02.5-T03, FR-081): replace (files
  only), keep both, or skip, with "apply to all" when the operation has
  more items. It lives in the layout, since uploads go on across pages.
  Escape skips this item, or stops the operation when it can be stopped.
-->
<script lang="ts">
  import Button from '$lib/components/Button.svelte';
  import Checkbox from '$lib/components/Checkbox.svelte';
  import Dialog from '$lib/components/Dialog.svelte';
  import { basename, parent } from '$lib/util/paths';
  import { conflicts, type ConflictAnswer } from './conflicts.svelte';

  let open = $state(false);
  let all = $state(false);

  const question = $derived(conflicts.current?.question);

  $effect(() => {
    open = question !== undefined;
    all = false; // each question starts unticked
  });

  const where = $derived.by(() => {
    const folder = question ? parent(question.target) : '/';
    return folder === '/' ? 'Files' : `“${basename(folder)}”`;
  });

  /** The name "keep both" is likely to give: "a (1).txt", "photos (1)". */
  const numbered = $derived.by(() => {
    if (!question) {
      return '';
    }
    const dot = question.name.lastIndexOf('.');
    return question.folder || dot <= 0
      ? `${question.name} (1)`
      : `${question.name.slice(0, dot)} (1)${question.name.slice(dot)}`;
  });

  function choose(choice: ConflictAnswer['choice']) {
    conflicts.answer({ choice, all });
  }

  // Closed without a choice (Escape): skip this item, or stop the run.
  function closed() {
    if (conflicts.current) {
      choose(conflicts.current.question.stoppable ? 'cancel' : 'skip');
    }
  }
</script>

<Dialog
  bind:open
  title="“{question?.name ?? ''}” is already there"
  description="{question?.folder
    ? 'A folder'
    : 'An item'} with this name is already in {where}. What should happen?"
  size="md"
  phone="full"
  persistent
  onclose={closed}
>
  {#if question}
    <div class="flex flex-col gap-2" role="group" aria-label="Choices">
      {#if !question.folder}
        <button
          type="button"
          class="flex flex-col items-start gap-0.5 rounded-md border border-border px-4 py-3 text-left hover:border-danger hover:bg-surface-2"
          onclick={() => choose('overwrite')}
        >
          <span class="font-medium">Replace</span>
          <span class="text-sm text-fg-muted"
            >The file in {where} is replaced by the new one. This cannot be undone.</span
          >
        </button>
      {/if}
      <button
        type="button"
        class="flex flex-col items-start gap-0.5 rounded-md border border-border px-4 py-3 text-left hover:border-accent hover:bg-surface-2"
        onclick={() => choose('rename')}
      >
        <span class="font-medium">Keep both</span>
        <span class="text-sm text-fg-muted"
          >The new one gets a numbered name, such as “{numbered}”.</span
        >
      </button>
      <button
        type="button"
        class="flex flex-col items-start gap-0.5 rounded-md border border-border px-4 py-3 text-left hover:border-accent hover:bg-surface-2"
        onclick={() => choose('skip')}
      >
        <span class="font-medium">Skip</span>
        <span class="text-sm text-fg-muted"
          >Leave the one that is there, and do nothing with this one.</span
        >
      </button>
    </div>
    {#if question.many}
      <Checkbox
        bind:checked={all}
        label="Do the same for the other {question.folder ? 'folders' : 'files'} with taken names"
      />
    {/if}
  {/if}
  {#snippet actions()}
    {#if question?.stoppable && question.many}
      <Button variant="ghost" onclick={() => choose('cancel')}>Stop the operation</Button>
    {/if}
  {/snippet}
</Dialog>
