<!--
  The files area (S02.3, FR-002): the folder at the address, as a list or
  a grid, with sorting, breadcrumbs, and the empty and error states. While
  items are selected (S02.5-T01), the toolbar shows the selection's count
  and actions instead; the operations open their dialogs (S02.5-T02).
-->
<script lang="ts">
  import Copy from '@lucide/svelte/icons/copy';
  import Download from '@lucide/svelte/icons/download';
  import FolderOpen from '@lucide/svelte/icons/folder-open';
  import FolderOutput from '@lucide/svelte/icons/folder-output';
  import FolderPlus from '@lucide/svelte/icons/folder-plus';
  import FolderUp from '@lucide/svelte/icons/folder-up';
  import LayoutGrid from '@lucide/svelte/icons/layout-grid';
  import List from '@lucide/svelte/icons/list';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Trash from '@lucide/svelte/icons/trash-2';
  import FolderUp2 from '@lucide/svelte/icons/folder-input';
  import Upload from '@lucide/svelte/icons/upload';
  import X from '@lucide/svelte/icons/x';
  import { onDestroy, untrack } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { api, unwrap } from '$lib/api/client';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import Button from '$lib/components/Button.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import ErrorPanel from '$lib/components/ErrorPanel.svelte';
  import IconButton from '$lib/components/IconButton.svelte';
  import Spinner from '$lib/components/Spinner.svelte';
  import DeleteDialog from '$lib/files/DeleteDialog.svelte';
  import DownloadAction from '$lib/files/DownloadAction.svelte';
  import { downloadArchive, downloadSelection } from '$lib/files/download';
  import FileView from '$lib/files/FileView.svelte';
  import FolderPicker from '$lib/files/FolderPicker.svelte';
  import { FolderListing, type PageLoader } from '$lib/files/listing.svelte';
  import NameDialog from '$lib/files/NameDialog.svelte';
  import {
    createFolder,
    itemsText,
    renameItem,
    runBulk,
    type BulkKind
  } from '$lib/files/operations';
  import { Selection } from '$lib/files/selection.svelte';
  import type { FileItem, SortKey, SortOrder } from '$lib/files/types';
  import { activity } from '$lib/shell/activity.svelte';
  import { toasts } from '$lib/shell/toasts.svelte';
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

  // The selection (S02.5-T01) belongs to the folder: another folder starts
  // with none. A new sort keeps it, since items are kept by path.
  const selection = new Selection();
  const selected = $derived(selection.count(listing.total));
  $effect(() => {
    void path;
    untrack(() => selection.clear());
  });

  // Ctrl/Cmd+A and Escape also work when nothing has the focus, as after
  // a click on the page's background.
  function pageKeys(event: KeyboardEvent) {
    if (document.activeElement !== document.body || listing.total === undefined) {
      return;
    }
    const ctrl = event.ctrlKey || event.metaKey;
    if (ctrl && event.key.toLowerCase() === 'a' && !event.shiftKey && !event.altKey) {
      event.preventDefault();
      selection.selectAll();
    } else if (event.key === 'Escape' && selected > 0) {
      selection.clear();
    }
  }

  // Operations (S02.5-T02). Each dialog works on the items that were
  // selected when it opened; afterwards the list refreshes, and items that
  // moved away or were deleted leave the selection.
  let newFolderOpen = $state(false);
  let renameOpen = $state(false);
  let moveOpen = $state(false);
  let copyOpen = $state(false);
  let deleteOpen = $state(false);
  let targets = $state.raw<FileItem[]>([]);

  /** Takes the selected items for a dialog; false when there are none. */
  async function takeSelection(): Promise<boolean> {
    const items = await selection.resolve(listing);
    if (!items) {
      toasts.push({ kind: 'error', message: 'The folder’s items could not be loaded. Try again.' });
      return false;
    }
    targets = items;
    return items.length > 0;
  }

  async function bulk(kind: BulkKind, folder = path) {
    const result = await runBulk(kind, targets, folder);
    if (kind !== 'copy') {
      selection.forget(result.done.map((item) => item.path));
    }
    await listing.refresh();
  }

  async function newFolder(name: string) {
    const folder = await createFolder(path, name);
    toasts.push({ kind: 'success', message: `Created the folder “${folder.name}”.` });
    await listing.refresh();
  }

  async function rename(name: string) {
    const [item] = targets;
    await renameItem(item, name);
    selection.forget([item.path]);
    await listing.refresh();
  }

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

<svelte:window onkeydown={pageKeys} />

<div class="flex h-full flex-col">
  <div class="flex flex-wrap items-center gap-3 border-b border-border px-4 py-2">
    <Breadcrumbs crumbs={crumbs(path)} label="Folder" />
    {#if selected > 0}
      <div
        class="ml-auto flex items-center gap-2"
        role="toolbar"
        aria-label="Selected items"
        data-testid="selection-bar"
      >
        <span class="text-sm font-medium" aria-live="polite">{formatCount(selected)} selected</span>
        <Button
          size="sm"
          variant="primary"
          onclick={() => void downloadSelection(selection, listing, path)}
        >
          <Download class="size-4" /> Download
        </Button>
        {#if selected === 1}
          <Button size="sm" onclick={async () => (renameOpen = await takeSelection())}>
            <Pencil class="size-4" /> Rename
          </Button>
        {/if}
        <Button size="sm" onclick={async () => (moveOpen = await takeSelection())}>
          <FolderOutput class="size-4" /> Move
        </Button>
        <Button size="sm" onclick={async () => (copyOpen = await takeSelection())}>
          <Copy class="size-4" /> Copy
        </Button>
        <Button
          size="sm"
          variant="danger"
          onclick={async () => (deleteOpen = await takeSelection())}
        >
          <Trash class="size-4" /> Delete
        </Button>
        {#if !selection.everything}
          <Button size="sm" onclick={() => selection.selectAll()}>Select all</Button>
        {/if}
        <IconButton label="Clear the selection" size="sm" onclick={() => selection.clear()}>
          <X class="size-4" />
        </IconButton>
      </div>
    {/if}
    <div class="ml-auto flex items-center gap-2 {selected > 0 ? 'hidden' : ''}">
      <Button variant="primary" size="sm" onclick={() => fileInput?.click()}>
        <Upload class="size-4" /> Upload
      </Button>
      <Button size="sm" onclick={() => folderInput?.click()}>
        <FolderUp2 class="size-4" /> Upload folder
      </Button>
      {#if listing.folder?.kind === 'dir'}
        <Button size="sm" onclick={() => (newFolderOpen = true)}>
          <FolderPlus class="size-4" /> New folder
        </Button>
      {/if}
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
      description="Upload files or a whole folder, drop them anywhere on this page, or make a new folder."
    >
      {#snippet actions()}
        <Button variant="primary" onclick={() => fileInput?.click()}>
          <Upload class="size-4" /> Upload files
        </Button>
        <Button onclick={() => folderInput?.click()}>
          <FolderUp2 class="size-4" /> Upload a folder
        </Button>
        <Button onclick={() => (newFolderOpen = true)}>
          <FolderPlus class="size-4" /> New folder
        </Button>
        {#if path !== '/'}
          <Button onclick={() => goto(filesHref(parent(path)))}>
            <FolderUp class="size-4" /> Parent folder
          </Button>
        {/if}
      {/snippet}
    </EmptyState>
  {:else}
    <FileView {listing} {selection} {mode} {sort} {order} onsort={setSort} onopen={open}>
      {#snippet actions(item)}
        <DownloadAction {item} />
      {/snippet}
    </FileView>
  {/if}
</div>

<NameDialog
  bind:open={newFolderOpen}
  title="New folder"
  label="Folder name"
  submitLabel="Create"
  initial="New folder"
  onsubmit={newFolder}
/>
<NameDialog
  bind:open={renameOpen}
  title="Rename {itemsText(targets)}"
  label="New name"
  submitLabel="Rename"
  initial={targets[0]?.name ?? ''}
  selectStem={targets[0]?.kind === 'file'}
  onsubmit={rename}
/>
<FolderPicker
  bind:open={moveOpen}
  title="Move {itemsText(targets)}"
  submitLabel="Move here"
  start={path}
  items={targets}
  move
  onpick={(folder) => void bulk('move', folder)}
/>
<FolderPicker
  bind:open={copyOpen}
  title="Copy {itemsText(targets)}"
  submitLabel="Copy here"
  start={path}
  items={targets}
  onpick={(folder) => void bulk('copy', folder)}
/>
<DeleteDialog bind:open={deleteOpen} items={targets} onconfirm={() => void bulk('delete')} />

<DropZone
  target={path === '/' ? 'Files' : basename(path)}
  ondrop={(dropped) => void startUpload(dropped.files, path, dropped.emptyFolders)}
/>
