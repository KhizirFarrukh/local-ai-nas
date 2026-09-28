# Cross-browser report: the web interface (S02)

Stage 2's criterion (plan S02.8, criterion 2): the GUI is checked in current Chrome, Edge, and Firefox. Safari is not checked in S02 (the user's decision, S006); a phone or a Mac cannot reach the server before login and HTTPS (stage 3).

## Browsers

| Browser | Version | How |
|---|---|---|
| Microsoft Edge | 154.0.4258.37 (installed on the test PC, Windows 11) | Playwright's `msedge` channel: the system tests, and the by-hand checks of S02.1–S02.7 |
| Chromium (the open-source build of Chrome) | 153.0.8010.12 (Playwright 1.63.0) | Playwright: the system tests; the unit and component tests (Vitest browser mode) |
| Firefox | 155.0 (Playwright 1.63.0) | Playwright: the system tests, and the by-hand checks of S02.4–S02.7 |
| Google Chrome | the version on GitHub's machines | Playwright's `chrome` channel in the CI system-test job (`E2E_CHROME=1`). Chrome is not installed on the test PC, so the local runs use Chromium, which is the same engine. |

## What was checked

**Automatically, in every browser** (`web/tests/e2e`, 42 tests per browser, against the real binary with the embedded interface):

- browsing, including a folder of 50,000 items (first page and a jump to the end each within 5 s, only the rows on screen in the page);
- uploads: files, a folder, drag and drop, pause and resume, a network drop, a reload followed by adding the same file again (continues at the server's offset), and a taken name;
- downloads: a file byte for byte, a folder and a selection as ZIP files whose entries were read back;
- operations: new folder, rename, move, copy, delete, cut and paste, and every choice of the conflict dialog;
- previews: image, video and audio with seeking by byte ranges (206), text with its 256 KB limit, PDF, the fallback card, and active content (HTML shown as text, SVG scripts never run);
- keyboard-only use and the focus after every dialog and menu; phone and tablet widths (360, 768, 1024 px); accessibility with axe in both themes;
- the regression tests of the bugs found in stage 2 (S02-B02, B03, B05, B06, B10, B11).

**Results on the test PC** (2026-09-28): Chromium, Firefox, and Edge, **125 of 126 passed, 1 skipped** (below), repeatedly. axe found **no violation of any impact** on the main screens in either theme in any browser. **In CI** (the stage-end run 36433983279, 2026-09-28), the same suite passed on Linux (Chromium, Firefox, Google Chrome) and on Windows (Chromium, Firefox, Edge, Google Chrome). The first stage-end run had found two timing races in the tests themselves, both in Chromium on Linux; they were fixed before this run (stage document, S02.8-T05).

**By hand during the stage** (details in the stage document and the session logs S006 and S007): layouts at phone and tablet widths, touch (tap and long press), contrast in both themes (axe `color-contrast` in Edge and Firefox: no violation), a 2 GiB video seeking with two range requests in Edge and Firefox, ZIP archives of 4.2 GiB opening in Windows Explorer and Info-ZIP `unzip`, and uploads of 1,000 files by folder button in Edge, Chromium, and Firefox.

## Differences found

| Area | Difference | Effect |
|---|---|---|
| Folder upload in tests | Playwright can fill a folder picker only in Chromium-based browsers. | The folder-button test is skipped in Firefox. Folder upload was checked by hand in Firefox (S02.4-T02, 1,000 files), and the drag-and-drop test runs in all three. |
| Empty folders by folder button | A folder picker lists files only (every browser). | Empty subfolders arrive only with drag and drop, which keeps them. Documented; not fixable in the page. |
| Video formats | Which formats play depends on the browser and the system (Playwright's Firefox builds lack H.264 on Linux). | Files a browser cannot play show the fallback card with a Download button. The tests use WebM (VP8, Opus), which all three play. |
| Focus rings after script focus | Firefox showed no focus ring after the app moved the focus (dialogs, folder changes). | Fixed in S02.7-T02: the app asks for the ring when the keyboard is in use. |

No other difference in behavior was found between Chrome/Chromium, Edge, and Firefox.
