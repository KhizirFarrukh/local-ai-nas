<!--
  The component gallery (S02.1-T04, development builds only): every base
  component in both themes, side by side, with working examples of the
  dialog, the menu, and the API client.
-->
<script lang="ts">
  import Copy from '@lucide/svelte/icons/copy';
  import FolderOpen from '@lucide/svelte/icons/folder-open';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';
  import { api, unwrap } from '$lib/api/client';
  import { ApiError } from '$lib/api/errors';
  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import Button from '$lib/components/Button.svelte';
  import Checkbox from '$lib/components/Checkbox.svelte';
  import Dialog from '$lib/components/Dialog.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import IconButton from '$lib/components/IconButton.svelte';
  import Menu, { type MenuEntry } from '$lib/components/Menu.svelte';
  import ProgressBar from '$lib/components/ProgressBar.svelte';
  import Spinner from '$lib/components/Spinner.svelte';
  import TextField from '$lib/components/TextField.svelte';
  import ThemeSwitch from '$lib/components/ThemeSwitch.svelte';
  import Toast from '$lib/components/Toast.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import { ApiError as DemoError } from '$lib/api/errors';
  import { tasks } from '$lib/shell/tasks.svelte';
  import { toasts } from '$lib/shell/toasts.svelte';

  let dialogOpen = $state(false);
  let menuOpen = $state(false);
  let menuX = $state(0);
  let menuY = $state(0);
  let name = $state('report.txt');
  let picked = $state('nothing yet');
  let apiResult = $state('');

  const items: MenuEntry[] = [
    { label: 'Open', icon: FolderOpen, shortcut: 'Enter', onselect: () => (picked = 'Open') },
    { label: 'Rename', icon: Pencil, shortcut: 'F2', onselect: () => (picked = 'Rename') },
    { label: 'Copy', icon: Copy, onselect: () => (picked = 'Copy') },
    'separator',
    {
      label: 'Delete',
      icon: Trash2,
      shortcut: 'Del',
      danger: true,
      onselect: () => (picked = 'Delete')
    }
  ];

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    const target = event.currentTarget as HTMLElement;
    const r = target.getBoundingClientRect();
    menuX = event.type === 'contextmenu' ? event.clientX : r.left;
    menuY = event.type === 'contextmenu' ? event.clientY : r.bottom + 4;
    menuOpen = true;
  }

  // A pretend long operation for the progress panel: ten steps of 300 ms.
  function demoTask(outcome: 'finish' | 'fail' | 'cancel') {
    let stopped = false;
    const task = tasks.start(`Demo task (${outcome})`, {
      total: 10,
      cancel: () => {
        stopped = true;
        task.cancelled();
      }
    });
    let step = 0;
    const tick = () => {
      if (stopped) return;
      step++;
      task.update(step, 10, `${step} of 10 steps`);
      if (step === 6 && outcome === 'fail') {
        task.fail(new DemoError({ status: 507, code: 'insufficient_storage', message: 'full' }));
      } else if (step === 10) {
        task.finish('Demo task finished.');
      } else {
        setTimeout(tick, 300);
      }
    };
    setTimeout(tick, 300);
  }

  async function callApi(path: string) {
    try {
      const page = await unwrap(api.GET('/files/items', { params: { query: { path } } }));
      apiResult = `OK: ${page.item.path} (${page.items?.length ?? 0} items)`;
    } catch (e) {
      apiResult =
        e instanceof ApiError
          ? `ApiError ${e.status} ${e.code} (request ${e.correlationId ?? 'none'}): ${e.message}`
          : String(e);
    }
  }
</script>

<svelte:head>
  <title>Components · local-ai-nas</title>
</svelte:head>

{#snippet gallery()}
  <div class="flex flex-col gap-6 bg-bg p-6">
    <section class="flex flex-col gap-2">
      <h2 class="font-semibold">Buttons</h2>
      <div class="flex flex-wrap items-center gap-2">
        <Button variant="primary"><Upload class="size-4" /> Upload</Button>
        <Button>Secondary</Button>
        <Button variant="danger">Delete</Button>
        <Button variant="ghost">Ghost</Button>
        <Button size="sm">Small</Button>
        <Button disabled>Disabled</Button>
        <IconButton label="Rename"><Pencil class="size-4" /></IconButton>
        <Tooltip text="Copies the selection">
          {#snippet children(tip)}
            <Button {...tip}><Copy class="size-4" /> With tooltip</Button>
          {/snippet}
        </Tooltip>
      </div>
    </section>

    <section class="flex flex-col gap-2">
      <h2 class="font-semibold">Fields</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <TextField label="Name" bind:value={name} hint="One name, no slashes." />
        <TextField
          label="New name"
          value="aux.txt"
          error="“aux.txt” is a reserved name on Windows."
        />
      </div>
      <div class="flex gap-4">
        <Checkbox label="Select all" />
        <Checkbox label="Some selected" indeterminate />
        <Checkbox label="Checked" checked />
      </div>
    </section>

    <section class="flex flex-col gap-2">
      <h2 class="font-semibold">Progress and status</h2>
      <ProgressBar value={0.42} label="Uploading report.txt" />
      <ProgressBar label="Preparing" />
      <div class="flex items-center gap-2 text-sm">
        <Spinner label="Loading the folder" /> Loading…
      </div>
    </section>

    <section class="flex flex-col gap-2">
      <h2 class="font-semibold">Notifications</h2>
      <Toast kind="success" message="Uploaded 3 files to /docs." onclose={() => {}} />
      <Toast kind="info" message="The upload continues when the connection is back." />
      <Toast
        kind="error"
        message="No space left on the disk."
        detail="Request 3f9c2a7e0b1d4c6f"
        onclose={() => {}}
      />
    </section>

    <section class="flex flex-col gap-2">
      <h2 class="font-semibold">Navigation and empty states</h2>
      <Breadcrumbs
        crumbs={[
          { label: 'Files', href: '/files/' },
          { label: 'docs', href: '/files/docs' },
          { label: 'reports 2026', href: '/files/docs/reports%202026' }
        ]}
      />
      <div class="rounded-lg border border-border bg-surface">
        <EmptyState
          icon={FolderOpen}
          title="This folder is empty"
          description="Drop files here, or use Upload."
        >
          {#snippet actions()}
            <Button variant="primary">Upload</Button>
            <Button>New folder</Button>
          {/snippet}
        </EmptyState>
      </div>
    </section>
  </div>
{/snippet}

<main class="flex flex-col gap-4 p-6">
  <header class="flex flex-wrap items-center justify-between gap-4">
    <h1 class="text-xl font-semibold">Component gallery</h1>
    <ThemeSwitch />
  </header>

  <section class="flex flex-wrap items-center gap-2">
    <Button variant="primary" onclick={() => (dialogOpen = true)}>Open a dialog</Button>
    <Button onclick={openMenu} oncontextmenu={openMenu}>Open a menu</Button>
    <span class="text-sm text-fg-muted">Picked: {picked}</span>
    <Button onclick={() => callApi('/')}>API: list /</Button>
    <Button onclick={() => callApi('/nope')}>API: list /nope</Button>
    <span class="text-sm text-fg-muted" data-testid="api-result">{apiResult}</span>
  </section>

  <section class="flex flex-wrap items-center gap-2">
    <Button onclick={() => demoTask('finish')}>Task that finishes</Button>
    <Button onclick={() => demoTask('fail')}>Task that fails</Button>
    <Button onclick={() => demoTask('cancel')}>Task to cancel</Button>
    <Button onclick={() => toasts.push({ message: 'A note that goes away after 5 s.' })}
      >Info note</Button
    >
  </section>

  <div class="grid overflow-hidden rounded-lg border border-border lg:grid-cols-2">
    <div data-theme="light">{@render gallery()}</div>
    <div data-theme="dark">{@render gallery()}</div>
  </div>
</main>

<Dialog bind:open={dialogOpen} title="Rename" description="Give report.txt a new name.">
  <TextField label="New name" bind:value={name} />
  {#snippet actions()}
    <Button onclick={() => (dialogOpen = false)}>Cancel</Button>
    <Button variant="primary" onclick={() => (dialogOpen = false)}>Rename</Button>
  {/snippet}
</Dialog>

<Menu bind:open={menuOpen} x={menuX} y={menuY} label="Actions" {items} />
