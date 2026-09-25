<!--
  Settings (S02.2-T01): the theme, and About. More settings come with the
  settings page of S10.5.
-->
<script lang="ts">
  import { api, unwrap } from '$lib/api/client';
  import ThemeSwitch from '$lib/components/ThemeSwitch.svelte';
  import components from '$lib/generated/components.json';

  const version = unwrap(api.GET('/system/health')).then((health) => health.version);
</script>

<svelte:head>
  <title>Settings · local-ai-nas</title>
</svelte:head>

<div class="mx-auto flex max-w-3xl flex-col gap-8 p-4 md:p-8">
  <h1 class="text-xl font-semibold">Settings</h1>

  <section class="flex flex-col gap-3" aria-labelledby="appearance">
    <h2 id="appearance" class="font-semibold">Appearance</h2>
    <p class="text-sm text-fg-muted">The theme follows your device unless you pick one here.</p>
    <ThemeSwitch />
  </section>

  <section class="flex flex-col gap-3" aria-labelledby="about">
    <h2 id="about" class="font-semibold">About</h2>
    <dl class="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-2 text-sm">
      <dt class="text-fg-muted">Version</dt>
      <dd>
        {#await version}
          …
        {:then v}
          {v}
        {:catch}
          unknown (the server could not be reached)
        {/await}
      </dd>
      <dt class="text-fg-muted">License</dt>
      <dd>
        GNU Affero General Public License v3.0 or later. You can get the source code of the version
        you use from the project's repository,
        <a class="text-accent underline" href="https://github.com/KhizirFarrukh/local-ai-nas"
          >github.com/KhizirFarrukh/local-ai-nas</a
        >.
      </dd>
    </dl>

    <h3 class="mt-2 text-sm font-semibold">Open-source components in the web interface</h3>
    <table class="w-full text-left text-sm">
      <thead class="text-fg-muted">
        <tr>
          <th class="py-1 font-medium">Component</th>
          <th class="font-medium">Version</th>
          <th class="font-medium">License</th>
        </tr>
      </thead>
      <tbody>
        {#each components as component (component.name)}
          <tr class="border-t border-border">
            <td class="py-1">{component.name}</td>
            <td>{component.version}</td>
            <td>{component.license}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </section>
</div>
