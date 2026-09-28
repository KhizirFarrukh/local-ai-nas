# Using the web interface

The web interface is how you use the NAS from a browser: browse your files, upload and download them, organize them, and look at them. This guide covers everything in it today (stage 2). Photos, search, sharing, and settings come in later stages.

## Opening it

Start the server (see the README's Development section), then open `http://127.0.0.1:8080/` in a browser on the same computer. Until login and HTTPS arrive (stage 3), the NAS only answers on this computer.

It works in current Chrome, Edge, and Firefox, on a computer, a tablet, or a phone-sized window.

## Finding your way

- **Sections:** Files, Photos (coming in stage 4), and Settings, in the bar on the left, or at the bottom on a phone.
- **The path** at the top shows where you are. Click any part of it to go back up.
- **Open a folder** by double-clicking it (a single tap on a touch screen), or select it and press Enter.
- **Go up** with Backspace or Alt+↑.
- **List or grid:** the two buttons on the right switch between a list with details and a grid of tiles. In the list, click a column title (Name, Size, Modified, Type) to sort by it; click it again to reverse the order. In the grid, pick the order from the Sort menu. The browser remembers your choice.
- **Large folders** stay fast: only what is on screen is loaded, so a folder of 50,000 items opens as quickly as a small one.

## Selecting

- **Click** an item to select it. **Ctrl+click** (⌘+click on a Mac) adds or removes one item. **Shift+click** selects everything from the last item you clicked.
- **Ctrl+A** selects everything in the folder. **Escape**, or a click on empty space, clears the selection.
- On a **touch screen**, a tap opens an item, and a **long press** selects it and opens its menu. While something is selected, taps add or remove items.

While items are selected, the bar at the top shows how many and what you can do with them: Download, Move, Copy, Delete, Select all, and more.

## Menus

Right-click an item, or empty space in the folder, to open its menu. On a touch screen, long-press. With the keyboard, press Shift+F10 or the menu key. The menu lists every action and its shortcut.

## Uploading

- **Upload** picks files; **Upload folder** picks a folder and uploads everything in it, with its subfolders.
- **Drag and drop** files or folders from your computer anywhere onto the page. An outline shows the folder they will go into.
- The **progress panel** at the bottom shows each upload with its progress and speed. You can **pause**, **resume**, **cancel**, and **retry** each one.
- Uploads survive trouble: if the network drops, they continue by themselves when it is back. If you close or reload the page, add the same file to the same folder again, and it continues from where it stopped instead of starting over.
- A file appears only once all of it has arrived, so a broken upload never leaves half a file behind.

## When a name is taken

When an upload, a move, a copy, or a new folder meets a name that is already there, the NAS asks:

- **Replace:** the new file replaces the old one. (Folders are never replaced or merged.)
- **Keep both:** the new one gets a numbered name, such as `report (1).txt`.
- **Skip:** leave the one that is there.

When several items are involved, tick **Do the same for the other files** to answer once for all of them, or choose **Stop the operation**.

## Downloading

- **A file:** the download button on its row, or select it and press Download.
- **A folder, or several items:** they download as one ZIP file. **Download folder** downloads the folder you are in.
- Downloads go through your browser's own download manager, so large ones are fine and the browser shows their progress.

## Organizing

| To | Do this |
|---|---|
| Make a folder | **New folder**, or Shift+N |
| Rename | Select the item, then F2 or **Rename…** in its menu. For a file, the name before the extension is selected, so you can type the new name at once. |
| Move or copy | **Move** or **Copy**, then choose the folder to put the items in. |
| Move or copy with the keyboard | Ctrl+X (cut) or Ctrl+C (copy), go to the other folder, then Ctrl+V. Cut items look faded until you paste them. |
| Delete | **Delete**, or the Delete key. The NAS asks first. |

> **Deleting is permanent for now.** There is no trash yet (it comes in stage 8), so deleted items cannot be brought back. Keep a backup of anything important.

Long operations show their progress in the panel at the bottom and can be cancelled between items. A note at the end says what was done, and explains anything that failed.

## Previews

Open a file to preview it without downloading it:

- **Images** (JPEG, PNG, GIF, WebP, AVIF, SVG, BMP, icons);
- **Videos and audio** in the formats your browser plays, with seeking;
- **Text and code** (the first 256 KB of very large files, with a note);
- **PDF** documents, a page at a time (Page Up and Page Down turn pages).

Other files show their name, type, and size, with a Download button. The arrow keys (← →) step to the previous or next file in the folder; Escape or the browser's Back button closes the preview.

Previews are safe: web pages are shown as their text, never run, and pictures with hidden scripts cannot run them.

## Keyboard shortcuts

Press **?** in the file list to see them all.

| Keys | Action |
|---|---|
| ↑ ↓ ← →, Home, End, Page Up, Page Down | Move, and select the item there |
| Shift + a move | Select a range |
| Ctrl + a move | Move without changing the selection |
| Space | Select or unselect the item |
| Ctrl+A | Select all |
| Escape | Clear the selection |
| Enter | Open |
| Alt+↑ or Backspace | Go to the parent folder |
| Shift+N | New folder |
| F2 | Rename |
| Delete | Delete (asks first) |
| Ctrl+C, Ctrl+X | Copy or cut, to paste in another folder |
| Ctrl+V | Paste into this folder |
| Shift+F10 or the menu key | Open the menu |
| ? | Show the list of shortcuts |

On a Mac, use ⌘ instead of Ctrl; ⌘+Backspace also deletes.

## Accessibility

- Everything works with the keyboard alone, and the focused item always shows a ring. After a dialog or a menu closes, the focus goes back to where you were.
- Screen readers hear each item's name, type, size, and date, whether it is selected, how many items are selected, and every notification (errors at once).
- Buttons and links are at least 44 pixels tall on touch screens, and the colors meet the WCAG 2.1 AA contrast levels in both themes.

## Settings

**Settings** has the **theme** (follow your device, light, or dark) and **About**: the version of the NAS, its license (GNU AGPL v3.0 or later, with a link to the source code), and the open-source components of the interface.

## When something goes wrong

- **"Can't reach the NAS"** at the top means the server stopped or the connection dropped. The app keeps checking and carries on by itself when the server is back; **Try now** checks at once.
- Errors say what happened and what to do, in plain words. For errors that come from inside the server, the message includes a request ID, which points to the matching entry in the server's log.
