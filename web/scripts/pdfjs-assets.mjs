// Copies the files pdf.js fetches at run time into static/pdfjs/, served by
// the app (S02.6-T04): the CMaps (text in Chinese, Japanese, Korean), the
// standard fonts, and the JavaScript image decoders for JBIG2 and JPEG 2000
// (used instead of WebAssembly, so the app's CSP needs no exception). Each
// folder keeps its license file. It runs before dev and build.
//
// The Liberation Sans fonts in pdfjs-dist are not copied: their license is
// GPL-2.0 with a font exception, which docs/licensing.md does not allow in
// shipped files. pdf.js then uses the system's sans-serif font instead.
import { copyFileSync, mkdirSync, readdirSync, rmSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const web = join(dirname(fileURLToPath(import.meta.url)), '..');
const from = join(web, 'node_modules', 'pdfjs-dist');
const to = join(web, 'static', 'pdfjs');

/** The files to copy per folder: a filter on the names. */
const folders = {
  cmaps: () => true, // Adobe's CMaps: BSD-3-Clause (cmaps/LICENSE)
  standard_fonts: (name) => name.startsWith('Foxit') || name === 'LICENSE_FOXIT', // BSD-3-Clause
  wasm: (name) =>
    [
      'jbig2_nowasm_fallback.js',
      'openjpeg_nowasm_fallback.js',
      'LICENSE_JBIG2', // BSD-3-Clause (PDFium)
      'LICENSE_OPENJPEG', // BSD-2-Clause
      'LICENSE_PDFJS_JBIG2', // Apache-2.0 (Mozilla's port)
      'LICENSE_PDFJS_OPENJPEG' // BSD-2-Clause (Mozilla's port)
    ].includes(name)
};

rmSync(to, { recursive: true, force: true });
let copied = 0;
for (const [folder, wanted] of Object.entries(folders)) {
  mkdirSync(join(to, folder), { recursive: true });
  for (const name of readdirSync(join(from, folder)).filter(wanted)) {
    copyFileSync(join(from, folder, name), join(to, folder, name));
    copied++;
  }
}
console.log(`pdfjs-assets: ${copied} files in static/pdfjs/`);
