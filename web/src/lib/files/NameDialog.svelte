<!--
  Asks for a name (S02.5-T02): for a new folder, or to rename an item. When
  the server refuses the name, its rule shows under the field in plain words
  (docs/api/errors.md, Name rules) and the dialog stays open to fix it. For a
  file, the part before the extension is selected, as file managers do.
-->
<script lang="ts">
  import { describe } from '$lib/api/messages';
  import Button from '$lib/components/Button.svelte';
  import Dialog from '$lib/components/Dialog.svelte';
  import TextField from '$lib/components/TextField.svelte';

  interface Props {
    open: boolean;
    title: string;
    label: string;
    /** The button that does it, such as "Create" or "Rename". */
    submitLabel: string;
    /** The name the field starts with. */
    initial?: string;
    /** Select only the part before the extension (for files). */
    selectStem?: boolean;
    /** Does the work; what it throws is shown under the field. */
    onsubmit: (name: string) => Promise<void>;
  }

  let {
    open = $bindable(false),
    title,
    label,
    submitLabel,
    initial = '',
    selectStem = false,
    onsubmit
  }: Props = $props();

  const uid = $props.id();
  let name = $state('');
  let error = $state<string | undefined>(undefined);
  let busy = $state(false);
  let input: HTMLInputElement | undefined = $state();

  // Each opening starts from the initial name, selected. The dialog shows
  // itself in an effect too, so the selection waits for that.
  $effect(() => {
    if (!open) {
      return;
    }
    name = initial;
    error = undefined;
    busy = false;
    const timer = setTimeout(() => {
      const dot = initial.lastIndexOf('.');
      input?.focus();
      input?.setSelectionRange(0, selectStem && dot > 0 ? dot : initial.length);
    }, 0);
    return () => clearTimeout(timer);
  });

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    if (name === '' || busy) {
      return;
    }
    busy = true;
    error = undefined;
    try {
      await onsubmit(name);
      open = false;
    } catch (err) {
      const m = describe(err);
      error = `${m.title}. ${m.message}`;
      input?.focus();
    } finally {
      busy = false;
    }
  }
</script>

<Dialog bind:open {title} persistent>
  <form id="{uid}-form" onsubmit={submit}>
    <TextField
      {label}
      bind:value={name}
      bind:input
      {error}
      autocomplete="off"
      spellcheck="false"
      oninput={() => (error = undefined)}
    />
  </form>
  {#snippet actions()}
    <Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
    <Button variant="primary" type="submit" form="{uid}-form" disabled={name === '' || busy}>
      {submitLabel}
    </Button>
  {/snippet}
</Dialog>
