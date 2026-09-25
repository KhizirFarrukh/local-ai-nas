<!--
  The files area (S02.3, FR-002): the folder at the address, as a list or
  a grid, with sorting, breadcrumbs, and the empty and error states.
-->
<script lang="ts">
  import Download from '@lucide/svelte/icons/download';
  import FolderOpen from '@lucide/svelte/icons/folder-open';
  import FolderUp from '@lucide/svelte/icons/folder-up';
  import LayoutGrid from '@lucide/svelte/icons/layout-grid';
  import List from '@lucide/svelte/icons/list';
  import FolderUp2 from '@lucide/svelte/icons/folder-input';
  import Upload from '@lucide/svelte/icons/upload';
  import { onDestroy } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { api, unwrap } from '$lib/api/client';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import Button from '$lib/components/Button.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import ErrorPanel from '$lib/components/ErrorPanel.svelte';
  import IconButton from '$lib/components/IconButton.svelte';
  import Spinner from '$lib/components/Spinner.svelte';
  import DownloadAction from '$lib/files/DownloadAction.svelte';
  import { downloadArchive } from '$lib/files/download';
  import FileView from '$lib/files/FileView.svelte';
  import { FolderListing, type PageLoader } from '$lib/files/listing.svelte';
  import type { FileItem, SortKey, SortOrder } from '$lib/files/types';
  import { activity } from '$lib/shell/activity.svelte';
  import DropZone from '$lib/uploads/DropZone.svelte';
  import { pickedFromInput } from '$lib/uploads/plan';
  import { startUpload } from '$lib/uploads/start';
  import { getUploader } from '$lib/uploads/state.svelte';
  import { formatCount } from '$lib/util/format';
  import { basename, crumbs, filesHref, parent, pathFromParam } from '$lib/util/paths';
  import { readStored, writeStored } from '$lib/util/stored';

  const path = $derived(pathFromParam(page.params.path));

  const sortKeys: readonly SortKey[] = ['name', 'size', 'mod_time', 'kind'];
  let mode = $state(readStored('view', ['list', 'grid'] as const, 'list'));
  let sort = $state(readStored('sort', sortKeys, 'name'));
  let order = $state(readStored('order', ['asc', 'desc'] as const, 'asc'));
  let attempt = $state(0);

  const load: PageLoader = (query) => unwrap(api.GET('/files/items', { params: { query } }));

  const listing = $derived.by(() => {
    void attempt; // "Try again" starts over
    return new FolderListing(load, path, sort, order);
  });

  // Loading writes state, so it starts in an effect (bug S02-B03).
  $effect(() => {
    void activity.track(listing.start());
  });

  // Uploads (S02.4-T01): each file goes to this folder. When one lands in
  // the folder on screen, the list refreshes in place (at most every 400 ms).
  let fileInput: HTMLInputElement | undefined = $state();
  let folderInput: HTMLInputElement | undefined = $state();
  let refreshTimer: ReturnType<typeof setTimeout> | undefined;
  let stopListening: (() => void) | undefined;

  void getUploader().then((manager) => {
    stopListening = manager.onFinished((entry) => {
      if (parent(entry.itemPath ?? entry.target) === path) {
        clearTimeout(refreshTimer);
        refreshTimer = setTimeout(() => void listing.refresh(), 400);
      }
    });
  });
  onDestroy(() => {
    stopListening?.();
    clearTimeout(refreshTimer);
  });

  // Files and folders (S02.4-T02): folders are created first, then the files
  // follow into them.
  function upload(files: File[]) {
    if (files.length > 0) {
      void startUpload(pickedFromInput(files), path);
    }
  }

  function setMode(next: 'list' | 'grid') {
    mode = next;
    writeStored('view', next);
  }

  function setSort(key: SortKey) {
    order = key === sort && order === 'asc' ? 'desc' : 'asc';
    sort = key;
    writeStored('sort', sort);
    writeStored('order', order);
  }

  function open(item: FileItem) {
    if (item.kind === 'dir') {
      goto(filesHref(item.path));
    }
    // Files open in the preview (S02.6).
  }

  const sortOptions: { value: string; label: string }[] = [
    { value: 'name:asc', label: 'Name, A to Z' },
    { value: 'name:desc', label: 'Name, Z to A' },
    { value: 'size:desc', label: 'Size, largest first' },
    { value: 'size:asc', label: 'Size, smallest first' },
    { value: 'mod_time:desc', label: 'Modified, newest first' },
    { value: 'mod_time:asc', label: 'Modified, oldest first' },
    { value: 'kind:asc', label: 'Type' }
  ];

  function pickSort(value: string) {
    const [key, dir] = value.split(':') as [SortKey, SortOrder];
    sort = key;
    order = dir;
    writeStored('sort', sort);
    writeStored('order', order);
  }
</script>

<svelte:head>
  <title>{path === '/' ? 'Files' : basename(path)} · local-ai-nas</title>
</svelte:head>

<div class="flex h-full flex-col">
  <div class="flex flex-wrap items-center gap-3 border-b border-border px-4 py-2">
    <Breadcrumbs crumbs={crumbs(path)} label="Folder" />
    <div class="ml-auto flex items-center gap-2">
      <Button variant="primary" size="sm" onclick={() => fileInput?.click()}>
        <Upload class="size-4" /> Upload
      </Button>
      <Button size="sm" onclick={() => folderInput?.click()}>
        <FolderUp2 class="size-4" /> Upload folder
      </Button>
      {#each [{ folder: false }, { folder: true }] as kind (kind.folder)}
        <input
          type="file"
          multiple
          class="hidden"
          aria-hidden="true"
          tabindex="-1"
          data-testid={kind.folder ? 'folder-input' : 'file-input'}
          {...kind.folder ? { webkitdirectory: true } : {}}
          {@attach (node: HTMLInputElement) => {
            if (kind.folder) folderInput = node;
            else fileInput = node;
          }}
          onchange={(e) => {
            // Copy the list first: clearing the input empties it (bug S02-B06).
            const files = [...(e.currentTarget.files ?? [])];
            e.currentTarget.value = '';
            upload(files);
          }}
        />
      {/each}
      {#if listing.total && listing.folder?.kind === 'dir'}
        <!-- S02.4-T04: the folder on screen as one ZIP file. -->
        <Button size="sm" onclick={() => void downloadArchive([path])}>
          <Download class="size-4" /> Download folder
        </Button>
      {/if}
      {#if listing.total !== undefined && listing.folder?.kind === 'dir'}
        <span class="text-sm text-fg-muted" data-testid="item-count"
          >{formatCount(listing.total)} {listing.total === 1 ? 'item' : 'items'}</span
        >
      {/if}
      {#if mode === 'grid'}
        <label class="flex items-center gap-2 text-sm">
          <span class="sr-only">Sort</span>
          <select
            class="h-8 rounded-md border border-border-strong bg-surface px-2 text-sm"
            value="{sort}:{order}"
            onchange={(e) => pickSort(e.currentTarget.value)}
          >
            {#each sortOptions as option (option.value)}
              <option value={option.value}>{option.label}</option>
            {/each}
          </select>
        </label>
      {/if}
      <div class="flex rounded-md border border-border-strong">
        <IconButton
          label="List view"
          size="sm"
          aria-pressed={mode === 'list'}
          onclick={() => setMode('list')}
        >
          <List class="size-4" />
        </IconButton>
        <IconButton
          label="Grid view"
          size="sm"
          aria-pressed={mode === 'grid'}
          onclick={() => setMode('grid')}
        >
          <LayoutGrid class="size-4" />
        </IconButton>
      </div>
    </div>
  </div>

  {#if listing.error}
    <ErrorPanel error={listing.error} onretry={() => attempt++} />
  {:else if listing.total === undefined}
    <div class="flex justify-center p-8"><Spinner label="Loading the folder" /></div>
  {:else if listing.folder && listing.folder.kind !== 'dir'}
    <EmptyState
      icon={FolderOpen}
      title="“{listing.folder.name}” is a file"
      description="This address shows a file, not a folder."
    >
      {#snippet actions()}
        <Button onclick={() => goto(filesHref(parent(path)))}>Show its folder</Button>
      {/snippet}
    </EmptyState>
  {:else if listing.total === 0}
    <EmptyState
      icon={FolderOpen}
      title="This folder is empty"
      description="Upload files or a whole folder, or drop them anywhere on this page."
    >
      {#snippet actions()}
        <Button variant="primary" onclick={() => fileInput?.click()}>
          <Upload class="size-4" /> Upload files
        </Button>
        <Button onclick={() => folderInput?.click()}>
          <FolderUp2 class="size-4" /> Upload a folder
        </Button>
        {#if path !== '/'}
          <Button onclick={() => goto(filesHref(parent(path)))}>
            <FolderUp class="size-4" /> Parent folder
          </Button>
        {/if}
      {/snippet}
    </EmptyState>
  {:else}
    <FileView {listing} {mode} {sort} {order} onsort={setSort} onopen={open}>
      {#snippet actions(item)}
        <DownloadAction {item} />
      {/snippet}
    </FileView>
  {/if}
</div>

<DropZone
  target={path === '/' ? 'Files' : basename(path)}
  ondrop={(dropped) => void startUpload(dropped.files, path, dropped.emptyFolders)}
/>
