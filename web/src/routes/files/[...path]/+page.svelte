<!--
  The files area (S02.3, FR-002): the folder at the address, as a list or
  a grid, with sorting, breadcrumbs, and the empty and error states. While
  items are selected (S02.5-T01), the toolbar shows the selection's count
  and actions instead; the operations open their dialogs (S02.5-T02).
  Every action is also in the menus (right-click, long press, the menu
  key) and has a shortcut; "?" lists them (S02.5-T04).
-->
<script lang="ts">
  import ClipboardPaste from '@lucide/svelte/icons/clipboard-paste';
  import Copy from '@lucide/svelte/icons/copy';
  import CopyPlus from '@lucide/svelte/icons/copy-plus';
  import Download from '@lucide/svelte/icons/download';
  import Eye from '@lucide/svelte/icons/eye';
  import FolderOpen from '@lucide/svelte/icons/folder-open';
  import FolderOutput from '@lucide/svelte/icons/folder-output';
  import FolderPlus from '@lucide/svelte/icons/folder-plus';
  import FolderUp from '@lucide/svelte/icons/folder-up';
  import LayoutGrid from '@lucide/svelte/icons/layout-grid';
  import List from '@lucide/svelte/icons/list';
  import Keyboard from '@lucide/svelte/icons/keyboard';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Scissors from '@lucide/svelte/icons/scissors';
  import SquareCheck from '@lucide/svelte/icons/square-check-big';
  import Trash from '@lucide/svelte/icons/trash-2';
  import FolderUp2 from '@lucide/svelte/icons/folder-input';
  import Upload from '@lucide/svelte/icons/upload';
  import X from '@lucide/svelte/icons/x';
  import { onDestroy, untrack } from 'svelte';
  import { goto, pushState, replaceState } from '$app/navigation';
  import { page } from '$app/state';
  import { api, unwrap } from '$lib/api/client';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import Button from '$lib/components/Button.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import ErrorPanel from '$lib/components/ErrorPanel.svelte';
  import IconButton from '$lib/components/IconButton.svelte';
  import Menu, { type MenuEntry } from '$lib/components/Menu.svelte';
  import Spinner from '$lib/components/Spinner.svelte';
  import DeleteDialog from '$lib/files/DeleteDialog.svelte';
  import DownloadAction from '$lib/files/DownloadAction.svelte';
  import { downloadArchive, downloadSelection } from '$lib/files/download';
  import FileView from '$lib/files/FileView.svelte';
  import FolderPicker from '$lib/files/FolderPicker.svelte';
  import { isMac, keyAction, shortcutLabel } from '$lib/files/keys';
  import { FolderListing, type PageLoader } from '$lib/files/listing.svelte';
  import { clipboard } from '$lib/files/clipboard.svelte';
  import { ConflictBatch, conflicts } from '$lib/files/conflicts.svelte';
  import NameDialog from '$lib/files/NameDialog.svelte';
  import {
    createFolder,
    itemsText,
    renameItem,
    runBulk,
    type BulkKind
  } from '$lib/files/operations';
  import { Selection } from '$lib/files/selection.svelte';
  import PreviewFrame from '$lib/previews/PreviewFrame.svelte';
  import ShortcutsDialog from '$lib/files/ShortcutsDialog.svelte';
  import type { FileItem, SortKey, SortOrder } from '$lib/files/types';
  import { activity } from '$lib/shell/activity.svelte';
  import { ApiError } from '$lib/api/errors';
  import { toasts } from '$lib/shell/toasts.svelte';
  import DropZone from '$lib/uploads/DropZone.svelte';
  import { pickedFromInput } from '$lib/uploads/plan';
  import { startUpload } from '$lib/uploads/start';
  import { getUploader } from '$lib/uploads/state.svelte';
  import { formatCount } from '$lib/util/format';
  import { basename, child, crumbs, filesHref, parent, pathFromParam } from '$lib/util/paths';
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

  // Shortcuts (S02.5-T04, keys.ts) work while the file view has the focus,
  // or the page itself (the body, or the main area a click on the
  // background focuses); never while typing, on a button, or in a dialog or
  // menu. Ctrl/Cmd+A and Escape work on the page too; in the view, it
  // handles them itself.
  const mac = isMac();
  function pageKeys(event: KeyboardEvent) {
    const target = event.target as HTMLElement;
    const onBody = target === document.body || target.tagName === 'MAIN';
    if (event.defaultPrevented || listing.total === undefined) {
      return;
    }
    if (!onBody && !target.closest('[role="grid"]')) {
      return;
    }
    const ctrl = event.ctrlKey || event.metaKey;
    if (onBody && ctrl && event.key.toLowerCase() === 'a' && !event.shiftKey && !event.altKey) {
      event.preventDefault();
      selection.selectAll();
      return;
    }
    if (onBody && event.key === 'Escape' && selected > 0) {
      selection.clear();
      return;
    }
    const action = keyAction(event, mac);
    if (!action || listing.folder?.kind !== 'dir') {
      return;
    }
    event.preventDefault();
    switch (action) {
      case 'rename':
        if (selected === 1) void openDialog('rename');
        break;
      case 'delete':
        if (selected > 0) void openDialog('delete');
        break;
      case 'new-folder':
        newFolderOpen = true;
        break;
      case 'parent':
        if (path !== '/') {
          focusView = fromView() || onBody;
          void goto(filesHref(parent(path)));
        }
        break;
      case 'copy':
      case 'cut':
        if (selected > 0) void toClipboard(action === 'copy' ? 'copy' : 'move');
        break;
      case 'paste':
        void paste();
        break;
      case 'help':
        helpOpen = true;
        break;
      case 'menu':
        showMenu(undefined, 280, 120); // nothing focused: the folder's menu
        break;
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

  type DialogKind = 'rename' | 'move' | 'copy' | 'delete';

  /** Opens an operation's dialog for the selected items. */
  async function openDialog(kind: DialogKind) {
    if (!(await takeSelection())) {
      return;
    }
    if (kind === 'rename') renameOpen = true;
    else if (kind === 'move') moveOpen = true;
    else if (kind === 'copy') copyOpen = true;
    else deleteOpen = true;
  }

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

  // A taken name asks the conflict dialog (S02.5-T03); "apply to all"
  // holds for the rest of this run.
  async function bulk(kind: BulkKind, folder = path, items: FileItem[] = targets) {
    const batch = new ConflictBatch(conflicts, items.length, true);
    const result = await runBulk(kind, items, folder, {
      onConflict: (item) => batch.choose(item.name, item.kind === 'dir', child(folder, item.name))
    });
    if (kind !== 'copy') {
      selection.forget(result.done.map((item) => item.path));
    }
    await listing.refresh();
  }

  // The app's clipboard (S02.5-T04): Ctrl/Cmd+C or X here, Ctrl/Cmd+V in
  // any folder. Cut items show dimmed until they are moved.
  async function toClipboard(mode: 'copy' | 'move') {
    const items = await selection.resolve(listing);
    if (!items || items.length === 0) {
      return;
    }
    clipboard.set(mode, items);
    toasts.push({
      message: `${itemsText(items)} ready to ${mode === 'copy' ? 'copy' : 'move'}. In the folder they should go to, press ${shortcutLabel('paste', mac)} or use Paste in its menu.`
    });
  }

  async function paste(folder = path) {
    const { mode, items } = clipboard;
    if (!mode) {
      toasts.push({ message: 'Nothing to paste: copy or cut items first.' });
      return;
    }
    await bulk(mode, folder, items);
    if (mode === 'move') {
      clipboard.clear();
    }
  }

  // Menus (S02.5-T04): the selection's, or the folder's on empty space.
  let menuOpen = $state(false);
  let menuX = $state(0);
  let menuY = $state(0);
  let menuItems = $state.raw<MenuEntry[]>([]);
  let helpOpen = $state(false);

  function showMenu(item: FileItem | undefined, x: number, y: number) {
    menuItems = item ? selectionMenu(item) : folderMenu();
    menuX = x;
    menuY = y;
    menuOpen = true;
  }

  function selectionMenu(item: FileItem): MenuEntry[] {
    const one = selected === 1;
    const entries: MenuEntry[] = [];
    if (one && (item.kind === 'dir' || item.kind === 'file')) {
      entries.push({
        label: item.kind === 'dir' ? 'Open' : 'Preview',
        icon: item.kind === 'dir' ? FolderOpen : Eye,
        shortcut: 'Enter',
        onselect: () => open(item, focusedIndex)
      });
    }
    entries.push(
      {
        label: 'Download',
        icon: Download,
        onselect: () => void downloadSelection(selection, listing, path)
      },
      'separator',
      {
        label: 'Cut',
        icon: Scissors,
        shortcut: shortcutLabel('cut', mac),
        onselect: () => void toClipboard('move')
      },
      {
        label: 'Copy',
        icon: Copy,
        shortcut: shortcutLabel('copy', mac),
        onselect: () => void toClipboard('copy')
      }
    );
    if (one && item.kind === 'dir' && clipboard.mode) {
      entries.push({
        label: `Paste into “${item.name}”`,
        icon: ClipboardPaste,
        onselect: () => void paste(item.path)
      });
    }
    entries.push('separator');
    if (one) {
      entries.push({
        label: 'Rename…',
        icon: Pencil,
        shortcut: shortcutLabel('rename', mac),
        onselect: () => void openDialog('rename')
      });
    }
    entries.push(
      { label: 'Move to…', icon: FolderOutput, onselect: () => void openDialog('move') },
      { label: 'Copy to…', icon: CopyPlus, onselect: () => void openDialog('copy') },
      'separator',
      {
        label: 'Delete…',
        icon: Trash,
        shortcut: shortcutLabel('delete', mac),
        danger: true,
        onselect: () => void openDialog('delete')
      }
    );
    return entries;
  }

  function folderMenu(): MenuEntry[] {
    return [
      {
        label: 'New folder…',
        icon: FolderPlus,
        shortcut: shortcutLabel('new-folder', mac),
        onselect: () => (newFolderOpen = true)
      },
      { label: 'Upload files…', icon: Upload, onselect: () => fileInput?.click() },
      { label: 'Upload a folder…', icon: FolderUp2, onselect: () => folderInput?.click() },
      'separator',
      {
        label: clipboard.mode ? `Paste ${itemsText(clipboard.items)}` : 'Paste',
        icon: ClipboardPaste,
        shortcut: shortcutLabel('paste', mac),
        disabled: !clipboard.mode,
        onselect: () => void paste()
      },
      {
        label: 'Select all',
        icon: SquareCheck,
        shortcut: mac ? '⌘A' : 'Ctrl+A',
        disabled: !listing.total,
        onselect: () => selection.selectAll()
      },
      'separator',
      {
        label: 'Keyboard shortcuts',
        icon: Keyboard,
        shortcut: '?',
        onselect: () => (helpOpen = true)
      }
    ];
  }

  async function newFolder(name: string) {
    let folder;
    try {
      folder = await createFolder(path, name);
    } catch (error) {
      if (!(error instanceof ApiError && error.code === 'conflict')) {
        throw error;
      }
      const choice = await new ConflictBatch(conflicts, 1).choose(name, true, child(path, name));
      if (choice !== 'rename') {
        return; // skipped: the folder that is there stays
      }
      folder = await createFolder(path, name, 'rename');
    }
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

  // After opening a folder or going up from the file view, the new
  // folder's view takes the focus (S02.5-T04): the old view is gone while
  // the folder loads, and the focus would fall back to the page.
  let focusView = $state(false);
  let focusedIndex = $state(0);
  function fromView(): boolean {
    const active = document.activeElement;
    return !!active?.closest('[role="grid"]');
  }

  function open(item: FileItem, index: number) {
    if (item.kind === 'dir') {
      focusView = fromView();
      goto(filesHref(item.path));
    } else if (item.kind === 'file') {
      // The preview (S02.6-T01) is a history entry, so Back closes it.
      pushState('', { preview: { path: item.path, index } });
    }
  }

  // The preview belongs to this folder's listing; stepping replaces its
  // history entry, and closing goes back past it.
  const preview = $derived(
    page.state.preview && parent(page.state.preview.path) === path ? page.state.preview : undefined
  );

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
        class="ml-auto flex flex-wrap items-center gap-2"
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
          <Button size="sm" onclick={() => void openDialog('rename')}>
            <Pencil class="size-4" /> Rename
          </Button>
        {/if}
        <Button size="sm" onclick={() => void openDialog('move')}>
          <FolderOutput class="size-4" /> Move
        </Button>
        <Button size="sm" onclick={() => void openDialog('copy')}>
          <Copy class="size-4" /> Copy
        </Button>
        <Button size="sm" variant="danger" onclick={() => void openDialog('delete')}>
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
    <div class="ml-auto flex flex-wrap items-center gap-2 {selected > 0 ? 'hidden' : ''}">
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
    <FileView
      {listing}
      {selection}
      {mode}
      {sort}
      {order}
      onsort={setSort}
      onopen={open}
      onmenu={showMenu}
      dimmed={(p) => clipboard.isCut(p)}
      autofocus={focusView}
      bind:focused={focusedIndex}
    >
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
<ShortcutsDialog bind:open={helpOpen} />
{#if preview}
  <PreviewFrame
    {listing}
    index={preview.index}
    onstep={(index, item) => replaceState('', { preview: { path: item.path, index } })}
    onclose={() => history.back()}
  />
{/if}
<Menu
  bind:open={menuOpen}
  x={menuX}
  y={menuY}
  label={menuItems.some((e) => e !== 'separator' && e.label === 'Delete…')
    ? 'Selected items'
    : 'This folder'}
  items={menuItems}
/>

<DropZone
  target={path === '/' ? 'Files' : basename(path)}
  ondrop={(dropped) => void startUpload(dropped.files, path, dropped.emptyFolders)}
/>
