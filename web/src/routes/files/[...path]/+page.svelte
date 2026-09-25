<!-- The files area (FR-002). The file browser itself comes in S02.3. -->
<script lang="ts">
  import { page } from '$app/state';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import { basename, crumbs, pathFromParam } from '$lib/util/paths';

  const path = $derived(pathFromParam(page.params.path));
</script>

<svelte:head>
  <title>{path === '/' ? 'Files' : basename(path)} · local-ai-nas</title>
</svelte:head>

<div class="flex flex-col gap-4 p-4">
  <Breadcrumbs crumbs={crumbs(path)} label="Folder" />
</div>
