# Moving the storage root

The storage root is the folder that holds everything the NAS keeps: the
`files/` and `photos/` areas and the internal data in `.local-ai-nas/`
(database, logs, unfinished uploads). It can be moved to another folder or
disk, for example onto a larger drive. Nothing inside it records where it
is, so a copy works in its new place.

1. **Stop the NAS** (stop the service, or press Ctrl+C where it runs).
2. **Copy or move the whole folder**, including the hidden `.local-ai-nas/`
   folder. Keep everything inside one folder on one drive: uploads are
   finished by renaming them into the files area, which needs one file
   system (the health check reports it if not).

   Linux:

   ```sh
   sudo rsync -a /srv/nas/ /mnt/big-disk/nas/
   ```

   Windows (PowerShell):

   ```powershell
   robocopy D:\nas E:\nas /E /COPY:DAT
   ```

3. **Point the NAS at the new folder** with `storage.root`: in the config
   file (`root = "/mnt/big-disk/nas"` under `[storage]`), with the
   environment variable `LOCALAINAS_STORAGE_ROOT`, or with
   `--storage-root`. If `storage.db_dir` or `storage.logs_dir` are set, they
   are separate folders: leave them, or move them too and change them.
4. **Start the NAS** and check `GET /api/v1/system/health`: every check is
   `ok`. Uploads that were not finished continue from where they stopped;
   the NAS points them at the new folder when it starts.
5. **Delete the old folder** once you have checked your files in the new
   place. Until then the NAS never reads or writes it.
