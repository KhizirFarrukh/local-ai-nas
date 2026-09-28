<!--
  A PDF preview (S02.6-T04): pdf.js, loaded only when a PDF is opened. Its
  core draws pages on a canvas and never runs a PDF's JavaScript (that
  needs the viewer's scripting sandbox, which the app does not load), and
  pdf.js 6 has no eval. The worker, the fonts, the CMaps, and the image
  decoders come from the app, and WebAssembly is off, so the app's CSP
  needs no exception (the needs are recorded for S03.5). The file is read
  by ranges, a page at a time.

  One page is shown, fitted to the window. Page Up and Page Down (also
  ↑ and ↓, Home, End) turn pages; ← and → still step between files.
-->
<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import type { PDFDocumentLoadingTask, PDFDocumentProxy, RenderTask } from 'pdfjs-dist';
  import workerSrc from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
  import Spinner from '$lib/components/Spinner.svelte';
  import type { FileItem } from '$lib/files/types';

  let { item, src, onfail }: { item: FileItem; src: string; onfail: (why: string) => void } =
    $props();

  let doc = $state.raw<PDFDocumentProxy | undefined>(undefined);
  let pageNumber = $state(1);
  let canvas: HTMLCanvasElement | undefined = $state();
  let width = $state(0);
  let height = $state(0);
  let size = $state({ width: 0, height: 0 });

  /** A plain reason for a PDF that cannot be shown. */
  function reason(error: unknown): string {
    const name = (error as { name?: string })?.name;
    if (name === 'PasswordException') {
      return 'This PDF is protected with a password. Download it to open it.';
    }
    if (name === 'InvalidPDFException') {
      return 'This file is not a valid PDF, or it is damaged.';
    }
    return `This PDF cannot be shown here (${error instanceof Error ? error.message : String(error)}).`;
  }

  // Open the document; closing the preview stops the worker.
  $effect(() => {
    let task: PDFDocumentLoadingTask | undefined;
    let stopped = false;
    void (async () => {
      try {
        const pdfjs = await import('pdfjs-dist');
        pdfjs.GlobalWorkerOptions.workerSrc = workerSrc;
        task = pdfjs.getDocument({
          url: src,
          cMapUrl: '/pdfjs/cmaps/',
          cMapPacked: true,
          standardFontDataUrl: '/pdfjs/standard_fonts/',
          wasmUrl: '/pdfjs/wasm/',
          useWasm: false,
          enableXfa: false,
          // Read what a page needs, by ranges, and nothing ahead.
          disableStream: true,
          disableAutoFetch: true
        });
        const opened = await task.promise;
        if (!stopped) {
          doc = opened;
        }
      } catch (error) {
        if (!stopped) {
          onfail(reason(error));
        }
      }
    })();
    return () => {
      stopped = true;
      void task?.destroy();
    };
  });

  // Draw the page, fitted to the area, sharp on high-density screens.
  $effect(() => {
    const current = doc;
    const target = canvas;
    const number = pageNumber;
    const [areaWidth, areaHeight] = [width, height];
    if (!current || !target || areaWidth === 0 || areaHeight === 0) {
      return;
    }
    let render: RenderTask | undefined;
    let stopped = false;
    void (async () => {
      try {
        const page = await current.getPage(number);
        if (stopped) {
          return;
        }
        const natural = page.getViewport({ scale: 1 });
        const scale = Math.min(areaWidth / natural.width, areaHeight / natural.height);
        const viewport = page.getViewport({ scale });
        const ratio = window.devicePixelRatio || 1;
        target.width = Math.floor(viewport.width * ratio);
        target.height = Math.floor(viewport.height * ratio);
        size = { width: Math.floor(viewport.width), height: Math.floor(viewport.height) };
        render = page.render({
          canvas: target,
          viewport,
          transform: ratio === 1 ? undefined : [ratio, 0, 0, ratio, 0, 0]
        });
        await render.promise;
      } catch (error) {
        if (!stopped && (error as { name?: string })?.name !== 'RenderingCancelledException') {
          onfail(reason(error));
        }
      }
    })();
    return () => {
      stopped = true;
      render?.cancel();
    };
  });

  function turn(to: number) {
    if (doc) {
      pageNumber = Math.max(1, Math.min(doc.numPages, to));
    }
  }

  function keys(event: KeyboardEvent) {
    const moves: Record<string, number> = {
      PageDown: pageNumber + 1,
      ArrowDown: pageNumber + 1,
      PageUp: pageNumber - 1,
      ArrowUp: pageNumber - 1,
      Home: 1,
      End: doc?.numPages ?? 1
    };
    if (doc && event.key in moves && !event.altKey && !event.ctrlKey && !event.metaKey) {
      event.preventDefault();
      turn(moves[event.key]);
    }
  }
</script>

<svelte:window onkeydown={keys} />

<div class="flex h-full w-full flex-col items-center gap-2">
  <div
    class="relative flex min-h-0 w-full flex-1 items-center justify-center"
    bind:clientWidth={width}
    bind:clientHeight={height}
  >
    {#if !doc}
      <Spinner label="Loading the PDF" />
    {/if}
    <canvas
      bind:this={canvas}
      class="absolute bg-white shadow-lg {doc ? '' : 'invisible'}"
      style:width="{size.width}px"
      style:height="{size.height}px"
      aria-label="Page {pageNumber} of {item.name}"
      data-testid="preview-pdf"
    ></canvas>
  </div>
  {#if doc}
    <div class="flex items-center gap-3 text-sm" data-testid="preview-pdf-pages">
      <button
        type="button"
        class="inline-flex size-8 items-center justify-center rounded-md hover:bg-white/15 disabled:opacity-30 pointer-coarse:size-11"
        aria-label="Previous page"
        title="Previous page (Page Up)"
        disabled={pageNumber <= 1}
        onclick={() => turn(pageNumber - 1)}
      >
        <ChevronUp class="size-5" aria-hidden="true" />
      </button>
      <span aria-live="polite">Page {pageNumber} of {doc.numPages}</span>
      <button
        type="button"
        class="inline-flex size-8 items-center justify-center rounded-md hover:bg-white/15 disabled:opacity-30 pointer-coarse:size-11"
        aria-label="Next page"
        title="Next page (Page Down)"
        disabled={pageNumber >= doc.numPages}
        onclick={() => turn(pageNumber + 1)}
      >
        <ChevronDown class="size-5" aria-hidden="true" />
      </button>
    </div>
  {/if}
</div>
