<!-- The files area (FR-002). The file browser itself comes in S02.3. -->
<script lang="ts">
  import { page } from '$app/state';
  import { api, unwrap } from '$lib/api/client';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import ErrorPanel from '$lib/components/ErrorPanel.svelte';
  import Spinner from '$lib/components/Spinner.svelte';
  import { activity } from '$lib/shell/activity.svelte';
  import { basename, crumbs, pathFromParam } from '$lib/util/paths';

  const path = $derived(pathFromParam(page.params.path));

  let attempt = $state(0);
  let folder: ReturnType<typeof load> | undefined = $state();

  function load(path: string) {
    return unwrap(api.GET('/files/items', { params: { query: { path, limit: 1 } } }));
  }

  // Loading changes state (the activity counter), so it runs in an effect,
  // never in a $derived (bug S02-B03).
  $effect(() => {
    void attempt; // "Try again" loads again
    const pending = load(path);
    folder = pending;
    activity.track(pending).catch(() => {});
  });
</script>

<svelte:head>
  <title>{path === '/' ? 'Files' : basename(path)} · local-ai-nas</title>
</svelte:head>

<div class="flex flex-col gap-4 p-4">
  <Breadcrumbs crumbs={crumbs(path)} label="Folder" />
  {#if folder}
    {#await folder}
      <Spinner label="Loading the folder" />
    {:then result}
      <p class="text-sm text-fg-muted">{result.item.path}</p>
    {:catch error}
      <ErrorPanel {error} onretry={() => attempt++} />
    {/await}
  {/if}
</div>
